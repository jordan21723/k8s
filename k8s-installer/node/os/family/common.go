package family

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strings"

	blockDevice "k8s-installer/pkg/block_device"
	depMap "k8s-installer/pkg/dep"

	osInfoProvider "k8s-installer/node/os"
	"k8s-installer/pkg/command"
	config "k8s-installer/pkg/config/downloader"
	"k8s-installer/pkg/downloader"
	backdown "k8s-installer/pkg/downloader/back_downloader"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util"
	"k8s-installer/pkg/util/fileutils"
	"k8s-installer/schema"

	natsLib "github.com/nats-io/nats.go"
)

func StartSystemdService(enabled, restart bool, unit string) error {
	var stdErr bytes.Buffer
	var err error
	if enabled {
		err = EnableSystemdService(unit)
		if err != nil {
			return err
		}
	}

	startOrRestart := "start"
	if restart {
		startOrRestart = "restart"
	}

	log.Debugf("Try to %s service %s", startOrRestart, unit)
	_, stdErr, err = command.RunCmd("systemctl", startOrRestart, unit)
	if err != nil {
		// systemd unit name does not exists or some error occurred
		log.Errorf("Failed to %v service due to error %v", startOrRestart, stdErr.String())
	}
	return err
}

func StopSystemdService(unit string) error {
	var stdErr bytes.Buffer
	var err error
	err = DisableSystemdService(unit)
	if err != nil {
		return err
	}

	log.Debugf("Try to stop service %s", unit)
	_, stdErr, err = command.RunCmd("systemctl", "stop", unit)
	if err != nil {
		// systemd unit name does not exists or some error occurred
		log.Errorf("Failed to stop service due to error %v", stdErr.String())
	}
	return err
}

func EnableSystemdService(unit string) error {
	var stdErr bytes.Buffer
	var err error
	log.Debugf("Enable service %s", unit)
	_, stdErr, err = command.RunCmd("systemctl", "enable", unit)
	if err != nil {
		// systemd unit name does not exists or some error occurred
		log.Errorf("Failed to enable service due to error %v", stdErr.String())
		return err
	}
	return nil
}

func DisableSystemdService(unit string) error {
	var stdErr bytes.Buffer
	var err error
	log.Debugf("Disable service %s", unit)
	_, stdErr, err = command.RunCmd("systemctl", "disable", unit)
	if err != nil {
		// systemd unit name does not exists or some error occurred
		log.Errorf("Failed to disable service due to error %s", stdErr.String())
		return err
	}
	return nil
}

func CheckSystemdService(unit string) (bool, error) {
	var stdOut, stdErr bytes.Buffer
	var err error
	log.Debugf("Checking service %s...", unit)
	stdOut, stdErr, err = command.RunCmd("systemctl", "check", unit)
	if err != nil {
		// systemd unit name does not exists or some error occurred
		log.Errorf("Failed to check service error %s", stdErr.String())
		return false, err
	}
	return stdOut.String() == "active\n", nil
}

func CheckSystemdServiceExists(unit string) (bool, error) {
	var stdErr bytes.Buffer
	var err error
	log.Debugf("Checking service %s exists", unit)
	_, stdErr, err = command.RunCmd("systemctl", "status", unit)
	if err != nil {
		// systemd unit name does not exists or some error occurred
		log.Errorf("Failed to check service error %s", stdErr.String())
		return false, err
	}
	return true, nil
}

func DisableFirewalld() error {
	var stdErr bytes.Buffer
	var err error
	_, stdErr, err = command.RunCmd("systemctl", "disable", "firewalld")
	if err != nil {
		log.Error("Failed to disable firewalld due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	log.Debug("Successfully disable firewalld")
	_, stdErr, err = command.RunCmd("systemctl", "stop", "firewalld")
	if err != nil {
		log.Error("Failed to stop firewalld due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	log.Debug("Successfully disable firewalld")
	return nil
}

func GetEnforce() (bool, error) {
	var stdOut, stdErr bytes.Buffer
	var err error
	stdOut, stdErr, err = command.RunCmd("getenforce")
	if err != nil {
		log.Error("Failed to run command getenforce due to error as following")
		log.Errorf("Error message %s", err)
		log.Errorf("StdErr %s", stdErr.String())
		return false, err
	}
	log.Debug(stdOut.String())
	return stdOut.String() == "Enforcing\n" || stdOut.String() == "Permissive\n", nil
}

func EnableKernelOptions() error {
	var _, stdErr bytes.Buffer
	var err error
	log.Debug("Try to enable kernel modular br_netfilter")
	_, stdErr, err = command.RunCmd("modprobe", "br_netfilter")
	if err != nil {
		log.Errorf("Failed to load kernel modular br_netfilter due to error %s", stdErr.String())
		return err
	}
	log.Debug("Try to enable kernel modular nf_conntrack")
	_, stdErr, err = command.RunCmd("modprobe", "nf_conntrack")
	if err != nil {
		log.Errorf("Failed to load kernel modular nf_conntrack due to error %s", stdErr.String())
		return err
	}
	return nil
}

func EnableIPV46Forwarding() error {
	content := `
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward=1
`
	log.Debugf("Configuring /etc/sysctl.d/k8s.conf")
	err := util.WriteTxtToFile("/etc/sysctl.d/k8s.conf", content)
	if err != nil {
		log.Errorf("Failed to Configuring /etc/sysctl.d/k8s.conf due to error %s", err)
		return err
	}
	content = `net.ipv6.conf.all.forwarding=1`
	log.Debugf("Configuring /etc/sysctl.conf")
	err = util.WriteTxtToFile("/etc/sysctl.conf", content)
	if err != nil {
		log.Errorf("Failed to Configuring /etc/sysctl.conf due to error %s", err)
		return err
	}
	var _, stdErr bytes.Buffer
	log.Debug("Try to apply the setting with sysctl")
	_, stdErr, err = command.RunCmd("sysctl", "--system")
	if err != nil {
		log.Errorf("Failed to apply sysctl setting due to error %s", stdErr.String())
		return err
	}
	return nil
}

func DisableSwapForever() error {
	var _, stdErr bytes.Buffer
	var err error
	log.Debug("Try to turn off swap")
	_, stdErr, err = command.RunCmd("swapoff", "-a")
	if err != nil {
		log.Errorf("Failed to turn off swap due to error %s", stdErr.String())
		return err
	}
	log.Debugf("Try to disable swap in /etc/fstab")
	_, stdErr, err = command.RunCmd("sed", "-i", "/swap/d", "/etc/fstab")
	if err != nil {
		log.Errorf("Failed to disable swap in /etc/fstab due to error %s", stdErr.String())
		return err
	}
	return nil
}

func IncreaseMaximumNumberOfFileDescriptors() error {
	cf := `
#IncreaseMaximumNumberOfFileDescriptors
fs.file-max = 100000
vm.max_map_count=262144
#IncreaseMaximumNumberOfFileDescriptors
`
	if err := fileutils.AppendFile("/etc/sysctl.conf", cf, "#IncreaseMaximumNumberOfFileDescriptors"); err != nil {
		return err
	}

	cf = `
#IncreaseMaximumNumberOfFileDescriptors
* soft nproc 65535
* hard nproc 65535
* soft nofile 65535
* hard nofile 65535
#IncreaseMaximumNumberOfFileDescriptors
`
	if err := fileutils.AppendFile("/etc/security/limits.conf", cf, "#IncreaseMaximumNumberOfFileDescriptors"); err != nil {
		return err
	}

	_, stdErr, err := command.RunCmd("sysctl", "-p")
	if err != nil {
		log.Errorf("Failed to IncreaseMaximumNumberOfFileDescriptors due to error %s", stdErr.String())
		return err
	}
	return nil
}

// var dockerVersion = map[string]map[string]map[string]map[string]string {
// 	"centos": {
// 		constants.V1_18_6:{
// 			"x86_64":{
// 				"audit-libs-python": "audit-libs-python-2.8.5-4.el7.x86_64.rpm",
// 			},
// 			"aarch64": {
// 				"audit-libs-python": "audit-libs-python-2.8.5-4.el7.aarch64.rpm",
// 			},
// 		},
// 	},
// 	"ubuntu": {
// 	},
// }

// Download from /resourceServer/{k8sVersion}/{osInfo.OS.Vendor}/{osInfo.OS.Version}/{osInfo.Kernel.Architecture}/"package"/{file}
// file = dep[osInfo.OS.Vendor][k8sVersion][osInfo.Kernel.Architecture]
// Save to: {saveTo}/{file}
func CommonDownloadDep(resourceServerURL string, dep depMap.DepMap, saveTo string, k8sVersion string, md5Dep depMap.DepMap) error {

	osInfo, errOSInfo := osInfoProvider.GetAllSystemInformation()
	if errOSInfo != nil {
		log.Errorf("Failed to get node cpu arch due to error %s", errOSInfo.Error())
		return errOSInfo
	}

	_m, errMarshal := json.Marshal(dep)
	if errMarshal != nil {
		log.Errorf("Failed to marshal download source schema. Download abort!!!")
		return errMarshal
	}
	var m depMap.DepMap
	if errUnmarshal := json.Unmarshal(_m, &m); errUnmarshal != nil {
		log.Errorf("Failed to unmarshal object before downloading starts!!")
		return errUnmarshal
	}

	var fileList []string
	var md5List []string
	log.Debugf("Try to range %v[%v][%v][%v]", reflect.TypeOf(m).String(), osInfo.OS.Vendor, k8sVersion, osInfo.Kernel.Architecture)
	if _, ok := m[osInfo.OS.Vendor][k8sVersion][osInfo.Kernel.Architecture]; !ok {
		return fmt.Errorf("The %v do not support %v[%v][%v][%v]", reflect.TypeOf(m).String(), reflect.TypeOf(m).String(), osInfo.OS.Vendor, k8sVersion, osInfo.Kernel.Architecture)
	}

	for k, v := range m[osInfo.OS.Vendor][k8sVersion][osInfo.Kernel.Architecture] {
		fileList = append(fileList, v)
		if md5Dep != nil {
			md5List = append(md5List, md5Dep[osInfo.OS.Vendor][k8sVersion][osInfo.Kernel.Architecture][k])
		}
		// log.Debugf("md5Dep:%v md5List:%v", md5Dep, md5List)
	}

	errCommonDownload := CommonDownload(resourceServerURL, "", k8sVersion, fileList, saveTo, true, md5List)
	if errCommonDownload != nil {
		return errCommonDownload
	}
	return nil
}

func CommonDownload(resourceServerURL string, FromDir string, k8sVersion string, fileList []string, saveTo string, isUseDefaultPath bool, md5List []string) error {

	osInfo, errOSInfo := osInfoProvider.GetAllSystemInformation()
	if errOSInfo != nil {
		log.Errorf("Failed to get node cpu arch due to error %s", errOSInfo.Error())
		return errOSInfo
	}

	if errCreateDirIfNotExists := util.CreateDirIfNotExists(saveTo); errCreateDirIfNotExists != nil {
		log.Errorf("Failed to create dir %s due to error: %v", saveTo, errCreateDirIfNotExists.Error())
		return errCreateDirIfNotExists
	}

	DefaultDownloadDir := ""
	if isUseDefaultPath {
		DefaultDownloadDir = path.Join(k8sVersion, osInfo.OS.Vendor, osInfo.OS.Version, osInfo.Kernel.Architecture, "package")
	}

	for k, f := range fileList {
		cfg := config.NewDownloadConfig()
		cfg.URL = resourceServerURL + "/" + path.Join(DefaultDownloadDir, f)
		cfg.RV.RealTarget = path.Join(saveTo, strings.Replace(f, FromDir, "", 1))
		if md5List != nil {
			cfg.Md5 = md5List[k]
		}
		saveToDir, _ := path.Split(cfg.RV.RealTarget)
		if errCreateDirIfNotExists := util.CreateDirIfNotExists(saveToDir); errCreateDirIfNotExists != nil {
			log.Errorf("Failed to create dir %s due to error %v", saveToDir, errCreateDirIfNotExists.Error())
			return errCreateDirIfNotExists
		}
		timeout := downloader.CalculateTimeout(cfg)
		getter := backdown.NewBackDownloader(cfg)

		// log.Debugf("Try to download file from %s to %s", cfg.URL, cfg.RV.RealTarget)
		if errDoDownloadTimeout := downloader.DoDownloadTimeout(getter, timeout); errDoDownloadTimeout != nil {
			log.Errorf("Failed to down file %s to %s due to error: %s", cfg.URL, cfg.RV.RealTarget, errDoDownloadTimeout.Error())
			return errDoDownloadTimeout
		}
	}
	return nil
}

func ReplyMsg(reply schema.QueueReply, msg *natsLib.Msg) error {
	if data, err := json.Marshal(&reply); err != nil {
		log.Debugf("Failed to Marshal reply QueueReply due to error %s", err.Error())
		log.Debug("Failed to reply message to notify server due to json Marshal error")
		return err
	} else {
		if err := msg.Respond(data); err != nil {
			log.Debug("Failed to reply message to notify server due to json error %", err.Error())
			return err
		}
	}
	return nil
}

func CreateReplyBody(operationId, nodeId, stat, message, nodeStepId string, returnData map[string]string) schema.QueueReply {
	return schema.QueueReply{
		OperationId: operationId,
		Stat:        stat,
		Message:     message,
		NodeId:      nodeId,
		NodeStepId:  nodeStepId,
		ReturnData:  returnData,
	}
}

func MountOrUmount(dev string, targetDir string, fsType string, enableBootCheck bool, umount bool, force bool, backupFilePath string) error {
	log.Debugf("Attempt to create cri data dir %s", targetDir)
	if err := util.CreateDirIfNotExists(targetDir); err != nil {
		log.Errorf("Failed to create cri data dir %s due to error: %s", targetDir, err)
		return err
	}
	log.Debugf("Attempt to remake fs ensure to clean all data on the block device...")
	if err := blockDevice.MakeFS(dev, force, fsType); err != nil {
		log.Errorf("Failed to mkfs.%s of device %s due to error: %s", fsType, dev, err)
		return err
	}
	log.Debugf("Attempt to mount device %s to target dir %s", dev, targetDir)
	if err := blockDevice.Mount(dev, targetDir, blockDevice.XFS, enableBootCheck, umount, force, backupFilePath); err != nil {
		log.Errorf("Failed to mount device %s to target dir for cri docker %s due to error: %s", dev, "/var/lib/docker", err)
		return err
	}
	return nil
}
