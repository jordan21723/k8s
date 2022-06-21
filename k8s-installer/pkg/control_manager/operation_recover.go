package control_manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
)

/*
OperationRecover is the one handle recover failed operation runtime stat
*/

type OnAllTaskDoneHandler func(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache) error
type OnStepErrorHandler func(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache, err error) error

type IOperationRecover interface {
	Recover(operation *schema.Operation, cluster *schema.Cluster) (OnAllTaskDoneHandler, OnStepErrorHandler, error)
}

type CreateOrDestroyClusterRecover struct {
}

type AddOrRemoveNodesToClusterRecover struct {
}

type UpgradeRecover struct {
}

func createRecover(operationType string) (IOperationRecover, error) {
	switch operationType {
	case constants.OperationTypeClusterSetupOrDestroy:
		return CreateOrDestroyClusterRecover{}, nil
	case constants.OperationTypeAddOrRemoveNodesToCluster:
		return AddOrRemoveNodesToClusterRecover{}, nil
	case constants.OperationTypeUpgradeCluster:
		return UpgradeRecover{}, nil
	default:
		return nil, errors.New(fmt.Sprintf("Operation type %s is not support", operationType))
	}
}

func (CreateOrDestroyClusterRecover CreateOrDestroyClusterRecover) Recover(operation *schema.Operation, cluster *schema.Cluster) (OnAllTaskDoneHandler, OnStepErrorHandler, error) {
	if cluster.Action == constants.ActionDelete {
		cluster.Status = constants.StatusDeleting
	} else {
		cluster.Status = constants.StatusInstalling
	}
	operation.Status = constants.StatusProcessing
	return clusterOperationDoneHandler, clusterOperationErrorHandler, nil
}

func (AddOrRemoveNodesToClusterRecover AddOrRemoveNodesToClusterRecover) Recover(operation *schema.Operation, cluster *schema.Cluster) (OnAllTaskDoneHandler, OnStepErrorHandler, error) {
	if val, exists := operation.RequestParameter.RuntimeParameters["addOrRemove"]; exists {
		addOrRemove := val.(string)
		var nodesToAddOrRemove []schema.ClusterNode
		data, err := json.Marshal(&operation.RequestParameter.BodyParameters)
		if err != nil {
			return nil, nil, err
		}
		errConvert := json.Unmarshal(data, &nodesToAddOrRemove)
		if errConvert != nil {
			return nil, nil, err
		}
		if addOrRemove == constants.ActionDelete {
			cluster.Status = constants.StatusDeletingNode
		} else {
			cluster.Status = constants.StatusAddingNode
		}
		operation.Status = constants.StatusProcessing
		return addOrRemoveNodeDoneHandlerCreator(addOrRemove, nodesToAddOrRemove), clusterOperationIgnoreErrorHandler, nil
	} else {
		return nil, nil, errors.New("Unable to find runtime parameter: 'addOrRemove' ")
	}
}

func (upgrade UpgradeRecover) Recover(operation *schema.Operation, cluster *schema.Cluster) (OnAllTaskDoneHandler, OnStepErrorHandler, error) {
	if val, exists := operation.RequestParameter.QueryParameters["planId"]; exists {
		rtc := cache.GetCurrentCache()
		plan, errPlan := rtc.GetUpgradePlan(val.(string))
		if errPlan != nil {
			return nil, nil, errPlan
		}
		upgradeVersion, errVer := rtc.GetUpgradeVersion(plan.TargetVersion)
		if errVer != nil {
			return nil, nil, errVer
		}

		cluster.Status = constants.StatusUpgrading
		operation.Status = constants.StatusProcessing
		return applyUpgradeTaskDoneHandler(plan, upgradeVersion), applyUpgradeTaskErrorHandler(plan), nil
	} else {
		return nil, nil, errors.New("Unable to find runtime parameter: 'planId' ")
	}
}
