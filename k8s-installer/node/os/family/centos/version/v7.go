package version

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"k8s-installer/components/kubesphere"
	"k8s-installer/node/reportor"
	bd "k8s-installer/pkg/block_device"
	"k8s-installer/pkg/cache"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	natsLib "github.com/nats-io/nats.go"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/clientcmd"

	virtualKubelet "k8s-installer/components/virtual_kubelet"
	containerRuntime "k8s-installer/node/container_runtime"
	"k8s-installer/node/container_runtime/containerd"
	"k8s-installer/node/container_runtime/docker"
	"k8s-installer/node/k8s"
	"k8s-installer/node/loadbalancer"
	osInfoProvider "k8s-installer/node/os"
	"k8s-installer/node/os/family"
	linuxFamily "k8s-installer/node/os/family"
	"k8s-installer/node/os/family/centos"
	blockDevice "k8s-installer/pkg/block_device"
	"k8s-installer/pkg/command"
	"k8s-installer/pkg/config/client"
	"k8s-installer/pkg/constants"
	depMap "k8s-installer/pkg/dep"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/network"
	"k8s-installer/pkg/util"
	"k8s-installer/pkg/util/fileutils"
	"k8s-installer/schema"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	hostsHelper "github.com/txn2/txeh"
)

const (
	apiServerPort        = 6443
	proxiedAPIServerPort = 9443
)

/*
stands for centos 7
*/
type V7 struct {
	client.Config
}

func (v V7) TaskSetHost(hosts schema.TaskSetHosts) (map[string]string, error) {
	helper, err := hostsHelper.NewHostsDefault()
	if err != nil {
		log.Errorf("Failed to set /etc/hosts due to error: %s", err.Error())
		return nil, err
	}
	if hosts.Action == constants.ActionCreate {
		for ip, hostnames := range hosts.Hosts {
			helper.AddHosts(ip, hostnames)
		}
		if err := helper.Save(); err != nil {
			log.Errorf("Failed to write /etc/hosts due to error: %s", err.Error())
			return nil, err
		}
	} else {
		for ip, _ := range hosts.Hosts {
			helper.RemoveAddress(ip)
		}
		if err := helper.Save(); err != nil {
			log.Errorf("Failed to write /etc/hosts due to error: %s", err.Error())
			return nil, err
		}
	}
	return nil, nil
}

func (v V7) BasicNodeSetup(basicSetup schema.TaskBasicConfig, clusterInformation schema.Cluster) (map[string]string, error) {
	if basicSetup.Action == constants.ActionCreate {
		return v.basicNodeSetup(basicSetup, clusterInformation)
	} else {
		return v.basicNodeSetupDestroy(basicSetup, clusterInformation)
	}
}

func (v V7) basicNodeSetupDestroy(basicSetup schema.TaskBasicConfig, clusterInformation schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V7) basicNodeSetup(basicSetup schema.TaskBasicConfig, clusterInformation schema.Cluster) (map[string]string, error) {
	var err error
	var isActive bool
	returnData := map[string]string{}
	// disable firewall if it`s on
	isActive, err = linuxFamily.CheckSystemdService("firewalld")

	if err != nil {
		log.Debugf("Systemd not found skip!!!")
		// skip check when systemd unit already be disabled
		// this will results in error status so we skip it
	} else {
		if isActive {
			log.Debug("Service firewalld is active. Run stop and disable script!!!")
			err = linuxFamily.DisableFirewalld()
			if err != nil {
				return returnData, err
			}
		} else {
			log.Debug("Service firewalld is inactive. Skip!!!")
		}
	}

	// disable selinux and reboot if it`s on
	isActive, err = linuxFamily.GetEnforce()
	if err != nil {
		return returnData, err
	} else {
		if isActive {
			log.Debugf("Selinux is not disabled. Run disable script!!!")
			err = centos.DisableSelinux()
			if err != nil {
				return returnData, err
			}
		} else {
			log.Debugf("Selinux already disabled. Skip!!!")
		}
	}

	// load kernel modular
	err = linuxFamily.EnableKernelOptions()
	if err != nil {
		return returnData, err
	}

	// enable ipv4 ipv6 forwarding
	err = linuxFamily.EnableIPV46Forwarding()
	if err != nil {
		return returnData, err
	}

	err = linuxFamily.DisableSwapForever()
	if err != nil {
		return returnData, err
	}

	//  Increase the Maximum Number of File Descriptors
	err = linuxFamily.IncreaseMaximumNumberOfFileDescriptors()
	if err != nil {
		return returnData, err
	}

	return returnData, nil
}

func (v V7) PurgeAll(cluster schema.Cluster) (map[string]string, error) {
	if err := linuxFamily.StopSystemdService("kubelet"); err != nil {
		return nil, err
	}

	return nil, nil
}

func (v V7) InstallOrRemoveContainerRuntime(taskCri schema.TaskCRI, clusterInformation schema.Cluster, resourceServerURL string, md5Dep depMap.DepMap) (map[string]string, error) {
	taskCri.K8sVersion = util.StringDefaultIfNotSet(taskCri.K8sVersion, constants.V1_18_6)
	// When server process is running on the same node with client,
	// skip the task for remaining docker/containerd.
	if v.Config.IsPlexing {
		return map[string]string{}, nil
	}

	if taskCri.Action == constants.ActionCreate {
		return v.installContainerRuntime(taskCri, clusterInformation, resourceServerURL, taskCri.K8sVersion, md5Dep)
	}
	return v.removeContainerRuntime(taskCri)
}

func (v V7) installContainerRuntime(taskCri schema.TaskCRI, cluster schema.Cluster, resourceServerURL, k8sVersion string, md5Dep depMap.DepMap) (map[string]string, error) {
	var err error
	var isExists bool

	returnData := map[string]string{}

	// check docker already installed
	isExists, err = linuxFamily.CheckSystemdServiceExists(taskCri.CRIType.CRIType)

	if isExists && taskCri.CRIType.ReinstallIfAlreadyInstall {
		log.Debug("Config says we need reinstall. Go remove container runtime first !!!")
		if returnData, err = v.removeContainerRuntime(taskCri); err != nil {
			log.Errorf("Failed to remove container runtime docker during reinstall due to error %s", err.Error())
			return returnData, err
		}
		log.Debug("Done with remove container runtime")
	}
	rtc := cache.GetCurrentCache()
	clientConfig := rtc.GetClientRuntimeConfig(cache.NodeId)

	switch taskCri.CRIType.CRIType {
	case constants.CRITypeDocker:
		criRootDir := clientConfig.CRIRootDir + "/docker"
		if clientConfig.CRIRootDir != "/var/lib" {
			if err := util.CreateDirIfNotExists(criRootDir); err != nil {
				log.Errorf("Failed to create docker root dir with location '%s'", criRootDir)
			}
		}

		cri := docker.CRIDockerCentos{}
		// mount the cri device if config CRIMountDev is set
		if err := mountIfCriDevIsSet(criRootDir, v.Config, true, false, true); err != nil {
			return returnData, err
		}
		err = cri.Install(v.Config.Offline, taskCri, cluster, v.Config, resourceServerURL, k8sVersion, md5Dep)
	case constants.CRITypeContainerd:
		criRootDir := clientConfig.CRIRootDir + "/containerd"
		if clientConfig.CRIRootDir != "/var/lib" {
			if err := util.CreateDirIfNotExists(criRootDir); err != nil {
				log.Errorf("Failed to create docker root dir with location '%s'", criRootDir)
			}
		}

		cri := containerd.ContainerD{}
		// mount the cri device if config CRIMountDev is set
		if err := mountIfCriDevIsSet(criRootDir, v.Config, true, false, true); err != nil {
			return returnData, err
		}
		err = cri.Install(v.Config.Offline, taskCri, cluster, v.Config, resourceServerURL, k8sVersion, md5Dep)
	default:
		err = fmt.Errorf("Container runtime %s currently not support yet", taskCri.CRIType.CRIType)
	}

	if err != nil {
		log.Errorf("Failed to set up container runtime %s due to error %s", taskCri.CRIType.CRIType, err.Error())
	}

	return returnData, err
}

func (v V7) removeContainerRuntime(taskCri schema.TaskCRI) (map[string]string, error) {
	log.Debugf("Try to stop kubelet or container will be recreate after we kill it")
	if err := linuxFamily.StopSystemdService("kubelet"); err != nil {
		log.Errorf("Failed to stop systemd service kubelet due to error: %s", err.Error())
		log.Error("Skip it and try to remove container runtime directly")
	}

	var err error
	//var stdErr bytes.Buffer
	var isExists bool
	var cri containerRuntime.INodeContainerRuntime
	returnData := map[string]string{}
	isExists, err = linuxFamily.CheckSystemdServiceExists(taskCri.CRIType.CRIType)
	rtc := cache.GetCurrentCache()
	clientConfig := rtc.GetClientRuntimeConfig(cache.NodeId)
	dataDir := ""
	switch taskCri.CRIType.CRIType {
	case constants.CRITypeDocker:
		cri = docker.CRIDockerCentos{}
		dataDir = clientConfig.CRIRootDir + "/docker"
	case constants.CRITypeContainerd:
		cri = containerd.ContainerD{}
		dataDir = clientConfig.CRIRootDir + "/containerd"
	default:
		err = fmt.Errorf("Container runtime %s currently not support yet ", taskCri.CRIType.CRIType)
	}
	if isExists && cri != nil {
		err = cri.Remove(taskCri)
		if err != nil {
			return returnData, err
		}
		return returnData, umountOrRemoveCRIDataDir(v.Config.CRIMountDev, dataDir, v.Config.YamlDataDir, cri)
	}

	if err := umountOrRemoveCRIDataDir(v.Config.CRIMountDev, dataDir, v.Config.YamlDataDir, cri); err != nil {
		return returnData, err
	}
	// when service cri service does not exists ,which will result in error.
	// that said we should ignore the error and move on
	log.Debugf("Service %s does not found nothing need to be done...", taskCri.CRIType.CRIType)
	return returnData, nil
}

func (v V7) InstallOrRemoveLoadBalance(loadBalance schema.TaskLoadBalance, clusterInformation schema.Cluster, resourceServerURL string, md5Dep depMap.DepMap) (map[string]string, error) {
	if loadBalance.Action == constants.ActionCreate {
		return v.installLoadBalance(loadBalance, clusterInformation, resourceServerURL, md5Dep)
	} else {
		return v.removeLoadBalance(loadBalance)
	}
}

func (v V7) installLoadBalance(loadBalance schema.TaskLoadBalance, clusterInformation schema.Cluster, resourceServerURL string, md5Dep depMap.DepMap) (map[string]string, error) {
	var proxy loadbalancer.IWebServer
	proxy = loadbalancer.CreateProxy(loadBalance.ProxyType)

	osInfo, errOSInfo := osInfoProvider.GetAllSystemInformation()

	if errOSInfo != nil {
		log.Errorf("Failed to get node cpu arch due to error %s", errOSInfo.Error())
		return nil, errOSInfo
	}

	log.Debug("Try to install proxy")
	if err := proxy.Install(v.Config.Offline, loadBalance, clusterInformation, v.Config, resourceServerURL, "7", osInfo.Kernel.Architecture, md5Dep); err != nil {
		return nil, err
	}

	log.Debugf("Try to enable service %s", proxy.GetSystemdServiceName())
	if err := linuxFamily.StartSystemdService(true, true, proxy.GetSystemdServiceName()); err != nil {
		return nil, err
	}
	return nil, nil
}

func (v V7) removeLoadBalance(loadBalance schema.TaskLoadBalance) (map[string]string, error) {
	var proxy loadbalancer.IWebServer
	proxy = loadbalancer.CreateProxy(loadBalance.ProxyType)
	if err := proxy.Remove(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (v V7) InitOrDestroyFirstControlPlane(kubeadm schema.TaskKubeadm, resourceServerURL string, clusterInformation schema.Cluster, md5Dep depMap.DepMap) (map[string]string, error) {
	if kubeadm.Action == constants.ActionCreate {
		return v.initFirstControlPlane(kubeadm, resourceServerURL, clusterInformation, md5Dep)
	}
	return v.destroyFirstControlPlane(kubeadm, resourceServerURL, clusterInformation)
}

func (v V7) destroyFirstControlPlane(kubeadm schema.TaskKubeadm, resourceServerURL string, cluster schema.Cluster) (map[string]string, error) {
	return v.destroyK8sNode(kubeadm, cluster)
}

func (v V7) initFirstControlPlane(kubeadm schema.TaskKubeadm, resourceServerURL string, cluster schema.Cluster, md5Dep depMap.DepMap) (map[string]string, error) {
	returnData := map[string]string{}
	if v.Config.Offline {
		// do offline install package from rpms
		if err := offlineInstallK8sYumDep(v.Config, cluster, resourceServerURL, md5Dep); err != nil {
			return nil, err
		}
	} else {
		if err := centos.OnlineInstallK8sYumDep(cluster.ControlPlane.EnableIPVS); err != nil {
			return nil, err
		}
	}

	if err := linuxFamily.EnableSystemdService("kubelet"); err != nil {
		return nil, err
	}
	if data, err := k8s.InitFirstControlPlane(kubeadm, v.Config); err != nil {
		return nil, err
	} else {
		returnData = data
	}

	log.Debug("Creating .kube dir")
	if err := util.CreateDirIfNotExists("/root/.kube"); err != nil {
		log.Errorf("Fail to create dir %s due to error: %v", "/root/.kube", err.Error())
		return nil, err
	}

	var stdErr bytes.Buffer
	var err error

	log.Debug("Remove old config if exists")
	_, _, _ = command.RunCmd("rm", "-f", "/root/.kube/config")

	log.Debug("Copy admin.conf to .kube/config")
	_, stdErr, err = command.RunCmd("cp", "/etc/kubernetes/admin.conf", "/root/.kube/config")
	if err != nil {
		log.Error("Failed to copy admin.conf to .kube/config due to following error:")
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
	}

	if err := k8s.ApplyCNIConfig(kubeadm, v.Config); err != nil {
		return nil, err
	}
	return returnData, nil
}

func offlineInstallK8sYumDep(config client.Config, cluster schema.Cluster, resourceServerURL string, md5Dep depMap.DepMap) error {

	if len(cluster.ControlPlane.KubernetesVersion) == 0 {
		cluster.ControlPlane.KubernetesVersion = constants.V1_18_6
	}
	saveTo := path.Join(config.YamlDataDir, "kubernetes-"+cluster.ControlPlane.KubernetesVersion)

	mergedDep := k8s.KubeDepMapping

	//if err := family.CommonDownloadDep(resourceServerURL, k8s.KubeDepMapping, saveTo, constants.V1_18_6); err != nil {
	if len(cluster.AdditionalVersionDep) > 0 && cluster.ControlPlane.KubernetesVersion != constants.V1_18_6 {
		if _, found := cluster.AdditionalVersionDep["centos"]; found {
			for key, version := range cluster.AdditionalVersionDep["centos"] {
				if _, found := mergedDep["centos"]; found {
					mergedDep["centos"][key] = version
				}
			}
		}
		if _, found := cluster.AdditionalVersionDep["ubuntu"]; found {
			for key, version := range cluster.AdditionalVersionDep["ubuntu"] {
				if _, found := mergedDep["ubuntu"]; found {
					mergedDep["ubuntu"][key] = version
				}
			}
		}
	}
	if err := family.CommonDownloadDep(resourceServerURL, mergedDep, saveTo, cluster.ControlPlane.KubernetesVersion, md5Dep); err != nil {
		return err
	}

	var stdErr bytes.Buffer
	var err error
	// install yum-utils
	log.Debug("Installing all docker rpms")
	_, stdErr, err = command.RunCmd("rpm", "-ivh", "--replacefiles", "--replacepkgs", "--nodeps", saveTo+"/*.rpm")
	if err != nil {
		log.Errorf("Failed to install docker related rpm due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	return nil
}

func (v V7) JoinOrDestroyControlPlane(kubeadm schema.TaskKubeadm, resourceServerURL string, preReturnData map[string]string, clusterInformation schema.Cluster, md5Dep depMap.DepMap) (map[string]string, error) {
	if kubeadm.Action == constants.ActionCreate {
		return v.joinControlPlane(preReturnData, clusterInformation, resourceServerURL, md5Dep)
	}
	return v.destroyControlPlane(kubeadm, clusterInformation)
}

func (v V7) destroyControlPlane(kubeadm schema.TaskKubeadm, cluster schema.Cluster) (map[string]string, error) {
	return v.destroyK8sNode(kubeadm, cluster)
}

func (v V7) joinControlPlane(preReturnData map[string]string, cluster schema.Cluster, resourceServerURL string, md5Dep depMap.DepMap) (map[string]string, error) {
	if v.Config.Offline {
		// do offline install pcakge from rpms
		if err := offlineInstallK8sYumDep(v.Config, cluster, resourceServerURL, md5Dep); err != nil {
			return nil, err
		}
	} else {
		if err := centos.OnlineInstallK8sYumDep(cluster.ControlPlane.EnableIPVS); err != nil {
			return nil, err
		}

	}

	if err := linuxFamily.EnableSystemdService("kubelet"); err != nil {
		return nil, err
	}

	return map[string]string{}, k8s.JoinControlPlane(preReturnData)
}

func (v V7) JoinOrDestroyWorkNode(kubeadm schema.TaskKubeadm, resourceServerURL string, preReturnData map[string]string, clusterInformation schema.Cluster, md5Dep depMap.DepMap) (map[string]string, error) {
	if kubeadm.Action == constants.ActionCreate {
		return v.joinWorkNode(kubeadm, resourceServerURL, preReturnData, clusterInformation, md5Dep)
	}
	return v.destroyWorkNode(kubeadm, clusterInformation)
}

func (v V7) joinWorkNode(kubeadm schema.TaskKubeadm, resourceServerURL string, preReturnData map[string]string, cluster schema.Cluster, md5Dep depMap.DepMap) (map[string]string, error) {
	if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
		return map[string]string{}, k8s.JoinRancherNode(preReturnData)
	}

	if v.Config.Offline {
		// do offline install pcakge from rpms
		if err := offlineInstallK8sYumDep(v.Config, cluster, resourceServerURL, md5Dep); err != nil {
			return nil, err
		}
	} else {
		if err := centos.OnlineInstallK8sYumDep(cluster.ControlPlane.EnableIPVS); err != nil {
			return nil, err
		}

	}

	if err := linuxFamily.EnableSystemdService("kubelet"); err != nil {
		return nil, err
	}

	return map[string]string{}, k8s.JoinWorker(preReturnData)
}

func (v V7) destroyWorkNode(kubeadm schema.TaskKubeadm, cluster schema.Cluster) (map[string]string, error) {
	return v.destroyK8sNode(kubeadm, cluster)
}

func (v V7) RenameHostname(renameHostnameTask schema.TaskRenameHostName, clusterInformation schema.Cluster) (map[string]string, error) {
	var err error
	var stdErr bytes.Buffer
	returnData := map[string]string{}
	log.Debug("Attempt to change hostname...")
	_, stdErr, err = command.RunCmd("hostname", renameHostnameTask.Hostname)
	if err != nil {
		log.Error("Failed to rename hostname due to following error:")
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
	}

	log.Debug("Attempt to change hostname forever...")
	if err := util.WriteTxtToFile("/etc/hostname", renameHostnameTask.Hostname); err != nil {
		log.Errorf("Failed to set /etc/hostname due to following error: %s", err.Error())
		return returnData, err
	}

	log.Debug("Attempt to set /etc/hosts...")
	if err := util.AppendTxtToFile("/etc/hosts", "\n"+renameHostnameTask.Hosts, 0644); err != nil {
		log.Errorf("Failed to set /etc/hosts due to following error: %s", err.Error())
		return returnData, err
	}

	return returnData, err
}

func (v V7) RunKubectl(kubectl schema.TaskKubectl, preReturnData map[string]string, clusterInformation schema.Cluster) (map[string]string, error) {
	return k8s.KubectlExecutor(kubectl, v.Config)
}

func (v V7) RunCommand(commandToRun schema.TaskRunCommand, preReturnData map[string]string, clusterInformation schema.Cluster) (map[string]string, error) {
	var err error
	var stdErr, stdOut bytes.Buffer
	result := map[string]string{}

	indexes := make([]int, 0, len(commandToRun.Commands))
	for k := range commandToRun.Commands {
		indexes = append(indexes, k)
	}
	sort.Ints(indexes)

	for _, index := range indexes {
		log.Debugf("Try to run command %v", commandToRun.Commands[index])
		if len(commandToRun.Commands[index]) == 0 {
			continue
		}
		stdOut, stdErr, err = command.RunCmd(commandToRun.Commands[index][0], commandToRun.Commands[index][1:]...)
		if err != nil {
			log.Errorf("Failed to run command %v due to following error:", commandToRun.Commands[index])
			errMsg := fmt.Sprintf("Global error: %s. Std error: %s", err.Error(), stdErr.String())
			log.Errorf("cmd err: %s", err.Error())
			log.Errorf("StdErr %s", errMsg)
			if commandToRun.CommandRunId == "" {
				result[strconv.Itoa(index)] = fmt.Sprintf("cmd err: %s", err.Error()) + "|" + fmt.Sprintf("StdErr %s", errMsg)
			} else {
				result[commandToRun.CommandRunId+"-"+strconv.Itoa(index)] = fmt.Sprintf("cmd err: %s", err.Error()) + "|" + fmt.Sprintf("StdErr %s", errMsg)
			}

			if commandToRun.IgnoreError {
				log.Warnf("TaskRunCommand.IgnoreError is set to true. Ignore error and run next command")
				continue
			}
			err = errors.New(errMsg)
			return nil, err
		}
		if commandToRun.CommandRunId == "" {
			result[strconv.Itoa(index)] = stdOut.String()
		} else {
			result[commandToRun.CommandRunId+"-"+strconv.Itoa(index)] = stdOut.String()
		}

	}

	// so far we don`t set return value so we return nil
	// if we need stdout then we can set stdout on a map[string]string
	if commandToRun.RequireResult {
		return result, nil
	} else {
		return nil, nil
	}
}

func (v V7) AsyncRunCommand(operationId string, nodeId string, nodeStepId string, msg *natsLib.Msg, commandToRun schema.TaskRunCommand, preReturnData map[string]string, clusterInformation schema.Cluster) (map[string]string, error) {
	var err error
	var stdErr, stdOut bytes.Buffer
	stat := constants.StatusSuccessful
	var message string
	result := map[string]string{}
	indexes := make([]int, 0, len(commandToRun.Commands))
	for k := range commandToRun.Commands {
		indexes = append(indexes, k)
	}
	sort.Ints(indexes)

	for _, index := range indexes {
		log.Debugf("Try to run command %v", commandToRun.Commands[index])
		if len(commandToRun.Commands[index]) == 0 {
			continue
		}
		// ignore stdout
		stdOut, stdErr, err = command.RunCmd(commandToRun.Commands[index][0], commandToRun.Commands[index][1:]...)
		if err != nil {
			log.Errorf("Failed to run command %s due to following error:", commandToRun.Commands[index][0])
			errMsg := fmt.Sprintf("Global error: %s. Std error:%s", err.Error(), stdErr.String())
			log.Errorf("cmd err: %s", err.Error())
			log.Errorf("StdErr %s", errMsg)
			err = errors.New(errMsg)
			message = fmt.Sprintf("Operation %s excution failed with node task %s on node %s due to error %v",
				operationId,
				constants.TaskDownloadDep,
				nodeId,
				err)
			log.Errorf(message)
			stat = constants.StatusError
		}
		result[strconv.Itoa(index)] = stdOut.String()
		break
	}

	if err == nil {
		// so far we don`t set return value so we return nil
		// if we need stdout then we can set stdout on a map[string]string
		message = fmt.Sprintf("Operation %v AsyncRunCommand %v done on node %v", operationId, commandToRun, nodeId)
		if commandToRun.RequireResult {
			return result, nil
		}
	}

	log.Debug("Send back message to notify server.")
	reply := family.CreateReplyBody(operationId, nodeId, stat, message, nodeStepId, nil)
	family.ReplyMsg(reply, msg)

	return nil, err
}

func (v V7) CopyTextFile(fileToCopy schema.TaskCopyTextBaseFile, preReturnData map[string]string, clusterInformation schema.Cluster) (map[string]string, error) {
	for filePath, fileToWrite := range fileToCopy.TextFiles {
		if err := util.CreateDirIfNotExists(filePath[0:strings.LastIndex(filePath, "/")]); err != nil {
			return nil, err
		}
		log.Debugf("Try to write file %s", filePath)
		if err := util.WriteTxtToFile(filePath, string(fileToWrite)); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (v V7) InstallOrDestroyVirtualKubelet(vk schema.TaskVirtualKubelet, resourceServerURL string, cluster schema.Cluster, md5Dep depMap.DepMap) (map[string]string, error) {
	if cluster.Action == constants.ActionCreate {
		return v.InstallVirtualKubelet(vk, resourceServerURL, cluster, md5Dep)
	}
	return v.DestroyVirtualKubelet(vk, resourceServerURL, cluster)
}

func (v V7) DestroyVirtualKubelet(vk schema.TaskVirtualKubelet, resourceServerURL string, cluster schema.Cluster) (map[string]string, error) {
	k8sVersion := constants.V1_18_6
	if cluster.ControlPlane.KubernetesVersion != "" {
		k8sVersion = cluster.ControlPlane.KubernetesVersion
	}

	osInfo, errOSInfo := osInfoProvider.GetAllSystemInformation()

	if errOSInfo != nil {
		log.Errorf("Failed to get node cpu arch due to error %s", errOSInfo.Error())
		return nil, errOSInfo
	}

	targetBinary := "/usr/bin/" + virtualKubelet.VKMockDep["centos"][k8sVersion][osInfo.Kernel.Architecture]["virtual-kubelet"]
	if _, stdErr, err := command.RunCmd("rm", "-f", targetBinary); err != nil {
		log.Errorf("Failed to run rm -f %s due to following error:", targetBinary)
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
		return nil, err
	}

	if err := linuxFamily.StopSystemdService("virtual-kubelet"); err != nil {
		return nil, err
	}

	if _, stdErr, err := command.RunCmd("rm", "-f", "/etc/systemd/system/virtual-kubelet.service"); err != nil {
		log.Errorf("Failed to run rm -f %s due to following error:", "/etc/systemd/system/virtual-kubelet.service")
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
		return nil, err
	}

	return nil, nil
}

func (v V7) InstallVirtualKubelet(vk schema.TaskVirtualKubelet, resourceServerURL string, cluster schema.Cluster, md5Sum depMap.DepMap) (map[string]string, error) {

	saveTo := "/usr/bin/"

	if err := family.CommonDownloadDep(resourceServerURL, virtualKubelet.VKMockDep, saveTo, constants.V1_18_6, md5Sum); err != nil {
		return nil, err
	}

	if _, stdErr, err := command.RunCmd("chmod", "+x", saveTo+"virtual-kubelet"); err != nil {
		log.Error("Failed to run chmod +x /usr/bin/virtual-kubelet due to following error:")
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
	}

	if err := linuxFamily.StopSystemdService("kubelet"); err != nil {
		return nil, err
	}

	saveTo = "/etc/virtual_kubelet"
	if err := util.CreateDirIfNotExists(saveTo); err != nil {
		log.Errorf("Fail to create dir %s due to error: %v", saveTo, err.Error())
		return nil, err
	}

	saveTo += "/config.json"

	if err := util.WriteTxtToFileByte(saveTo, vk.Config); err != nil {
		log.Errorf("Failed to save virtual kubelet config to %s due to error %s", saveTo, err.Error())
		return nil, err
	}

	if err := util.WriteTxtToFileByte("/etc/systemd/system/virtual-kubelet.service", vk.SystemdTemplate); err != nil {
		log.Errorf("Failed to write systemd file /etc/systemd/system/virtual-kubelet.service due to error %s", err.Error())
		return nil, err
	}

	if err := linuxFamily.StartSystemdService(true, false, "virtual-kubelet"); err != nil {
		return nil, err
	}
	return nil, nil
}

func (v V7) PrintJoinString(printJoinString schema.TaskPrintJoin) (map[string]string, error) {
	return k8s.PrintJoinString(printJoinString)
}

func (v V7) CommonLink(from depMap.DepMap, saveTo string, linkTo string) (map[string]string, error) {

	osInfo, _ := osInfoProvider.GetAllSystemInformation()
	for k, v := range from["centos"][constants.V1_18_6][osInfo.Kernel.Architecture] {
		log.Debugf("Create SymbolicLink from %v to %v", path.Join(saveTo, v), path.Join(linkTo, k))
		err := fileutils.SymbolicLink(path.Join(saveTo, v), path.Join(linkTo, k))
		if err != nil {
			log.Errorf("Create SymbolicLink from %v to %v err: %v", path.Join(saveTo, v), path.Join(linkTo, k), err)
			return nil, err
		}
	}
	return nil, nil
}

func (v V7) destroyK8sNode(kubeadm schema.TaskKubeadm, cluster schema.Cluster) (map[string]string, error) {
	_, err := k8s.DestroyKubeNode(kubeadm, v.Config, cluster)
	if err != nil {
		log.Errorf("Failed to destroy kube node due to error %s", err.Error())
		log.Error("But we are going to report node stat anyway")
	}
	//ignore error and do node reporting
	log.Debug("Reporting new node stat to server immediately")
	// only consider Reporting node stat as an error to fail task
	return nil, reportor.ReportIn(v.Config)
}

func (v V7) GoCurl(task schema.TaskCurl) (map[string]string, error) {
	body, err := json.Marshal(task.Body)
	if err != nil {
		return nil, err
	}
	ret, code, err := util.CommonRequest(task.URL, task.Method, task.NameServer,
		body, task.Headers, task.SkipTLS, true, time.Duration(task.TimeOut)*time.Second)
	if err != nil {
		log.Errorf("HTTP request failed, err: %s", err.Error())
		return nil, err
	}
	return map[string]string{
		"return_body": string(ret),
		"status_code": strconv.Itoa(code),
	}, nil
}

func (v V7) CommonDownloadDep(operationId string, nodeId string, nodeStepId string, msg *natsLib.Msg, resourceServerURL string, dep depMap.DepMap, saveTo string, k8sVersion string, md5Sum depMap.DepMap) (map[string]string, error) {

	stat := constants.StatusSuccessful
	var message string
	err := family.CommonDownloadDep(resourceServerURL, dep, saveTo, k8sVersion, md5Sum)
	if err != nil {
		message = fmt.Sprintf("Operation %s excution failed with node task %s on node %s due to error %v",
			operationId,
			constants.TaskDownloadDep,
			nodeId,
			err)
		log.Errorf(message)
		stat = constants.StatusError
	} else {
		message = fmt.Sprintf("Operation %v CommonDownloadDep %v done on node %v", operationId, dep, nodeId)
	}
	log.Debug("Send back message to notify server.")
	reply := family.CreateReplyBody(operationId, nodeId, stat, message, nodeStepId, nil)
	family.ReplyMsg(reply, msg)
	return nil, err
}

func (v V7) CommonDownload(operationId string, nodeId string, nodeStepId string, msg *natsLib.Msg, resourceServerURL string, FromDir string, k8sVersion string, fileList []string, saveTo string, isUseDefaultPath bool) (map[string]string, error) {

	stat := constants.StatusSuccessful
	var message string
	err := family.CommonDownload(resourceServerURL, FromDir, k8sVersion, fileList, saveTo, isUseDefaultPath, nil)
	if err != nil {
		message = fmt.Sprintf("Operation %s excution failed with node task %s on node %s due to error %v",
			operationId,
			constants.TaskDownload,
			nodeId,
			err)
		log.Errorf(message)
		stat = constants.StatusError
	} else {
		message = fmt.Sprintf("Operation %v CommonDownload %v done on node %v", operationId, fileList, nodeId)
	}
	log.Debug("Send back message to notify server.")
	reply := family.CreateReplyBody(operationId, nodeId, stat, message, nodeStepId, nil)
	family.ReplyMsg(reply, msg)
	return nil, err
}

func (v V7) GenerateLBKubeConfig(vip string) (*clientcmdapi.Config, error) {
	currentConfig, err := clientcmd.LoadFromFile("/root/.kube/config")
	if err != nil {
		return nil, err
	}
	currentCtx, exists := currentConfig.Contexts[currentConfig.CurrentContext]
	if !exists {
		return nil, fmt.Errorf("failed to find CurrentContext in Contexts of the kubeconfig file")
	}

	_, exists = currentConfig.Clusters[currentCtx.Cluster]
	if exists {
		currentConfig.Clusters[currentCtx.Cluster].Server = vip
	}
	return currentConfig, nil
}

func (v V7) GenerateKSClusterConfig(cluster schema.Cluster, ipAddress string) (map[string]string, error) {
	if cluster.KsClusterConf == nil {
		return nil, errors.New("Cluster does not contain any ks information")
	}
	template := `
apiVersion: cluster.kubesphere.io/v1alpha1
kind: Cluster
metadata:
  annotations:
    kubesphere.io/description: %s
    kubesphere.io/creator: admin
  finalizers:
    - finalizer.cluster.kubesphere.io
  generation: 3
  labels:
    cluster.kubesphere.io/group: %s
  name: %v
spec:
  connection:
    kubeconfig: %v
    kubernetesAPIEndpoint: https://%v:%d
    type: direct
  joinFederation: true
status:
  kubernetesVersion: v1.18.6
  nodeCount: %v
`

	log.Debug("GenerateKSClusterConfig")
	port := apiServerPort // default apiserver port
	vip := ipAddress
	if cluster.ClusterLB != nil {
		vip = cluster.ClusterLB.VIP
		// if the cluster is deployed with LB, the port of proxy is 9443
		port = proxiedAPIServerPort
	}
	kubeConfig, err := util.GenerateLBKubeConfig(vip, cluster.ClusterId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Debug("Get kubeconfig", kubeConfig)
	t, err := clientcmd.Write(*kubeConfig)
	if err != nil {
		log.Error("Failed to write kubeconfig", err)
		return nil, err
	}
	cb6 := base64.StdEncoding.EncodeToString(t)

	var masterNodeIdList []string
	for _, node := range cluster.Masters {
		masterNodeIdList = append(masterNodeIdList, node.NodeId)
	}
	var workerNodeIdList []string
	for _, node := range cluster.Workers {
		workerNodeIdList = append(workerNodeIdList, node.NodeId)
	}

	clusters := UnionCluter(masterNodeIdList, workerNodeIdList)

	log.Debug("Get union cluster", clusters)

	clusterType := kubesphere.KsClusterTypeTesting
	// set clusterType if post cluster cluster.KsClusterConf.Multicluster.ClusterType is proper set
	if cluster.KsClusterConf.MultiClusterConfig.ClusterType != "" {
		clusterType = cluster.KsClusterConf.MultiClusterConfig.ClusterType
	}

	clusterId := cluster.ClusterId
	if len(clusterId) > 17 {
		clusterId = cluster.ClusterId[0:16]
	}

	config := fmt.Sprintf(template, cluster.Description ,clusterType, clusterId, cb6, vip, port, len(clusters))

	p := path.Join(v.Config.YamlDataDir, fmt.Sprintf("KSConnect-%v.yaml", cluster.ClusterId))
	r := fileutils.CreateTestFileWithMD5(p, config)
	if r == "" {
		return nil, fmt.Errorf("fail to create file %v", p)
	}

	args := []string{"apply", "-f", p}
	_, stdErr, err := command.RunCmd("kubectl", args...)
	if err != nil {
		log.Errorf("Failed to run command %v due to following error:")
		errMsg := fmt.Sprintf("Global error: %s. Std error: %s", err.Error(), stdErr.String())
		log.Errorf("cmd err: %s", err.Error())
		log.Errorf("StdErr %s", errMsg)
	}

	return nil, nil
}

func UnionCluter(slice1, slice2 []string) []string {
	m := make(map[string]int)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		times, _ := m[v]
		if times == 0 {
			slice1 = append(slice1, v)
		}
	}
	return slice1
}

type Client struct {
	URL string `yaml:"url"`
}

type Labels struct {
	Job     string `yaml:"job"`
	Node    string `yaml:"node"`
	Path    string `yaml:"__path__"`
	Cluster string `yaml:"cluster"`
}

type StaticConfig struct {
	Targets []string `yaml:"targets"`
	Labels  Labels   `yaml:"labels"`
}

type ScrapeConfig struct {
	JobName       string         `yaml:"job_name"`
	StaticConfigs []StaticConfig `yaml:"static_configs"`
}

type PromTail struct {
	Server struct {
		HTTPListenPort int `yaml:"http_listen_port"`
		GrpcListenPort int `yaml:"grpc_listen_port"`
	} `yaml:"server"`
	Positions struct {
		Filename string `yaml:"filename"`
	} `yaml:"positions"`
	Clients       []Client       `yaml:"clients"`
	ScrapeConfigs []ScrapeConfig `yaml:"scrape_configs"`
}

func (v V7) ConfigPromtail(cluster schema.Cluster) (map[string]string, error) {
	yamlFile, err := ioutil.ReadFile("/etc/loki/promtail-local-config.yaml")
	if err != nil {
		log.Errorf("yamlFile.Get err #%v ", err)
		return nil, err
	}
	var pconfig PromTail
	err = yaml.Unmarshal(yamlFile, &pconfig)
	if err != nil {
		log.Errorf("Unmarshal: %v", err)
	}
	pconfig.ScrapeConfigs[0].StaticConfigs[0].Labels.Cluster = cluster.ClusterName
	if pconfig.ScrapeConfigs[0].StaticConfigs[0].Labels.Node == "CLIENT_NODE_IP" {
		pconfig.ScrapeConfigs[0].StaticConfigs[0].Labels.Node = network.GetDefaultIP(true).String()
	}
	pyaml, err := yaml.Marshal(pconfig)
	if err != nil {
		log.Errorf("yamlFile.Get err #%v ", err)
		return nil, err
	}
	err = fileutils.WriteToFile("/etc/loki/promtail-local-config.yaml", string(pyaml))
	if err != nil {
		log.Errorf("yamlFile.Get err #%v ", err)
		return nil, err
	}

	var stdErr bytes.Buffer
	_, stdErr, err = command.RunCmd("systemctl", "daemon-reload")
	if err != nil {
		log.Error("Failed to run systemctl daemon-reload due to following error:")
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
		return nil, err
	}

	cmd := exec.Command("systemctl", "restart", "promtail")
	buf := &bytes.Buffer{}
	cmd.Stdout = buf

	if err := cmd.Start(); err != nil {
		log.Error("Failed to run systemctl restart promtail due to following error:")
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
		return nil, err
	}

	return nil, nil
}

func (v V7) PreLoadImage(preLoad schema.TaskPreLoadImage, containerRuntimeType string) (map[string]string, error) {
	if preLoad.ImageDirPath == "" {
		if v.Config.LocalImagePath != "" {
			preLoad.ImageDirPath = v.Config.LocalImagePath
		} else {
			preLoad.ImageDirPath = constants.DepResourceDir + "/images/"
		}
	}

	if err := containerRuntime.LoadContainerRuntimeTar(preLoad.ImageDirPath, containerRuntimeType); err != nil {
		return nil, err
	}
	return nil, nil
}

func umountOrRemoveCRIDataDir(dev, targetDir, backupFilePath string, cri containerRuntime.INodeContainerRuntime) error {
	if cri != nil {
		if isMount, err := blockDevice.IsMountPoint(targetDir); err != nil || !isMount {
			// simply remove the container data dir
			if err := cri.CleanDataDir(); err != nil {
				return err
			}
		} else {
			// umount the device
			if err := blockDevice.Mount(dev, "", blockDevice.XFS, true, true, true, backupFilePath); err != nil {
				return err
			}
		}
	}
	return nil
}

func mountIfCriDevIsSet(mountPoint string, config client.Config, enableBootCheck bool, umount bool, force bool) error {

	if config.CRIMountDev == "" {
		log.Warn("Cri device is not set, skip mount use os disk")
		return nil
	}

	if result, err := bd.DeviceIsReady(config.CRIMountDev); err != nil {
		log.Errorf("Failed to check Cri Device %s due to error %s.", config.CRIMountDev, err.Error())
		return err
	} else if !result {
		log.Warnf("Cri Device %s is not ready please ensure the device is ready", mountPoint)
		return err
	}
	if err := linuxFamily.MountOrUmount(config.CRIMountDev, mountPoint, blockDevice.XFS, enableBootCheck, umount, force, config.YamlDataDir); err != nil {
		return err
	}
	return nil
}
