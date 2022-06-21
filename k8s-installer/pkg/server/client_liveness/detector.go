package client_liveness

import (
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/config/server"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/network"
	"time"
)

func ClientAgentDaemon(stopChan <-chan struct{}) {
	runtimeCache := cache.GetCurrentCache()
	serverConfig := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
	ticker := time.NewTicker(time.Second * time.Duration(serverConfig.AgentLivenessDetectorFrequency))
	first := make(chan bool, 1)
	first <- true
	go func() {
		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				Detector(serverConfig, runtimeCache)
			case <-first:
				Detector(serverConfig, runtimeCache)
			}
		}
	}()
}

func Detector(config server.Config, runtimeCache cache.ICache) {
	nodes, err := runtimeCache.GetNodeInformationCollection()
	if err != nil {
		log.Errorf("Detect Agent: Failed to get nodes collection due to error: %s", err.Error())
		return
	}
	for _, node := range nodes {
		if err := network.CheckIpIsReachable(node.Ipv4DefaultIp, config.SignalPort, "tcp", 2*time.Second); err != nil {
			log.Warnf("Detect Agent: agent is dead of node: %s", node.Id)
			node.AgentStatus = "unReachable"
			if err := runtimeCache.SaveOrUpdateNodeInformation(node.Id, node); err != nil {
				log.Errorf("Detect Agent: Failed to update agent status  of node: %s  due to error: %s", node.Id, err.Error())
				return
			}
		} else {
			log.Debugf("Detect Agent: agent is alive of node: %s", node.Id)
		}
	}
}
