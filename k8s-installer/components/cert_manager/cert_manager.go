package cert_manager

import (
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
)

const (
	DefaultNamespace = "cert-manager"
)

type DeployCertManager struct {
	ImageRegistry string `json:"-"`
	Namespace     string `json:"namespace"`
}

/**
 * Set image registry for certificate manager deployment.
 */
func (dc *DeployCertManager) SetImageRegistry(registry string) {
	dc.ImageRegistry = registry
}

/**
 * Complete namespace with default one.
 */
func (dc *DeployCertManager) CompleteCertManagerDeploy() {
	dc.Namespace = util.StringDefaultIfNotSet(dc.Namespace, DefaultNamespace)
}

/**
 * Renderer out target yaml string from template.
 */
func (dc *DeployCertManager) TemplateRender() (string, error) {
	dc.CompleteCertManagerDeploy()
	return template.New("cert manager").Render(Template, dc)
}
