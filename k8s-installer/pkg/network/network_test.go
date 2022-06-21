package network

import (
	"testing"

	"k8s-installer/pkg/log"
)

func TestGetDefaultIP(t *testing.T) {
	t.Log("default ip:", GetDefaultIP(true).String())
}

func TestGetDefaultGateway(t *testing.T) {
	_, nic, err := GetDefaultGateway(true)
	if err != nil {
		log.Fatal(err)
	}
	t.Log("nic name:", nic.Name)
}
