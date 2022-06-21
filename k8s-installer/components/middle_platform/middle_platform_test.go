package middle_platform

import (
	"testing"

	"k8s-installer/components/openldap"
	postgresOperator "k8s-installer/components/postgres_operator"
)

const (
	TestRegistry            = "docker.io"
	TestHarborAddr          = "1.2.3.4:5000"
	TestLDAPAddr            = "4.3.2.1:389"
	TestHarborAdminPassword = "harbor_test_admin_password"
)

func TestTemplateRender(t *testing.T) {
	db := DB{}.CompleteDB()
	pgo := &postgresOperator.DeployPostgresOperator{}
	pgo.CompletePostgresOperatorDeploy()
	do := openldap.DeployOpenLDAP{}.CompleteOpenLDAPDeploy()
	ldap := &LDAP{
		Enable:     true,
		CaaSDeploy: &do,
	}

	dm := &DeployMiddlePlatform{
		ImageRegistry:       TestRegistry,
		Namespace:           DefaultNamespace,
		SystemAdminPassword: DefaultSystemAdminPassword,
		DB:                  db,
		LDAP:                ldap,
		EnableLicense:       true,
	}
	tmpl, err := dm.TemplateRender()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("middle platform deployment yaml: \n%s", tmpl)
}

func TestCompleteMiddlePlatformDeploy(t *testing.T) {
	do := openldap.DeployOpenLDAP{}.CompleteOpenLDAPDeploy()
	ldap := &LDAP{
		Enable:     true,
		CaaSDeploy: &do,
	}
	dm := &DeployMiddlePlatform{
		LDAP: ldap,
	}
	dm.SetImageRegistry(TestRegistry)
	dm.CompleteMiddlePlatformDeploy()
	t.Logf("Middle Platform Deploy struct: %+v\n", dm)
	t.Logf("CaaS LDAP struct: %+v\n", *&dm.LDAP.CaaSDeploy)
}
