package addons

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"k8s-installer/internal/apiserver/utils"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/control_manager"
	"k8s-installer/schema"
	"net/http"
	"strings"
)

func ManageClusterAddons(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster-id"))
		return
	}

	addons := schema.Addons{}
	if err := request.ReadEntity(&addons); err != nil {
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

	if cluster == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to find cluster id %s", clusterId))
		return
	}

	if cluster.Status != constants.ClusterStatusRunning {
		utils.ResponseError(response, http.StatusPreconditionFailed, "Cluster is not ready to perform this action. Is cluster in cluster-running stat?")
		return
	}

	// ensure only one deploy addons operation is running
	if operations, err := runtimeCache.GetClusterOperation(*cluster); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, "Failed to check last operation stat")
		return
	} else {
		for _, operation := range operations {
			if operation.OperationType == constants.OperationTypeManageAddonsToCluster && operation.Status == constants.StatusProcessing {
				utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Last deploy addons operation with id %s is still running. Please wait until last operation is complete?", operation.Id))
				return
			}
		}
	}

	if err := control_manager.CheckNodeReachable(cluster.Masters[0].NodeId); err != nil {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("First master %s unreachable. Is client daemon running?", cluster.Masters[0].NodeId))
		return
	}

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	if err := control_manager.AddonsManage(cluster, addons, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}
	response.WriteAsJson(&addons)
}

func ListClusterAddons(request *restful.Request, response *restful.Response) {
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
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to find cluster id %s", clusterId))
		return
	}

	response.WriteAsJson(schema.Addons{
		MiddlePlatform: cluster.MiddlePlatform,
		CloudProvider:  cluster.CloudProvider,
		EFK:            cluster.EFK,
		Console:        cluster.Console,
	},
	)
}
