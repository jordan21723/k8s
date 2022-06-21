package velero

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
	"k8s-installer/schema/plugable"
)

const (
	DefaultName      = "velero"
	DefaultNamespace = "velero"

	DefaultChartPath = "/usr/share/k8s-installer/velero-chart"
)

type DeployVelero struct {
	Enable        bool
	ImageRegistry string                        `json:"-"`
	Namespace     string                        `json:"namespace,omitempty" validate:"omitempty,k8s_namespace" description:"namespace"`
	Dependencies  map[string]plugable.IPlugAble `json:"-"`
	Status        string                        `json:"status,omitempty" description:"status"`
}

func NewDeployVelero(registry string) *DeployVelero {
	return &DeployVelero{
		Enable:        true,
		ImageRegistry: registry,
		Namespace:     DefaultNamespace,
	}
}

func (d *DeployVelero) GetNamespace() string {
	return d.Namespace
}

func (d *DeployVelero) SetImageRegistry(registry string) {
	d.ImageRegistry = registry
}

func (d *DeployVelero) IsEnable() bool {
	return d.Enable
}

func (d *DeployVelero) GetName() string {
	return DefaultName
}

func (d *DeployVelero) GetStatus() string {
	return d.Status
}

func (d *DeployVelero) SetDependencies(deps map[string]plugable.IPlugAble) {
	d.Dependencies = deps
}

func (d *DeployVelero) GetDependencies() map[string]plugable.IPlugAble {
	return d.Dependencies
}

func (d *DeployVelero) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(d)
}

func (d *DeployVelero) CompleteVeleroDeploy() {
	d.Namespace = util.StringDefaultIfNotSet(d.Namespace, DefaultNamespace)
}

func (d *DeployVelero) GetLicenseLabel() uint16 {
	return constants.LLVelero
}

func (d *DeployVelero) TemplateMinioSecrteText(str string) (string, error) {
	return template.New("credentials-velero").Render(str, d)
}

type Deps struct {
}

func (k Deps) GetDeps() dep.DepMap {
	return VeleroDep
}

var VeleroDep = dep.DepMap{
	"centos": {
		constants.V1_18_6: {
			"x86_64": {
				"velero": "velero",
			},
			"aarch64": {
				"velero": "velero",
			},
		},
	},
	"ubuntu": {},
}
