package app

import (
	"k8s-installer/pkg/log"
	"strings"
)

func commandArgsValidation() {
	if strings.ToLower(currentConfig.MessageQueue.AuthMode) != "basic" && strings.ToLower(currentConfig.MessageQueue.AuthMode) != "tls" {
		log.Fatalf("Message queue auth mode %s is not a valid only basic or tls is support", currentConfig.MessageQueue.AuthMode)
	}
}
