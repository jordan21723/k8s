package coredns

import (
	"fmt"
	"k8s-installer/pkg/message_queue/nats"
)

func SendToDNS(DNSDomain string, subject, senderId string, mq nats.MessageQueueConfig) error {
	if err := nats.SendingMessage(subject, DNSDomain, senderId); err != nil {
		return fmt.Errorf("Failed to send data to dns due to error: %s ", err.Error())
	}
	return nil
}
