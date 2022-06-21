package reportor

import (
	"encoding/json"
	"fmt"
	bd "k8s-installer/pkg/block_device"
	"k8s-installer/pkg/constants"
	"strings"
	"time"

	"k8s-installer/pkg/config/client"

	"k8s-installer/node/os"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/message_queue/nats"
	"k8s-installer/pkg/network"
	"k8s-installer/schema"
)

func NodeStatReport(stopChan <-chan struct{}) {
	runtimeCache := cache.GetCurrentCache()
	clientConfig := runtimeCache.GetClientRuntimeConfig(cache.NodeId)
	ticker := time.NewTicker(time.Second * time.Duration(clientConfig.StatReportInFrequency))
	first := make(chan bool, 1)
	first <- true
	go func() {
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				ReportIn(clientConfig)
			case <-first:
				ReportIn(clientConfig)
			}
		}
	}()
}

func ReportIn(clientConfig client.Config) error {
	nodeInformation := gatherNodeInformation()
	if clientConfig.ProxyIpv4CIDR != "" {
		nodeInformation.ProxyIpv4CIDR = clientConfig.ProxyIpv4CIDR
	}
	nodeInformation.Region = &schema.Region{
		ID:   clientConfig.Region,
		Name: clientConfig.Region,
	}
	data, err := json.Marshal(&nodeInformation)
	if err != nil {
		log.Debugf("Failed to parse node information data due to error %s", err)
		log.Debugf("Failed to report stat of node %s ,aborting report", cache.NodeId)
		return err
	}
	if err := nats.SendingMessage(clientConfig.MessageQueue.ServerNodeStatusReportInSubject, string(data), cache.NodeId); err != nil {
		log.Debugf("Failed to send node stat to server due to error %s", err)
		return err
	}
	return nil
}

func gatherNodeInformation() schema.NodeInformation {
	var fatalErr error
	runtimeCache := cache.GetCurrentCache()
	clientConfig := runtimeCache.GetClientRuntimeConfig(cache.NodeId)
	nodeInformation := schema.NodeInformation{
		Status:      constants.StateReady,
		AgentStatus: constants.StateReady,
	}
	nodeInformation.ClusterInstaller = clientConfig.ClusterInstaller
	nodeInformation.Id = clientConfig.ClientId
	nodeInformation.NodeIPV4AddressList, nodeInformation.NodeIPV6AddressList, fatalErr = network.GetAllIpAddress()
	if fatalErr != nil {
		log.Fatalf("Failed to gather node information due to error %s aborting", fatalErr.Error())
	}
	defaultIp := network.GetDefaultIP(true)
	nodeInformation.Ipv4DefaultIp = defaultIp.String()
	if defaultGw, _, err := network.GetDefaultGateway(true); err != nil {
		nodeInformation.Status = constants.StatusError
	} else {
		nodeInformation.Ipv4DefaultGw = defaultGw
	}
	if osInfo, err := os.GetAllSystemInformation(); err != nil {
		nodeInformation.Status = constants.StatusError
		log.Debugf("Failed to get node operation system information due to error %s", err)
	} else {
		nodeInformation.SystemInfo = *osInfo
	}
	_, nic, err := network.GetDefaultGateway(true)
	if err == nil {
		nodeInformation.DefaultMetworkInterface = nic.Name
	}

	if !clientConfig.IsTestNode {
		if !CheckBlockDeviceIsReadyForProduction(&nodeInformation, clientConfig) {
			nodeInformation.Status = constants.StatusError
		}
	}

	return nodeInformation
}

// here we need ensure all block device is proper mount to following dir
// a. /var/lib/kubelet/
// b. /var/lib/etcd
// c. CRIMountDev is ready to be mount
func CheckBlockDeviceIsReadyForProduction(information *schema.NodeInformation, clientConfig client.Config) bool {
	var errorList []string
	if result, err := bd.DeviceIsReady(clientConfig.CRIMountDev); err != nil {
		errorList = append(errorList, fmt.Sprintf("Failed to check cri device %s due to error: %s", clientConfig.CRIMountDev, err.Error()))
	} else if !result {
		errorList = append(errorList, "Cri device %s already been mount. You can ignore it if the node is member of a k8s cluster.", clientConfig.CRIMountDev)
	} else {
		deviceNameWithoutPath := clientConfig.CRIMountDev[strings.LastIndex(clientConfig.CRIMountDev, "/")+1:]
		found := false
		for _, dev := range information.SystemInfo.Storage {
			if deviceNameWithoutPath == dev.Name {
				found = true
				if dev.Size < clientConfig.CRIDevMinDiskSize {
					errorList = append(errorList, fmt.Sprintf("Cri device %s request size %dG but got %d", clientConfig.CRIMountDev, clientConfig.CRIDevMinDiskSize+1, dev.Size))
				}
				break
			}
		}
		if !found {
			errorList = append(errorList, fmt.Sprintf("Unable to find device %s", clientConfig.CRIMountDev))
		}
	}

	// check /var/lib/kubelet/
	if result, err := bd.CheckDirWithMountSize("/var/lib/kubelet/", clientConfig.KubeletDevMinDiskSize); err != nil {
		errorList = append(errorList, fmt.Sprintf("Failed to check dir /var/lib/kubelet/ mount point. Does it proper mounted and has %dG disk space available?", clientConfig.KubeletDevMinDiskSize+1))
	} else if !result {
		errorList = append(errorList, fmt.Sprintf("Failed to check dir /var/lib/kubelet/ mount point. Does it proper mounted and has %dG disk space available?", clientConfig.KubeletDevMinDiskSize+1))
	}

	// check /var/lib/etcd
	if result, err := bd.CheckDirWithMountSize("/var/lib/etcd", clientConfig.EtcdDevMinDiskSize); err != nil {
		errorList = append(errorList, fmt.Sprintf("Failed to check dir /var/lib/etcd mount point. Does it proper mounted and has %dG disk space available?", clientConfig.EtcdDevMinDiskSize+1))
	} else if !result {
		errorList = append(errorList, fmt.Sprintf("Failed to check dir /var/lib/etcd mount point. Does it proper mounted and has %dG disk space available?", clientConfig.EtcdDevMinDiskSize+1))
	}

	if len(errorList) == 0 {
		return true
	} else {
		for _, msg := range errorList {
			log.Error(msg)
		}
		information.IssueList = errorList
		return false
	}
}
