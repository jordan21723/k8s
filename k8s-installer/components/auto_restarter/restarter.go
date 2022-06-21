package restarter

import (
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
	"k8s-installer/schema/plugable"
)

const (
	DefaultNamespace = "kube-system"
	DefaultReplicas  = 2
	DefaultVersion   = "v1.0.0"
)

type DeployAutoRestarter struct {
	Enable        bool                          `json:"enable"`
	ImageRegistry string                        `json:"-"`
	Namespace     string                        `json:"namespace,omitempty" validate:"omitempty,k8s_namespace"` // optional
	Status        string                        `json:"status,omitempty"`                                       // optional
	VersionTag    string                        `json:"version,omitempty"`                                      // optional
	Replicas      int                           `json:"replicas,omitempty" validate:"omitempty,gte=1,lte=9`     // optional
	Dependencies  map[string]plugable.IPlugAble `json:"-"`
}

func (dr *DeployAutoRestarter) SetImageRegistry(registry string) {
	dr.ImageRegistry = registry
}

func (dr *DeployAutoRestarter) CompleteRestarterDeploy() {
	dr.Namespace = util.StringDefaultIfNotSet(dr.Namespace, DefaultNamespace)
	dr.VersionTag = util.StringDefaultIfNotSet(dr.VersionTag, DefaultVersion)
	dr.Replicas = util.IntDefaultIfZero(dr.Replicas, DefaultReplicas)
}

func (dr *DeployAutoRestarter) TemplateRender() (string, error) {
	dr.CompleteRestarterDeploy()
	return template.New("auto restarter").Render(Template, dr)
}

func (dr *DeployAutoRestarter) IsEnable() bool {
	return dr.Enable
}

func (dr *DeployAutoRestarter) GetStatus() string {
	return dr.Status
}

func (dr *DeployAutoRestarter) GetName() string {
	return "AutoRestarter"
}

func (dr *DeployAutoRestarter) GetDependencies() map[string]plugable.IPlugAble {
	return dr.Dependencies
}

func (dr *DeployAutoRestarter) SetDependencies(deps map[string]plugable.IPlugAble) {
	dr.Dependencies = deps
}

func (dr *DeployAutoRestarter) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(dr)
}

func (dr *DeployAutoRestarter) GetLicenseLabel() uint16 {
	return 1
}
