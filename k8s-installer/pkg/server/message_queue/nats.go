package message_queue

import (
	"encoding/json"
	"fmt"

	natClient "github.com/nats-io/nats.go"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
	k8sCore "k8s.io/api/core/v1"
)

func GeneratorNodeSubscribe(node schema.NodeInformation, suffix string) string {
	return node.Ipv4DefaultIp + "." + suffix
}

func CreateMsgBody(operationId, resourceServerURL string, task schema.ITask, cluster schema.Cluster, stepReturnData map[string]string) (string, error) {
	msgBody := &schema.QueueBody{
		OperationId:       operationId,
		TaskType:          task.GetTaskType(),
		Cluster:           cluster,
		ResourceServerURL: resourceServerURL,
		StepReturnData:    stepReturnData,
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

func ParseKubectlGetNodes(nodes map[string]*schema.NodeInformation,
	doneOrErrorSignal chan struct {
		Error   error
		Message string
	}) func(msg *natClient.Msg) error {
	return func(msg *natClient.Msg) error {
		var reply schema.QueueReply
		if err := json.Unmarshal(msg.Data, &reply); err != nil {
			doneOrErrorSignal <- struct {
				Error   error
				Message string
			}{Error: err, Message: fmt.Sprintf("Failed to parse reply message %s to QueueReply due to error: %s", reply.Message, err.Error())}
			return err
		}
		var k8sNodeLists k8sCore.NodeList
		if err := json.Unmarshal([]byte(reply.ReturnData[constants.ReturnDataKeyKubectl]), &k8sNodeLists); err != nil {
			doneOrErrorSignal <- struct {
				Error   error
				Message string
			}{Error: err, Message: fmt.Sprintf("Failed to parse reply message %s to k8sCore.NodeList due to error: %v", reply.Message, err.Error())}
			return err
		}
		for _, kubeNode := range k8sNodeLists.Items {
			if node, found := nodes[kubeNode.Name]; found {
				if kubeNode.Status.Conditions[len(kubeNode.Status.Conditions)-1].Status == "True" {
					node.KubeNodeStat = "Ready"
				} else {
					node.KubeNodeStat = "NotReady"
				}
			}
		}

		// no error
		doneOrErrorSignal <- struct {
			Error   error
			Message string
		}{Error: nil, Message: ""}
		return nil
	}
}
