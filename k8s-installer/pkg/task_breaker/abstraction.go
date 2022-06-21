package task_breaker

import (
	"encoding/json"
	"errors"
	"fmt"
	runtimeCache "k8s-installer/pkg/cache"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/license"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
)

type IAddOns interface {
	DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error
	DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error
	DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error
	LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error)
	GetAddOnName() string
	SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns
	IsEnable() bool
}

type ITaskBreakDown interface {
	BreakDownTask() (schema.Operation, error)
}

func SetOperationStep(operation schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeCollectionList schema.NodeInformationCollection) (schema.Operation, error) {
	switch operation.OperationType {
	case constants.OperationTypeClusterSetupOrDestroy:
		sumLicenseLabel, err := license.SumLicenseLabel()
		if err != nil {
			return operation, err
		}
		clusterTaskBreakDown := ClusterTaskBreakDown{
			Operation:          operation,
			Cluster:            cluster,
			Config:             config,
			NodeCollectionList: nodeCollectionList,
			SumLicenseLabel:    sumLicenseLabel,
		}
		return clusterTaskBreakDown.BreakDownTask()
	case constants.OperationTypeAddOrRemoveNodesToCluster:
		sumLicenseLabel, err := license.SumLicenseLabel()
		var nodesToAddOrRemove []schema.ClusterNode
		data, err := json.Marshal(&operation.RequestParameter.BodyParameters)
		if err != nil {
			return schema.Operation{}, err
		}
		errConvert := json.Unmarshal(data, &nodesToAddOrRemove)
		if errConvert != nil {
			return schema.Operation{}, errConvert
		}
		clusterNodeTaskBreakDown := ClusterNodeTaskBreakDown{
			Operation:          operation,
			Cluster:            cluster,
			Config:             config,
			NodesToAddOrRemove: nodesToAddOrRemove,
			NodeCollectionList: nodeCollectionList,
			Action:             operation.RequestParameter.RuntimeParameters["addOrRemove"].(string),
			SumLicenseLabel:    sumLicenseLabel,
		}
		return clusterNodeTaskBreakDown.BreakDownTask()
	case constants.OperationTypeManageAddonsToCluster:
		var addons schema.Addons
		data, err := json.Marshal(&operation.RequestParameter.BodyParameters)
		if err != nil {
			return schema.Operation{}, err
		}
		errConvert := json.Unmarshal(data, &addons)
		if errConvert != nil {
			return schema.Operation{}, errConvert
		}
		batchAddonsTaskBreakdown := BatchAddonsTaskBreakDown{
			Operation: &operation,
			Cluster:   &cluster,
			NewAddons: addons,
		}
		return batchAddonsTaskBreakdown.BreakDownTask()
	case constants.OperationTypeUpgradeCluster:
		if val, exists := operation.RequestParameter.QueryParameters["planId"]; exists {
			rtc := runtimeCache.GetCurrentCache()
			plan, errPlan := rtc.GetUpgradePlan(val.(string))
			if errPlan != nil {
				return operation, errPlan
			}
			upgradeVersion, errVer := rtc.GetUpgradeVersion(plan.TargetVersion)
			if errVer != nil {
				return operation, errVer
			}

			upgradeTaskBreakDown := UpgradingTaskBreakDown{
				Cluster:          cluster,
				Operation:        operation,
				NodesCollection:  nodeCollectionList,
				TargetK8sVersion: plan.TargetVersion,
				ApplyVersion:     *upgradeVersion,
				Dep:              upgradeVersion.ToDep(),
				Config:           rtc.GetServerRuntimeConfig(runtimeCache.NodeId),
			}
			return upgradeTaskBreakDown.BreakDownTask()
		} else {
			return operation, errors.New("Unable to find runtime parameter: 'planId' ")
		}
	case constants.OperationTypeSingleTask:
		return operation, nil
	default:
		return schema.Operation{}, errors.New(fmt.Sprintf("Operation type %s is not valid operation type", operation.OperationType))
	}
}
