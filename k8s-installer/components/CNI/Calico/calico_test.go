package Calico

import (
	"testing"

	"k8s-installer/schema"
)

const (
	Registry   = "docker.io"
	EnableIPIP = "Always"
)

func TestV4OnlyTemplateRender(t *testing.T) {
	cni := schema.CNI{
		Calico: schema.Calico{
			EnableDualStack: false,
			//Mode:            EnableIPIP,
		},
	}
	tmpl, err := NewCalicoDeploy(Registry, cni).TemplateRender()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("calico IPv4 only yaml: \n%s", tmpl)
}

func TestDualStackTemplateRender(t *testing.T) {
	cni := schema.CNI{
		Calico: schema.Calico{
			EnableDualStack: true,
			//Mode:            EnableIPIP,
		},
		PodV6CIDR: "fd20::0/112",
	}
	tmpl, err := NewCalicoDeploy(Registry, cni).TemplateRender()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("calico IPv4 + IPv6 yaml: \n%s", tmpl)
}
