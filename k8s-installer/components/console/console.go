package console

import (
	"k8s-installer/components/gap"
	mp "k8s-installer/components/middle_platform"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
	"k8s-installer/schema/plugable"
)

const (
	DefaultNamespace         = "caas4-console"
	DefaultCertManagerIssuer = "console-cert-manager-issuer"
	DefaultReplicas          = 2
	DefaultConsoleTitle      = "CaaS"
)

type (
	DeployConsole struct {
		Enable                  bool                          `json:"enable"`
		ImageRegistry           string                        `json:"-"`
		Namespace               string                        `json:"namespace,omitempty" validate:"omitempty,k8s_namespace"`
		VendorTag               string                        `json:"vendor_tag" validate:"required"`
		MiddlePlatformNamespace string                        `json:"-"`
		PrometheusNamespace     string                        `json:"-"`
		TLSEnable               bool                          `json:"tls_enable"`
		CertManagerIssuer       string                        `json:"cert_manager_issuer,omitempty"`
		Replicas                *int                          `json:"replicas,omitempty" validate:"omitempty,gte=1,lte=9"`
		Dependencies            map[string]plugable.IPlugAble `json:"-"`
		Status                  string                        `json:"status,omitempty"`
		ConsoleTitle            string                        `json:"-"`
		ConsoleResourcePath     string                        `json:"-"`
	}
)

func (dc *DeployConsole) SetImageRegistry(registry string) {
	dc.ImageRegistry = registry
}

/**
 * Complete empty options of console with default settings.
 */
func (dc *DeployConsole) CompleteConsoleDeploy() {
	dc.Namespace = util.StringDefaultIfNotSet(dc.Namespace, DefaultNamespace)
	dc.MiddlePlatformNamespace = util.StringDefaultIfNotSet(dc.MiddlePlatformNamespace, mp.DefaultNamespace)
	dc.PrometheusNamespace = util.StringDefaultIfNotSet(dc.PrometheusNamespace, gap.DefaultNamespace)
	dc.CertManagerIssuer = util.StringDefaultIfNotSet(dc.CertManagerIssuer, DefaultCertManagerIssuer)
	dc.Replicas = util.IntPDefaultIfNotSet(dc.Replicas, DefaultReplicas)
	dc.ConsoleTitle = util.StringDefaultIfNotSet(dc.ConsoleTitle, DefaultConsoleTitle)
}

/**
 * Renderer out target yaml string from template.
 */
func (dc *DeployConsole) TemplateRender() (string, error) {
	dc.CompleteConsoleDeploy()
	return template.New("caas console").Render(Template, dc)
}

func (dc DeployConsole) IsEnable() bool {
	return dc.Enable
}

func (dc DeployConsole) GetStatus() string {
	return dc.Status
}

func (dc DeployConsole) GetName() string {
	return "Console"
}

func (dc *DeployConsole) SetDependencies(deps map[string]plugable.IPlugAble) {
	dc.Dependencies = deps
}

func (dc *DeployConsole) GetDependencies() map[string]plugable.IPlugAble {
	return dc.Dependencies
}

func (dc *DeployConsole) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(dc)
}

func (dc *DeployConsole) GetLicenseLabel() uint16 {
	return constants.LLConsole
}
