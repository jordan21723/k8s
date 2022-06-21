package efk

import (
	"testing"
)

const (
	Registry         = "172.20.150.76:4000"
	StorageClassName = "cinder"
	Replicas         = 3
)

func TestTemplateRenderBase(t *testing.T) {
	e := NewDeployEFK(Registry, StorageClassName, Replicas)
	tmpl, err := e.TemplateRender(EFKBase)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("efk base yaml: \n%s", tmpl)
}

func TestTemplateRenderDynamicPV(t *testing.T) {
	e := NewDeployEFK(Registry, StorageClassName, Replicas)
	tmpl, err := e.TemplateRender(EFKDynamicPV)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("efk base yaml: \n%s", tmpl)
}
