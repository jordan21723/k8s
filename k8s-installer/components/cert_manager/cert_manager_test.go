package cert_manager

import "testing"

const (
	TestRegistry = "docker.io"
)

func TestTemplateRender(t *testing.T) {
	dc := &DeployCertManager{}
	dc.SetImageRegistry(TestRegistry)
	dc.CompleteCertManagerDeploy()
	tmpl, err := dc.TemplateRender()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("cert manager deployment yaml: \n%s", tmpl)
}
