package control_manager

import (
	"encoding/json"
	"errors"
	"fmt"
	br "k8s-installer/pkg/backup_restore"
	"k8s-installer/pkg/license"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s-installer/node/os"
	"k8s-installer/pkg/cache"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/pkg/log"
	taskBreaker "k8s-installer/pkg/task_breaker"
	"k8s-installer/pkg/util"
	"k8s-installer/schema"

	"github.com/google/uuid"
	natsLib "github.com/nats-io/nats.go"
)

func NodeStateReportInHandler(msg *natsLib.Msg) {
	log.Debugf("Got incoming msg subject %s", msg.Subject)
	// log.Debugf("Got incoming msg content %s", string(msg.Data))
	runtimeCache := cache.GetCurrentCache()
	var nodeInformation schema.NodeInformation
	err := json.Unmarshal(msg.Data, &nodeInformation)
	if err != nil {
		log.Errorf("Failed to parse reported in node stat due to error %s", err)
	} else {
		config := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
		if config.HostRequirement.CheckOSFamily {
			if os.ValidationOsFamilyWithGivenSysInfo(strings.Split(config.HostRequirement.SupportOSFamily, ","),
				strings.Split(config.HostRequirement.SupportOSFamilyVersion, ","),
				nodeInformation.SystemInfo,
				config.HostRequirement.CheckKernelVersion) {
				nodeInformation.Supported = true
			} else {
				nodeInformation.IssueList = append(nodeInformation.IssueList, "Operation system is not support yet")
			}
			log.Debugf("Attempt to get node existing data: %s", nodeInformation.Id)
			lastStat, errGetNodeInfo := runtimeCache.GetNodeInformation(nodeInformation.Id)
			if errGetNodeInfo != nil {
				log.Errorf("Failed to retrieve node last report in status due to error: %s", errGetNodeInfo.Error())
				return
			}
			if lastStat != nil {
				// keep following stat
				log.Debugf("Got existing data with node %s. So keep role , cluster relationship and some key stat", nodeInformation.Id)
				nodeInformation.Role = lastStat.Role
				nodeInformation.BelongsToCluster = lastStat.BelongsToCluster
				nodeInformation.KubeNodeStat = lastStat.KubeNodeStat
				nodeInformation.IsDisabled = lastStat.IsDisabled
				//nodeInformation.Region = lastStat.DeepCopyRegion()
			}
		}
		log.Debugf("Attempt to update node: %s", nodeInformation.Id)
		nodeInformation.LastReportInDate = time.Now().Format("2006-01-02T15:04:05Z07:00")
		if err := runtimeCache.SaveOrUpdateNodeInformation(nodeInformation.Id, nodeInformation); err != nil {
			log.Errorf("Failed to save reported in node stat to cache due to error %s", err)
		}
		handlerReportInRegion(nodeInformation.Region, runtimeCache)
	}
}

func handlerReportInRegion(region *schema.Region, runtimeCache cache.ICache) {
	existRegion, err := runtimeCache.GetRegion(region.ID)
	if err != nil {
		log.Errorf("Failed to found existing region with id %s due to error %s", region.ID, err.Error())
		return
	}
	if existRegion == nil {
		// it`s a new region,save to db
		region.CreationTimestamp = metav1.Now()
		err = runtimeCache.CreateOrUpdateRegion(*region)
		if err != nil {
			log.Errorf("Failed to save region with id %s due to error %s", region.ID, err.Error())
		}
	} else {
		log.Debugf("Region with id %s already exists do nothing", region.ID)
	}
}

func ApplyUpgrade(plan *schema.UpgradePlan, cluster *schema.Cluster, runByUser string) error {

	if cluster.Mock {
		return fmt.Errorf("Unable to upgrade a mocked cluster '%s' ", cluster.ClusterId)
	}

	operation := createOperation("UpgradeCluster", cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeUpgradeCluster
	operation.RunByUser = runByUser
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"planId": plan.Id,
	}
	runtimeCache := cache.GetCurrentCache()
	upgradeVersion, err := runtimeCache.GetUpgradeVersion(plan.TargetVersion)
	plan.Operations = append(plan.Operations, operation.Id)
	plan.OperationId = operation.Id
	plan.ApplyDate = time.Now().Format("2006-01-02T15:04:05Z07:00")
	plan.Status = constants.StatusUpgrading
	if err := runtimeCache.CreateOrUpdateUpgradePlan(*plan); err != nil {
		log.Errorf("Failed to save upgrade plan due to error: %s", err.Error())
		return err
	}
	if err != nil {
		log.Errorf("Failed to get upgrade version %s information due to error: %s", plan.TargetVersion, err.Error())
		return err
	}
	if upgradeVersion == nil {
		return errors.New(fmt.Sprintf("Failed to get upgrade version %s information due to not found", plan.TargetVersion))
	}

	nodesCollections, nodeCollectionErr := runtimeCache.GetNodeInformationCollection()

	if nodeCollectionErr != nil {
		log.Errorf("Failed to get node collection due to error: %s", nodeCollectionErr.Error())
		return nodeCollectionErr
	}

	upgradeTaskBreakDown := taskBreaker.UpgradingTaskBreakDown{
		Cluster:          *cluster,
		Operation:        operation,
		NodesCollection:  nodesCollections,
		TargetK8sVersion: plan.TargetVersion,
		ApplyVersion:     *upgradeVersion,
		Dep:              upgradeVersion.ToDep(),
		Config:           runtimeCache.GetServerRuntimeConfig(cache.NodeId),
	}

	operationBD, errBreakDownTask := upgradeTaskBreakDown.BreakDownTask()

	if errBreakDownTask != nil {
		log.Errorf("Failed to break down upgrading task due to error: %s", errBreakDownTask.Error())
		return err
	}

	if err := runtimeCache.SaveOrUpdateOperationCollection(operation.Id, operation); err != nil {
		log.Errorf("Failed to save operation information to db due to error: %s", err.Error())
		return err
	}

	cluster.Status = constants.StatusUpgrading
	if cluster.CurrentOperation == nil {
		cluster.CurrentOperation = map[string]byte{}
	}
	cluster.CurrentOperation[operation.Id] = 0
	cluster.ClusterOperationIDs = append(cluster.ClusterOperationIDs, operation.Id)

	if err := runtimeCache.SaveOrUpdateClusterCollection(cluster.ClusterId, *cluster); err != nil {
		log.Errorf("Failed to save cluster information to db due to error: %s", err.Error())
		return err
	}

	go doTask(&operationBD, cluster, &nodesCollections, applyUpgradeTaskDoneHandler(plan, upgradeVersion), applyUpgradeTaskErrorHandler(plan))
	return nil
}

func UpgradeDownloadNewKubeadm(cluster *schema.Cluster, runByUser string, upgradeToVersion string, deps dep.DepMap) ([]byte, error) {

	operation := createOperation("DownloadNewKubeadm", cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeSingleTask
	operation.RunByUser = runByUser
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"clusterId": cluster.ClusterId,
	}
	operation.Step = map[int]*schema.Step{}
	var err error
	taskBreakerDown := taskBreaker.MakeUpgradePlanTaskBreakDown{
		Cluster:          *cluster,
		Operation:        operation,
		DepMapping:       deps,
		TargetK8sVersion: upgradeToVersion,
	}
	operation, err = taskBreakerDown.BreakDownTask()

	if err != nil {
		log.Errorf("Failed bread down task for DownloadNewKubeadm task due to error: %s", err.Error())
		return nil, err
	}
	runtimeCache := cache.GetCurrentCache()
	if err := saveOperationStatus(runtimeCache, operation); err != nil {
		log.Errorf("Failed to save operations status to db due to error: %s", err.Error())
		return nil, err
	}

	return DoSingleTask(&operation, cluster, map[string]string{})
}

func AddonsManage(cluster *schema.Cluster, addons schema.Addons, runByUser string) error {
	operation := createOperation("DeployAddons", cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeManageAddonsToCluster
	operation.RunByUser = runByUser
	operation.RequestParameter.BodyParameters = addons
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"clusterId": cluster.ClusterId,
	}
	runtimeCache := cache.GetCurrentCache()

	// we should keep cluster addons stat not changed until related task is done
	copyCluster := *cluster

	sumLicenseLabel, errLicense := license.SumLicenseLabel()
	if errLicense != nil {
		return errLicense
	}

	batchAddonsTaskBreakdown := taskBreaker.BatchAddonsTaskBreakDown{
		Operation:       &operation,
		Cluster:         &copyCluster,
		OriginalCluster: *cluster,
		NewAddons:       addons,
		SumLicenseLabel: sumLicenseLabel,
	}

	var err error
	operation, err = batchAddonsTaskBreakdown.BreakDownTask()

	if err != nil {
		log.Errorf("Failed to break down task due to error: %s", err.Error())
		return err
	}

	nodeInfoCollection, errGetNode := runtimeCache.GetNodeInformationCollection()
	if errGetNode != nil {
		return errGetNode
	}

	// put operation in db and cache
	if err := saveOperationStatus(runtimeCache, operation); err != nil {
		return err
	}

	go doTask(&operation, &copyCluster, &nodeInfoCollection, nil, nil)

	return nil
}

func OperationContinue(cluster schema.Cluster, operation schema.Operation, runByUser string) error {
	operation.LastRun = time.Now().Format("2006-01-02T15:04:05Z07:00")
	operation.RunByUser = runByUser
	recover, err := createRecover(operation.OperationType)
	if err != nil {
		return err
	}

	if onDone, onError, err2 := recover.Recover(&operation, &cluster); err2 != nil {
		return err
	} else {
		runtimeContext := cache.GetCurrentCache()
		if err := saveClusterStatus(runtimeContext, cluster); err != nil {
			return err
		}
		operation.Logs = append(operation.Logs, util.LogStyleMessage("debug", fmt.Sprintf("Operation %s is going to resume from error step %s", operation.Id, operation.Step[operation.CurrentStep].Name)))
		if err := saveOperationStatus(runtimeContext, operation); err != nil {
			return err
		}

		nodeInfoCollection, errGetNode := runtimeContext.GetNodeInformationCollection()
		if errGetNode != nil {
			return errGetNode
		}

		go doTask(&operation, &cluster, &nodeInfoCollection, onDone, onError)
		return nil
	}
}

func SetupOrDestroyCluster(cluster *schema.Cluster, runByUser string) error {
	runtimeCache := cache.GetCurrentCache()
	operationName := "SetupCluster"
	if cluster.Action == constants.ActionDelete {
		operationName = "DestroyCluster"
	}

	// mock meaning do not actually do setup or destroy cluster instead just setup db stat as normal setup
	if cluster.Mock {
		operationName = "Mock" + operationName
	}

	SetReclaimNamespaces(cluster)

	nodeInfoCollection, errGetNode := runtimeCache.GetNodeInformationCollection()
	if errGetNode != nil {
		return errGetNode
	}

	sumLicenseLabel, err := license.SumLicenseLabel()
	if err != nil {
		return err
	}

	operation, err := createOperationFromCluster(operationName, *cluster, runtimeCache.GetServerRuntimeConfig(cache.NodeId), nodeInfoCollection, constants.StatusProcessing, sumLicenseLabel)

	if err != nil {
		return err
	}

	operation.RunByUser = runByUser
	operation.OperationType = constants.OperationTypeClusterSetupOrDestroy

	cluster.CurrentOperation = map[string]byte{}
	// set current operation id in order to recover it from error stat in future
	cluster.CurrentOperation[operation.Id] = 0

	// add operation id to it`s related operations
	cluster.ClusterOperationIDs = append(cluster.ClusterOperationIDs, operation.Id)

	if cluster.Action == constants.ActionCreate {
		// during the cluster installation
		// we lock down node immediately
		nodeRoleList := map[string]int{}
		tagNodeRole(nodeRoleList, cluster.Masters, constants.NodeRoleMaster)
		tagNodeRole(nodeRoleList, cluster.Workers, constants.NodeRoleWorker)
		if cluster.Ingress != nil {
			tagNodeRole(nodeRoleList, cluster.Ingress.NodeIds, constants.NodeRoleIngress)
		}
		if cluster.ClusterLB != nil {
			tagNodeRole(nodeRoleList, cluster.ClusterLB.Nodes, constants.NodeRoleExternalLB)
		}
		var nodeToClusterRelation []string
		// update node
		for nodeId, role := range nodeRoleList {
			nodeToClusterRelation = append(nodeToClusterRelation, nodeId)
			nodeData, errGetNodeInfo := runtimeCache.GetNodeInformation(nodeId)
			if errGetNodeInfo != nil {
				log.Errorf("Failed to retrieve node information due to error: %s", errGetNodeInfo.Error())
				return errGetNodeInfo
			}
			if nodeData == nil {
				log.Errorf("Unable to get node %s information", nodeId)
				return fmt.Errorf("Unable to get node %s information ", nodeId)
			}
			nodeData.Role = role
			nodeData.BelongsToCluster = cluster.ClusterId
			if (role&constants.NodeRoleMaster == constants.NodeRoleMaster) || (role&constants.NodeRoleWorker == constants.NodeRoleWorker) {
				// only master and worker need to set kube stat
				nodeData.KubeNodeStat = constants.UnAvailable
			}
			if err := saveNodeStatus(runtimeCache, *nodeData); err != nil {
				return err
			}
		}

		// save cluster node relation ship
		if err := saveRelationshipStatus(runtimeCache, cluster.ClusterId, nodeToClusterRelation); err != nil {
			return err
		}
	} else {
		if err := saveMasterWorkerKubeStat(*cluster, ""); err != nil {
			return err
		}
	}

	// put cluster in db and cache
	if err := saveClusterStatus(runtimeCache, *cluster); err != nil {
		return err
	}

	// put operation in db and cache
	if err := saveOperationStatus(runtimeCache, operation); err != nil {
		return err
	}

	// if mock
	if cluster.Mock {
		// always reset the operation step to nothing but wait 1 seconds
		mockStep := taskBreaker.CreateWaitStep(1, "MockStep")
		operation.Step = map[int]*schema.Step{
			0: &mockStep,
		}
	}

	go doTask(&operation, cluster, &nodeInfoCollection, clusterOperationDoneHandler, clusterOperationErrorHandler)

	return nil
}

func RemoveNodeToClusterDBOnly(cluster schema.Cluster, nodesToAddOrRemove []schema.ClusterNode, addOrRemove string, runByUser string) error {
	runtimeCache := cache.GetCurrentCache()
	var err error

	initState := constants.StatusDeletingNode
	operationName := "RemoveNodeFromClusterDB"

	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.RunByUser = runByUser
	// set current operation id in order to recover it from error stat in future

	if cluster.CurrentOperation == nil {
		cluster.CurrentOperation = map[string]byte{}
	}

	cluster.CurrentOperation[operation.Id] = 0
	cluster.Status = initState

	// add operation id to it`s related operations
	cluster.ClusterOperationIDs = append(cluster.ClusterOperationIDs, operation.Id)

	operation.OperationType = constants.OperationTypeAddOrRemoveNodesToCluster
	// save request parameters in operation
	operation.RequestParameter.BodyParameters = nodesToAddOrRemove
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"operationId": operation.Id,
	}
	operation.RequestParameter.RuntimeParameters = map[string]interface{}{
		"addOrRemove": addOrRemove,
	}

	// during the remove node phase
	// we do not unlock node immediately by set role to 0 only remove the relation with node only for data consistency reason
	for _, node := range nodesToAddOrRemove {
		nodeData, errGetNodeInfo := runtimeCache.GetNodeInformation(node.NodeId)
		if errGetNodeInfo != nil {
			log.Errorf("Failed to retrieve node information due to error: %s", errGetNodeInfo.Error())
			return errGetNodeInfo
		}
		nodeData.KubeNodeStat = ""
		if err := saveNodeStatus(runtimeCache, *nodeData); err != nil {
			return err
		}
	}

	sumLicenseLabel, err := license.SumLicenseLabel()
	nodeInfoCollection, errGetNode := runtimeCache.GetNodeInformationCollection()
	if errGetNode != nil {
		return errGetNode
	}

	clusterNodeTaskBreakDown := taskBreaker.ClusterNodeTaskBreakDown{
		Operation:          operation,
		Cluster:            cluster,
		Config:             runtimeCache.GetServerRuntimeConfig(cache.NodeId),
		NodesToAddOrRemove: nodesToAddOrRemove,
		NodeCollectionList: nodeInfoCollection,
		Action:             addOrRemove,
		SumLicenseLabel:    sumLicenseLabel,
	}

	operation, err = clusterNodeTaskBreakDown.BreakDownTask()
	if err != nil {
		log.Errorf("Failed to break down task due to error %s", err.Error())
		return err
	}

	// replace step
	mockStep := taskBreaker.CreateWaitStep(1, "DoNothing")
	operation.Step = map[int]*schema.Step{
		0: &mockStep,
	}

	// put operation in db and cache
	if err := saveOperationStatus(runtimeCache, operation); err != nil {
		return err
	}

	// put cluster in db and cache
	if err := saveClusterStatus(runtimeCache, cluster); err != nil {
		return err
	}
	go doTask(&operation, &cluster, &nodeInfoCollection, addOrRemoveNodeDoneHandlerCreator(addOrRemove, nodesToAddOrRemove), clusterOperationIgnoreErrorHandler)
	return nil
}

func AddOrRemoveNodeToCluster(cluster schema.Cluster, nodesToAddOrRemove []schema.ClusterNode, addOrRemove string, runByUser string) error {
	runtimeCache := cache.GetCurrentCache()
	var err error
	initState := ""
	operationName := ""
	if addOrRemove == constants.ActionCreate {
		initState = constants.StatusAddingNode
		operationName = "AddNodeToCluster"
	} else {
		initState = constants.StatusDeletingNode
		operationName = "RemoveNodeFromCluster"
	}

	if cluster.Mock {
		operationName = "Mock" + operationName
	}

	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.RunByUser = runByUser
	// set current operation id in order to recover it from error stat in future

	if cluster.CurrentOperation == nil {
		cluster.CurrentOperation = map[string]byte{}
	}

	cluster.CurrentOperation[operation.Id] = 0
	cluster.Status = initState

	// add operation id to it`s related operations
	cluster.ClusterOperationIDs = append(cluster.ClusterOperationIDs, operation.Id)

	operation.OperationType = constants.OperationTypeAddOrRemoveNodesToCluster
	// save request parameters in operation
	operation.RequestParameter.BodyParameters = nodesToAddOrRemove
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"operationId": operation.Id,
	}
	operation.RequestParameter.RuntimeParameters = map[string]interface{}{
		"addOrRemove": addOrRemove,
	}

	if addOrRemove == constants.ActionCreate {
		// during the add node phase
		// we lock down node immediately

		nodeRoleList := map[string]int{}
		tagNodeRole(nodeRoleList, nodesToAddOrRemove, constants.NodeRoleWorker)

		var nodeToClusterRelation []string

		for nodeId, role := range nodeRoleList {
			nodeToClusterRelation = append(nodeToClusterRelation, nodeId)
			nodeData, errGetNodeInfo := runtimeCache.GetNodeInformation(nodeId)
			if errGetNodeInfo != nil {
				log.Errorf("Failed to retrieve node information due to error: %s", errGetNodeInfo.Error())
				return errGetNodeInfo
			}
			nodeData.Role = role
			nodeData.BelongsToCluster = cluster.ClusterId
			nodeData.KubeNodeStat = constants.UnAvailable
			if err := saveNodeStatus(runtimeCache, *nodeData); err != nil {
				return err
			}
		}

		// save cluster node relation ship
		if err := saveRelationshipStatus(runtimeCache, cluster.ClusterId, nodeToClusterRelation); err != nil {
			return err
		}
	} else {
		// during the remove node phase
		// we do not unlock node immediately by set role to 0 only remove the relation with node only for data consistency reason
		for _, node := range nodesToAddOrRemove {
			nodeData, errGetNodeInfo := runtimeCache.GetNodeInformation(node.NodeId)
			if errGetNodeInfo != nil {
				log.Errorf("Failed to retrieve node information due to error: %s", errGetNodeInfo.Error())
				return errGetNodeInfo
			}
			nodeData.KubeNodeStat = ""
			if err := saveNodeStatus(runtimeCache, *nodeData); err != nil {
				return err
			}
		}
	}

	sumLicenseLabel, err := license.SumLicenseLabel()
	nodeInfoCollection, errGetNode := runtimeCache.GetNodeInformationCollection()
	if errGetNode != nil {
		return errGetNode
	}

	clusterNodeTaskBreakDown := taskBreaker.ClusterNodeTaskBreakDown{
		Operation:          operation,
		Cluster:            cluster,
		Config:             runtimeCache.GetServerRuntimeConfig(cache.NodeId),
		NodesToAddOrRemove: nodesToAddOrRemove,
		NodeCollectionList: nodeInfoCollection,
		Action:             addOrRemove,
		SumLicenseLabel:    sumLicenseLabel,
	}

	operation, err = clusterNodeTaskBreakDown.BreakDownTask()
	if err != nil {
		log.Errorf("Failed to break down task due to error %s", err.Error())
		return err
	}
	// if mock
	if cluster.Mock {
		if cluster.ClusterInstaller == constants.ClusterInstallerRancher {
			// always reset the operation step to nothing but wait 1 seconds
			rancherOp, err := clusterNodeTaskBreakDown.BreakDownRancherTask()
			if err != nil {
				log.Errorf("Failed to rancher break down task due to error %s", err.Error())
				return err
			}
			operation.Step = rancherOp.Step
		} else {
			// always reset the operation step to nothing but wait 1 seconds
			mockStep := taskBreaker.CreateWaitStep(1, "MockStep")
			operation.Step = map[int]*schema.Step{
				0: &mockStep,
			}
		}
	}

	// put operation in db and cache
	if err := saveOperationStatus(runtimeCache, operation); err != nil {
		return err
	}

	// put cluster in db and cache
	if err := saveClusterStatus(runtimeCache, cluster); err != nil {
		return err
	}
	go doTask(&operation, &cluster, &nodeInfoCollection, addOrRemoveNodeDoneHandlerCreator(addOrRemove, nodesToAddOrRemove), clusterOperationIgnoreErrorHandler)
	return nil
}

func recordDebugLogToOperationLog(operation *schema.Operation, msg string) {
	log.Debug(msg)
	operation.Logs = append(operation.Logs, util.LogStyleMessage("debug", msg))
}

func recordErrorLogToOperationLog(operation *schema.Operation, msg string) {
	log.Error(msg)
	operation.Logs = append(operation.Logs, util.LogStyleMessage("error", msg))
}

func saveOperationStatus(runtimeCache cache.ICache, operation schema.Operation) error {
	if err := runtimeCache.SaveOrUpdateOperationCollection(operation.Id, operation); err != nil {
		log.Errorf("Failed to set operation status to %s due to error %s", constants.StatusSuccessful, err.Error())
		return err
	}
	return nil
}

func saveClusterStatus(runtimeCache cache.ICache, cluster schema.Cluster) error {
	if err := runtimeCache.SaveOrUpdateClusterCollection(cluster.ClusterId, cluster); err != nil {
		log.Errorf("Failed to set cluster status to %s due to error %s", constants.StatusSuccessful, err.Error())
		return err
	}
	return nil
}

func saveNodeStatus(runtimeCache cache.ICache, node schema.NodeInformation) error {
	if err := runtimeCache.SaveOrUpdateNodeInformation(node.Id, node); err != nil {
		log.Errorf("Failed to set node status to %s due to error %s", constants.StatusSuccessful, err.Error())
		return err
	}
	return nil
}

func saveRelationshipStatus(runtimeCache cache.ICache, clusterId string, nodes []string) error {
	if err := runtimeCache.AddClusterNodeRelationShip(clusterId, nodes); err != nil {
		log.Errorf("Failed to build cluster and nodes relationship due to error %s", err.Error())
		return err
	}
	return nil
}

func cleanNodeStat(runtimeCache cache.ICache, nodeId string) error {
	node, errGetNodeInfo := runtimeCache.GetNodeInformation(nodeId)
	if errGetNodeInfo != nil {
		log.Errorf("Failed to retrieve node information due to error: %s", errGetNodeInfo.Error())
		return errGetNodeInfo
	}
	node.BelongsToCluster = ""
	node.Role = 0
	return saveNodeStatus(runtimeCache, *node)
}

func releaseClusterNode(runtimeCache cache.ICache, cluster *schema.Cluster) error {
	for _, master := range cluster.Masters {
		if err := cleanNodeStat(runtimeCache, master.NodeId); err != nil {
			return err
		}
	}

	for _, worker := range cluster.Workers {
		if err := cleanNodeStat(runtimeCache, worker.NodeId); err != nil {
			return err
		}
	}

	if cluster.ClusterLB != nil {
		for _, lb := range cluster.ClusterLB.Nodes {
			if err := cleanNodeStat(runtimeCache, lb.NodeId); err != nil {
				return err
			}
		}
	}

	if cluster.Ingress != nil {
		for _, ingress := range cluster.Ingress.NodeIds {
			if err := cleanNodeStat(runtimeCache, ingress.NodeId); err != nil {
				return err
			}
		}
	}

	return nil
}

func createOperationFromCluster(operationName string, cluster schema.Cluster, config serverConfig.Config, nodeInformationCollection schema.NodeInformationCollection, initStat string, sumLicenseLabel uint16) (schema.Operation, error) {
	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), initStat)

	clusterTaskBreakDown := taskBreaker.ClusterTaskBreakDown{
		Operation:          operation,
		Cluster:            cluster,
		Config:             config,
		NodeCollectionList: nodeInformationCollection,
		SumLicenseLabel:    sumLicenseLabel,
	}

	return clusterTaskBreakDown.BreakDownTask()
}

func createOperation(operationName, clusterId, operationId, initStatus string) schema.Operation {
	now := time.Now().Format("2006-01-02T15:04:05Z07:00")
	operation := schema.Operation{
		Id:        "operation-" + operationId,
		ClusterId: clusterId,
		Name:      operationName,
		Status:    initStatus,
		Created:   now,
		LastRun:   now,
	}
	return operation
}

func CreateOperation(operationName, clusterId, operationId, initStatus string) schema.Operation {
	return createOperation(operationName, clusterId, operationId, initStatus)
}

func tagNodeRole(nodeRoleList map[string]int, nodes []schema.ClusterNode, role int) {
	for _, node := range nodes {
		if _, exists := nodeRoleList[node.NodeId]; exists {
			nodeRoleList[node.NodeId] |= role
		} else {
			nodeRoleList[node.NodeId] = role
		}
	}
}

func saveMasterWorkerKubeStat(cluster schema.Cluster, kubeStat string) error {
	runtimeCache := cache.GetCurrentCache()
	reducedNode := map[string]int{}
	for _, master := range cluster.Masters {
		reducedNode[master.NodeId] = 0
	}
	for _, worker := range cluster.Workers {
		reducedNode[worker.NodeId] = 0
	}
	for nodeId := range reducedNode {
		nodeData, errGetNodeInfo := runtimeCache.GetNodeInformation(nodeId)
		if errGetNodeInfo != nil {
			log.Errorf("Failed to retrieve node information due to error: %s", errGetNodeInfo.Error())
			return errGetNodeInfo
		}
		nodeData.KubeNodeStat = kubeStat
		if err := saveNodeStatus(runtimeCache, *nodeData); err != nil {
			return err
		}
	}
	return nil
}

func clusterOperationDoneHandler(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache) error {
	if cluster.Action == constants.ActionCreate {
		cluster.Status = constants.ClusterStatusRunning
		return nil
	} else {
		cluster.Status = constants.ClusterStatusDestroyed
		if err := releaseClusterNode(runtimeCache, cluster); err != nil {
			return err
		}
		if err := runtimeCache.RemoveAllClusterNodeRelationShip(cluster.ClusterId); err != nil {
			return err
		}

	}
	return nil
}

func addOrRemoveNodeDoneHandlerCreator(action string, nodesToRemove []schema.ClusterNode) func(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache) error {
	return func(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache) error {
		cluster.Status = constants.ClusterStatusRunning
		nodesToRemoveList := []string{}
		if action == constants.ActionDelete {
			cluster.Workers = RemoveNodeFromNodeList(cluster.Workers, nodesToRemove)
			for _, node := range nodesToRemove {
				nodesToRemoveList = append(nodesToRemoveList, node.NodeId)
				if err := cleanNodeStat(runtimeCache, node.NodeId); err != nil {
					return err
				}
			}
			if err := runtimeCache.RemoveClusterNodeRelationShip(cluster.ClusterId, nodesToRemoveList); err != nil {
				log.Errorf("Failed remove cluster node relationship due to error %s", err.Error())
				return err
			}
		} else if action == constants.ActionCreate {
			cluster.Workers = append(cluster.Workers, nodesToRemove...)
		}
		return nil
	}
}

func clusterOperationErrorHandler(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache, err error) error {
	cluster.Status = constants.StatusError
	return saveClusterStatus(runtimeCache, *cluster)
}

func clusterOperationIgnoreErrorHandler(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache, err error) error {
	cluster.Status = constants.ClusterStatusRunning
	return saveClusterStatus(runtimeCache, *cluster)
}

func RemoveNodeFromNodeList(nodes []schema.ClusterNode, nodesToRemove []schema.ClusterNode) []schema.ClusterNode {
	remainList := nodes
	for _, nodeToRemove := range nodesToRemove {
		for index, node := range nodes {
			if nodeToRemove.NodeId == node.NodeId {
				remainList[index] = remainList[len(remainList)-1]    // Copy last element to index i.
				remainList[len(remainList)-1] = schema.ClusterNode{} // Erase last element (write zero value).
				remainList = remainList[:len(remainList)-1]          // Truncate slice.
			}
		}
	}
	return remainList
}

func SetReclaimNamespaces(cluster *schema.Cluster) {

	// always clean ReclaimNamespaces when create cluster
	if cluster.EFK != nil {
		if cluster.EFK.Enable {
			cluster.ReclaimNamespaces = append(cluster.ReclaimNamespaces, cluster.EFK.Namespace)
		}
	}
	if cluster.GAP != nil {
		if cluster.GAP.Enable {
			cluster.ReclaimNamespaces = append(cluster.ReclaimNamespaces, cluster.GAP.GetNamespace())
		}
	}
	if cluster.MiddlePlatform != nil {
		if cluster.MiddlePlatform.Enable {
			cluster.ReclaimNamespaces = append(cluster.ReclaimNamespaces, cluster.MiddlePlatform.GetNamespace())
		}
	}
	if cluster.PostgresOperator != nil {
		if cluster.PostgresOperator.Enable {
			cluster.ReclaimNamespaces = append(cluster.ReclaimNamespaces, cluster.PostgresOperator.GetNamespace())
			for _, ns := range cluster.PostgresOperator.GetWatchNamespaces() {
				cluster.ReclaimNamespaces = append(cluster.ReclaimNamespaces, ns)
			}
		}
	}
}

func applyUpgradeTaskDoneHandler(plan *schema.UpgradePlan, upgradeVersion *schema.UpgradableVersion) func(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache) error {
	return func(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache) error {
		cluster.Status = constants.ClusterStatusRunning
		plan.Status = constants.StatusSuccessful
		cluster.ControlPlane.KubernetesVersion = plan.TargetVersion
		cluster.AdditionalVersionDep = upgradeVersion.ToDep()
		if err := runtimeCache.CreateOrUpdateUpgradePlan(*plan); err != nil {
			log.Errorf("Failed to save upgrade plan due to error: %s", err.Error())
			return err
		}
		return nil
	}
}

func applyUpgradeTaskErrorHandler(plan *schema.UpgradePlan) func(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache, err error) error {
	// if upgrade failed we need consider it`s critical issue
	// mark cluster status to error prevent further action from platform
	return func(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache, err error) error {
		cluster.Status = constants.StatusError
		plan.Status = constants.StatusError
		if err := runtimeCache.CreateOrUpdateUpgradePlan(*plan); err != nil {
			log.Errorf("Failed to save upgrade plan due to error: %s", err.Error())
			return err
		}
		if err := runtimeCache.SaveOrUpdateClusterCollection(cluster.ClusterId, *cluster); err != nil {
			log.Errorf("Failed to save cluster due to error: %s", err.Error())
			return err
		}
		return nil
	}
}

func BackupCreate(v schema.Backup, cluster *schema.Cluster, runByUser string) error {
	runtimeCache := cache.GetCurrentCache()
	operationName := "SetupClusterBackupCreate"

	masterNode := &schema.NodeInformation{}
	var err error
	for _, val := range cluster.Masters {
		if CheckNodeReachable(val.NodeId) == nil {
			masterNode, err = runtimeCache.GetNodeInformation(val.NodeId)
			if err != nil {
				return err
			}
			if masterNode != nil {
				break
			}
		}
	}
	if masterNode == nil {
		return errors.New("master node not exist or all dead")
	}

	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeClusterBackupOrRestore
	operation.RunByUser = runByUser
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"clusterId": cluster.ClusterId,
	}
	operation.Step = make(map[int]*schema.Step)
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:   "SetupClusterBackupCreate-" + operation.Id,
		Name: "SetupClusterBackupCreate",
		NodeSteps: []schema.NodeStep{
			taskBreaker.CommonRunWithSingleNode(
				masterNode.Id, "BackupAndRestore", "", 100000000000,
				br.NewBackupCmd(v.Action,
					fmt.Sprintf("%s%s", v.BackupName, time.Now().Format("2006-01-02-15.04.05")),
					v.Output, v.Args).ActionCmd(), true)},
	}

	cluster.CurrentOperation = map[string]byte{}
	// set current operation id in order to recover it from error stat in future
	cluster.CurrentOperation[operation.Id] = 0

	// add operation id to it`s related operations
	cluster.ClusterOperationIDs = append(cluster.ClusterOperationIDs, operation.Id)
	// put cluster in db and cache
	if err := saveClusterStatus(runtimeCache, *cluster); err != nil {
		return err
	}

	// put operation in db and cache
	if err := saveOperationStatus(runtimeCache, operation); err != nil {
		return err
	}

	go doTask(&operation, cluster, &schema.NodeInformationCollection{masterNode.Id: *masterNode}, backupRestoreTaskDoneHandler, backupRestoreTaskErrorHandler)

	return nil
}

func BackupCreateWait(v schema.Backup, cluster *schema.Cluster, runByUser string) error {
	runtimeCache := cache.GetCurrentCache()
	operationName := "SetupClusterUpgradeBackupCreate"

	masterNode := &schema.NodeInformation{}
	var err error
	for _, val := range cluster.Masters {
		if CheckNodeReachable(val.NodeId) == nil {
			masterNode, err = runtimeCache.GetNodeInformation(val.NodeId)
			if err != nil {
				return err
			}
			if masterNode != nil {
				break
			}
		}
	}
	if masterNode == nil {
		return errors.New("master node not exist or all dead")
	}

	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeClusterBackupOrRestore
	operation.RunByUser = runByUser
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"clusterId": cluster.ClusterId,
	}
	operation.Step = make(map[int]*schema.Step)
	resp := make(chan string, 1)
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:   "SetupClusterUpgradeBackupCreate-" + operation.Id,
		Name: "SetupClusterUpgradeBackupCreate",
		NodeSteps: []schema.NodeStep{
			taskBreaker.CommonRunWithSingleNode(
				masterNode.Id, "BackupAndRestore", "", 100000000000,
				br.NewBackupCmd(v.Action,
					fmt.Sprintf("%s%s", v.BackupName, time.Now().Format("2006-01-02-15.04.05")),
					v.Output, v.Args).ActionCmd(), true)},
		OnStepDoneOrErrorHandler: taskBreaker.OnErrorIgnoreHandler{
			OnAllDoneCustomerHandler: func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				if v, ok := returnData["0"]; ok {
					resp <- v
				} else {
					resp <- reply.Message
				}
			},
		},
	}

	cluster.CurrentOperation = map[string]byte{}
	// set current operation id in order to recover it from error stat in future
	cluster.CurrentOperation[operation.Id] = 0

	// add operation id to it`s related operations
	cluster.ClusterOperationIDs = append(cluster.ClusterOperationIDs, operation.Id)
	// put cluster in db and cache
	if err := saveClusterStatus(runtimeCache, *cluster); err != nil {
		return err
	}

	// put operation in db and cache
	if err := saveOperationStatus(runtimeCache, operation); err != nil {
		return err
	}

	go doTask(&operation, cluster, &schema.NodeInformationCollection{masterNode.Id: *masterNode}, backupRestoreTaskDoneHandler, backupRestoreTaskErrorHandler)

	return veleroError(<-resp)
}

func veleroError(str string) error {
	i := strings.Index(str, "An error occurred")

	if i != -1 {
		r := []rune(str)
		r = r[i:]
		return errors.New(string(r))
	}
	return nil
}

func BackupDelete(v schema.Backup, cluster *schema.Cluster, runByUser string) error {
	runtimeCache := cache.GetCurrentCache()
	operationName := "SetupClusterBackUpDelete"

	masterNode := &schema.NodeInformation{}
	var err error
	for _, val := range cluster.Masters {
		if CheckNodeReachable(val.NodeId) == nil {
			masterNode, err = runtimeCache.GetNodeInformation(val.NodeId)
			if err != nil {
				return err
			}
			if masterNode != nil {
				break
			}
		}
	}
	if masterNode == nil {
		return errors.New("master node not exist or all dead")
	}

	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeClusterBackupOrRestore
	operation.RunByUser = runByUser
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"clusterId": cluster.ClusterId,
	}
	operation.Step = make(map[int]*schema.Step)
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:   "SetupClusterBackUpDelete-" + operation.Id,
		Name: "SetupClusterBackUpDelete",
		NodeSteps: []schema.NodeStep{
			taskBreaker.CommonRunWithSingleNode(
				masterNode.Id, "BackupAndRestore", "", 100000000000,
				br.NewBackupCmd(v.Action, v.BackupName, v.Output, v.Args).ActionCmd(), true)},
	}

	go doTask(&operation, cluster, &schema.NodeInformationCollection{masterNode.Id: *masterNode}, backupRestoreTaskDoneHandler, backupRestoreTaskErrorHandler)

	return nil
}

func BackupList(v schema.Backup, cluster *schema.Cluster, runByUser string) (string, error) {
	runtimeCache := cache.GetCurrentCache()
	operationName := "SetupBackUpList"

	masterNode := &schema.NodeInformation{}
	var err error
	for _, val := range cluster.Masters {
		if CheckNodeReachable(val.NodeId) == nil {
			masterNode, err = runtimeCache.GetNodeInformation(val.NodeId)
			if err != nil {
				return "", err
			}
			if masterNode != nil {
				break
			}
		}
	}
	if masterNode == nil {
		return "", errors.New("master node not exist or all dead")
	}

	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeClusterBackupOrRestore
	operation.RunByUser = runByUser
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"clusterId": cluster.ClusterId,
	}
	operation.Step = make(map[int]*schema.Step)
	resp := make(chan string, 1)
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:   "SetupBackUpList-" + operation.Id,
		Name: "SetupBackUpList",
		OnStepDoneOrErrorHandler: taskBreaker.OnErrorIgnoreHandler{
			OnAllDoneCustomerHandler: func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				if v, ok := returnData["0"]; ok {
					resp <- v
				} else {
					resp <- reply.Message
				}
			},
		},
		NodeSteps: []schema.NodeStep{
			taskBreaker.CommonRunWithSingleNode(
				masterNode.Id, "BackupAndRestore", "", 5,
				br.NewBackupCmd(v.Action, v.BackupName, v.Output, v.Args).ActionCmd(), true)},
	}

	go doTask(&operation, cluster, &schema.NodeInformationCollection{masterNode.Id: *masterNode}, nil, nil)
	select {
	case result := <-resp:
		return result, veleroError(result)
	case <-time.After(5 * time.Second):
		return "", errors.New("get backup data time out, please retry")
	}
}

func RestoreCreate(v schema.Restore, cluster *schema.Cluster, runByUser string) error {
	runtimeCache := cache.GetCurrentCache()
	operationName := "SetupClusterRestoreCreate"

	masterNode := &schema.NodeInformation{}
	var err error
	for _, val := range cluster.Masters {
		if CheckNodeReachable(val.NodeId) == nil {
			masterNode, err = runtimeCache.GetNodeInformation(val.NodeId)
			if err != nil {
				return err
			}
			if masterNode != nil {
				break
			}
		}
	}
	if masterNode == nil {
		return errors.New("master node not exist or all dead")
	}

	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeClusterBackupOrRestore
	operation.RunByUser = runByUser
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"clusterId": cluster.ClusterId,
	}
	operation.Step = make(map[int]*schema.Step)
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:   "SetupClusterRestoreCreate-" + operation.Id,
		Name: "SetupClusterRestoreCreate",
		NodeSteps: []schema.NodeStep{
			taskBreaker.CommonRunWithSingleNode(
				masterNode.Id, "BackupAndRestore", "", 100000000000,
				br.NewRestoreCmd(v.Action, v.BackupName, v.RestoreName, v.Args).ActionCmd(), true)},
	}

	cluster.CurrentOperation = map[string]byte{}
	// set current operation id in order to recover it from error stat in future
	cluster.CurrentOperation[operation.Id] = 0

	// add operation id to it`s related operations
	cluster.ClusterOperationIDs = append(cluster.ClusterOperationIDs, operation.Id)
	// put cluster in db and cache
	if err := saveClusterStatus(runtimeCache, *cluster); err != nil {
		return err
	}

	// put operation in db and cache
	if err := saveOperationStatus(runtimeCache, operation); err != nil {
		return err
	}

	go doTask(&operation, cluster, &schema.NodeInformationCollection{masterNode.Id: *masterNode}, backupRestoreTaskDoneHandler, backupRestoreTaskErrorHandler)

	return nil
}

func RestoreList(v schema.Restore, cluster *schema.Cluster, runByUser string) (string, error) {
	runtimeCache := cache.GetCurrentCache()
	operationName := "SetupClusterRestoreList"

	masterNode := &schema.NodeInformation{}
	var err error
	for _, val := range cluster.Masters {
		if CheckNodeReachable(val.NodeId) == nil {
			masterNode, err = runtimeCache.GetNodeInformation(val.NodeId)
			if err != nil {
				return "", err
			}
			if masterNode != nil {
				break
			}
		}
	}
	if masterNode == nil {
		return "", errors.New("master node not exist or all dead")
	}

	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeClusterBackupOrRestore
	operation.RunByUser = runByUser
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"clusterId": cluster.ClusterId,
	}
	operation.Step = make(map[int]*schema.Step)
	resp := make(chan string, 1)
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:   "SetupClusterRestoreList-" + operation.Id,
		Name: "SetupClusterRestoreList",
		OnStepDoneOrErrorHandler: taskBreaker.OnErrorIgnoreHandler{
			OnAllDoneCustomerHandler: func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				if v, ok := returnData["0"]; ok {
					resp <- v
				} else {
					resp <- reply.Message
				}
			},
		},
		NodeSteps: []schema.NodeStep{
			taskBreaker.CommonRunWithSingleNode(
				masterNode.Id, "BackupAndRestore", "", 100000000000,
				br.NewRestoreCmd(v.Action, v.BackupName, v.RestoreName, v.Args).ActionCmd(), true)},
	}

	cluster.CurrentOperation = map[string]byte{}
	// set current operation id in order to recover it from error stat in future
	cluster.CurrentOperation[operation.Id] = 0

	// add operation id to it`s related operations
	cluster.ClusterOperationIDs = append(cluster.ClusterOperationIDs, operation.Id)

	go doTask(&operation, cluster, &schema.NodeInformationCollection{masterNode.Id: *masterNode}, nil, nil)
	select {
	case result := <-resp:
		return result, veleroError(result)
	case <-time.After(5 * time.Second):
		return "", errors.New("get backup data time out, please retry")
	}
}

func RestoreDelete(v schema.Restore, cluster *schema.Cluster, runByUser string) error {
	runtimeCache := cache.GetCurrentCache()
	operationName := "SetupClusterRestoreDelete"

	masterNode := &schema.NodeInformation{}
	var err error
	for _, val := range cluster.Masters {
		if CheckNodeReachable(val.NodeId) == nil {
			masterNode, err = runtimeCache.GetNodeInformation(val.NodeId)
			if err != nil {
				return err
			}
			if masterNode != nil {
				break
			}
		}
	}
	if masterNode == nil {
		return errors.New("master node not exist or all dead")
	}

	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeClusterBackupOrRestore
	operation.RunByUser = runByUser
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"clusterId": cluster.ClusterId,
	}
	operation.Step = make(map[int]*schema.Step)
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:   "SetupClusterRestoreDelete-" + operation.Id,
		Name: "SetupClusterRestoreDelete",
		NodeSteps: []schema.NodeStep{
			taskBreaker.CommonRunWithSingleNode(
				masterNode.Id, "BackupAndRestore", "", 100000000000,
				br.NewRestoreCmd(v.Action, v.BackupName, v.RestoreName, v.Args).ActionCmd(), true)},
	}
	go doTask(&operation, cluster, &schema.NodeInformationCollection{masterNode.Id: *masterNode}, backupRestoreTaskDoneHandler, backupRestoreTaskErrorHandler)

	return nil
}

func BackupRegularCreate(v schema.BackupRegular, cluster *schema.Cluster, runByUser string) error {
	runtimeCache := cache.GetCurrentCache()
	operationName := "SetupClusterBackupRegularCreate"

	masterNode := &schema.NodeInformation{}
	var err error
	for _, val := range cluster.Masters {
		if CheckNodeReachable(val.NodeId) == nil {
			masterNode, err = runtimeCache.GetNodeInformation(val.NodeId)
			if err != nil {
				return err
			}
			if masterNode != nil {
				break
			}
		}
	}
	if masterNode == nil {
		return errors.New("master node not exist or all dead")
	}

	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeClusterBackupOrRestore
	operation.RunByUser = runByUser
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"clusterId": cluster.ClusterId,
	}

	if cluster.ContainerRuntime.PrivateRegistryAddress != "" && cluster.ContainerRuntime.PrivateRegistryPort != 0 {
		v.Registry = fmt.Sprintf("%s:%d", cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort)
	}

	step, err := taskBreaker.BackupCronJobCreate(masterNode.Id, &v, 10)
	if err != nil {
		return err
	}

	operation.Step = make(map[int]*schema.Step)
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:        "SetupClusterBackupRegularCreate-" + operation.Id,
		Name:      "SetupClusterBackupRegularCreate",
		NodeSteps: []schema.NodeStep{*step},
	}

	cluster.CurrentOperation = map[string]byte{}
	// set current operation id in order to recover it from error stat in future
	cluster.CurrentOperation[operation.Id] = 0

	// add operation id to it`s related operations
	cluster.ClusterOperationIDs = append(cluster.ClusterOperationIDs, operation.Id)
	// put cluster in db and cache
	if err := saveClusterStatus(runtimeCache, *cluster); err != nil {
		return err
	}

	// put operation in db and cache
	if err := saveOperationStatus(runtimeCache, operation); err != nil {
		return err
	}

	go doTask(&operation, cluster, &schema.NodeInformationCollection{masterNode.Id: *masterNode}, backupRestoreTaskDoneHandler, backupRestoreTaskErrorHandler)

	return nil
}

func BackupRegularDelete(backupRegularName string, cluster *schema.Cluster, runByUser string) error {
	runtimeCache := cache.GetCurrentCache()
	operationName := "SetupClusterBackupRegularDelete"

	masterNode := &schema.NodeInformation{}
	var err error
	for _, val := range cluster.Masters {
		if CheckNodeReachable(val.NodeId) == nil {
			masterNode, err = runtimeCache.GetNodeInformation(val.NodeId)
			if err != nil {
				return err
			}
			if masterNode != nil {
				break
			}
		}
	}
	if masterNode == nil {
		return errors.New("master node not exist or all dead")
	}

	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeClusterBackupOrRestore
	operation.RunByUser = runByUser
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"clusterId": cluster.ClusterId,
	}

	operation.Step = make(map[int]*schema.Step)
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:   "SetupClusterBackupRegularDelete-" + operation.Id,
		Name: "SetupClusterBackupRegularDelete",
		NodeSteps: []schema.NodeStep{
			taskBreaker.CommonRunWithSingleNode(
				masterNode.Id, "BackupAndRestore", "", 100000000000,
				[]string{"kubectl", "delete", "cronjob", backupRegularName, "-n", "velero"}, false),
		},
	}

	cluster.CurrentOperation = map[string]byte{}
	// set current operation id in order to recover it from error stat in future
	cluster.CurrentOperation[operation.Id] = 0

	// add operation id to it`s related operations
	cluster.ClusterOperationIDs = append(cluster.ClusterOperationIDs, operation.Id)
	// put cluster in db and cache
	if err := saveClusterStatus(runtimeCache, *cluster); err != nil {
		return err
	}

	// put operation in db and cache
	if err := saveOperationStatus(runtimeCache, operation); err != nil {
		return err
	}

	go doTask(&operation, cluster, &schema.NodeInformationCollection{masterNode.Id: *masterNode}, nil, nil)

	return nil
}

func BackupRegularGet(backupRegularName string, cluster *schema.Cluster, runByUser string) (string, error) {
	runtimeCache := cache.GetCurrentCache()
	operationName := "SetupClusterBackupRegularGet"

	masterNode := &schema.NodeInformation{}
	var err error
	for _, val := range cluster.Masters {
		if CheckNodeReachable(val.NodeId) == nil {
			masterNode, err = runtimeCache.GetNodeInformation(val.NodeId)
			if err != nil {
				return "", err
			}
			if masterNode != nil {
				break
			}
		}
	}
	if masterNode == nil {
		return "", errors.New("master node not exist or all dead")
	}

	operation := createOperation(operationName, cluster.ClusterId, uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeClusterBackupOrRestore
	operation.RunByUser = runByUser
	operation.RequestParameter.QueryParameters = map[string]interface{}{
		"clusterId": cluster.ClusterId,
	}

	operation.Step = make(map[int]*schema.Step)
	resp := make(chan string, 1)
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:   "SetupClusterBackupRegularGet-" + operation.Id,
		Name: "SetupClusterBackupRegularGet",
		OnStepDoneOrErrorHandler: taskBreaker.OnErrorIgnoreHandler{
			OnAllDoneCustomerHandler: func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				if v, ok := returnData["0"]; ok {
					resp <- v
				} else {
					resp <- reply.Message
				}
			},
		},
		NodeSteps: []schema.NodeStep{
			taskBreaker.CommonRunWithSingleNode(
				masterNode.Id, "BackupAndRestore", "", 100000000000,
				[]string{"kubectl", "get", "cronjob", backupRegularName, "-n", "velero", "-o", "json"},
				true),
		},
	}

	cluster.CurrentOperation = map[string]byte{}
	// set current operation id in order to recover it from error stat in future
	cluster.CurrentOperation[operation.Id] = 0

	// add operation id to it`s related operations
	cluster.ClusterOperationIDs = append(cluster.ClusterOperationIDs, operation.Id)
	// put cluster in db and cache
	if err := saveClusterStatus(runtimeCache, *cluster); err != nil {
		return "", err
	}

	// put operation in db and cache
	if err := saveOperationStatus(runtimeCache, operation); err != nil {
		return "", err
	}

	go doTask(&operation, cluster, &schema.NodeInformationCollection{masterNode.Id: *masterNode}, nil, nil)

	select {
	case result := <-resp:
		return result, nil
	case <-time.After(5 * time.Second):
		return "", errors.New("get backup data time out, please retry")
	}
}

func backupRestoreTaskDoneHandler(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache) error {
	cluster.Status = constants.ClusterStatusRunning
	if err := saveOperationStatus(runtimeCache, *operation); err != nil {
		log.Errorf("Failed to record error data on all step is done due to error %s", err.Error())
	}

	if err := saveClusterStatus(runtimeCache, *cluster); err != nil {
		log.Errorf("Failed to save cluster data to db when all step is done due to error %s", err.Error())
	}
	return nil
}

func backupRestoreTaskErrorHandler(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache, err error) error {
	cluster.Status = constants.ClusterStatusRunning
	if err := saveOperationStatus(runtimeCache, *operation); err != nil {
		log.Errorf("Failed to record error data on all step is done due to error %s", err.Error())
	}

	if err := saveClusterStatus(runtimeCache, *cluster); err != nil {
		log.Errorf("Failed to save cluster data to db when all step is done due to error %s", err.Error())
	}
	return nil
}
