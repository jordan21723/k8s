package control_manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	messageQueue "k8s-installer/pkg/message_queue/nats"
	mqHelper "k8s-installer/pkg/server/message_queue"
	taskBreaker "k8s-installer/pkg/task_breaker"
	"k8s-installer/schema"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

func createTaskCallBackHandler(step *schema.Step, taskDoneSignal chan int, returnData map[string]string, stepIndex int, runtimeCache cache.ICache, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) func(msg *nats.Msg) error {
	// closure to hold input step and chan object to future operation
	return func(msg *nats.Msg) error {
		reply := schema.QueueReply{}
		if err := json.Unmarshal(msg.Data, &reply); err != nil {
			log.Error("Failed to Unmarshal reply message data to QueueReply")
			return err
		} else {
			log.Debugf("Got reply message from node %s with operation id %s with message %v", reply.NodeId, reply.OperationId, reply.Message)
		}
		if step.OnStepDoneOrErrorHandler == nil {
			log.Debugf("Step %d on done or error handler is not set fall back to default handler onErrorAbort", stepIndex)
			OnStepErrorHandler := taskBreaker.OnErrorAbortHandler{}
			OnStepErrorHandler.OnStepDoneOrErrorHandler(reply, step, taskDoneSignal, returnData, cluster, operation, nodeCollection, func(cluster schema.Cluster) error {
				return saveClusterStatus(runtimeCache, cluster)
			}, func(operation schema.Operation) error {
				return saveOperationStatus(runtimeCache, operation)
			})
		} else {
			log.Debugf("Step %d on done or error handler is set use it", stepIndex)
			step.OnStepDoneOrErrorHandler.OnStepDoneOrErrorHandler(reply, step, taskDoneSignal, returnData, cluster, operation, nodeCollection, func(cluster schema.Cluster) error {
				return saveClusterStatus(runtimeCache, cluster)
			}, func(operation schema.Operation) error {
				return saveOperationStatus(runtimeCache, operation)
			})
		}
		return nil
	}
}

func createTaskTimeOutHandler(step *schema.Step, taskTimeoutSignal chan string, taskDoneSignal chan int, abortWhenCount int, stepIndex int, runtimeCache cache.ICache, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) func(nodeId string, nodeStepId string) {
	return func(nodeId string, nodeStepId string) {
		log.Errorf("Step %d running into timeout stat. Leave decision to step time out handler", stepIndex)
		if step.OnStepTimeOutHandler == nil {
			log.Debugf("Step %d on timeout handler not set fall back to default handler onTimeoutAbort", stepIndex)
			onTimeoutAbortHandler := taskBreaker.OnTimeOutAbortHandler{}
			onTimeoutAbortHandler.OnStepTimeOutHandler(step, abortWhenCount, taskTimeoutSignal, cluster, operation, nodeCollection, func(cluster schema.Cluster) error {
				return saveClusterStatus(runtimeCache, cluster)
			}, func(operation schema.Operation) error {
				return saveOperationStatus(runtimeCache, operation)
			}, nodeStepId)
		} else {
			log.Debugf("Step %d on timeout handler is set use it", stepIndex)
			step.OnStepTimeOutHandler.OnStepTimeOutHandler(step, abortWhenCount, taskTimeoutSignal, taskDoneSignal, cluster, operation, nodeCollection, func(cluster schema.Cluster) error {
				return saveClusterStatus(runtimeCache, cluster)
			}, func(operation schema.Operation) error {
				return saveOperationStatus(runtimeCache, operation)
			}, nodeStepId)
		}
	}
}

func combineResourceServer(config server.Config) string {
	protocol := "http"
	if config.ApiServer.EnableTLS {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s:%s", protocol, config.ApiServer.ResourceServerCIDR, strconv.FormatUint(uint64(config.ApiServer.ResourceServerPort), 10))
}

func DoSingleTask(operation *schema.Operation, cluster *schema.Cluster, metaData map[string]string) ([]byte, error) {

	if operation.Step[0].WaitBeforeRun > 0 {
		log.Warnf("Step %d says we should wait for %d seconds before run. So we wait :(", 0, operation.Step[0].WaitBeforeRun)
		time.Sleep(time.Duration(operation.Step[0].WaitBeforeRun) * time.Second)
	}
	if len(operation.Step[0].NodeSteps) == 0 {
		recordDebugLogToOperationLog(operation, fmt.Sprintf("There is no node steps in step %d skip it!", 0))
		return nil, nil
	}
	runtimeCache := cache.GetCurrentCache()
	config := runtimeCache.GetServerRuntimeConfig(cache.NodeId)

	// always send one master that handler this task as resource server
	resourceServerURL := combineResourceServer(config)

	if metaData == nil {
		metaData = map[string]string{}
	}
	msgData, err := createMsgBody(operation.Id, resourceServerURL, operation.Step[0].Id, operation.Step[0].NodeSteps[0].Tasks[0], *cluster, metaData)
	if err != nil {
		SetOperationToFailedStat(operation, runtimeCache, fmt.Sprintf("Failed to marshal data for node task %d of step %d for operation %s due to error %s", 0, 0, operation.Id, err.Error()))
		return nil, err
	}
	nodeData, errGetNodeInfo := runtimeCache.GetNodeInformation(operation.Step[0].NodeSteps[0].NodeID)
	if errGetNodeInfo != nil || nodeData == nil {
		SetOperationToFailedStat(operation, runtimeCache, fmt.Sprintf("Failed to retrieve node information due to error: %s or node not found", errGetNodeInfo.Error()))
		return nil, errGetNodeInfo
	}

	result, errSentMsg := messageQueue.SendingMessageAndWaitReply(mqHelper.GeneratorNodeSubscribe(*nodeData, config.MessageQueue.SubjectSuffix),
		operation.Step[0].NodeSteps[0].NodeID,
		operation.Step[0].NodeSteps[0].Id,
		uuid.New().String(),
		config.MessageQueue,
		time.Duration(operation.Step[0].NodeSteps[0].Tasks[0].GetTaskTimeOut())*time.Second,
		[]byte(msgData))

	if errSentMsg != nil {
		SetOperationToFailedStat(operation, runtimeCache, fmt.Sprintf("Task %s with index %d result in error state...abort!!!!", operation.Step[0].Name, 0))
		return nil, errSentMsg
	} else {
		reply := schema.QueueReply{}
		if err := json.Unmarshal(result, &reply); err != nil {
			log.Error("Failed to Unmarshal reply message data to QueueReply")
			SetOperationToFailedStat(operation, runtimeCache, fmt.Sprintf("Task %s with index %d result in error state...abort!!!!", operation.Step[0].Name, 0))
			return nil, err
		} else {
			log.Debugf("Got reply message from node %s with operation id %s", reply.NodeId, reply.OperationId)
		}
		// check remote result
		if reply.Stat == constants.StatusSuccessful {
			recordDebugLogToOperationLog(operation, fmt.Sprintf("All step done set operation status to %s", constants.StatusSuccessful))
			operation.Status = constants.StatusSuccessful
		} else {
			SetOperationToFailedStat(operation, runtimeCache, fmt.Sprintf("Task %s with index %d result in error state...abort!!!!", operation.Step[0].Name, 0))
			return nil, errors.New(fmt.Sprintf("Got client reply error :%s", reply.Message))
		}
	}

	if err := saveOperationStatus(runtimeCache, *operation); err != nil {
		log.Errorf("Failed to record error data on all step is done due to error %s", err.Error())
		return nil, err
	}
	return result, nil
}

/*
step flow control
*/
func doTask(operation *schema.Operation, cluster *schema.Cluster, nodeCollection *schema.NodeInformationCollection,
	taskCompleteHandler func(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache) error,
	taskErrorHandler func(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache, err error) error) {

	runtimeCache := cache.GetCurrentCache()
	config := runtimeCache.GetServerRuntimeConfig(cache.NodeId)

	// ensure to record which handle this operation
	operation.Host = cache.NodeId

	// always send one master that handler this task as resource server
	resourceServerURL := combineResourceServer(config)

	log.Debugf("Got resourceServerURL %s", resourceServerURL)

	// collect all step return data in this object
	// for those task  consume previously task return data
	var stepReturnData = map[string]string{}

	if len(operation.PreStepReturnData) > 0 {
		stepReturnData = operation.PreStepReturnData
	}
	recordDebugLogToOperationLog(operation, fmt.Sprintf("Operation %s task starts", operation.Id))
	for stepIndex := operation.CurrentStep; stepIndex < len(operation.Step); stepIndex++ {
		if operation.Step[stepIndex].WaitBeforeRun > 0 {
			log.Warnf("Step %d says we should wait for %d seconds before run. So we wait :(", stepIndex, operation.Step[stepIndex].WaitBeforeRun)
			time.Sleep(time.Duration(operation.Step[stepIndex].WaitBeforeRun) * time.Second)
		}
		taskDoneSignal := make(chan int)
		taskTimeOutSignal := make(chan string)
		// use closure to create a message handler
		taskDoneHandler := createTaskCallBackHandler(operation.Step[stepIndex], taskDoneSignal, stepReturnData, stepIndex, runtimeCache, cluster, operation, nodeCollection)
		// use closure to create a timeout handler
		taskTimeOutSignalHandler := createTaskTimeOutHandler(operation.Step[stepIndex], taskTimeOutSignal, taskDoneSignal, 1, stepIndex, runtimeCache, cluster, operation, nodeCollection)
		stepErrMsg := fmt.Sprintf("Task %s with index %d result in error state...abort!!!!", operation.Step[stepIndex].Name, stepIndex)
		operation.CurrentStep = stepIndex
		operation.PreStepReturnData = stepReturnData
		if operation.Step[stepIndex].DynamicNodeSteps != nil {
			log.Debugf("Step %s contains dynamic steps bring it on then", operation.Step[stepIndex].Name)
			dynamicSteps, err := operation.Step[stepIndex].DynamicNodeSteps(stepReturnData, *cluster, *operation)
			if err != nil {
				if operation.Step[stepIndex].IgnoreDynamicStepCreationError {
					msg := fmt.Sprintf("(ignore) Failed to create dynamic steps from step %s due to error: %s", operation.Step[stepIndex].Name, err.Error())
					recordDebugLogToOperationLog(operation, msg)
				} else {
					handlerStepError(operation, cluster, runtimeCache, operation.Step[stepIndex], stepIndex, taskErrorHandler, stepErrMsg)
					return
				}
			}
			operation.Step[stepIndex].NodeSteps = append(operation.Step[stepIndex].NodeSteps, dynamicSteps...)
		}
		if len(operation.Step[stepIndex].NodeSteps) == 0 {
			recordDebugLogToOperationLog(operation, fmt.Sprintf("There is no node steps in step %d skip it!", stepIndex))
			continue
		}
		for nodeTaskIndex, nodeTask := range operation.Step[stepIndex].NodeSteps {
			msgData, err := createMsgBody(operation.Id, resourceServerURL, nodeTask.Id, nodeTask.Tasks[0], *cluster, stepReturnData)
			if err != nil {
				logString := fmt.Sprintf("Failed to marshal data for node task %d of step %d for operation %s due to error %s", nodeTaskIndex, stepIndex, operation.Id, err.Error())
				handlerStepError(operation, cluster, runtimeCache, operation.Step[stepIndex], stepIndex, taskErrorHandler, logString)
				return
			}

			nodeData, foundNode := (*nodeCollection)[nodeTask.NodeID]
			if !foundNode {
				log.Error("Unable to find node information with id", nodeTask.NodeID)
				taskDoneSignal <- 1
			} else {
				// send message to all node task to do actual job such as set up cri
				logMsg := fmt.Sprintf("Processing step %s by send a message for task %s to target node %s and wait %d seconds for reply", operation.Step[stepIndex].Name, nodeTask.Tasks[0].GetTaskType(), nodeData.Ipv4DefaultIp+fmt.Sprintf("(proxy:%s)", nodeData.ProxyIpv4CIDR), nodeTask.Tasks[0].GetTaskTimeOut())
				recordDebugLogToOperationLog(operation, logMsg)
				go messageQueue.SendingMessageWithReply(mqHelper.GeneratorNodeSubscribe(nodeData, config.MessageQueue.SubjectSuffix),
					cache.NodeId,
					nodeTask.NodeID,
					nodeTask.Id,
					config.MessageQueue,
					time.Duration(nodeTask.Tasks[0].GetTaskTimeOut())*time.Second,
					[]byte(msgData),
					taskDoneHandler,
					taskTimeOutSignalHandler)
			}
		}
		recordDebugLogToOperationLog(operation, fmt.Sprintf("Waiting for node reply for task %s with index %d", operation.Step[stepIndex].Name, stepIndex))
		select {
		case doneOrError := <-taskDoneSignal:
			if doneOrError == 0 {
				if stepIndex < len(operation.Step)-1 {
					recordDebugLogToOperationLog(operation, fmt.Sprintf("All node task has been done. Processing next step: %s", operation.Step[stepIndex+1].Name))
				}
				if config.DisableLazyOperationLog {
					if err := saveOperationStatus(runtimeCache, *operation); err != nil {
						log.Errorf("Failed to record error data on task failure due to error %s", err.Error())
						handlerStepError(operation, cluster, runtimeCache, operation.Step[stepIndex], stepIndex, taskErrorHandler, stepErrMsg)
						return
					}
				}
				continue
			} else {
				handlerStepError(operation, cluster, runtimeCache, operation.Step[stepIndex], stepIndex, taskErrorHandler, stepErrMsg)
				return
			}
		case msg := <-taskTimeOutSignal:
			// got abort signal aborting ...
			errMsg := fmt.Sprintf("Task %s with index %d reach maxmuim allowed time out count...abort!!!!", operation.Step[stepIndex].Name, stepIndex)
			recordErrorLogToOperationLog(operation, msg)
			handlerStepError(operation, cluster, runtimeCache, operation.Step[stepIndex], stepIndex, taskErrorHandler, errMsg)
			return
		}
	}
	recordDebugLogToOperationLog(operation, fmt.Sprintf("All step done set operation status to %s", constants.StatusSuccessful))
	operation.Status = constants.StatusSuccessful
	//remove current operation id from current operations map
	delete(cluster.CurrentOperation, operation.Id)

	if taskCompleteHandler != nil {
		if err := taskCompleteHandler(operation, cluster, runtimeCache); err != nil {
			log.Errorf("(ignore) Task complete handler run into error stat: %s", err.Error())
		}
	}

	if err := saveOperationStatus(runtimeCache, *operation); err != nil {
		log.Errorf("Failed to record error data on all step is done due to error %s", err.Error())
	}

	if err := saveClusterStatus(runtimeCache, *cluster); err != nil {
		log.Errorf("Failed to save cluster data to db when all step is done due to error %s", err.Error())
	}
}

func createMsgBody(operationId, resourceServerURL, nodeStepId string, task schema.ITask, cluster schema.Cluster, stepReturnData map[string]string) (string, error) {
	msgBody := &schema.QueueBody{
		OperationId:       operationId,
		TaskType:          task.GetTaskType(),
		Cluster:           cluster,
		ResourceServerURL: resourceServerURL,
		StepReturnData:    stepReturnData,
		NodeStepId:        nodeStepId,
	}
	if data, err := json.Marshal(task); err != nil {
		return "", err
	} else {
		msgBody.TaskData = data
	}
	if data, err := json.Marshal(msgBody); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func SetOperationToFailedStat(operation *schema.Operation, runtimeCache cache.ICache, logString string) {
	recordErrorLogToOperationLog(operation, logString)
	operation.Status = constants.StatusError
	operation.OperationLog = logString
	if err := saveOperationStatus(runtimeCache, *operation); err != nil {
		log.Errorf("Failed to record error data on creating message body failure due to error %s", err.Error())
	}
}

func handlerStepError(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache, step *schema.Step, stepIndex int, taskErrorHandler func(operation *schema.Operation, cluster *schema.Cluster, runtimeCache cache.ICache, err error) error, msg string) {
	//errMsg := fmt.Sprintf("Task %s with index %d result in error state...abort!!!!", step.Name, stepIndex)
	recordErrorLogToOperationLog(operation, msg)
	recordDebugLogToOperationLog(operation, fmt.Sprintf("Operation %s abort. Saving error state to db", operation.Id))
	operation.Status = constants.StatusError
	if taskErrorHandler != nil {
		if err := taskErrorHandler(operation, cluster, runtimeCache, errors.New(msg)); err != nil {
			log.Errorf("(ignore) Task error handler run into error stat: %s", err.Error())
		}
	}
	if err := saveOperationStatus(runtimeCache, *operation); err != nil {
		log.Errorf("Task control flow error handler result in error status,no more action can be taken,please see following error,this may leads to operation lost his correct status and certain operation log. which normally caused by db connection issue")
		log.Errorf("Failed to record error data on task failure due to error %s", err.Error())
	}
}
