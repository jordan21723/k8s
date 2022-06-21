package task_breaker

import (
	"fmt"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util"
	"k8s-installer/schema"
)

type OnErrorAbortHandler struct {
	OnStepReturnCustomerHandler func(reply schema.QueueReply,
		step *schema.Step, taskDoneSignal chan int,
		returnData map[string]string,
		cluster *schema.Cluster,
		operation *schema.Operation,
		nodeCollection *schema.NodeInformationCollection)
	OnAllDoneCustomerHandler func(reply schema.QueueReply,
		step *schema.Step, taskDoneSignal chan int,
		returnData map[string]string,
		cluster *schema.Cluster,
		operation *schema.Operation,
		nodeCollection *schema.NodeInformationCollection)
}

func (OnErrorAbortHandler OnErrorAbortHandler) OnStepDoneOrErrorHandler(reply schema.QueueReply,
	step *schema.Step, taskDoneSignal chan int,
	returnData map[string]string,
	cluster *schema.Cluster,
	operation *schema.Operation,
	nodeCollection *schema.NodeInformationCollection,
	saveClusterToDB func(cluster schema.Cluster) error,
	saveOperationToDB func(operation schema.Operation) error) {

	if reply.Stat == constants.StatusSuccessful {
		step.OnSuccessNodes = append(step.OnSuccessNodes, schema.ClusterNode{NodeId: step.NodeSteps[0].NodeID})
	} else {
		nodeIp := "not found"
		nodeFound, found := (*nodeCollection)[reply.NodeId]
		if found {
			nodeIp = nodeFound.Ipv4DefaultIp
		}
		step.OnFailedNodes = append(step.OnFailedNodes, schema.ClusterNode{NodeId: step.NodeSteps[0].NodeID})
		msg := fmt.Sprintf("Node %s in step %s result in error state", nodeIp, step.Id)
		log.Error(msg)
		operation.Logs = append(operation.Logs, util.LogStyleMessage("error", msg))
	}

	if OnErrorAbortHandler.OnStepReturnCustomerHandler != nil {
		OnErrorAbortHandler.OnStepReturnCustomerHandler(reply, step, taskDoneSignal, returnData, cluster, operation, nodeCollection)
	}

	log.Debugf("Got return data %v message %v", reply.ReturnData, reply.Message)

	for key, val := range reply.ReturnData {
		returnData[key] = val
	}

	if len(step.OnFailedNodes) > 0 {
		taskDoneSignal <- constants.MessageSignalError
		return
	}
	if len(step.NodeSteps) == len(step.OnSuccessNodes) {
		// if all reply done
		// say let`s move to next task
		if OnErrorAbortHandler.OnAllDoneCustomerHandler != nil {
			OnErrorAbortHandler.OnAllDoneCustomerHandler(reply, step, taskDoneSignal, returnData, cluster, operation, nodeCollection)
		}
		taskDoneSignal <- constants.MessageSignalDone
	}
}

type OnErrorIgnoreHandler struct {
	OnStepReturnCustomerHandler func(reply schema.QueueReply,
		step *schema.Step, taskDoneSignal chan int,
		returnData map[string]string,
		cluster *schema.Cluster,
		operation *schema.Operation,
		nodeCollection *schema.NodeInformationCollection)
	OnAllDoneCustomerHandler func(reply schema.QueueReply,
		step *schema.Step, taskDoneSignal chan int,
		returnData map[string]string,
		cluster *schema.Cluster,
		operation *schema.Operation,
		nodeCollection *schema.NodeInformationCollection)
}

func (OnErrorIgnoreHandler OnErrorIgnoreHandler) OnStepDoneOrErrorHandler(reply schema.QueueReply,
	step *schema.Step,
	taskDoneSignal chan int,
	returnData map[string]string,
	cluster *schema.Cluster,
	operation *schema.Operation,
	nodeCollection *schema.NodeInformationCollection,
	saveClusterToDB func(cluster schema.Cluster) error,
	saveOperationToDB func(operation schema.Operation) error) {

	// consider all reply is success
	step.OnSuccessNodes = append(step.OnSuccessNodes, schema.ClusterNode{NodeId: step.NodeSteps[0].NodeID})
	if reply.Stat != constants.StatusSuccessful {
		// only to record it
		nodeIp := "not found"
		nodeFound, found := (*nodeCollection)[reply.NodeId]
		if found {
			nodeIp = nodeFound.Ipv4DefaultIp
		}
		msg := fmt.Sprintf("Node %s in step %s result in error state but we are going to ignore it.", nodeIp, step.Id)
		log.Error(msg)
		step.OnFailedNodes = append(step.OnFailedNodes, schema.ClusterNode{NodeId: step.NodeSteps[0].NodeID})
		operation.Logs = append(operation.Logs, util.LogStyleMessage("warning", msg))
	}

	if OnErrorIgnoreHandler.OnStepReturnCustomerHandler != nil {
		OnErrorIgnoreHandler.OnStepReturnCustomerHandler(reply, step, taskDoneSignal, returnData, cluster, operation, nodeCollection)
	}

	log.Debugf("Got return data %v message %v", reply.ReturnData, reply.Message)

	for key, val := range reply.ReturnData {
		returnData[key] = val
	}

	if len(step.NodeSteps) == len(step.OnSuccessNodes) {
		// if all node step is done
		// say let`s move to next task
		if OnErrorIgnoreHandler.OnAllDoneCustomerHandler != nil {
			OnErrorIgnoreHandler.OnAllDoneCustomerHandler(reply, step, taskDoneSignal, returnData, cluster, operation, nodeCollection)
		}
		taskDoneSignal <- constants.MessageSignalDone
	}
}

type OnTimeOutAbortHandler struct {
	OnStepTimeoutCustomerHandler func(step *schema.Step,
		cluster *schema.Cluster,
		operation *schema.Operation,
		nodeStepId string,
		nodeCollection *schema.NodeInformationCollection)
	OnAllDoneCustomerHandler func(step *schema.Step,
		cluster *schema.Cluster,
		operation *schema.Operation,
		nodeStepId string,
		nodeCollection *schema.NodeInformationCollection)
}

func (OnTimeOutAbortHandler OnTimeOutAbortHandler) OnStepTimeOutHandler(step *schema.Step,
	abortWhenCount int,
	taskTimeoutSignal chan string,
	cluster *schema.Cluster,
	operation *schema.Operation,
	nodeCollection *schema.NodeInformationCollection,
	saveClusterToDB func(cluster schema.Cluster) error,
	saveOperationToDB func(operation schema.Operation) error,
	nodeStepId string) {

	nodeIp := "not found"
	nodeFound, found := (*nodeCollection)[step.NodeSteps[0].NodeID]
	if found {
		nodeIp = nodeFound.Ipv4DefaultIp
	}
	msg := fmt.Sprintf("Node %s in step %s result in timeout state.", nodeIp, step.Id)
	log.Error(msg)
	operation.Logs = append(operation.Logs, util.LogStyleMessage("error", msg))
	step.OnTimeoutNodes = append(step.OnTimeoutNodes, schema.ClusterNode{NodeId: step.NodeSteps[0].NodeID})

	if OnTimeOutAbortHandler.OnStepTimeoutCustomerHandler != nil {
		OnTimeOutAbortHandler.OnStepTimeoutCustomerHandler(step, cluster, operation, nodeStepId, nodeCollection)
	}
	// if reach allowed timeout count send signal to time out channel
	if len(step.OnTimeoutNodes) >= abortWhenCount {
		if OnTimeOutAbortHandler.OnAllDoneCustomerHandler != nil {
			OnTimeOutAbortHandler.OnAllDoneCustomerHandler(step, cluster, operation, nodeStepId, nodeCollection)
		}
		taskTimeoutSignal <- "timeout"
	}
}

type OnTimeOutIgnoreHandler struct {
	OnStepTimeoutCustomerHandler func(step *schema.Step,
		cluster *schema.Cluster,
		operation *schema.Operation,
		nodeStepId string,
		nodeCollection *schema.NodeInformationCollection)
	OnAllDoneCustomerHandler func(step *schema.Step,
		cluster *schema.Cluster,
		operation *schema.Operation,
		nodeStepId string,
		nodeCollection *schema.NodeInformationCollection)
}

func (OnTimeOutIgnoreHandler OnTimeOutIgnoreHandler) OnStepTimeOutHandler(step *schema.Step,
	abortWhenCount int,
	taskTimeoutSignal chan string,
	taskDoneSignal chan int,
	cluster *schema.Cluster,
	operation *schema.Operation,
	nodeCollection *schema.NodeInformationCollection,
	saveClusterToDB func(cluster schema.Cluster) error,
	saveOperationToDB func(operation schema.Operation) error,
	nodeStepId string) {

	// do not care reply status consider all reply is success
	nodeIp := "not found"
	nodeFound, found := (*nodeCollection)[step.NodeSteps[0].NodeID]
	if found {
		nodeIp = nodeFound.Ipv4DefaultIp
	}
	msg := fmt.Sprintf("Node %s in step %s result in timeout state but we are going to ignore it.", nodeIp, step.Id)
	log.Error(msg)
	operation.Logs = append(operation.Logs, util.LogStyleMessage("warning", msg))
	step.OnSuccessNodes = append(step.OnSuccessNodes, schema.ClusterNode{NodeId: step.NodeSteps[0].NodeID})
	// only record it
	step.OnTimeoutNodes = append(step.OnTimeoutNodes, schema.ClusterNode{NodeId: step.NodeSteps[0].NodeID})

	if OnTimeOutIgnoreHandler.OnStepTimeoutCustomerHandler != nil {
		OnTimeOutIgnoreHandler.OnStepTimeoutCustomerHandler(step, cluster, operation, nodeStepId, nodeCollection)
	}

	if len(step.NodeSteps) == len(step.OnSuccessNodes) {
		// if all node step is done
		// say let`s move to next task
		if OnTimeOutIgnoreHandler.OnAllDoneCustomerHandler != nil {
			OnTimeOutIgnoreHandler.OnAllDoneCustomerHandler(step, cluster, operation, nodeStepId, nodeCollection)
		}
		taskDoneSignal <- constants.MessageSignalDone
	}
}
