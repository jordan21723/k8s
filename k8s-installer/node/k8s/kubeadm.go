package k8s

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"k8s-installer/node/os/family"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/command"
	"k8s-installer/pkg/config/client"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util"
	"k8s-installer/pkg/util/fileutils"
	"k8s-installer/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeProxyConfigSchema "k8s.io/kube-proxy/config/v1alpha1"
	"k8s.io/kubelet/config/v1beta1"
	kubeadmConfigSchema "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta2"
)

func PrintJoinString(kubeadm schema.TaskPrintJoin) (map[string]string, error) {
	result := map[string]string{}
	var err error
	var stdErr, stdOut bytes.Buffer
	stdOut, stdErr, err = command.RunCmd("kubeadm", "token", "create", "--print-join-command")
	if err != nil {
		log.Errorf("Fail to run  kubeadm token create --print-join-command due to error %s", errors.New(stdErr.String()))
		return result, err
	} else {
		log.Debug(stdOut.String())
	}
	out := stdOut.String()

	log.Debugf("Got print result %s", out)
	out = strings.Replace(out, "    ", "", -1)
	out = out[:len(out)-2]
	result[constants.ReturnDataKeyJoinWorkerCMD] = strings.Replace(out, "\\\n", "", -1)
	return result, nil
}

func InitFirstControlPlane(kubeadm schema.TaskKubeadm, config client.Config) (map[string]string, error) {
	returnData := map[string]string{}

	var err error
	var controlJoinCommand, workerJoinCommand string
	retry := 0

	for retry < 5 {
		if retry > 0 {
			// wait 3 seconds before retry
			time.Sleep(3 * time.Second)
			log.Warnf("Failed to init first control plane for %d time, we are going to retry with following step: 1. kubeadm reset -f 2. run kubeadm init again", retry)
			kubeadmReset()
		}
		log.Debugf("Attempt to init first control plane with command")
		if controlJoinCommand, workerJoinCommand, err = kubeadmInit(string(kubeadm.KubeadmConfig), config.YamlDataDir); err != nil {
			log.Errorf("Failed to init first control plane due to error %s", err.Error())
			retry += 1
		} else {
			returnData[constants.ReturnDataKeyJoinControlPlaneCMD] = controlJoinCommand
			returnData[constants.ReturnDataKeyJoinWorkerCMD] = workerJoinCommand
			log.Debugf("Got join control plan command %s", controlJoinCommand)
			log.Debugf("Git join worker command %s", workerJoinCommand)
			return returnData, nil
		}
	}
	return returnData, err
}

func DestroyKubeNode(kubeadm schema.TaskKubeadm, config client.Config, cluster schema.Cluster) (map[string]string, error) {
	var stdErr bytes.Buffer
	var err error

	if err := clusterIPClearUp(cluster.ControlPlane.EnableIPVS); err != nil {
		return nil, err
	}

	// stop kubelet service
	if err := family.StopSystemdService("kubelet.service"); err != nil {
		return nil, err
	}

	if err := kubeadmReset(); err != nil {
		return nil, err
	}

	log.Debug("Ensure /etc/kubernetes/ directory is deleted")
	_, stdErr, err = command.RunCmd("rm", "-rf", "/etc/kubernetes")
	if err != nil {
		log.Errorf("Failed to run command rm -rf /etc/kubernetes due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return nil, err
	}

	if cluster.ControlPlane.KubernetesVersion == constants.V1_18_6 {
		packages := fileutils.GetDepList(KubeDepMapping)
		subCommand := []string{"-e", "--nodeps"}
		subCommand = append(subCommand, packages...)
		// run setenforce 0
		log.Debug("Try to run command yum remove")
		_, stdErr, err = command.RunCmd("rpm", subCommand...)
		if err != nil {
			log.Warnf("Failed to removed all kubernetes package due to following error:")
			log.Warnf("StdErr %s", stdErr.String())
		}
	} else {
		err := util.RemoveInstalledRPM()
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	return nil, nil
}

func clusterIPClearUp(ipvs bool) error {
	var stdErr bytes.Buffer
	var err error

	if ipvs {
		log.Debug("Try to run command ipvsadm --clear")
		_, stdErr, err = command.RunCmd("ipvsadm", "--clear")
		if err != nil {
			log.Errorf("Failed to run command ipvsadm --clear due to following error:")
			log.Errorf("StdErr %s", stdErr.String())
			return err
		}
		log.Debug("Try to run command ip link delete kube-ipvs0")
		_, stdErr, err = command.RunCmd("ip", "link", "delete", "kube-ipvs0")
		if err != nil {
			log.Errorf("(Ignored) Failed to run command ip link delete kube-ipvs0 due to following error:")
			log.Errorf("(Ignored) StdErr %s", stdErr.String())
			// ignore error when device not exists
		}
	}

	// ensure vxlan.calico is delete when calico use vxlan mode
	_, stdErr, err = command.RunCmd("ip", "link", "delete", "vxlan.calico")
	if err != nil {
		log.Errorf("(Ignored) Failed to run command ip link delete vxlan.calico due to following error:")
		log.Errorf("(Ignored) StdErr %s", stdErr.String())
		//return err
	}

	// ensure all route is released completely
	_, stdErr, err = command.RunCmd("modprobe", "-r", "ipip")
	if err != nil {
		log.Errorf("(Ignored) Failed to run command modprobe -r ipip due to following error:")
		log.Errorf("(Ignored) StdErr %s", stdErr.String())
		//return err
	}

	log.Debug("Try to clean iptables with commands iptables -F && iptables -t nat -F && iptables -t mangle -F && iptables -X ")
	_, stdErr, err = command.RunCmd("iptables", "-F")
	if err != nil {
		log.Errorf("(Ignored) Failed to run command iptables -F due to following error:")
		log.Errorf("(Ignored) StdErr %s", stdErr.String())
		//return err
	}

	_, stdErr, err = command.RunCmd("iptables", "-t", "nat", "-F")
	if err != nil {
		log.Errorf("(Ignored) Failed to run command iptables -t nat -F due to following error:")
		log.Errorf("(Ignored) StdErr %s", stdErr.String())
		//return err
	}

	_, stdErr, err = command.RunCmd("iptables", "-t", "mangle", "-F")
	if err != nil {
		log.Errorf("(Ignored) Failed to run command iptables -t mangle -F due to following error:")
		log.Errorf("(Ignored) StdErr %s", stdErr.String())
		//return err
	}

	_, stdErr, err = command.RunCmd("iptables", "-X")
	if err != nil {
		log.Errorf("(Ignored) Failed to run command iptables -X due to following error:")
		log.Errorf("(Ignored) StdErr %s", stdErr.String())
		//return err
	}

	return nil
}

func kubeadmReset() error {
	log.Debug("Try to reset kubeadm with kubeadm reset -f")
	var err error
	var stdErr, out bytes.Buffer
	out, stdErr, err = command.RunCmd("kubeadm", "reset", "-f")
	if err != nil {
		log.Debugf("Failed to reset with kubeadm due to error: %s", stdErr.String())
		return err
	} else {
		log.Debug(out.String())
	}
	return nil
}

func JoinControlPlane(preStepReturnData map[string]string) error {

	log.Debugf("Print preStepReturnData %v", preStepReturnData)
	if _, exists := preStepReturnData[constants.ReturnDataKeyJoinControlPlaneCMD]; !exists {
		return errors.New("Cannot found join control plane command in return date ")
	}

	// remove kubeadm join from return join control plane command
	// it`s not needed because we use kubeadm join lib
	args := strings.Split(preStepReturnData[constants.ReturnDataKeyJoinControlPlaneCMD], " ")[1:]

	retry := 0
	var err error
	for retry < 5 {
		var stdErr, out bytes.Buffer
		out, stdErr, err = command.RunCmd("kubeadm", args...)
		if err == nil {
			log.Debug(out.String())
			return nil
		}

		retry++
		log.Debugf("Failed to join control plane due to error %s", stdErr.String())
		log.Warnf("Failed to init first control plane for %d time, we are going to retry with following step: 1. kubeadm reset -f 2. run kubeadm init again", retry)
		time.Sleep(3 * time.Second)
		kubeadmReset()
	}

	return err
}

func JoinWorker(preStepReturnData map[string]string) error {
	log.Debugf("Print preStepReturnData %v", preStepReturnData)
	if _, exists := preStepReturnData[constants.ReturnDataKeyJoinWorkerCMD]; !exists {
		return errors.New("Cannot found join worker command in return date ")
	}

	// remove kubeadm join from return join worker command
	// it`s not needed because we use kubeadm join lib
	args := strings.Split(preStepReturnData[constants.ReturnDataKeyJoinWorkerCMD], " ")[1:]

	retry := 0

	var err error
	for retry < 5 {
		if retry > 0 {
			// wait 3 seconds before retry
			time.Sleep(3 * time.Second)
			log.Warnf("Failed to join worker we are going to retry %d time with following step: 1. kubeadm reset -f 2. run kubeadm join again", retry)
			kubeadmReset()
		}
		// replace api address to local vip 127.0.0.1:6443
		args[1] = "127.0.0.1:6443"

		var stdErr, out bytes.Buffer
		out, stdErr, err = command.RunCmd("kubeadm", args...)
		if err != nil {
			log.Debugf("Failed to join worker due to error %s", stdErr.String())
			retry += 1
		} else {
			log.Debug(out.String())
			return nil
		}
	}

	return err
}

func JoinRancherNode(preStepReturnData map[string]string) error {
	args := strings.Split(preStepReturnData[constants.ReturnDataKeyJoinRancherNodeCMD], " ")[1:]

	log.Debugf("Print Rancher command string: %s", preStepReturnData[constants.ReturnDataKeyJoinRancherNodeCMD])

	var err error
	var stdErr, out bytes.Buffer
	out, stdErr, err = command.RunCmd("docker", args...)
	if err != nil {
		log.Errorf("Failed to join rancher node due to error %s", stdErr.String())
	} else {
		log.Debug("rancher join node command output:", out.String())
		return nil
	}

	return err
}

/*
kubeadmInit should return join control plane command and join worker command
*/
func kubeadmInit(kubeadmConfig, configSavePath string) (string, string, error) {
	saveTo := configSavePath + "/kubeadm-config.yaml"
	joinControlPlaneCMD := ""
	joinWorkerCMD := ""
	var err error

	log.Debugf("Ensure dir %s exists, create it if not", configSavePath)

	if err := util.CreateDirIfNotExists(configSavePath); err != nil {
		log.Errorf("Fail to create dir %s due to error: %v", configSavePath, err.Error())
		return "", "", err
	}

	log.Debugf("Try to save kubeadmConfig to location %s", saveTo)
	// before we kubeadm init --config xxx.yaml we should save kubeadm config to
	// local file system which is pre config in every client config file property: YamlDataDir
	if err = util.WriteTxtToFile(saveTo, kubeadmConfig); err != nil {
		log.Debugf("Failed to save kubeadmConfig to %s due to error %s", saveTo, err.Error())
		return "", "", err
	}
	var stdErr, out bytes.Buffer
	out, stdErr, err = command.RunCmd("kubeadm", []string{"init", "--config", saveTo, "--upload-certs"}...)
	if err != nil {
		log.Errorf("Fail to run kubeadm init --config %s --upload-certs due to error %s", saveTo, stdErr.String())
		return "", "", errors.New(stdErr.String())
	} else {
		log.Debugf(out.String())
	}

	joinControlPlaneCMD = getJoinFromStdOut(out.String(), "You can now join any number of the control-plane node running the following command on each as root:")
	joinWorkerCMD = getJoinFromStdOut(out.String(), "Then you can join any number of worker nodes by running the following on each as root:")
	return joinControlPlaneCMD, joinWorkerCMD, err
}

func getJoinFromStdOut(output string, cutBegin string) string {
	outputSlice := strings.Split(output, "\n")
	controlPlaneJoinStartIndex := 0
	for index, val := range outputSlice {
		if val == cutBegin {
			controlPlaneJoinStartIndex = index
		}
	}
	result := strings.Join([]string{strings.TrimSpace(outputSlice[controlPlaneJoinStartIndex+2]), strings.TrimSpace(outputSlice[controlPlaneJoinStartIndex+3]), strings.TrimSpace(outputSlice[controlPlaneJoinStartIndex+4])}, "\n")
	return strings.Replace(result, "\\\n", "", -1)
}

func CreateKubeadmConfigFromCluster(cluster schema.Cluster, nodeInformationCollection schema.NodeInformationCollection) (string, error) {
	var configJson []string
	clusterConfiguration := kubeadmConfigSchema.ClusterConfiguration{}
	clusterConfiguration.APIVersion = "kubeadm.k8s.io/v1beta2"
	clusterConfiguration.Kind = "ClusterConfiguration"
	clusterConfiguration.ClusterName = cluster.ClusterName
	clusterConfiguration.Etcd.Local = new(kubeadmConfigSchema.LocalEtcd)

	serverConfig := cache.GetCurrentCache().GetServerRuntimeConfig(cache.NodeId)
	clusterConfiguration.Etcd.Local.DataDir = filepath.Join(serverConfig.K8sEtcdDataDir, "etcd")
	clusterConfiguration.Etcd.Local.ExtraArgs = map[string]string{"auto-compaction-retention": "1", "snapshot-count": "5000", "quota-backend-bytes": "8589934592", "heartbeat-interval": "300", "election-timeout": "1500"}
	if cluster.ControlPlane.KubernetesVersion == "" {
		clusterConfiguration.KubernetesVersion = constants.V1_18_6
	} else {
		clusterConfiguration.KubernetesVersion = cluster.ControlPlane.KubernetesVersion
	}
	clusterConfiguration.Scheduler.ExtraArgs = map[string]string{"bind-address": "0.0.0.0", "port": "10251"}
	if cluster.ContainerRuntime.PrivateRegistryAddress != "" && cluster.ContainerRuntime.PrivateRegistryPort != 0 {
		clusterConfiguration.ImageRepository = fmt.Sprintf("%s:%d", cluster.ContainerRuntime.PrivateRegistryAddress,
			cluster.ContainerRuntime.PrivateRegistryPort)
	} else {
		log.Warnf("Private registry not proper set fallback to default image registry.")
	}
	clusterConfiguration.FeatureGates = map[string]bool{}
	for key, val := range cluster.ControlPlane.FeatureGates {
		clusterConfiguration.FeatureGates[key] = val
	}
	clusterConfiguration.ControlPlaneEndpoint = nodeInformationCollection[cluster.Masters[0].NodeId].Ipv4DefaultIp + ":6443"

	if cluster.ClusterLB != nil {
		clusterConfiguration.APIServer = kubeadmConfigSchema.APIServer{
			CertSANs: []string{"127.0.0.1", cluster.ClusterLB.VIP},
		}
	} else {
		clusterConfiguration.APIServer = kubeadmConfigSchema.APIServer{
			CertSANs: []string{"127.0.0.1"},
		}
	}

	localtimeMount := kubeadmConfigSchema.HostPathMount{}
	localtimeMount.Name = "localtime"
	localtimeMount.HostPath = "/etc/localtime"
	localtimeMount.MountPath = "/etc/localtime"
	localtimeMount.ReadOnly = true
	localtimeMount.PathType = "File"
	clusterConfiguration.APIServer.ExtraVolumes = []kubeadmConfigSchema.HostPathMount{localtimeMount}
	clusterConfiguration.ControllerManager.ExtraVolumes = []kubeadmConfigSchema.HostPathMount{localtimeMount}
	clusterConfiguration.Scheduler.ExtraVolumes = []kubeadmConfigSchema.HostPathMount{localtimeMount}

	createNetworkingConfig(cluster, &clusterConfiguration)

	if data, err := json.Marshal(&clusterConfiguration); err != nil {
		log.Errorf("Failed to Marshal ClusterConfiguration to json due to error: %s", err.Error())
		return "", err
	} else {
		configJson = append(configJson, string(data))
	}
	if cluster.ControlPlane.EnableIPVS {
		if data, err := createKubeProxyConfig(cluster); err != nil {
			log.Errorf("Failed to Marshal KubeProxyConfiguration to json due to error: %s", err.Error())
			return "", err
		} else {
			configJson = append(configJson, data)
		}
	}
	if cluster.ContainerRuntime.CRIType == constants.CRITypeContainerd {
		if data, err := createKubeletInitConfiguration(cluster, &clusterConfiguration); err != nil {
			log.Errorf("Failed to Marshal KubeletInitConfiguration to json due to error: %s", err.Error())
			return "", err
		} else {
			configJson = append(configJson, data)
		}

		if data, err := createKubeadmInitConfiguration(cluster, &clusterConfiguration); err != nil {
			log.Errorf("Failed to Marshal createKubeadmInitConfiguration to json due to error: %s", err.Error())
			return "", err
		} else {
			configJson = append(configJson, data)
		}
	}

	return strings.Join(configJson, "\n---\n"), nil
}

func createNetworkingConfig(cluster schema.Cluster, configuration *kubeadmConfigSchema.ClusterConfiguration) {
	networkConfig := kubeadmConfigSchema.Networking{}
	if cluster.CNI.CNIType == constants.CNITypeCalico {
		if cluster.CNI.Calico.EnableDualStack {
			networkConfig.ServiceSubnet = cluster.ControlPlane.ServiceV4CIDR + "," + cluster.ControlPlane.ServiceV6CIDR
			networkConfig.PodSubnet = cluster.CNI.PodV4CIDR + "," + cluster.CNI.PodV6CIDR
			// ensure IPv6DualStack FeatureGates enabled
			configuration.FeatureGates["IPv6DualStack"] = true
		} else {
			networkConfig.ServiceSubnet = cluster.ControlPlane.ServiceV4CIDR
			networkConfig.PodSubnet = cluster.CNI.PodV4CIDR
		}
	} else {
		networkConfig.ServiceSubnet = cluster.ControlPlane.ServiceV4CIDR
		networkConfig.PodSubnet = cluster.CNI.PodV4CIDR
	}
	networkConfig.DNSDomain = cluster.ClusterName
	configuration.Networking = networkConfig
}

func createKubeProxyConfig(cluster schema.Cluster) (string, error) {
	var mode kubeProxyConfigSchema.ProxyMode
	mode = "ipvs"
	if !cluster.ControlPlane.EnableIPVS {
		mode = "iptables"
	}
	proxyConfig := kubeProxyConfigSchema.KubeProxyConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeProxyConfiguration",
			APIVersion: "kubeproxy.config.k8s.io/v1alpha1",
		},
		Mode: mode,
	}
	if data, err := json.Marshal(&proxyConfig); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func createKubeletInitConfiguration(cluster schema.Cluster, configuration *kubeadmConfigSchema.ClusterConfiguration) (string, error) {
	KubeletConfiguration := v1beta1.KubeletConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeletConfiguration",
			APIVersion: "kubelet.config.k8s.io/v1beta1",
		},
	}
        if cluster.CNI.Calico.EnableDualStack {
                KubeletConfiguration.FeatureGates["IPv6DualStack"] = true
        }

	KubeletConfiguration.CgroupDriver = "systemd"

	if cluster.ContainerRuntime.CRIType != constants.CRITypeDocker {
		rtc := cache.GetCurrentCache()
		config := rtc.GetServerRuntimeConfig(cache.NodeId)
		KubeletConfiguration.ContainerLogMaxFiles = &config.MaxContainerLogFile
		KubeletConfiguration.ContainerLogMaxSize = fmt.Sprintf("%dMi", config.MaxContainerLogFileSize)
	}

	if data, err := json.Marshal(&KubeletConfiguration); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func createKubeadmInitConfiguration(cluster schema.Cluster, configuration *kubeadmConfigSchema.ClusterConfiguration) (string, error) {
	clusterConfiguration := kubeadmConfigSchema.InitConfiguration{}
	clusterConfiguration.APIVersion = "kubeadm.k8s.io/v1beta2"
	clusterConfiguration.Kind = "InitConfiguration"
	clusterConfiguration.NodeRegistration = kubeadmConfigSchema.NodeRegistrationOptions{
		CRISocket: "/run/containerd/containerd.sock",
	}
	if data, err := json.Marshal(&clusterConfiguration); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}
