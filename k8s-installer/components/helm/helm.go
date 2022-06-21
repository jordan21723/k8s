package helm

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/schema/plugable"
)

type Helm struct {
	Enable       bool                          `json:"enable" description:"enable or disable helm,default=false"`
	Version      int                           `json:"helm_version" validate:"required" description:"version of helm"`
	Dependencies map[string]plugable.IPlugAble `json:"-"`
	Status       string                        `json:"status,omitempty" description:"auto generated,do not input and all input will be ignore"`
}

func (helm Helm) IsEnable() bool {
	return helm.Enable
}

func (helm Helm) GetName() string {
	return "Helm"
}

func (helm Helm) GetStatus() string {
	return helm.Status
}

func (helm *Helm) SetDependencies(deps map[string]plugable.IPlugAble) {
	helm.Dependencies = deps
}
func (helm *Helm) GetDependencies() map[string]plugable.IPlugAble {
	return helm.Dependencies
}

func (helm *Helm) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(helm)
}

func (helm *Helm) GetLicenseLabel() uint16 {
	return constants.LLHelm
}

type Deps struct {
}

func (k Deps) GetDeps() dep.DepMap {
	return HelmDep
}

var HelmDep = dep.DepMap{
	"centos": {
		constants.V1_18_6: {
			"x86_64": {
				"helm": "helm",
			},
			"aarch64": {
				"helm": "helm",
			},
		},
	},
	"ubuntu": {},
}
