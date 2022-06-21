package Calico

import (

	// "text/template"

	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
	"k8s-installer/schema"
)

type DeployCalico struct {
	ImageRegistry   string // image registry
	Version         string // calico version
	CNIVer          string // cni version
	EnableIPv4      bool   // if enable IPv4
	V4DetectMethod  string
	PodV4CIDR       string // pod IPv4 CIDR
	Mode            string // if enable IPIP
	EnableDualStack bool   // if enable IPv4 + IPv6
	EnableIPv6      bool   // if enable IPv6
	V6DetectMethod  string
	PodV6CIDR       string // pod IPv6 CIDR
	IPam            bool   // IP address management
	MTU             int
}

func NewCalicoDeploy(registry string, cni schema.CNI) DeployCalico {
	return DeployCalico{
		ImageRegistry: registry,
		Version:       CalicoVer,
		//CNIVer:          util.StringDefaultIfNotSet(cni.CniVersion, CNIVer),
		CNIVer:          CNIVer,
		EnableIPv4:      true, // enable IPv4 by default
		V4DetectMethod:  util.StringDefaultIfNotSet(cni.Calico.IPAutoDetection, constants.CalicoV4DetectMethodAuto),
		PodV4CIDR:       cni.PodV4CIDR,
		Mode:            cni.Calico.CalicoMode,
		IPam:            true, // enable IPAM by default
		EnableDualStack: cni.Calico.EnableDualStack,
		EnableIPv6:      cni.Calico.EnableDualStack,
		PodV6CIDR:       cni.PodV6CIDR,
		V6DetectMethod:  util.StringDefaultIfNotSet(cni.Calico.IP6AutoDetection, constants.CalicoV6DetectMethodAuto),
		MTU:             util.IntDefaultIfZero(cni.MTU, 1440),
	}
}

/**
 * Renderer out target yaml string from template.
 */
func (dc DeployCalico) TemplateRender() (string, error) {
	return template.New("calico").Render(Template, dc)
}
