package schema

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/schema/plugable"
)

type ClusterLB struct {
	Enable       bool                          `json:"enable" description:"enable or disable"`
	Nodes        []ClusterNode                 `json:"nodes" validate:"required,gt=1" description:"related nodes"`
	VIP          string                        `json:"vip" validate:"required" description:"external lb vip"`
	Dependencies map[string]plugable.IPlugAble `json:"-"`
	Status       string                        `json:"status,omitempty" description:"auto generated,do not input and all input will be ignore"`
	RouterID     *int                          `json:"router_id,omitempty" description:"auto generated,do not input and all input will be ignore"`
}

func (c *ClusterLB) GetName() string {
	return "Cluster Load Balancer"
}

func (c *ClusterLB) GetStatus() string {
	return c.Status
}

func (c *ClusterLB) IsEnable() bool {
	return c.Enable
}

func (c *ClusterLB) SetDependencies(deps map[string]plugable.IPlugAble) {}

func (c *ClusterLB) GetDependencies() map[string]plugable.IPlugAble {
	return c.Dependencies
}

func (c *ClusterLB) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(c)
}

func (c *ClusterLB) GetLicenseLabel() uint16 {
	return constants.LLClusterLB
}

func (c *ClusterLB) GetAllNodeIDs() []string {
	ids := make([]string, len(c.Nodes))
	for i, node := range c.Nodes {
		ids[i] = node.NodeId
	}
	return ids
}
