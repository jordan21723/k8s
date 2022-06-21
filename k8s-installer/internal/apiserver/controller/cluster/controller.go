package cluster

import (
	"encoding/json"
	"errors"
	"fmt"
	restarter "k8s-installer/components/auto_restarter"
	"k8s-installer/components/velero"
	"k8s-installer/pkg/backup_restore"
	"net/http"
	"strings"
	"time"

	"k8s-installer/components/minio"
	"k8s-installer/internal/apiserver/controller/common"

	"k8s-installer/pkg/log"
	"k8s-installer/pkg/message_queue/nats"
	"k8s-installer/pkg/network"
	"k8s-installer/pkg/restplus"
	mqHelper "k8s-installer/pkg/server/message_queue"

	"github.com/emicklei/go-restful"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"k8s-installer/internal/apiserver/utils"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/control_manager"
	"k8s-installer/schema"
)

func ListKsHostCluster(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	clusterMap, err := runtimeCache.GetClusterCollection()
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to get ks cluster list with err %s", err))
		return
	}
	var clusters []schema.Cluster
	for _, cluster := range clusterMap {
		if cluster.Status == constants.ClusterStatusRunning {
			if cluster.KsClusterConf != nil && cluster.KsClusterConf.MultiClusterConfig != nil && cluster.KsClusterConf.MultiClusterConfig.ClusterRole == constants.ClusterRoleHost {
				clusters = append(clusters, cluster)
			}
		}

	}
	response.WriteAsJson(clusters)
}

func UpdateClusterStatus(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	clusterPost := struct {
		Status string `json:"status"`
	}{}

	if err := request.ReadEntity(&clusterPost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	runtimeCache := cache.GetCurrentCache()
	// if cluster id is given
	// check id is unique
	existingCluster, err := runtimeCache.GetCluster(clusterId)

	// recreate cluster expected cluster already exist
	if err != nil || existingCluster == nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to find cluster with id %s", clusterId))
		return
	}

	existingCluster.Status = clusterPost.Status

	if err := runtimeCache.SaveOrUpdateClusterCollection(existingCluster.ClusterId, *existingCluster); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to update cluster with id %s due to error: %s ", clusterId, err.Error()))
		return
	}

	response.WriteAsJson(&clusterPost)
}

func UpdateClusterInstaller(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	clusterPost := struct {
		ClusterInstaller string `json:"cluster_installer"`
	}{}

	if err := request.ReadEntity(&clusterPost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	runtimeCache := cache.GetCurrentCache()
	// if cluster id is given
	// check id is unique
	existingCluster, err := runtimeCache.GetCluster(clusterId)

	// recreate cluster expected cluster already exist
	if err != nil || existingCluster == nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to find cluster with id %s", clusterId))
		return
	}

	existingCluster.ClusterInstaller = clusterPost.ClusterInstaller
	if clusterPost.ClusterInstaller == constants.ClusterInstallerRancher {
		existingCluster.Mock = true
	}

	if err := runtimeCache.SaveOrUpdateClusterCollection(existingCluster.ClusterId, *existingCluster); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to update cluster with id %s due to error: %s ", clusterId, err.Error()))
		return
	}

	response.WriteAsJson(&clusterPost)
}

func ReCreateCluster(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	// if cluster id is given
	// check id is unique
	existingCluster, err := runtimeCache.GetCluster(clusterId)
	serverConfig := runtimeCache.GetServerRuntimeConfig(cache.NodeId)

	// recreate cluster expected cluster already exist
	if err != nil || existingCluster == nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to find cluster with id %s", clusterId))
		return
	}

	if existingCluster.ClusterInstaller == constants.ClusterInstallerRancher {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to re-create rancher cluster with id %s", clusterId))
		return
	}

	if existingCluster.Status != constants.ClusterStatusDestroyed {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to re-create cluster which is not currently in status %s", constants.ClusterStatusDestroyed))
		return
	}

	existingCluster.Action = constants.ActionCreate
	existingCluster.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")
	existingCluster.MinIO = minio.NewDeployMinIO(fmt.Sprintf("%s:%d",
		existingCluster.ContainerRuntime.PrivateRegistryAddress,
		existingCluster.ContainerRuntime.PrivateRegistryPort))

	existingCluster.Velero = velero.NewDeployVelero(fmt.Sprintf("%s:%d",
		existingCluster.ContainerRuntime.PrivateRegistryAddress,
		existingCluster.ContainerRuntime.PrivateRegistryPort))

	existingCluster.MinIO.CompleteMinIODeploy(existingCluster.ClusterId, serverConfig.BackupPath)
	existingCluster.Velero.CompleteVeleroDeploy()

	// before setup check
	if canProcess, errList, _, err := control_manager.PreCheck(*existingCluster); err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.WriteAsJson(schema.HttpErrorResult{
			ErrorCode: -1,
			Message:   err.Error(),
		})
		return
	} else if !canProcess {
		response.WriteHeader(http.StatusBadRequest)
		response.WriteAsJson(schema.HttpErrorResult{
			ErrorCode: 1,
			Message:   "PreCheckFailed",
			ErrorList: errList,
		})
		return
	}

	common.SetDependencies(existingCluster)
	existingCluster.Status = constants.StatusInstalling

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	if err := control_manager.SetupOrDestroyCluster(existingCluster, runByUser); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.WriteAsJson(schema.HttpErrorResult{
			ErrorCode: 2,
			Message:   err.Error(),
			ErrorList: []string{},
		})
	}
	response.WriteAsJson(&existingCluster)
}

func CreateCluster(request *restful.Request, response *restful.Response) {
	clusterPost := schema.Cluster{}
	if err := request.ReadEntity(&clusterPost); err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   err.Error(),
		})
		return
	}

	err := utils.Validate(clusterPost)
	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			if e.Tag() == "storage_size" {
				err = fmt.Errorf("storage size should greater than 5 and less than 50")
			}
		}
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	runtimeCache := cache.GetCurrentCache()
	if clusterPost.ClusterId == "" {
		// if cluster id is not set
		// generator one
		clusterPost.ClusterId = "cluster-" + uuid.New().String()
	} else {
		// if cluster id is given
		// check id is unique
		existingCluster, err := runtimeCache.GetCluster(clusterPost.ClusterId)

		if err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
			return
		}

		if existingCluster != nil {
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cluster ID: %s already been taken.Abort... !!!", existingCluster.ClusterId))
			return
		}
	}

	if clusterPost.CNI.CNIType == constants.CNITypeCalico {
		err = utils.CalicoValidate(&clusterPost)
		if err != nil {
			utils.ResponseError(response, http.StatusBadRequest, err.Error())
			return
		}
	}

	if clusterPost.ClusterInstaller != constants.ClusterInstallerRancher {
		clusterPost.ClusterInstaller = constants.ClusterInstallerKubeadm
	} else {
		if clusterPost.Rancher.RancherAddr == "" {
			utils.ResponseError(response, http.StatusBadRequest, errors.New("template:rancher addr is required").Error())
			return
		}
		if clusterPost.Rancher.RancherToken == "" {
			utils.ResponseError(response, http.StatusBadRequest, errors.New("template: rancher api token is required").Error())
			return
		}
		if clusterPost.Rancher.ManagedByClusterName == "" {
			utils.ResponseError(response, http.StatusBadRequest, errors.New("rancher Managed by cluster name is required").Error())
			return
		}
	}

	if clusterPost.KsClusterConf != nil {
		clusterPost.KsClusterConf = clusterPost.KsClusterConf.CompleteDeploy()
		if clusterPost.KsClusterConf.Enabled {
			if (clusterPost.KsClusterConf.MultiClusterConfig.ClusterRole == constants.ClusterRoleMember) && clusterPost.KsClusterConf.IsManagedByKubefed {
				clusterPost.KsInstaller = *clusterPost.KsInstaller.CompleteDeploy()
				runtimeCache := cache.GetCurrentCache()
				clusterCollection, err := runtimeCache.GetClusterCollection()
				if err != nil {
					log.Error(err)
				}
				IsHostClusterExist := false
				IsHostClusterInstalling := false
				for _, c := range clusterCollection {
					if c.KsClusterConf == nil {
						continue
					}
					if !c.KsClusterConf.Enabled || c.KsClusterConf.MultiClusterConfig.ClusterRole == constants.ClusterRoleMember {
						continue
					}

					if c.KsClusterConf.MultiClusterConfig.ClusterRole == constants.ClusterRoleHost {
						if c.Status == constants.ClusterStatusRunning {
							log.Debugf("Check Host Cluster Exist, %#v", c.KsClusterConf)
							IsHostClusterExist = true
							break
						} else if c.Status == constants.StatusInstalling {
							log.Debugf("Host cluster is not in 'cluster-running' state got cluster stat %s", c.Status)
							IsHostClusterInstalling = true
							break
						}
					}
				}
				if IsHostClusterInstalling {
					utils.ResponseError(response, http.StatusBadRequest, "Host cluster is currentlly not ready yet, please ensure host cluster is ready and try again.")
					return
				}

				if !IsHostClusterExist {
					utils.ResponseError(response, http.StatusBadRequest, "Please create host cluster first.")
					return
				}
			}

			if clusterPost.CloudProvider == nil && clusterPost.Storage.NFS == nil {
				utils.ResponseError(response, http.StatusBadRequest, "kubesphere hosts cluster require at least one storage backend.")
				return
			}
			// install kubesphere hosts cluster and kubesphere crd service
			clusterPost.ClusterRole = clusterPost.KsClusterConf.MultiClusterConfig.ClusterRole
		}
	}
	// enable MinIO & Velero by default
	serverConfig := runtimeCache.GetServerRuntimeConfig(cache.NodeId)

	clusterPost.MinIO = minio.NewDeployMinIO(fmt.Sprintf("%s:%d",
		clusterPost.ContainerRuntime.PrivateRegistryAddress,
		clusterPost.ContainerRuntime.PrivateRegistryPort))

	clusterPost.Velero = velero.NewDeployVelero(fmt.Sprintf("%s:%d",
		clusterPost.ContainerRuntime.PrivateRegistryAddress,
		clusterPost.ContainerRuntime.PrivateRegistryPort))

	clusterPost.MinIO.CompleteMinIODeploy(clusterPost.ClusterId, serverConfig.BackupPath)
	clusterPost.Velero.CompleteVeleroDeploy()

	// enable Auto Restarter by default
	clusterPost.AutoRestarter = &restarter.DeployAutoRestarter{
		Enable: true,
	}
	clusterPost.AutoRestarter.CompleteRestarterDeploy()

	clusterPost.Action = constants.ActionCreate
	// ignore user input operation ids
	// operation ids should only be set by program
	clusterPost.ClusterOperationIDs = []string{}
	// do not accept input value for AdditionalVersionDep
	clusterPost.AdditionalVersionDep = nil

	//if clusterPost.ControlPlane.KubernetesVersion == "" {
	// always set to 1.18.6
	clusterPost.ControlPlane.KubernetesVersion = constants.V1_18_6
	clusterPost.ControlPlane.AllowVirtualKubelet = false
	//}

	clusterPost.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")
	clusterPost.Created = time.Now().Format("2006-01-02T15:04:05Z07:00")
	// do basic data validation here

	common.SetDependencies(&clusterPost)

	// before setup check
	if canProcess, errList, _, err := control_manager.PreCheck(clusterPost); err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.WriteAsJson(schema.HttpErrorResult{
			ErrorCode: -1,
			Message:   err.Error(),
		})
		return
	} else if !canProcess {
		response.WriteHeader(http.StatusBadRequest)
		response.WriteAsJson(schema.HttpErrorResult{
			ErrorCode: 1,
			Message:   "PreCheckFailed",
			ErrorList: errList,
		})
		return
	}

	clusterPost.Status = constants.StatusInstalling

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	if clusterPost.Description == "" {
		clusterPost.Description = clusterPost.Region
		if clusterPost.KsClusterConf != nil {
			if clusterPost.KsClusterConf.MultiClusterConfig.ClusterRole == constants.ClusterRoleHost {
				clusterPost.Description += "-Host Cluster"
			} else {
				clusterPost.Description += fmt.Sprintf("-MemberOf-%s", clusterPost.KsClusterConf.MemberOfCluster)
			}
		}
	}

	if err := control_manager.SetupOrDestroyCluster(&clusterPost, runByUser); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.WriteAsJson(schema.HttpErrorResult{
			ErrorCode: 2,
			Message:   err.Error(),
			ErrorList: []string{},
		})
	}

	response.WriteAsJson(&clusterPost)
}

/*
delete cluster api should be use with extremely caution
it might results in some get fired
*/
func DeleteCluster(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	cluster, err := runtimeCache.GetCluster(clusterId)

	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if cluster == nil || cluster.Status == constants.ClusterStatusDestroyed {
		response.WriteHeader(http.StatusBadRequest)
		response.WriteAsJson(schema.HttpErrorResult{
			ErrorCode: 2,
			Message:   fmt.Sprintf("Unable to find cluster with id %s or it has been deleted already", clusterId),
			ErrorList: []string{},
		})
		return
	}

	if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
		utils.ResponseError(response, http.StatusPreconditionFailed, "Delete a rancher cluster is not allowed.")
		return
	}

	if cluster.IsProtected {
		utils.ResponseError(response, http.StatusPreconditionFailed, "Unable to delete protected cluster.")
		return
	}

	if !(cluster.Status == constants.ClusterStatusRunning || cluster.Status == constants.StatusError) {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to delete cluster with status %s. Is cluster still running some task?", cluster.Status))
		return
	}

	reducedNode := utils.RemoveDuplicateNode(*cluster)

	for _, checkNode := range reducedNode {
		if err := control_manager.CheckNodeReachable(checkNode.NodeId); err != nil {
			utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Node %s client is not activate. Is client daemon running ?", checkNode.NodeId))
			return
		}
	}

	if request.SelectedRoutePath() == "/api/cluster/v1/gracefully_delete/{cluster-id}" {
		namespaceToRelease := []string{}
		if err := request.ReadEntity(&namespaceToRelease); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
			return
		} else {
			if len(namespaceToRelease) > 0 {
				cluster.ReclaimNamespaces = append(cluster.ReclaimNamespaces, namespaceToRelease...)
			}
		}
	}

	cluster.Action = constants.ActionDelete
	cluster.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")

	cluster.Status = constants.StatusDeleting

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	if err := control_manager.SetupOrDestroyCluster(cluster, runByUser); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.WriteAsJson(schema.HttpErrorResult{
			ErrorCode: 2,
			Message:   err.Error(),
			ErrorList: []string{},
		})
	}

}

func DeleteClusterV2(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"),
		})
		return
	}

	recliamNamespace := []string{}
	if err := request.ReadEntity(&recliamNamespace); err != nil {
		log.Debug("No graceful shutdown body detected or parse body error. Skipping !!!")
	}

	cluster, err := runtimeCache.GetCluster(clusterId)

	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
		utils.ResponseError(response, http.StatusPreconditionFailed, "Delete a rancher cluster is not allowed.")
		return
	}

	if cluster.IsProtected {
		utils.ResponseError(response, http.StatusPreconditionFailed, "Unable to delete protected cluster.")
		return
	}

	if cluster == nil || cluster.Status == constants.ClusterStatusDestroyed {
		response.WriteHeader(http.StatusBadRequest)
		response.WriteAsJson(schema.HttpErrorResult{
			ErrorCode: 2,
			Message:   fmt.Sprintf("Unable to find cluster with id %s or it has been deleted already", clusterId),
			ErrorList: []string{},
		})
		return
	}

	if !(cluster.Status == constants.ClusterStatusRunning || cluster.Status == constants.StatusError) {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to delete cluster with status %s. Is cluster still running some task?", cluster.Status))
		return
	}

	if len(recliamNamespace) > 0 {
		cluster.ReclaimNamespaces = append(cluster.ReclaimNamespaces, recliamNamespace...)
	}

	reducedNode := utils.RemoveDuplicateNode(*cluster)

	for _, checkNode := range reducedNode {
		if err := control_manager.CheckNodeReachable(checkNode.NodeId); err != nil {
			utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Node %s client is not activate. Is client daemon running ?", checkNode.NodeId))
			return
		}
	}

	if restplus.GetBoolValueWithDefault(request, "gracefully_delete", false) {
		namespaceToRelease := []string{}
		if err := request.ReadEntity(&namespaceToRelease); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
			return
		} else {
			if len(namespaceToRelease) > 0 {
				cluster.ReclaimNamespaces = append(cluster.ReclaimNamespaces, namespaceToRelease...)
			}
		}
	}

	cluster.Action = constants.ActionDelete
	cluster.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")

	cluster.Status = constants.StatusDeleting

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	if err := control_manager.SetupOrDestroyCluster(cluster, runByUser); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.WriteAsJson(schema.HttpErrorResult{
			ErrorCode: 2,
			Message:   err.Error(),
			ErrorList: []string{},
		})
	}

}

func GetClusterInfo(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}
	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}
	if cluster == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find cluster with id %s", clusterId))
		return
	}
	response.WriteAsJson(cluster)
}

func ListCluster(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	clusterMap, err := runtimeCache.GetClusterCollection()
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to get cluster list with err %s", err))
		return
	}
	regionId := request.QueryParameter("region_id")
	clusters := []schema.Cluster{}
	if regionId == "" {
		for _, cluster := range clusterMap {
			clusters = append(clusters, cluster)
		}
	} else {
		for _, cluster := range clusterMap {
			if cluster.Region == regionId {
				clusters = append(clusters, cluster)
			}
		}
	}

	response.WriteAsJson(clusters)
}

func ListRunningCluster(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	clusterMap, err := runtimeCache.GetClusterCollection()
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to get cluster list with err %s", err))
		return
	}
	regionId := request.QueryParameter("region_id")
	clusterInstaller := request.QueryParameter("cluster_installer")

	if clusterInstaller != constants.ClusterInstallerRancher {
		clusterInstaller = constants.ClusterInstallerKubeadm
	}

	clusters := []schema.Cluster{}
	for _, cluster := range clusterMap {
		if cluster.Status == constants.ClusterStatusRunning && (regionId == "" || cluster.Region == regionId) && cluster.ClusterInstaller == clusterInstaller {
			clusters = append(clusters, cluster)
		}

	}
	response.WriteAsJson(clusters)
}

func AddNodeToCluster(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	nodes := []schema.ClusterNode{}
	if err := request.ReadEntity(&nodes); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}
	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}
	// cluster not found return
	if cluster == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to find cluster id %s", clusterId))
		return
	}
	// if cluster not in running stat return
	if cluster.Status != constants.ClusterStatusRunning {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster is not ready to accept a new node, is that cluster in %s stat ?", constants.ClusterStatusRunning))
		return
	}
	// node not found return
	for _, node := range nodes {
		node, errGetNodeInfo := runtimeCache.GetNodeInformation(node.NodeId)
		if errGetNodeInfo != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve node information due to error %s", errGetNodeInfo.Error()))
			return
		}
		if node == nil {
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to find node id %s", node.Id))
			return
		}

		if node.ClusterInstaller != cluster.ClusterInstaller {
			utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("ClusterInstaller of node '%s' does not match cluster ClusterInstaller", node.ClusterInstaller))
			return
		}

		if node.Supported == false {
			utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cannot add node that does not supported yet to cluster %s", fmt.Sprintf("%s:%s:%s", node.SystemInfo.OS.Vendor, node.SystemInfo.OS.Version, node.SystemInfo.OS.Architecture)))
			return
		}
		if node.IsDisabled == true {
			utils.ResponseError(response, http.StatusPreconditionFailed, "Cannot add node that currently disabled to cluster")
			return
		}
		if node.SystemInfo.Memory.Size < 4096 {
			utils.ResponseError(response, http.StatusPreconditionFailed, "Worker node require an least 4096 MB to join cluster")
			return
		}
		// check node adds to cluster region is proper set and match cluster region
		if node.Region == nil {
			utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Node with id: %s region is not proper set", node.Id))
			return
		} else if node.Region.ID != cluster.Region {
			utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Node with id: %s region id '%s' does not match cluster region id '%s'", node.Id, node.Region.ID, cluster.Region))
			return
		}
	}

	// do node check , is node can add to cluster
	if err := control_manager.CheckNodeExitsAndReachable(nodes, *cluster); err != nil {
		utils.ResponseError(response, http.StatusPreconditionFailed, err.Error())
		return
	}

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	if err := control_manager.AddOrRemoveNodeToCluster(*cluster, nodes, constants.ActionCreate, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteAsJson(&nodes)
}

func RemoveNodeFromClusterDBOnly(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	nodes := []schema.ClusterNode{}
	if err := request.ReadEntity(&nodes); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}
	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}
	// cluster not found return
	if cluster == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to find cluster id %s", clusterId))
		return
	}

	// if cluster not in running stat return
	if cluster.Status != constants.ClusterStatusRunning {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster is not ready to remove nodes, is that cluster in %s stat ?", constants.ClusterStatusRunning))
		return
	}

	// get cluster nodes
	/*	relatedNodes, errRelatedNodes := runtimeCache.GetClusterNodeRelationShip(clusterId)
		if errRelatedNodes != nil {
			utils.ResponseError(response, http.StatusInternalServerError, errRelatedNodes.Error())
			return
		}*/

	// node not found return
	for _, node := range nodes {
		nodeInfo, errGetNodeInfo := runtimeCache.GetNodeInformation(node.NodeId)
		if errGetNodeInfo != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve node information due to error %s", errGetNodeInfo.Error()))
			return
		}
		if nodeInfo == nil {
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to find cluster id %s", clusterId))
			return
		}

		// if find any nodes that does not belong to this cluster return
		/*		if _, found := relatedNodes[node.NodeId]; !found {
				utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Node %s does not belong to cluster %s", node.NodeId, clusterId))
				return
			}*/
		// check nodes contain master ?
		// we do not allow to remove master due to etcd FT as long as master etcd coexist
		// May be in future we will move etcd cluster out of master node
		// move master would involve a lot job to do such as re-config all worker node haproxy config
		for _, master := range cluster.Masters {
			if node.NodeId == master.NodeId {
				utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cannot remove master node %s", node.NodeId))
				return
			}
		}
	}

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	if err := control_manager.RemoveNodeToClusterDBOnly(*cluster, nodes, constants.ActionDelete, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteHeader(http.StatusOK)
}

/*
remove node api should be use with caution
let`s just say shit might happen
*/
func RemoveNodeFromCluster(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	nodes := []schema.ClusterNode{}
	if err := request.ReadEntity(&nodes); err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   err.Error(),
		})
		return
	}
	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}
	// cluster not found return
	if cluster == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to find cluster id %s", clusterId))
		return
	}

	// if cluster not in running stat return
	if cluster.Status != constants.ClusterStatusRunning {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster is not ready to remove nodes, is that cluster in %s stat ?", constants.ClusterStatusRunning))
		return
	}

	// get cluster nodes
	relatedNodes, errRelatedNodes := runtimeCache.GetClusterNodeRelationShip(clusterId)
	if errRelatedNodes != nil {
		utils.ResponseError(response, http.StatusInternalServerError, errRelatedNodes.Error())
		return
	}

	workerCount := 0
	existingWorker := map[string]byte{}

	for _, worker := range cluster.Workers {
		existingWorker[worker.NodeId] = 0
	}
	// node not found return
	for _, node := range nodes {
		nodeInfo, errGetNodeInfo := runtimeCache.GetNodeInformation(node.NodeId)
		if errGetNodeInfo != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve node information due to error %s", errGetNodeInfo.Error()))
			return
		}
		if nodeInfo == nil {
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to find cluster id %s", clusterId))
			return
		}

		// check whether node is a worker
		if nodeInfo.Role&constants.NodeRoleWorker == constants.NodeRoleWorker {
			// only consider it`s a worker for those node already added into cluster
			if _, alreadyAddToCluster := existingWorker[node.NodeId]; alreadyAddToCluster {
				workerCount += 1
			}

		}

		// if find any nodes that does not belong to this cluster return
		if _, found := relatedNodes[node.NodeId]; !found {
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Node %s does not belong to cluster %s", node.NodeId, clusterId))
			return
		}
		// check nodes contain master ?
		// we do not allow to remove master due to etcd FT as long as master etcd coexist
		// May be in future we will move etcd cluster out of master node
		// move master would involve a lot job to do such as re-config all worker node haproxy config
		for _, master := range cluster.Masters {
			if node.NodeId == master.NodeId {
				utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cannot remove master node %s", node.NodeId))
				return
			}
		}
	}

	if len(cluster.Workers) <= workerCount {
		utils.ResponseError(response, http.StatusPreconditionFailed, "Cannot remove nodes because cluster require at least one worker to run applications")
		return
	}

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	if err := control_manager.AddOrRemoveNodeToCluster(*cluster, nodes, constants.ActionDelete, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteHeader(http.StatusOK)
}

func ListClusterNodes(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}
	// cluster not found return
	if cluster == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to find cluster id %s", clusterId))
		return
	}

	nodes, err := runtimeCache.GetNodeInformationWithCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}

	// now try to send a message to request kubectl get nodes
	// ensue the actual kubernetes cluster is ok
	firstMaster, errGetNode := runtimeCache.GetNodeInformation(cluster.Masters[0].NodeId)
	if errGetNode != nil || firstMaster == nil {
		// print error and return nodes data do not listen to master reply for kubectl get nodes
		log.Errorf("Failed to prepare node message %v due to error: %s or unable to find first master", firstMaster, err.Error())
	} else {
		config := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
		doneOrErrorSignal := make(chan struct {
			Error   error
			Message string
		})
		task := schema.TaskKubectl{
			TaskType:        constants.TaskTypeKubectl,
			SubCommand:      constants.KubectlSubCommandGet,
			TimeOut:         2,
			CommandToRun:    []string{"nodes -o json"},
			RequestResponse: true,
		}
		data, err := mqHelper.CreateMsgBody("kubeclt-"+uuid.New().String(), "", task, *cluster, map[string]string{})
		if err != nil {
			// print error and return nodes data do not listen to master reply for kubectl get nodes
			log.Errorf("Failed to prepare node message %s due to error: %s", data, err.Error())
		} else {
			machineIDMap := map[string]*schema.NodeInformation{}
			for index, node := range nodes {
				if node.Role == constants.NodeRoleExternalLB {
					// if the node is only a load balance node , then skip it
					// because it will never appears in result of command "kubectl get nodes"
					nodes[index].KubeNodeStat = "N/A"
					continue
				}
				machineIDMap[node.SystemInfo.Node.Hostname] = &nodes[index]

			}
			go nats.SendingMessageWithReply(mqHelper.GeneratorNodeSubscribe(*firstMaster, config.MessageQueue.SubjectSuffix),
				uuid.New().String(),
				firstMaster.Id,
				"",
				config.MessageQueue, 3*time.Second,
				[]byte(data),
				mqHelper.ParseKubectlGetNodes(machineIDMap, doneOrErrorSignal),
				func(chan struct {
					Error   error
					Message string
				}) func(nodeId string, nodeStepId string) {
					return func(nodeId string, nodeStepId string) {
						doneOrErrorSignal <- struct {
							Error   error
							Message string
						}{Error: errors.New("Message queue timeout "), Message: fmt.Sprintf("Time out when try to reach node %s", nodeId)}
					}
				}(doneOrErrorSignal))
			result := <-doneOrErrorSignal
			if result.Error != nil {
				// print error and return nodes data do not listen to master reply for kubectl get nodes
				log.Errorf("Got api internal server error: %s", result.Message)
			}
		}
	}
	response.WriteAsJson(&nodes)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func CreateTemplate(request *restful.Request, response *restful.Response) {
	clusterPost := schema.ClusterTemplatePost{}
	if err := request.ReadEntity(&clusterPost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	err := utils.Validate(clusterPost)
	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			if e.Tag() == "storage_size" {
				err = fmt.Errorf("storage size should greater than 5 and less than 50.")
			}
		}
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	for _, c := range clusterPost.ContainerRuntimes {
		err := network.CheckIpIsReachable(c.PrivateRegistryAddress, c.PrivateRegistryPort, "tcp", 2*time.Second)
		if err != nil {
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("private registry address %v:%v is not reachable", c.PrivateRegistryAddress, c.PrivateRegistryPort))
			return
		}
	}

	runtimeCache := cache.GetCurrentCache()
	clusterNew := new(schema.ClusterTemplate)

	// TODO: fix nil point
	clusterNew.ContainerRuntimes = clusterPost.ContainerRuntimes
	clusterNew.CloudProvider = clusterPost.CloudProvider
	// TODO: remove useless components
	// clusterNew.PostgresOperator = clusterPost.PostgresOperator
	// clusterNew.PostgresOperator.CompletePostgresOperatorDeploy()
	// clusterNew.MiddlePlatform = clusterPost.MiddlePlatform
	// clusterNew.MiddlePlatform.CompleteMiddlePlatformDeploy()
	// clusterNew.EFK = clusterPost.EFK
	// clusterNew.EFK.CompleteEFKDeploy()
	// clusterNew.GAP = &gap.DeployGAP{}
	// clusterNew.GAP.CompleteGAPDeploy()
	clusterNew.KSHost = clusterPost.KSHost
	// clusterNew.Console = clusterPost.Console
	// clusterNew.Console.CompleteConsoleDeploy()
	clusterNew.ClusterDnsUpstream = clusterPost.DNSUpStreamServer
	if clusterPost.KSClusterConf != nil {
		clusterNew.KSClusterConf = clusterPost.KSClusterConf
		clusterNew.KSClusterConf.CompleteDeploy()
	}
	err = runtimeCache.SaveClusterTemplate(clusterNew)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}
	response.WriteAsJson(clusterNew)
}

func GetTemplate(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	template, err := runtimeCache.GetTemplate()
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}
	if template == nil {
		utils.WriteSingleEmptyStructWithStatus(response)
	} else {
		response.WriteAsJson(&template)
	}

}

// TODO: add impl, used by createCluster
func validateClusterNodeRegion(c *schema.Cluster) error {
	return nil
}

func BackUpCreate(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	backupPost := schema.Backup{}
	if err := request.ReadEntity(&backupPost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if cluster == nil || cluster.Status != constants.ClusterStatusRunning {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to find cluster with id %s or cluster status is not running", clusterId))
		return
	}
	if cluster.Mock {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("This operation is not allowed for mock cluster"))
		return
	}
	if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to create backup rancher cluster with id %s", clusterId))
		return
	}

	cluster.Action = constants.ActionBackup
	cluster.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")
	cluster.Status = constants.ClusterStatusBackingUp

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}
	if err := control_manager.BackupCreate(schema.Backup{Action: backup_restore.Create, BackupName: backupPost.BackupName},
		cluster, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
}

func BackUpDelete(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}
	if cluster == nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to find cluster with id %s or cluster status is not running", clusterId))
		return
	}
	if cluster.Mock {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("This operation is not allowed for mock cluster"))
		return
	}
	if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Create backup for rancher cluster is not allowed"))
		return
	}
	cluster.Action = constants.ActionBackup
	cluster.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	list, _ := control_manager.BackupList(schema.Backup{Action: backup_restore.List, Output: backup_restore.OutTable},
		cluster, runByUser)

	arr := strings.Split(list, "\n")
	ok := true
	for _, val := range arr {
		if strings.Contains(val, request.PathParameter("backup-name")) &&
			(strings.Contains(val, "InProgress") || strings.Contains(val, "Deleting")) {
			ok = !ok
		}
	}
	if !ok {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("unable to delete backup file which currently using by progress: %s", request.PathParameter("backup-name")))
		return
	}
	if err := control_manager.BackupDelete(schema.Backup{Action: backup_restore.Delete, BackupName: request.PathParameter("backup-name")},
		cluster, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
}

func BackUpList(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if cluster == nil || cluster.Status == constants.ClusterStatusDestroyed {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to find cluster with id %s or cluster status is not running", clusterId))
		return
	}

	if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
		result := schema.BackupList{}
		result.Kind = "BackupList"
		result.Items = []schema.BackupDetail{}
		result.ClusterName = cluster.ClusterName
		result.Region = cluster.Region
		response.WriteAsJson(&result)
		return
	}
	if cluster.Mock {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("This operation is not allowed for mock cluster"))
		return
	}
	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}
	res, err := control_manager.BackupList(schema.Backup{Action: backup_restore.List, Output: backup_restore.OutJson}, cluster, runByUser)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
	var result schema.BackupList
	if strings.Contains(res, "BackupList") {
		if err = listStructer(res, &result); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		detail := schema.BackupDetail{}
		if err = listStructer(res, &detail); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, err.Error())
			return
		}
		result.Kind = "BackupList"
		result.Items = []schema.BackupDetail{}
		result.Items = append(result.Items, detail)
	}
	result.ClusterName = cluster.ClusterName
	result.Region = cluster.Region
	response.WriteAsJson(&result)
}

func listStructer(str string, v interface{}) error {
	if str == "" {
		return nil
	}

	if strings.Contains(str, "An error occurred") {
		return errors.New(str)
	}

	return json.Unmarshal([]byte(str), v)
}

func RestoreCreate(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if cluster == nil || cluster.Status != constants.ClusterStatusRunning {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to find cluster with id %s or cluster status is not running", clusterId))
		return
	}
	if cluster.Mock {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("This operation is not allowed for mock cluster"))
		return
	}
	if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Create restore for rancher cluster is not allowed"))
		return
	}
	cluster.Action = constants.ActionRestore
	cluster.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")
	cluster.Status = constants.ClusterStatusRestoreing
	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}
	if err := control_manager.RestoreCreate(schema.Restore{Action: backup_restore.Create, BackupName: request.PathParameter("backup-name")},
		cluster, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
}

func RestoreDelete(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if cluster == nil || cluster.Status != constants.ClusterStatusRunning {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to find cluster with id %s or cluster status is not running", clusterId))
		return
	}
	if cluster.Mock {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("This operation is not allowed for mock cluster"))
		return
	}
	if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Delete restore for rancher cluster is not allowed"))
		return
	}

	cluster.Action = constants.ActionRestore
	cluster.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}
	if err := control_manager.RestoreDelete(schema.Restore{Action: backup_restore.Delete, RestoreName: request.PathParameter("restore-name")},
		cluster, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
}

func RestoreList(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if cluster == nil || cluster.Status == constants.ClusterStatusDestroyed {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to find cluster with id %s or cluster status is not running", clusterId))
		return
	}

	if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
		result := schema.RestoreList{}
		result.Kind = "RestoreList"
		result.Items = []schema.RestoreDetail{}
		response.WriteAsJson(&result)
		return
	}

	if cluster.Mock {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("This operation is not allowed for mock cluster"))
		return
	}

	cluster.Action = constants.ActionRestore
	cluster.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}
	res, err := control_manager.RestoreList(schema.Restore{Action: backup_restore.List}, cluster, runByUser)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
	var result schema.RestoreList
	if strings.Contains(res, "RestoreList") {
		if err = listStructer(res, &result); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		detail := schema.RestoreDetail{}
		if err = listStructer(res, &detail); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, err.Error())
			return
		}
		result.Kind = "RestoreList"
		result.Items = []schema.RestoreDetail{}
		result.Items = append(result.Items, detail)
	}
	response.WriteAsJson(&result)
}

func BackupRegularCreate(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	body := schema.BackupRegular{}
	if err := request.ReadEntity(&body); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}
	if cluster.Mock {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("This operation is not allowed in mock cluster"))
		return
	}
	if cluster == nil || cluster.Status != constants.ClusterStatusRunning {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to find cluster with id %s or cluster status is not running", clusterId))
		return
	}

	if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Create regular backup restore for rancher cluster is not allowed"))
		return
	}

	cluster.Action = constants.ActionBackupRegular
	cluster.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")
	cluster.Status = constants.ClusterStatusBackingUp
	cluster.BackupRegularEnable = true
	cluster.BackupRegularName = body.BackupRegularName

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}
	if err := control_manager.BackupRegularCreate(
		schema.BackupRegular{
			Action: backup_restore.Create, CronjobTime: body.CronjobTime,
			BackupRegularName: body.BackupRegularName, Args: body.Args,
		},
		cluster, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
}

func BackupRegularDelete(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if cluster == nil || cluster.Status == constants.ClusterStatusDestroyed {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to find cluster with id %s or cluster status is not running", clusterId))
		return
	}
	if cluster.Mock {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("This operation is not allowed for mock cluster"))
		return
	}
	if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("This operation is not appropritate for rancher cluster"))
		return
	}

	cluster.Action = constants.ActionBackupRegular
	cluster.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")
	cluster.BackupRegularEnable = false
	cluster.BackupRegularName = ""

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}
	if err := control_manager.BackupRegularDelete(request.PathParameter("backup-regular-name"), cluster, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
}

func MockOrUnMockCluster(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if cluster == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find cluster with id", clusterId))
		return
	}

	if strings.Contains(request.SelectedRoutePath(), "/mock/") {
		cluster.Mock = true
	} else {
		cluster.Mock = false
	}

	if err := runtimeCache.SaveOrUpdateClusterCollection(cluster.ClusterId, *cluster); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to save cluster to db due to error: %s", err.Error()))
		return
	}

}

func BackupRegularGet(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if cluster == nil || cluster.Status != constants.ClusterStatusRunning {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find cluster with id %s or cluster status is not running", clusterId))
		return
	}

	if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
		result := schema.BackupRegularlyDetail{}
		response.WriteAsJson(&result)
		return
	}

	if cluster.Mock {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("This operation is not allowed for mock cluster"))
		return
	}

	cluster.Action = constants.ActionBackupRegular
	cluster.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}
	res, err := control_manager.BackupRegularGet(request.PathParameter("backup-regular-name"), cluster, runByUser)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
	result := schema.BackupRegularlyDetail{}
	err = listCronjobStructer(res, &result)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteAsJson(&result)
}

func listCronjobStructer(str string, v interface{}) error {
	if str == "" {
		return nil
	}

	if strings.Contains(strings.ToLower(str), "error") {
		return errors.New(str)
	}

	return json.Unmarshal([]byte(str), v)
}
