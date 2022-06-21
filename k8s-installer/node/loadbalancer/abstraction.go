package loadbalancer

import (
	"k8s-installer/pkg/config/client"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/schema"
)

type IWebServer interface {
	Install(offline bool, task schema.TaskLoadBalance, cluster schema.Cluster, config client.Config, resourceServerURL, osVersion, cpuArch string, md5Dep dep.DepMap) error
	Remove() error
	CreateConfig(sections []schema.ProxySection) string
	CreateAPIServerConfig(localIp, port, balanceAlgorithm string, servers map[string]schema.NodeInformation) string
	GetSystemdServiceName() string
}

func CreateProxy(proxyType string) IWebServer {
	switch proxyType {
	case constants.ProxyTypeHaproxy:
		return Haproxy{}
	default:
		return Haproxy{}
	}
}
