package containerd

import (
	"bytes"
	"errors"
	"fmt"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/containerd"
	"k8s-installer/pkg/dep"
	"strconv"

	"k8s-installer/node/container_runtime"
	osInfoProvider "k8s-installer/node/os"

	"k8s-installer/node/os/family"
	"k8s-installer/node/os/family/centos"
	"k8s-installer/pkg/command"
	"k8s-installer/pkg/config/client"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util"
	"k8s-installer/pkg/util/fileutils"
	"k8s-installer/schema"
)

type ContainerD struct {
}

func (C ContainerD) v7(offline bool, taskCRI schema.TaskCRI, cluster schema.Cluster, config client.Config, resourceServerURL string, k8sVersion string, md5Dep dep.DepMap) error {
	var err error
	if offline {
		log.Debug("Offline is set to true. Go offline docker installation script")
		installContainerdOffline(cluster, config, resourceServerURL, k8sVersion, md5Dep)
	} else {
		log.Debug("Offline is set to false. Go online docker installation script")
		err = installContainerdOnline()
	}
	if err != nil {
		log.Errorf("Failed to install docker due to error %s", err.Error())
		return err
	}

	if taskCRI.CRIType.PrivateRegistryAddress != "" {
		if err := setUpContainerdConfigFile(taskCRI, cluster); err != nil {
			return err
		}
	}

	err = family.StartSystemdService(true, true, "containerd")
	if err != nil {
		log.Errorf("Failed to start containerd daemon due to error %s", err.Error())
	}

	container_runtime.LoadContainerRuntimeTar(config.LocalImagePath, constants.CRITypeContainerd)

	return err
}

func (C ContainerD) Install(offline bool, taskCRI schema.TaskCRI, cluster schema.Cluster, config client.Config, resourceServerURL, k8sVersion string, md5Dep dep.DepMap) error {
	return C.v7(offline, taskCRI, cluster, config, resourceServerURL, k8sVersion, md5Dep)
}

func installContainerdOnline() error {

	osInfo, errOSInfo := osInfoProvider.GetAllSystemInformation()

	if errOSInfo != nil {
		log.Errorf("Failed to get node cpu arch due to error %s", errOSInfo.Error())
		return errOSInfo
	}

	cpuArch := osInfo.Kernel.Architecture

	var stdErr bytes.Buffer
	var err error
	log.Debug("Installing yum-utils device-mapper-persistent-data lvm2 for containerd")
	_, stdErr, err = command.RunCmd("yum", "install", "yum-utils", "device-mapper-persistent-data", "lvm2", "-y")
	if err != nil {
		log.Errorf("Failed to install containerd due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	log.Debug("yum-utils device-mapper-persistent-data lvm2 install completed")

	log.Debug("Download repo for containerd")
	_, stdErr, err = command.RunCmd("yum-config-manager", "--add-repo", "https://download.docker.com/linux/centos/docker-ce.repo")
	if err != nil {
		log.Errorf("Failed to install containerd due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	log.Debug("Download repo completed")

	log.Debug("Install containerd")
	_, stdErr, err = command.RunCmd("yum", "update", "-y", "&&", "yum", "install", "containerd.io", "-y")
	if err != nil {
		log.Errorf("Failed to install containerd due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	log.Debug("Download repo completed")

	log.Debug("Download repo for kata")
	_, stdErr, err = command.RunCmd("yum-config-manager", "--add-repo", fmt.Sprintf("http://download.opensuse.org/repositories/home:/katacontainers:/releases:/%s:/stable-1.11/CentOS_7/home:katacontainers:releases:%s:stable-1.11.repo", cpuArch, cpuArch))
	if err != nil {
		log.Errorf("Failed to download kata repo due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	log.Debug("Download repo kata completed")

	log.Debug("Install package for kata")
	_, stdErr, err = command.RunCmd("yum", "install", "kata-runtime", "kata-proxy", "kata-shim", "-y")
	if err != nil {
		log.Errorf("Failed to install kata due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	log.Debug("Install kata package completed")

	// TODO: Test on physical machine
	// log.Debug("check env for kata")
	// _, stdErr, err = command.RunCmd("kata-runtime", "kata-check")
	// if err != nil {
	// 	log.Errorf("Failed to check kata env due to following error:")
	// 	log.Errorf("StdErr %s", stdErr.String())
	// 	return err
	// }
	// log.Debug("Check kata env completed")
	return nil
}

func (c ContainerD) Remove(config schema.TaskCRI) error {
	var stdErr bytes.Buffer
	var err error
	log.Debugf("Attempt to remove all container and it`s task")
	if err := containerd.DeleteAllContainer("k8s.io"); err != nil {
		return err
	}

	family.StopSystemdService("containerd")
	family.StopSystemdService("docker")

	err = centos.UninstallPackage([]string{"device-mapper-persistent-data", "containerd.io", "kata-runtime", "kata-proxy", "kata-shim"})
	if err != nil {
		return err
	}

	log.Debug("Attempt to remove containerd folder /etc/containerd")
	_, stdErr, err = command.RunCmd("rm", "-rf", "/etc/containerd")
	if err != nil {
		log.Error("Failed to remove /etc/containerd due to following error:")
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
	}

	log.Debug("Attempt to remove containerd kubelet config /etc/systemd/system/kubelet.service.d/0-containerd.conf")
	_, stdErr, err = command.RunCmd("rm", "-rf", "/etc/systemd/system/kubelet.service.d/0-containerd.conf")
	if err != nil {
		log.Error("Failed to remove /etc/systemd/system/kubelet.service.d/0-containerd.conf due to following error:")
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
	}

	log.Debug("Attempt to remove containerd kubelet config /usr/lib/systemd/system/docker.service")
	_, stdErr, err = command.RunCmd("rm", "-rf", "/usr/lib/systemd/system/docker.service")
	if err != nil {
		log.Error("Failed to remove /usr/lib/systemd/system/docker.service due to following error:")
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
	}

	return err
}

func setUpContainerdConfigFile(config schema.TaskCRI, cluster schema.Cluster) error {

	httpTemplate :=
		`    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."%s"]
      endpoint = ["http://%s"]`

	newLine := `
`

	httpsTemplate :=
		`    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."%s"]
      endpoint = ["https://%s"]
    [plugins."io.containerd.grpc.v1.cri".registry.configs."%s".tls]
      ca_file = "/etc/ssl/certs/%s/ca.crt"`

	templateContainerdConfig :=
		`
version = 2
root = "%s/containerd"
state = "/run/containerd"
oom_score = 0

[grpc]
  address = "/run/containerd/containerd.sock"
  uid = 0
  gid = 0
  max_recv_message_size = 16777216
  max_send_message_size = 16777216

[debug]
  address = ""
  uid = 0
  gid = 0
  level = ""

[metrics]
  address = ""
  grpc_histogram = false

[cgroup]
  path = ""

[plugins."io.containerd.grpc.v1.cri"]
  stream_server_address = "127.0.0.1"
  stream_server_port = "0"
  enable_selinux = false
  sandbox_image = "%s:%d/pause:3.2"
  stats_collect_period = 10
  systemd_cgroup = false
  enable_tls_streaming = false
  max_container_log_line_size = 16384
  disable_proc_mount = false
  [plugins."io.containerd.grpc.v1.cri".containerd]
    snapshotter = "overlayfs"
    no_pivot = false
    default_runtime_name = "runc"
    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
      runtime_type = "io.containerd.runc.v2"
      pod_annotations = []
      container_annotations = []
      privileged_without_host_devices = false
      base_runtime_spec = ""
  [plugins."io.containerd.grpc.v1.cri".cni]
    bin_dir = "/opt/cni/bin"
    conf_dir = "/etc/cni/net.d"
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
      endpoint = ["https://registry-1.docker.io"]
%s
`

	fileutils.CreateDirectory("/etc/containerd")

	type addressPort struct {
		address string
		port    string
	}

	var registries = map[string][]addressPort{
		"http":  {},
		"https": {},
	}

	registries["http"] = append(registries["http"], addressPort{cluster.ContainerRuntime.PrivateRegistryAddress, strconv.Itoa(cluster.ContainerRuntime.PrivateRegistryPort)})

	if cluster.Harbor != nil {
		if cluster.Harbor.Enable {
			if !cluster.Harbor.EnableTls {
				registries["http"] = append(registries["http"], addressPort{cluster.Harbor.Ip, strconv.Itoa(cluster.Harbor.Port)})
				if len(cluster.Harbor.Host) > 0 {
					registries["http"] = append(registries["http"], addressPort{cluster.Harbor.Host, strconv.Itoa(cluster.Harbor.Port)})
				}
			} else {
				registries["https"] = append(registries["https"], addressPort{cluster.Harbor.Ip, strconv.Itoa(cluster.Harbor.Port)})
				if len(cluster.Harbor.Host) > 0 {
					registries["https"] = append(registries["https"], addressPort{cluster.Harbor.Host, strconv.Itoa(cluster.Harbor.Port)})
				}
			}
		}
	}

	first := true
	fullConfig := ""
	for _, registry := range registries["http"] {
		port := ""
		// hide default 80
		if registry.port != "80" {
			port = ":" + registry.port
		}
		regAddress := fmt.Sprintf("%s%s", registry.address, port)
		if first {
			fullConfig += fmt.Sprintf(httpTemplate, regAddress, regAddress)
			first = false
		} else {
			fullConfig += newLine
			fullConfig += fmt.Sprintf(httpTemplate, regAddress, regAddress)
		}
	}

	for _, registry := range registries["https"] {
		port := ""
		// hide default 443
		if registry.port != "443" {
			port = ":" + registry.port
		}
		regAddress := fmt.Sprintf("%s%s", registry.address, port)
		if first {
			fullConfig += fmt.Sprintf(httpsTemplate, regAddress, regAddress, regAddress, registry.address)
			first = false
		} else {
			fullConfig += newLine
			fullConfig += fmt.Sprintf(httpsTemplate, regAddress, regAddress, regAddress, registry.address)
		}
	}

	rtc := cache.GetCurrentCache()
	clientConfig := rtc.GetClientRuntimeConfig(cache.NodeId)

	err := util.WriteTxtToFile("/etc/containerd/config.toml", fmt.Sprintf(templateContainerdConfig, clientConfig.CRIRootDir, cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort, fullConfig))
	if err != nil {
		log.Debugf("Failed to set containerd config file /etc/containerd/config.toml due to error %s: ", err.Error())
		return err
	}
	log.Debug("Done with containerd config file /etc/containerd/config.toml")

	template :=
		`
#containerd
KUBELET_EXTRA_ARGS=--container-runtime=remote --runtime-request-timeout=15m --container-runtime-endpoint=unix:///run/containerd/containerd.sock --image-service-endpoint=unix:///run/containerd/containerd.sock --pod-infra-container-image=%s:%v/pause:3.2
#containerd
`
	template = fmt.Sprintf(template, cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort)
	err = fileutils.AppendFile("/etc/sysconfig/kubelet", template, "#containerd")
	if err != nil {
		log.Debugf("Failed to set kubelet config file /etc/sysconfig/kubelet due to error %s: ", err.Error())
		return err
	}
	log.Debug("Done with containerd config file /etc/sysconfig/kubelet")
	return nil
}

func installContainerdOffline(cluster schema.Cluster, config client.Config, resourceServerURL string, k8sVersion string, md5Dep dep.DepMap) error {

	saveTo := config.YamlDataDir + "/kata"

	if err := family.CommonDownloadDep(resourceServerURL, containerdVersion, saveTo,
		k8sVersion, md5Dep); err != nil {
		return err
	}

	var stdErr bytes.Buffer
	var err error

	log.Debug("Patch lvm install error.")
	_, stdErr, err = command.RunCmd("bash", "-c", `useradd -s /sbin/nologin mockbuild`)
	if err != nil {
		log.Debugf("Failed to run useradd -s /sbin/nologin mockbuild  to following error:")
		log.Debugf("StdErr %s", stdErr.String())
	}

	// install yum-utils
	log.Debug("Installing all containerd rpms")
	_, stdErr, err = command.RunCmd("rpm", "-ivh", "--replacefiles", "--replacepkgs", "--nodeps", saveTo+"/*.rpm")
	if err != nil {
		log.Errorf("Failed to install containerd related rpm due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}

	return nil
}

func (C ContainerD) CleanDataDir() error {
	var stdErr bytes.Buffer
	var err error
	rtc := cache.GetCurrentCache()
	clientConfig := rtc.GetClientRuntimeConfig(cache.NodeId)
	containerdRootDir := clientConfig.CRIRootDir + "/containerd"

	log.Debugf("Attempt to remove containerd root data dir folder %s", containerdRootDir)
	_, stdErr, err = command.RunCmd("rm", "-rf", containerdRootDir)
	if err != nil {
		log.Errorf("(Ignored) Failed to remove containerd root data dir %s due to following error:", containerdRootDir)
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
	}
	return nil
}
