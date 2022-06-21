package middle_platform

import (
	"k8s-installer/components/openldap"
	pgo "k8s-installer/components/postgres_operator"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
	"k8s-installer/schema/plugable"
)

const (
	DefaultNamespace           = "caas4-middle-platform"
	DefaultSystemAdminPassword = "password"
	DefaultLDAPPassword        = "admin"
	DefaultPGOAdminUser        = "admin"
	DefaultPGOAdminPassword    = "mep@123"
	DefaultLDAPDomainName      = "demo.com"
	DefaultPGONamespace        = "system-middleware-pg"
	DefaultHarborAdminUser     = "admin"
	DefaultDBUser              = "middle"
	DefaultDBPassword          = "enfefjk35k35j4k6"
	DefaultDBReplicas          = 2
	DefaultDBHost              = "middle-db"
	DefaultDBName              = "maccount"
)

type (
	LDAP struct {
		Enable        bool                     `json:"enable" description:"Deprecated, will be remove in future"`                   // enable LDAP authentication
		CaaSDeploy    *openldap.DeployOpenLDAP `json:"caas_deploy,omitempty" description:"Deprecated, will be remove in future"`    // deploy OpenLDAP on CaaS
		Addr          string                   `json:"addr,omitempty" description:"Deprecated, will be remove in future"`           // LDAP server address
		AdminUser     string                   `json:"admin_user,omitempty" description:"Deprecated, will be remove in future"`     // LDAP admin user
		AdminPassword string                   `json:"admin_password,omitempty" description:"Deprecated, will be remove in future"` // LDAP admin password
		DN            string                   `json:"domain_name,omitempty" description:"Deprecated, will be remove in future"`    // LDAP domain name
	}
	DeployMiddlePlatform struct {
		Enable              bool                          `json:"enable" description:"enable or disable middle platform"`                               // enable middle platform
		ImageRegistry       string                        `json:"-"`                                                                                    // image registry
		Namespace           string                        `json:"namespace,omitempty" validate:"omitempty,k8s_namespace" description:"middle platform namespace"` // k8s namespace
		SystemAdminPassword string                        `json:"sys_admin_password,omitempty" description:"middle platform admin's password"`          // password of system administrator
		DB                  DB                            `json:"db" description:"middle platform postgres section"`                                    // database options
		PGO                 pgo.DeployPostgresOperator    `json:"pgo" description:"middle platform postgres section"`
		LDAP                *LDAP                         `json:"ldap" description:"Deprecated, will be remove in future"`           // LDAP options
		EnableLicense       bool                          `json:"enable_license" description:"Deprecated, will be remove in future"` // enable license validation
		Dependencies        map[string]plugable.IPlugAble `json:"-"`
		Status              string                        `json:"status" description:"auto generated,do not input and all input will be ignore"`
		// Harbor              HarborOptions                 `json:"harbor"`                   // harbor registry options
	}
	HarborOptions struct {
		Addr          string `json:"addr" validate:"required,hostname_port" description:"Deprecated, will be remove in future"` // Harbor registry host address
		AdminUser     string `json:"admin_user" validate:"required" description:"Deprecated, will be remove in future"`         // Harbor registry admin user
		AdminPassword string `json:"admin_password" validate:"required" description:"Deprecated, will be remove in future"`     // Harbor registry admin password
	}
	DB struct {
		User     string `json:"user,omitempty" description:"username of postgres db created by PostgresOperator, default=middle"`                  // db user
		Password string `json:"password,omitempty" description:"password of postgres db created by PostgresOperator, default=enfefjk35k35j4k6"`    // db password
		Replicas *int   `json:"replicas,omitempty" validate:"omitempty,gte=1,lte=9" description:"auto generated, do not input and all input will be ignore"` // db replicas
		Host     string `json:"host,omitempty" description:"hostname of the db created by by PostgresOperator,default=middle-db"`                  // db host address
		Name     string `json:"name,omitempty" description:"database name of the db created by by PostgresOperator,default=maccount"`              // db name
	}
)

func (dm DeployMiddlePlatform) GetNamespace() string {
	if dm.Namespace == "" {
		return DefaultNamespace
	}
	return dm.Namespace
}

func (dm DeployMiddlePlatform) IsEnable() bool {
	return dm.Enable
}

func (dm DeployMiddlePlatform) GetStatus() string {
	return dm.Status
}

func (dm DeployMiddlePlatform) GetName() string {
	return "MiddlePlatform"
}

func (dm *DeployMiddlePlatform) SetDependencies(deps map[string]plugable.IPlugAble) {
	dm.Dependencies = deps
}
func (dm *DeployMiddlePlatform) GetDependencies() map[string]plugable.IPlugAble {
	return dm.Dependencies
}

func (dm *DeployMiddlePlatform) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(dm)
}

func (dm *DeployMiddlePlatform) GetLicenseLabel() uint16 {
	// we use llconsole on purpose, share label with console
	return constants.LLConsole
}

/**
 * Set image registry for middle platform deployment.
 */
func (dm *DeployMiddlePlatform) SetImageRegistry(registry string) {
	dm.ImageRegistry = registry
}

/**
 * Complete empty options of middle platform with default settings.
 */
func (dm *DeployMiddlePlatform) CompleteMiddlePlatformDeploy() {
	dm.Namespace = util.StringDefaultIfNotSet(dm.Namespace, DefaultNamespace)
	dm.SystemAdminPassword = util.StringDefaultIfNotSet(dm.SystemAdminPassword, DefaultSystemAdminPassword)
	dm.DB = dm.DB.CompleteDB() // complete database options with default settings
	dm.PGO.CompletePostgresOperatorDeploy()
	// There are no default settings for Harbor registry.
	if dm.LDAP != nil && len(dm.LDAP.Addr)*len(dm.LDAP.AdminUser)*len(dm.LDAP.AdminPassword)*len(dm.LDAP.DN) == 0 {
		// only one in Addr/AdminUser/AdminPassword/DN is empty
		if dm.LDAP.CaaSDeploy != nil {
			do := openldap.DeployOpenLDAP{}.SetImageRegistry(dm.ImageRegistry)
			dm.LDAP.CaaSDeploy = &do
		}
		cd := (*dm.LDAP.CaaSDeploy).CompleteOpenLDAPDeploy()
		dm.LDAP.CaaSDeploy = &cd
	}
}

/**
 * Complete empty options of database with default settings.
 */
func (db DB) CompleteDB() DB {
	db.User = util.StringDefaultIfNotSet(db.User, DefaultDBUser)
	db.Password = util.StringDefaultIfNotSet(db.Password, DefaultDBPassword)
	db.Replicas = util.IntPDefaultIfNotSet(db.Replicas, DefaultDBReplicas)
	db.Host = util.StringDefaultIfNotSet(db.Host, DefaultDBHost)
	db.Name = util.StringDefaultIfNotSet(db.Name, DefaultDBName)
	return db
}

/**
 * Renderer out target yaml string from template.
 */
func (dm *DeployMiddlePlatform) TemplateRender() (string, error) {
	dm.CompleteMiddlePlatformDeploy()
	return template.New("caas middle platform").Render(Template, dm)
}
