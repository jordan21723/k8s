package postgres_operator

import (
	"fmt"
	"k8s-installer/pkg/constants"

	"k8s-installer/schema/plugable"

	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
)

const (
	DefaultNamespace     = "system-middleware-pg"
	DefaultAdminUser     = "admin"
	DefaultAdminPassword = "mep@123"
	DefaultStorageSize   = 5
	DefaultDBUser        = "middle"
	DefaultDBPassword    = "enfefjk35k35j4k6"
	DefaultDBReplicas    = 2
	DefaultDBHost        = "middle-db"
	DefaultDBName        = "maccount"
)

var DefaultWatchNamespaces = []string{
	"mep-services",
	"pg-cluster1",
	"pg-cluster2",
	"pg-cluster3",
	"pg-cluster4",
	"pg-cluster5",
	"pg-cluster6",
	"pg-cluster7",
	"pg-cluster8",
	"pg-cluster9",
}

type (
	DeployPostgresOperator struct {
		Enable          bool                          `json:"enable" description:"enable or disable operator, default=false"`
		ImageRegistry   string                        `json:"-"`
		Namespace       string                        `json:"namespace,omitempty" validate:"omitempty,k8s_namespace" description:"postgres namespace, default=system-middleware-pg"`
		AdminUser       string                        `json:"admin_user,omitempty" description:"username of Postgres Operator, default=admin"`
		AdminPassword   string                        `json:"admin_password,omitempty" description:"password of Postgres Operator, default=mep@123"`
		StorageClass    string                        `json:"storage_class,omitempty" validate:"omitempty,k8s_storage_class" description:"required when enable=true,should input storage class name of openstack provider storage class or nfs storage class"`
		StorageSize     *int                          `json:"storage_size,omitempty" validate:"omitempty,storage_size" description:"size of the postgres db will use in k8s, it will fall back to default(5) when less than 5"`
		WatchNamespaces string                        `json:"-"`
		Dependencies    map[string]plugable.IPlugAble `json:"-"`
		Status          string                        `json:"status,omitempty" description:"auto generated, do not input and all input will be ignore"`
	}
)

func (dp *DeployPostgresOperator) SetImageRegistry(registry string) {
	dp.ImageRegistry = registry
}

func (dp *DeployPostgresOperator) CompletePostgresOperatorDeploy() {
	dp.Namespace = util.StringDefaultIfNotSet(dp.Namespace, DefaultNamespace)
	dp.AdminUser = util.StringDefaultIfNotSet(dp.AdminUser, DefaultAdminUser)
	dp.AdminPassword = util.StringDefaultIfNotSet(dp.AdminPassword, DefaultAdminPassword)
	dp.StorageSize = util.IntPDefaultIfNotSet(dp.StorageSize, DefaultStorageSize)
	var watchNamespaces string
	for _, ns := range DefaultWatchNamespaces {
		watchNamespaces += fmt.Sprintf("%s,", ns)
	}
	dp.WatchNamespaces = watchNamespaces + dp.Namespace
}

func (dp *DeployPostgresOperator) TemplateRender() (string, error) {
	dp.CompletePostgresOperatorDeploy()
	return template.New("postgres operator").Render(Template, dp)
}

func (dp *DeployPostgresOperator) IsEnable() bool {
	return dp.Enable
}

func (dp *DeployPostgresOperator) GetStatus() string {
	return dp.Status
}

func (dp *DeployPostgresOperator) GetName() string {
	return "PostgresOperator"
}

func (dp *DeployPostgresOperator) SetDependencies(deps map[string]plugable.IPlugAble) {
	dp.Dependencies = deps
}
func (dp *DeployPostgresOperator) GetDependencies() map[string]plugable.IPlugAble {
	return dp.Dependencies
}

func (dp *DeployPostgresOperator) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(dp)
}

func (dm *DeployPostgresOperator) GetNamespace() string {
	return util.StringDefaultIfNotSet(dm.Namespace, DefaultNamespace)
}

func (dp *DeployPostgresOperator) GetWatchNamespaces() []string {
	return DefaultWatchNamespaces
}

func (dp *DeployPostgresOperator) GetLicenseLabel() uint16 {
	// we use llconsole on purpose, share label with console
	return constants.LLConsole
}