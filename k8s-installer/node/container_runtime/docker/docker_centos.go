package docker

import (
	"bytes"
	"errors"
	"fmt"
	"k8s-installer/node/container_runtime"
	osInfoProvider "k8s-installer/node/os"
	"k8s-installer/pkg/cache"
	"strconv"
	"strings"
	"time"

	"k8s-installer/node/os/family"
	"k8s-installer/node/os/family/centos"
	"k8s-installer/pkg/command"
	"k8s-installer/pkg/config/client"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util"
	"k8s-installer/pkg/util/fileutils"
	"k8s-installer/schema"
)

type CRIDockerCentos struct {
}

func (C CRIDockerCentos) v7(offline bool, taskCRI schema.TaskCRI, cluster schema.Cluster, config client.Config, resourceServerURL string, k8sVersion string, md5Dep dep.DepMap) error {
	var err error
	if offline {
		log.Debug("Offline is set to true. Go offline docker installation script")
		err = installDockerOffline(cluster, config, resourceServerURL, k8sVersion, md5Dep)
	} else {
		log.Debug("Offline is set to false. Go online docker installation script")
		err = installDockerOnline()
	}
	if err != nil {
		log.Errorf("Failed to install docker due to error %s", err.Error())
		return err
	}

	if taskCRI.CRIType.PrivateRegistryAddress != "" {
		if err := setUpDockerConfigFile(taskCRI, cluster); err != nil {
			return err
		}
	}

	err = family.StartSystemdService(true, true, "docker")
	if err != nil {
		log.Errorf("Failed to start docker daemon due to error %s", err.Error())
	}

	container_runtime.LoadContainerRuntimeTar(config.LocalImagePath, constants.CRITypeDocker)
	return err
}

func (C CRIDockerCentos) Install(offline bool, taskCRI schema.TaskCRI, cluster schema.Cluster, config client.Config, resourceServerURL, k8sVersion string, md5Dep dep.DepMap) error {
	return C.v7(offline, taskCRI, cluster, config, resourceServerURL, k8sVersion, md5Dep)
}

func (C CRIDockerCentos) Remove(config schema.TaskCRI) error {
	if major, sub, version, err := osInfoProvider.GetKernelVersion(); err != nil {
		log.Errorf("Failed to get kernel version due to error: %s", err.Error())
		return err
	} else if major == 3 && sub >= 10 {
		// for old version < 3.10 to prevent rm -rf /var/lib/docker device or resource busy issue
		log.Debugf("Detected old kernel version %d.%d-%d go with MountFlags=slave", major, sub, version)
		if err := kernelM3S10patch(); err != nil {
			return err
		}
	}
	return removeDocker()
}

func installDockerOnline() error {
	var stdErr bytes.Buffer
	var err error
	// install yum-utils
	log.Debug("Installing yum-utils for docker")
	_, stdErr, err = command.RunCmd("yum", "install", "yum-utils", "-y")
	if err != nil {
		log.Errorf("Failed to install docker due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	log.Debug("yum-utils install completed")

	log.Debug("Download repo for docker")
	_, stdErr, err = command.RunCmd("yum-config-manager", "--add-repo", "https://download.docker.com/linux/centos/docker-ce.repo")
	if err != nil {
		log.Errorf("Failed to install docker due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	log.Debug("Download repo completed")

	log.Debug("Install docker-ce")
	_, stdErr, err = command.RunCmd("yum", "install", "docker-ce", "-y")
	if err != nil {
		log.Errorf("Failed to install docker due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	log.Debug("Download repo completed")
	return nil
}

func removeDocker() error {
	if err := family.StopSystemdService("docker.service"); err != nil {
		return err
	}

	if err := family.StopSystemdService("containerd"); err != nil {
		return err
	}

	time.Sleep(5*time.Second)

	log.Debug("Attempt to remove all related sock")
	_, stdErr, err := command.RunCmd("rm", "-f", "/var/run/docker.sock")
	if err != nil {
		log.Errorf("Failed to remove /var/run/docker.sock due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
	}

	_, stdErr, err = command.RunCmd("rm", "-f", "/var/run/dockershim.sock")
	if err != nil {
		log.Errorf("Failed to remove /var/run/dockershim.sock due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
	}

	_, stdErr, err = command.RunCmd("rm", "-f", "/var/run/docker")
	if err != nil {
		log.Errorf("Failed to remove /var/run/docker due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
	}

	fileList := fileutils.GetDepList(dockerVersion)
	return centos.UninstallPackage(fileList)
}

func setUpDockerConfigFile(cri schema.TaskCRI, cluster schema.Cluster) error {
	template := `{
  "data-root": "%s/docker",
  "insecure-registries": [%s],
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "%dm",
    "max-file": "%d"
  },
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.override_kernel_check=true"
  ]
}`
	fileutils.CreateDirectory("/etc/docker")
	mainRegistry := cri.CRIType.PrivateRegistryAddress + ":" + strconv.Itoa(cri.CRIType.PrivateRegistryPort)
	registries := fmt.Sprintf("\"%s\"", mainRegistry)
	if cluster.KsClusterConf != nil {
		ksRegistryServer := strings.TrimSpace(cluster.KsClusterConf.ServerConfig.LocalRegistryServer)
		// append ksRegistryServer if it`s set to different registry beside the cluster one
		if cluster.KsClusterConf.Enabled && ksRegistryServer != "" && ksRegistryServer != mainRegistry {
			registries = registries + "," + fmt.Sprintf("\"%s\"", ksRegistryServer)
		}
	}

	if cluster.Harbor != nil {
		// check harbor interfacing is set and not enable tls
		if cluster.Harbor.Enable {
			if !cluster.Harbor.EnableTls {
				// if port is 80 do not add port to docker config
				inSecAddress := fmt.Sprintf("%s:%d", cluster.Harbor.Ip, cluster.Harbor.Port)
				if cluster.Harbor.Port != 80 {
					inSecAddress = cluster.Harbor.Ip
				}
				registries += fmt.Sprintf(",\"%s\"", inSecAddress)
				// if port is 80 do not add port to docker config
				if len(strings.TrimSpace(cluster.Harbor.Host)) > 0 {
					inSecHostAddress := fmt.Sprintf("%s:%d", cluster.Harbor.Ip, cluster.Harbor.Port)
					if cluster.Harbor.Port != 80 {
						inSecHostAddress = cluster.Harbor.Ip
					}
					registries += fmt.Sprintf(",\"%s\"", inSecHostAddress)
				}

			}
		}
	}

	// combine docker root dir
	rtc := cache.GetCurrentCache()
	clientConf := rtc.GetClientRuntimeConfig(cache.NodeId)

	err := util.WriteTxtToFile("/etc/docker/daemon.json", fmt.Sprintf(template, clientConf.CRIRootDir, registries, cri.LogSize, cri.LogMaxFile))
	if err != nil {
		log.Errorf("Failed to set docker config file /etc/docker/daemon.json due to error %s: ", err.Error())
		return err
	}
	log.Debug("Done with docker config file /etc/docker/daemon.json")
	return nil
}

func installDockerOffline(cluster schema.Cluster, config client.Config, resourceServerURL string, k8sVersion string, md5Dep dep.DepMap) error {

	saveTo := config.YamlDataDir + "/docker"

	if err := family.CommonDownloadDep(resourceServerURL, dockerVersion, saveTo,
		k8sVersion, md5Dep); err != nil {
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

func kernelM3S10patch() error {
	dockerSystemdDropIn := `[Service]
# MountFlags "slave" helps to prevent "device busy" errors on RHEL/CentOS 7.3 kernels
MountFlags=slave
`
	if err := util.CreateDirIfNotExists("/etc/systemd/system/docker.service.d/"); err != nil {
		log.Errorf("Failed  to create dir /etc/systemd/system/docker.service.d/ due to error: %s", err.Error())
		return err
	}

	if err := util.WriteTxtToFile("/etc/systemd/system/docker.service.d/mountflags-slave.conf", dockerSystemdDropIn); err != nil {
		log.Errorf("Failed to write docker systemd dropIn file due to error: %s", err.Error())
		return err
	}
	var stdErr bytes.Buffer
	var err error
	_, stdErr, err = command.RunCmd("systemctl", "daemon-reload")
	if err != nil {
		log.Error("Failed to run systemctl daemon-reload due to following error:")
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
		return err
	}
	if err := family.StartSystemdService(true, true, "docker.service"); err != nil {
		return err
	}
	_, stdErr, err = command.RunCmd("rm", "-f", "/etc/systemd/system/docker.service.d/mountflags-slave.conf")
	if err != nil {
		log.Error("Failed to remove /etc/systemd/system/docker.service.d/mountflags-slave.conf due to following error:")
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
	}
	return err
}

func (C CRIDockerCentos) CleanDataDir() error {
	var stdErr bytes.Buffer
	var err error

	rtc := cache.GetCurrentCache()
	clientConfig := rtc.GetClientRuntimeConfig(cache.NodeId)
	dockerRootDir := clientConfig.CRIRootDir + "/docker"

	log.Debugf("Attempt to remove docker root data dir folder %s", dockerRootDir)
	_, stdErr, err = command.RunCmd("rm", "-rf", dockerRootDir)
	if err != nil {
		log.Errorf("(Ignored) Failed to remove docker root data dir %s due to following error:", dockerRootDir)
		errMsg := stdErr.String()
		log.Errorf("StdErr %s", errMsg)
		err = errors.New(errMsg)
	}
	return nil
}
