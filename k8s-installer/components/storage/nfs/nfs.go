package nfs

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/template"
	"k8s-installer/schema/plugable"
)

type NFS struct {
	Enable           bool                          `json:"enable" description:"enable or disable nfs interfacing,default=false"`
	NFSServerAddress string                        `json:"nfs_server_address" validate:"required" description:"required when enable=true,nfs server address"`
	NFSPath          string                        `json:"nfs_path" validate:"required" description:"required when enable=true,nfs server nfs path"`
	ReclaimPolicy    string                        `json:"reclaim_policy" validate:"required,oneof=Delete Retain" description:"required when enable=true,behavior of delete nfs storage class pvc"`
	ImageRegistry    string                        `json:"-"`
	StorageClassName string                        `json:"storage_class_name" validate:"required,k8s_storage_class" description:"required when enable=true,e.g. nfs-csi"`
	Dependencies     map[string]plugable.IPlugAble `json:"-"`
	Status           string                        `json:"status,omitempty" description:"auto generated,do not input and all input will be ignore"`
	IsDefaultSc      bool                          `json:"-"`
	MountOptions     []string                      `json:"mount_options,omitempty" description:"auto generated,do not input and all input will be ignore"`
}

func (nfs NFS) IsEnable() bool {
	return nfs.Enable
}

func (nfs NFS) GetName() string {
	return "NFS"
}

func (nfs NFS) GetStatus() string {
	return nfs.Status
}

func (nfs *NFS) SetDependencies(deps map[string]plugable.IPlugAble) {
	nfs.Dependencies = deps
}
func (nfs *NFS) GetDependencies() map[string]plugable.IPlugAble {
	return nfs.Dependencies
}

func (nfs *NFS) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(nfs)
}

func (nfs *NFS) GetLicenseLabel() uint16 {
	return constants.LLNFS
}

/**
 * Set image registry for middle platform deployment.
 */
func (nfs *NFS) SetImageRegistry(registry string) {
	nfs.ImageRegistry = registry
}

func (nfs *NFS) DeploymentTemplateRender() (string, error) {
	return template.New("nfs template").Render(Template, nfs)
}

func (nfs *NFS) StorageClassTemplateRender() (string, error) {
	if nfs.ReclaimPolicy == constants.ActionDelete {
		nfs.ReclaimPolicy = "true"
	} else {
		nfs.ReclaimPolicy = "false"
	}
	return template.New("nfs storage class").Render(StorageClassTemplate, nfs)
}
