package version

import (
	"k8s-installer/node/k8s"
	"k8s-installer/pkg/config/client"
	depMap "k8s-installer/pkg/dep"
	"k8s-installer/schema"

	natsLib "github.com/nats-io/nats.go"
)

type V1804 struct {
	client.Config
}

func (v V1804) TaskSetHost(hosts schema.TaskSetHosts) (map[string]string, error) {
	return nil, nil
}

func (v V1804) BasicNodeSetup(basicSetup schema.TaskBasicConfig, clusterInformation schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) basicNodeSetupDestroy(basicSetup schema.TaskBasicConfig, clusterInformation schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) basicNodeSetup(basicSetup schema.TaskBasicConfig, clusterInformation schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) PurgeAll(cluster schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) InstallOrRemoveContainerRuntime(taskCri schema.TaskCRI, clusterInformation schema.Cluster, resourceServerURL string, md5Dep depMap.DepMap) (map[string]string, error) {
	return nil, nil
}

func (v V1804) installContainerRuntime(taskCri schema.TaskCRI, cluster schema.Cluster, resourceServerURL, k8sVersion string) (map[string]string, error) {
	return nil, nil
}

func (v V1804) removeContainerRuntime(taskCri schema.TaskCRI) (map[string]string, error) {
	return nil, nil
}

func (v V1804) InstallOrRemoveLoadBalance(loadBalance schema.TaskLoadBalance, clusterInformation schema.Cluster, resourceServerURL string, md5Dep depMap.DepMap) (map[string]string, error) {
	return nil, nil
}

func (v V1804) installLoadBalance(loadBalance schema.TaskLoadBalance, clusterInformation schema.Cluster, resourceServerURL string) (map[string]string, error) {
	return nil, nil
}

func (v V1804) InitOrDestroyFirstControlPlane(kubeadm schema.TaskKubeadm, resourceServerURL string, clusterInformation schema.Cluster, md5Dep depMap.DepMap) (map[string]string, error) {
	return nil, nil
}

func (v V1804) destroyFirstControlPlane(kubeadm schema.TaskKubeadm, resourceServerURL string, cluster schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) initFirstControlPlane(kubeadm schema.TaskKubeadm, resourceServerURL string, cluster schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func offlineInstallK8sYumDep(cluster schema.Cluster, resourceServerURL string) error {
	return nil
}

func (v V1804) JoinOrDestroyControlPlane(kubeadm schema.TaskKubeadm, resourceServerURL string, preReturnData map[string]string, clusterInformation schema.Cluster, md5Dep depMap.DepMap) (map[string]string, error) {
	return nil, nil
}

func (v V1804) destroyControlPlane(kubeadm schema.TaskKubeadm, cluster schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) joinControlPlane(preReturnData map[string]string, cluster schema.Cluster, resourceServerURL string) (map[string]string, error) {
	return nil, nil
}

func (v V1804) JoinOrDestroyWorkNode(kubeadm schema.TaskKubeadm, resourceServerURL string, preReturnData map[string]string, clusterInformation schema.Cluster, md5Dep depMap.DepMap) (map[string]string, error) {
	return nil, nil
}

func (v V1804) joinWorkNode(kubeadm schema.TaskKubeadm, resourceServerURL string, preReturnData map[string]string, cluster schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) destroyWorkNode(kubeadm schema.TaskKubeadm, cluster schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) RenameHostname(renameHostnameTask schema.TaskRenameHostName, clusterInformation schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) RunKubectl(kubectl schema.TaskKubectl, preReturnData map[string]string, clusterInformation schema.Cluster) (map[string]string, error) {
	return k8s.KubectlExecutor(kubectl, v.Config)
}

func (v V1804) RunCommand(commandToRun schema.TaskRunCommand, preReturnData map[string]string, clusterInformation schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) AsyncRunCommand(operationId string, nodeId string, nodeStepId string, msg *natsLib.Msg, commandToRun schema.TaskRunCommand, preReturnData map[string]string, clusterInformation schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) CopyTextFile(fileToCopy schema.TaskCopyTextBaseFile, preReturnData map[string]string, clusterInformation schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) InstallOrDestroyVirtualKubelet(vk schema.TaskVirtualKubelet, resourceServerURL string, cluster schema.Cluster, md5Dep depMap.DepMap) (map[string]string, error) {
	return nil, nil
}

func (v V1804) DestroyVirtualKubelet(vk schema.TaskVirtualKubelet, resourceServerURL string, cluster schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) InstallVirtualKubelet(vk schema.TaskVirtualKubelet, resourceServerURL string, cluster schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) PrintJoinString(printJoinString schema.TaskPrintJoin) (map[string]string, error) {
	return nil, nil
}

func (v V1804) CommonLink(from depMap.DepMap, saveTo string, linkTo string) (map[string]string, error) {
	return nil, nil
}

func (v V1804) destroyK8sNode(kubeadm schema.TaskKubeadm, cluster schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) GoCurl(task schema.TaskCurl) (map[string]string, error) {
	return nil, nil
}

func (v V1804) CommonDownloadDep(operationId string, nodeId string, nodeStepId string, msg *natsLib.Msg, resourceServerURL string, dep depMap.DepMap, saveTo string, k8sVersion string, md5Dep depMap.DepMap) (map[string]string, error) {
	return nil, nil
}

func (v V1804) CommonDownload(operationId string, nodeId string, nodeStepId string, msg *natsLib.Msg, resourceServerURL string, FromDir string, k8sVersion string, fileList []string, saveTo string, isUseDefaultPath bool) (map[string]string, error) {
	return nil, nil
}

func (v V1804) ConfigPromtail(clusterInformation schema.Cluster) (map[string]string, error) {
	return nil, nil
}

func (v V1804) PreLoadImage(preLoad schema.TaskPreLoadImage, containerRuntimeType string) (map[string]string, error) {
	return nil, nil
}

func (v V1804) GenerateKSClusterConfig(cluster schema.Cluster, ipAddress string) (map[string]string, error) {
	return nil, nil
}
