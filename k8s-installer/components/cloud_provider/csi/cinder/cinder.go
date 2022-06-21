package cinder

import (
	"strings"

	"k8s-installer/pkg/util"
	"k8s-installer/schema"

	"k8s-installer/pkg/template"
)

type DeployCinder struct {
	KeyStoneEnableTLS bool
	ImageRegistry     string
	BackendType       string
	AvailabilityZone  string
	StorageClassName  string
	ReclaimPolicy     string
	Username          string
	Password          string
	AuthURL           string
	ProjectId         string
	DomainId          string
	Region            string
	IsDefaultSc       bool
}

func NewDeployCinder(registry string, openstack schema.CloudProviderOpenStack) DeployCinder {
	return DeployCinder{
		KeyStoneEnableTLS: strings.HasPrefix(openstack.AuthURL, "https"),
		ImageRegistry:     registry,
		BackendType:       openstack.Cinder.BackendType,
		AvailabilityZone:  openstack.Cinder.AvailabilityZone,
		StorageClassName:  util.StringDefaultIfNotSet(openstack.Cinder.StorageClassName, "cinder"),
		ReclaimPolicy:     util.StringDefaultIfNotSet(openstack.Cinder.ReclaimPolicy, "Delete"),
		Username:          openstack.Username,
		Password:          openstack.Password,
		AuthURL:           openstack.AuthURL,
		ProjectId:         openstack.ProjectId,
		DomainId:          openstack.DomainId,
		Region:            openstack.Region,
		IsDefaultSc:       openstack.Cinder.IsDefaultSc,
	}
}

func (dc DeployCinder) TemplateRender(templateType, templateVal string) (string, error) {
	return template.New(templateType).Render(templateVal, dc)
}
