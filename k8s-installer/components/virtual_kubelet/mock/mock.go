package mock

import (
	"k8s-installer/pkg/template"
	"k8s-installer/schema"
)

type DeployMock struct {
	Hostname  string                `json:"hostname,omitempty"`
	Mock      schema.VKProviderMock `json:"mock,omitempty"`
}

func NewDeployMock(hostname string, mock schema.VKProviderMock) DeployMock {
	return DeployMock{
		Hostname:  hostname,
		Mock:      mock,
	}
}

/**
 * Renderer out target yaml string from template.
 */
func (dc DeployMock) TemplateVKLimitRender() (string, error) {
	return template.New("vk-mock").Render(mockConfig, dc)
}

func (dc DeployMock) TemplateVKRoleBindingRender() (string, error) {
	return template.New("vk-mock").Render(clusterRoleBinding1, dc)
}

func (dc DeployMock) TemplateVKSystemdRender() (string, error) {
	return template.New("vk-mock").Render(systemdTemplate, dc)
}