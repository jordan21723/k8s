package openldap

import "testing"

const (
	TestRegistry     = "docker.io"
	TestStorageClass = "nfs"
)

func TestTemplateRender(t *testing.T) {
	do := DeployOpenLDAP{
		StorageClass: TestStorageClass,
	}.SetImageRegistry(TestRegistry)
	tmpl, err := do.TemplateRender()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("openldap deployment yaml: \n%s", tmpl)
}
