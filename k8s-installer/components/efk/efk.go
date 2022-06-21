package efk

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
	"k8s-installer/schema/plugable"
)

const (
	DefaultNamespace = "efk"
	DefaultReplicas  = 1
)

type DeployEFK struct {
	Enable            bool                          `json:"enable" description:"enable or disable EFK,default=false"`
	ImageRegistry     string                        `json:"-"`
	Namespace         string                        `json:"namespace,omitempty" validate:"omitempty,k8s_namespace" description:"namespace of EFK used in k8s,default=efk"`
	StorageClassName  string                        `json:"storage_class_name" validate:"required,k8s_storage_class" description:"required when enable=true,should input storage class name of openstack provider storage class or nfs storage class"`
	Replicas          *int                           `json:"replicas,omitempty" validate:"omitempty,gte=1,lte=9" description:"the copy of efk container,default=1"`
	Dependencies      map[string]plugable.IPlugAble `json:"-"`
	Status            string                        `json:"status,omitempty" description:"auto generated,do not input and all input will be ignore"`
	User              string                        `json:"user,omitempty" description:"only used when ingress is active, default=admin"`
	Password          string                        `json:"password,omitempty" description:"only used when ingress is active, default=admin"`
}

func NewDeployEFK(registry, storageClassName string, replicas int) DeployEFK {
	return DeployEFK{
		ImageRegistry:    registry,
		Namespace:        "efk",
		StorageClassName: storageClassName,
		Replicas:         &replicas,
	}
}

func (de *DeployEFK) SetImageRegistry(registry string) {
	de.ImageRegistry = registry
}

func (de *DeployEFK) TemplateRender(templateVal string) (string, error) {
	de.CompleteEFKDeploy()
	return template.New("efk").Render(templateVal, de)
}

func (de *DeployEFK) IsEnable() bool {
	return de.Enable
}

func (de *DeployEFK) GetName() string {
	return "EFK"
}

func (de *DeployEFK) GetStatus() string {
	return de.Status
}

func (de *DeployEFK) SetDependencies(deps map[string]plugable.IPlugAble) {
	de.Dependencies = deps
}
func (de *DeployEFK) GetDependencies() map[string]plugable.IPlugAble {
	return de.Dependencies
}

func (de *DeployEFK) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(de)
}

func (de *DeployEFK) GetLicenseLabel() uint16 {
	return constants.LLEFK
}

func (de *DeployEFK) CompleteEFKDeploy() {
	de.Namespace = util.StringDefaultIfNotSet(de.Namespace, DefaultNamespace)
	de.Replicas = util.IntPDefaultIfNotSet(de.Replicas, DefaultReplicas)
}
