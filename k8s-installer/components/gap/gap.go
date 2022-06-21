package gap

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/util"
	"k8s-installer/schema/plugable"
)

const (
	DefaultNamespace = "gap"
)

type DeployGAP struct {
	Enable        bool                          `json:"enable"`
	ImageRegistry string                        `json:"-"`
	Namespace     string                        `json:"namespace,omitempty" validate:"omitempty,k8s_namespace"`
	Dependencies  map[string]plugable.IPlugAble `json:"-"`
	Status        string                        `json:"status,omitempty"`
}

func NewDeployGAP(registry string) DeployGAP {
	return DeployGAP{
		ImageRegistry: registry,
		Namespace:     "gap",
	}
}

func (dg *DeployGAP) GetNamespace() string {
	return dg.Namespace
}

func (dg *DeployGAP) SetImageRegistry(registry string) {
	dg.ImageRegistry = registry
}

func (dg *DeployGAP) IsEnable() bool {
	return dg.Enable
}

func (dg *DeployGAP) GetName() string {
	return "GAP"
}

func (dg *DeployGAP) GetStatus() string {
	return dg.Status
}

func (dg *DeployGAP) SetDependencies(deps map[string]plugable.IPlugAble) {
	dg.Dependencies = deps
}
func (dg *DeployGAP) GetDependencies() map[string]plugable.IPlugAble {
	return dg.Dependencies
}

func (dg *DeployGAP) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(dg)
}

func (dg *DeployGAP) CompleteGAPDeploy() {
	dg.Namespace = util.StringDefaultIfNotSet(dg.Namespace, DefaultNamespace)
}

func (dg *DeployGAP) GetLicenseLabel() uint16 {
	return constants.LLGAP
}