package operation

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"k8s-installer/internal/apiserver/utils"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/control_manager"
	"k8s-installer/schema"
)

func ListOperationWithUpgradePlan(request *restful.Request, response *restful.Response) {
	upgradePlanId := request.PathParameter("upgrade-plan-id")
	if strings.TrimSpace(upgradePlanId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "upgrade-plan-id"))
		return
	}
	runtimeCache := cache.GetCurrentCache()
	upgradePlane, err := runtimeCache.GetUpgradePlan(upgradePlanId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if upgradePlane == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find upgrade plan with id %s", upgradePlanId))
		return
	}

	operationColection, errOperation := runtimeCache.GetOperationCollection()
	if errOperation != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	var result []schema.Operation
	if len(upgradePlane.Operations) == 0 || len(operationColection) == 0 {
		response.WriteAsJson([]schema.Operation{})
	} else {
		for _, operation := range upgradePlane.Operations {
			if foundUpgradePlane, found := operationColection[operation]; found {
				result = append(result, foundUpgradePlane)
			}
		}
		response.WriteAsJson(result)
	}

}

func ListOperation(request *restful.Request, response *restful.Response) {
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
		// if cluster not found return 400
		// cluster id is not found
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find cluster with id %s", clusterId))
		return
	}

	if operations, err := runtimeCache.GetClusterOperationDoNotLoadStep(*cluster); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	} else {
		if operations == nil {
			response.WriteAsJson([]schema.Operation{})
		} else {
			response.WriteAsJson(operations)
		}
	}
}

func ContinueOperation(request *restful.Request, response *restful.Response) {
	operationId := request.PathParameter("operation-id")
	if strings.TrimSpace(operationId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "operation-id"))
		return
	}
	runtimeCache := cache.GetCurrentCache()

	operation, errOperation := runtimeCache.GetOperation(operationId)

	// operation should be found properly
	// we consider it`s a server error if not found
	if errOperation != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", errOperation.Error()))
		return
	}

	if operation == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find operation with id %s ", operationId))
		return
	}

	if operation.Status != constants.StatusError {
		utils.ResponseError(response, http.StatusPreconditionFailed, "Operation already complete no need continue !")
		return
	}

	cluster, err := runtimeCache.GetCluster(operation.ClusterId)

	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if cluster == nil {
		// if cluster not found return 400
		// cluster id is not found
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find cluster with id %s", operation.ClusterId))
		return
	}

	// if cluster still working on cluster heavy work
	// cannot proceed abort
	if cluster.Status == constants.StatusInstalling || cluster.Status == constants.StatusDeleting || cluster.Status == constants.StatusAddingNode || cluster.Status == constants.StatusDeletingNode {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to continue operation because cluster currently still working on critical operation %v", cluster.CurrentOperation))
		return
	}

	// operation id not in current operation meaning operation is successfully completed
	if _, found := cluster.CurrentOperation[operationId]; !found {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to find operation %v which means operation is successfuly completed. No need continue abort !!! ", cluster.CurrentOperation))
		return
	}

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	// do continue
	if err := control_manager.OperationContinue(*cluster, *operation, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable continue operation due to error: %s", err.Error()))
		return
	}

	response.WriteHeader(http.StatusOK)
}
