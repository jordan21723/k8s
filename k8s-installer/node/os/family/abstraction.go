package family

import (
	"k8s-installer/pkg/dep"
	"k8s-installer/schema"

	natsLib "github.com/nats-io/nats.go"
)

type NodeClient struct {
	os IOSFamily
}

type IOSFamily interface {
	GetOSVersion() IOSVersion
	GetOSFamily() string
	GetOSFullName() string
}

type IOSVersion interface {
	BasicNodeSetup(basicSetup schema.TaskBasicConfig, clusterInformation schema.Cluster) (map[string]string, error)
	RenameHostname(renameHostnameTask schema.TaskRenameHostName, clusterInformation schema.Cluster) (map[string]string, error)
	PurgeAll(clusterInformation schema.Cluster) (map[string]string, error)
	InstallOrRemoveContainerRuntime(cri schema.TaskCRI, clusterInformation schema.Cluster, resourceServerURL string, md5Dep dep.DepMap) (map[string]string, error)
	InstallOrRemoveLoadBalance(vip schema.TaskLoadBalance, clusterInformation schema.Cluster, resourceServerURL string, md5Dep dep.DepMap) (map[string]string, error)
	InitOrDestroyFirstControlPlane(kubeadm schema.TaskKubeadm, resourceServerURL string, clusterInformation schema.Cluster, md5Dep dep.DepMap) (map[string]string, error)
	JoinOrDestroyControlPlane(kubeadm schema.TaskKubeadm, resourceServerURL string, preReturnData map[string]string, clusterInformation schema.Cluster, md5 dep.DepMap) (map[string]string, error)
	JoinOrDestroyWorkNode(kubeadm schema.TaskKubeadm, resourceServerURL string, preReturnData map[string]string, clusterInformation schema.Cluster, md5 dep.DepMap) (map[string]string, error)
	RunKubectl(kubectl schema.TaskKubectl, preReturnData map[string]string, clusterInformation schema.Cluster) (map[string]string, error)
	RunCommand(commandToRun schema.TaskRunCommand, preReturnData map[string]string, clusterInformation schema.Cluster) (map[string]string, error)
	AsyncRunCommand(operationId string, nodeId string, nodeStepId string, msg *natsLib.Msg, commandToRun schema.TaskRunCommand, preReturnData map[string]string, clusterInformation schema.Cluster) (map[string]string, error)
	CopyTextFile(fileToCopy schema.TaskCopyTextBaseFile, preReturnData map[string]string, clusterInformation schema.Cluster) (map[string]string, error)
	InstallOrDestroyVirtualKubelet(vk schema.TaskVirtualKubelet, resourceServerURL string, clusterInformation schema.Cluster, md5 dep.DepMap) (map[string]string, error)
	PrintJoinString(printJoinString schema.TaskPrintJoin) (map[string]string, error)
	CommonLink(from dep.DepMap, saveTo string, linkTo string) (map[string]string, error)
	GoCurl(task schema.TaskCurl) (map[string]string, error)
	CommonDownloadDep(operationId string, nodeId string, nodeStepId string, msg *natsLib.Msg, resourceServerURL string, dep dep.DepMap, saveTo string, k8sVersion string, md5 dep.DepMap) (map[string]string, error)
	CommonDownload(operationId string, nodeId string, nodeStepId string, msg *natsLib.Msg, resourceServerURL string, FromDir string, k8sVersion string, fileList []string, saveTo string, isUseDefaultPath bool) (map[string]string, error)
	GenerateKSClusterConfig(cluster schema.Cluster, ipAddress string) (map[string]string, error)
	ConfigPromtail(clusterInformatInstallOrDestroyVirtualKubeletion schema.Cluster) (map[string]string, error)
	// to ask client load container image
	// if imageDirPath is not set, then client will use client config LocalImagePath as the default location
	PreLoadImage(preLoad schema.TaskPreLoadImage, containerRuntimeType string) (map[string]string, error)
	TaskSetHost(hosts schema.TaskSetHosts) (map[string]string, error)
}
