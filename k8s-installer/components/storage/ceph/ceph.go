package ceph

import (
	"k8s-installer/pkg/template"
	"k8s-installer/schema/plugable"
)

type Ceph struct {
	Enable           bool                          `json:"enable" description:"enable or disable nfs interfacing,default=false"`
	CephClusterId    string                        `json:"ceph_cluster_id" validate:"required" description:"ceph cluster fsid"`
	MonitorIPList    []string                      `json:"monitor_ip_list" validate:"required" description:"ceph monitor node ip"`
	PoolUserId       string                        `json:"pool_user_id" validate:"required" description:"ceph osd pool user id"`
	PoolUserKey      string                        `json:"pool_user_key" validate:"required" description:"ceph osd pool user key"`
	StorageClassName string                        `json:"storage_class_name" validate:"required,k8s_storage_class" description:"required when enable=true,e.g. ceph-csi"`
	ReclaimPolicy    string                        `json:"reclaim_policy" validate:"required,oneof=Delete Retain" description:"required when enable=true,behavior of delete ceph storage class pvc"`
	ImageRegistry    string                        `json:"-"`
	Dependencies     map[string]plugable.IPlugAble `json:"-"`
	IsDefaultSc      bool                          `json:"-"`
	Status           string                        `json:"status,omitempty" description:"auto generated,do not input and all input will be ignore"`
}

func (c Ceph) IsEnable() bool {
	return c.Enable
}

func (c Ceph) GetName() string {
	return "Ceph"
}

func (c Ceph) GetStatus() string {
	return c.Status
}

func (c *Ceph) SetDependencies(deps map[string]plugable.IPlugAble) {
	c.Dependencies = deps
}
func (c *Ceph) GetDependencies() map[string]plugable.IPlugAble {
	return c.Dependencies
}

func (c *Ceph) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(c)
}

// no implementation
func (c *Ceph) GetLicenseLabel() uint16 {
	return 0
	//return constants.LLCeph
}

func (c *Ceph) DeploymentNamespaceTemplateRender() (string, error) {
	return template.New("ceph namespace template").Render(TemplateNamespace, c)
}

func (c *Ceph) DeploymentTemplateRender() (string, error) {
	if c.ReclaimPolicy == "" {
		c.ReclaimPolicy = "Delete"
	}
	return template.New("ceph template").Render(Template, c)
}

func (c *Ceph) SetImageRegistry(registry string) {
	c.ImageRegistry = registry
}

func (c *Ceph) GetMonitorIPList(monitorIPList []string) string {
	var str string
	for k, v := range monitorIPList {
		if k < len(monitorIPList)-1 {
			str = str + `"` + v + `"` + ","
		} else {
			str = str + `"` + v + `"`
		}
	}
	return str
}
