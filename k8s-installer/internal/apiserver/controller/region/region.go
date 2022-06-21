package region

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/google/uuid"
	"k8s-installer/internal/apiserver/utils"
	"k8s-installer/pkg/cache"
	"k8s-installer/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListRegions(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	regionMap, err := runtimeCache.GetRegionList()
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}
	regionList := []schema.Region{}
	for _, item := range regionMap {
		regionList = append(regionList, item)
	}

	_ = response.WriteAsJson(regionList)
}

func CreateRegion(request *restful.Request, response *restful.Response) {
	body := schema.Region{}
	if err := request.ReadEntity(&body); err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   err.Error(),
		})
		return
	}
	err := utils.Validate(body)
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}
	body.ID = uuid.New().String()
	body.CreationTimestamp = metav1.Now()
	if err = cache.GetCurrentCache().CreateOrUpdateRegion(body); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
	_ = response.WriteAsJson(&body)
}

func DeleteRegion(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	regionID := request.PathParameter("region")
	r, err := runtimeCache.GetRegion(regionID)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
	if r != nil {
		if err = runtimeCache.DeleteRegion(regionID); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, err.Error())
			return
		}
	}
	response.WriteHeader(http.StatusOK)
	return
}

func UpdateRegion(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	regionID := request.PathParameter("region")
	body := schema.Region{}
	if err := request.ReadEntity(&body); err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   err.Error(),
		})
		return
	}
	err := utils.Validate(body)
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}
	r, err := runtimeCache.GetRegion(regionID)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
	if r == nil || r.ID != body.ID {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find region with id %s or region id not equal with body", regionID),
		})
		return
	}
	body.CreationTimestamp = r.CreationTimestamp
	ts := metav1.Now()
	body.UpdateTimestamp = &ts
	if err = cache.GetCurrentCache().CreateOrUpdateRegion(body); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
	_ = response.WriteAsJson(&body)
}

func GetRegion(request *restful.Request, response *restful.Response) {
	regionID := request.PathParameter("region")
	r, err := cache.GetCurrentCache().GetRegion(regionID)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
	if r == nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find region with id %s", regionID),
		})
		return
	}
	_ = response.WriteAsJson(r)
}

func AssignNodeRegion(request *restful.Request, response *restful.Response) {
	nodeID := request.PathParameter("node")
	regionID := request.PathParameter("region")
	if strings.TrimSpace(nodeID) == "" {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find path parameter: %s", "node-id"),
		})
		return
	}
	runtimeCache := cache.GetCurrentCache()
	node, errGetNodeInfo := runtimeCache.GetNodeInformation(nodeID)
	if errGetNodeInfo != nil {
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, &schema.HttpErrorResult{
			ErrorCode: http.StatusInternalServerError,
			Message:   fmt.Sprintf("Failed to query node %s due to error %s", nodeID, errGetNodeInfo.Error()),
		})
		return
	}
	if node == nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find node with id %s", nodeID),
		})
		return
	}

	if node.IsDisabled {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to change node region %s because this node is disabled", nodeID))
		return
	}

	if node.BelongsToCluster != "" {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to change node status %s because this node is belong cluster %s", nodeID, node.BelongsToCluster))
		return
	}

	if node.Role != 0 {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to delete node %s because this node is not clean. The role still set to %d", nodeID, node.Role))
		return
	}

	r, err := runtimeCache.GetRegion(regionID)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
	if r == nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find region with id %s", regionID),
		})
		return
	}

	node.Region = &schema.Region{
		ID:                r.ID,
		Name:              r.Name,
		CreationTimestamp: r.CreationTimestamp,
		UpdateTimestamp:   r.UpdateTimestamp.DeepCopy(),
	}

	if err := runtimeCache.SaveOrUpdateNodeInformation(nodeID, *node); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to disable node %s due to error %s", nodeID, err.Error()))
		return
	}

	response.WriteAsJson(node)
}

func ClearNodeRegion(request *restful.Request, response *restful.Response) {
	nodeId := request.PathParameter("node")
	if strings.TrimSpace(nodeId) == "" {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find path parameter: %s", "node-id"),
		})
		return
	}
	runtimeCache := cache.GetCurrentCache()
	node, errGetNodeInfo := runtimeCache.GetNodeInformation(nodeId)
	if errGetNodeInfo != nil {
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, &schema.HttpErrorResult{
			ErrorCode: http.StatusInternalServerError,
			Message:   fmt.Sprintf("Failed to query node %s due to error %s", nodeId, errGetNodeInfo.Error()),
		})
		return
	}
	if node == nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find node with id %s", nodeId),
		})
		return
	}
	node.Region = nil
	if err := runtimeCache.SaveOrUpdateNodeInformation(nodeId, *node); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to disable node %s due to error %s", nodeId, err.Error()))
		return
	}
	response.WriteAsJson(node)
}
