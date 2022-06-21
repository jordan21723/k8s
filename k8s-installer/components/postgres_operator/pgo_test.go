package postgres_operator

import "testing"

const (
	TestRegistry     = "docker.io"
	TestStorageClass = "nfs"
)

func TestTemplateRender(t *testing.T) {
	d := DeployPostgresOperator{
		StorageClass: TestStorageClass,
	}
	tmpl, err := d.TemplateRender()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("postgres operator deployment yaml: \n%s", tmpl)
}

func TestCompletePostgresOperatorDeploy(t *testing.T) {
	d := DeployPostgresOperator{
		StorageClass: TestStorageClass,
	}
	d.CompletePostgresOperatorDeploy()
	t.Logf("Postgres Operator Deploy struct: %+v\n", d)
}
