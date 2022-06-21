package nginx

import (
	"fmt"

	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
)

const (
	DefaultNamespace = "ingress-nginx"
)

type DeployNginxIngressController struct {
	ImageRegistry                 string   // image registry
	Namespace                     string   // k8s namespace
	Replicas                      int      // replicas of nginx ingress controller pod
	NodeAffinityBlockList         []string //
	NginxIngressControllerVersion string   // version of nginx ingress controller
	CertGenWebhookVersion         string   // version of cert gen webhook
}

func NewNginxIngressDeploy(registry string, ingressNodeLabels []string) DeployNginxIngressController {
	return DeployNginxIngressController{
		ImageRegistry: registry,
		Replicas:      len(ingressNodeLabels),
		NodeAffinityBlockList: func(labels []string) []string {
			for i, l := range labels {
				labels[i] = fmt.Sprintf("%s- %s", ValueIndent, l)
				if i != len(labels)-1 {
					labels[i] += "\n"
				}
			}
			return labels
		}(ingressNodeLabels),
		NginxIngressControllerVersion: NginxIngressControllerVer,
		CertGenWebhookVersion:         CertGenWebhookVer,
	}
}

/**
 * Complete empty options with default settings.
 */
func (dn DeployNginxIngressController) CompleteNginxIngressControllerDeploy() DeployNginxIngressController {
	dn.Namespace = util.StringDefaultIfNotSet(dn.Namespace, DefaultNamespace)
	return dn
}

/**
 * Renderer out target yaml string from template.
 */
func (dn DeployNginxIngressController) TemplateRender() (string, error) {
	return template.New("nginx ingress controller").Render(Template, dn.CompleteNginxIngressControllerDeploy())
}
