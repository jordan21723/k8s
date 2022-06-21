package nginx

import (
	"testing"
)

const (
	Registry = "docker.io"
)

var IngressNodeLabels = []string{"foo", "bar", "baz"}

func TestTemplateRender(t *testing.T) {
	tmpl, err := NewNginxIngressDeploy(Registry, IngressNodeLabels).TemplateRender()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("nginx ingress controller yaml: \n%s", tmpl)
}
