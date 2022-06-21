package openldap

import (
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
)

const (
	DefaultNamespace     = "openldap"
	DefaultAdminPassword = "admin"
	DefaultStorageSize   = 5
	DefaultDN            = "demo.com"
)

type DeployOpenLDAP struct {
	ImageRegistry string `json:"-"`                                                   // image registry
	Namespace     string `json:"namespace,omitempty" validate:"omitempty,k8s_namespace"`        // k8s namespace
	AdminPassword string `json:"admin_password,omitempty"`                            // LDAP admin password
	StorageSize   int    `json:"storage_size,omitempty" validate:"omitempty,storage_size"`      // storage size
	StorageClass  string `json:"storage_class" validate:"required,k8s_storage_class"` // storage class                      // storage class
	DN            string `json:"domain_name,omitempty"`                               // domain name
}

/**
 * Set image registry for OpenLDAP deployment.
 */
func (do DeployOpenLDAP) SetImageRegistry(registry string) DeployOpenLDAP {
	do.ImageRegistry = registry
	return do
}

/**
 * Complete empty options with default settings.
 */
func (do DeployOpenLDAP) CompleteOpenLDAPDeploy() DeployOpenLDAP {
	do.Namespace = util.StringDefaultIfNotSet(do.Namespace, DefaultNamespace)
	do.AdminPassword = util.StringDefaultIfNotSet(do.AdminPassword, DefaultAdminPassword)
	do.DN = util.StringDefaultIfNotSet(do.DN, DefaultDN)
	return do
}

/**
 * Renderer out target yaml string from template.
 */
func (do DeployOpenLDAP) TemplateRender() (string, error) {
	// Check and complete with default settings
	return template.New("openldap").Render(Template, do.CompleteOpenLDAPDeploy())
}
