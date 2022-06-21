package common

import (
	"errors"
	"fmt"
	natClient "github.com/nats-io/nats.go"
	"k8s-installer/pkg/message_queue/nats"
	mqHelper "k8s-installer/pkg/server/message_queue"
	"k8s-installer/schema"
	"time"
)

func SendingSingleCommand(mqConfig nats.MessageQueueConfig, nodeInformation schema.NodeInformation, dataToSend []byte, resultHandler func(msg *natClient.Msg) error) error {
	var mqErr error
	mqErr = nats.SendingMessageWithReply(mqHelper.GeneratorNodeSubscribe(nodeInformation, mqConfig.SubjectSuffix),
		nodeInformation.Id,
		nodeInformation.Id,
		"kubectlExec",
		mqConfig,
		3*time.Second,
		dataToSend,
		resultHandler,
		func(nodeId string, nodeStepId string) {
			mqErr = errors.New(fmt.Sprintf("Node with id %s is not responding in 3 sec", nodeInformation.Id))
		})
	return mqErr
}
