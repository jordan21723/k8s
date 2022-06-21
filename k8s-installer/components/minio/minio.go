package minio

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
	"k8s-installer/schema/plugable"
	"path/filepath"
)

const (
	DefaultName      = "minio"
	DefaultNamespace = "minio"

	DefaultUserName = "minio"
	DefaultPassWord = "minio123"
)

type DeployMinIO struct {
	Enable        bool                          `json:"enable" description:"enable"`
	ImageRegistry string                        `json:"-"`
	Namespace     string                        `json:"namespace,omitempty" validate:"omitempty,k8s_namespace" description:"namespace"`
	Dependencies  map[string]plugable.IPlugAble `json:"-"`
	Status        string                        `json:"status,omitempty" description:"status"`

	UserName string `json:"-"`
	PassWord string `json:"-"`

	// currently only nfs(mount)/hostpath has been tested
	BackupPath string `json:"-"`
}

func NewDeployMinIO(registry string) *DeployMinIO {
	return &DeployMinIO{
		Enable:        true,
		ImageRegistry: registry,
		Namespace:     DefaultNamespace,
	}
}

func (d *DeployMinIO) GetNamespace() string {
	return d.Namespace
}

func (d *DeployMinIO) SetImageRegistry(registry string) {
	d.ImageRegistry = registry
}

func (d *DeployMinIO) IsEnable() bool {
	return d.Enable
}

func (d *DeployMinIO) GetName() string {
	return DefaultName
}

func (d *DeployMinIO) GetStatus() string {
	return d.Status
}

func (d *DeployMinIO) SetDependencies(deps map[string]plugable.IPlugAble) {
	d.Dependencies = deps
}
func (d *DeployMinIO) GetDependencies() map[string]plugable.IPlugAble {
	return d.Dependencies
}

func (d *DeployMinIO) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(d)
}

func (d *DeployMinIO) CompleteMinIODeploy(clusterId, backupPath string) {
	d.Namespace = util.StringDefaultIfNotSet(d.Namespace, DefaultNamespace)
	d.UserName = util.StringDefaultIfNotSet(d.UserName, DefaultUserName)
	d.PassWord = util.StringDefaultIfNotSet(d.PassWord, DefaultPassWord)
	d.BackupPath = util.StringDefaultIfNotSet(d.BackupPath, filepath.Join(backupPath, clusterId))
}

func (d *DeployMinIO) GetLicenseLabel() uint16 {
	return constants.LLMinio
}

func (d *DeployMinIO) TemplateRender(clusterId, backupPath string) (string, error) {
	d.CompleteMinIODeploy(clusterId, backupPath)
	return template.New("minio-deployment").Render(MinioDeployment, d)
}
