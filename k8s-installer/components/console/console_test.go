package console

import "testing"

const (
	TestRegistry = "docker.io"
)

func TestCompleteConsoleDeploy(t *testing.T) {
	dc := &DeployConsole{}
	dc.SetImageRegistry(TestRegistry)
	dc.CompleteConsoleDeploy()
	t.Logf("console deploy struct: %+v", dc)
}

func TestTemplateRender(t *testing.T) {
	dc := &DeployConsole{}
	dc.SetImageRegistry(TestRegistry)
	dc.CompleteConsoleDeploy()
	tmpl, err := dc.TemplateRender()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("console deployment yaml: \n%s", tmpl)
}
