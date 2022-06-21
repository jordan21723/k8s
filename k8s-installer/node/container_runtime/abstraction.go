package container_runtime

import (
	"k8s-installer/pkg/config/client"
	"k8s-installer/pkg/dep"
	"k8s-installer/schema"
)

type INodeContainerRuntime interface {
	Install(offline bool, taskCRI schema.TaskCRI, cluster schema.Cluster, config client.Config, resourceServerURL string, k8sVersion string, md5Dep dep.DepMap) error
	Remove(config schema.TaskCRI) error
	CleanDataDir() error
}
