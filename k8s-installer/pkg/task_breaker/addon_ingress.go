package task_breaker

import (
	"reflect"

	cm "k8s-installer/components/cert_manager"
	"k8s-installer/components/ingress/nginx"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"

	"github.com/google/uuid"
)

type (
	AddOnIngress struct {
		IngressNodes          []schema.ClusterNode
		NodeToHostNameMapping map[string]string
		FirstMasterId         string
		Name                  string
		TaskTimeOut           int
		enable                bool
	}
)

func (params AddOnIngress) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// implement you LivenessProbe step to detect you addons state
	// do not add your step to operation
	return schema.Step{
		Id:   "step-" + uuid.New().String(),
		Name: "Ingress Setup Liveness Probe",
	}, nil
}

func (params AddOnIngress) GetAddOnName() string {
	return params.Name
}

func (params AddOnIngress) IsEnable() bool {
	return params.enable
}

func (params AddOnIngress) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		params.enable = false
	} else {
		// do not look into license switch
		params.enable = plugAble.IsEnable()
	}
	return params
}

func (params AddOnIngress) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (params AddOnIngress) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return params.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func (params AddOnIngress) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if action == constants.ActionCreate {
		if len(params.IngressNodes) == 0 {
			return nil
		}
		registry := joinPrivateRegistry(cluster.ContainerRuntime.PrivateRegistryAddress,
			cluster.ContainerRuntime.PrivateRegistryPort)

		step, err := params.setupInstallCertManagerTask(registry, cm.DeployCertManager{})
		if err != nil {
			return err
		}
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "setupInstallCertManagerTask" + operation.Id,
			Name:      "setupInstallCertManagerTask",
			NodeSteps: []schema.NodeStep{step},
		}

		step, err = params.setupInstallNginxIngressController(registry)
		if err != nil {
			return err
		}
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "setupInstallNginxIngressController" + operation.Id,
			Name:      "setupInstallNginxIngressController",
			NodeSteps: []schema.NodeStep{step},
		}
		stepWait := CreateWaitStep(120, "waiting-for-ingress-controller")
		operation.Step[len(operation.Step)] = &stepWait

		return nil
	} else if action == constants.ActionDelete {
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "setupUninstallNginxIngressController" + operation.Id,
			Name:      "setupUninstallNginxIngressController",
			NodeSteps: []schema.NodeStep{params.setupUninstallIngressController()},
		}
	}
	return nil
}

func (params AddOnIngress) setupInstallNginxIngressController(registry string) (schema.NodeStep, error) {
	ingressNodeLabelsList := make([]string, len(params.IngressNodes))
	for i, node := range params.IngressNodes {
		ingressNodeLabelsList[i] = params.NodeToHostNameMapping[node.NodeId]
	}
	dn := nginx.NewNginxIngressDeploy(registry, ingressNodeLabelsList)
	tmpl, err := dn.TemplateRender()
	if err != nil {
		return schema.NodeStep{}, err
	}
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + params.FirstMasterId,
		NodeID: params.FirstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"ingress": []byte(tmpl),
				},
				TimeOut: params.TaskTimeOut,
			},
		},
	}
	return step, nil
}

func (params AddOnIngress) setupInstallCertManagerTask(registry string, dc cm.DeployCertManager) (schema.NodeStep, error) {
	dc.SetImageRegistry(registry)
	tmpl, err := dc.TemplateRender()
	if err != nil {
		return schema.NodeStep{}, err
	}
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + params.FirstMasterId,
		NodeID: params.FirstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"cert-manager": []byte(tmpl),
				},
				TimeOut: params.TaskTimeOut,
			},
		},
	}
	return step, nil
}

func (params AddOnIngress) setupUninstallIngressController() schema.NodeStep {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + params.FirstMasterId,
		NodeID: params.FirstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:     constants.TaskTypeKubectl,
				SubCommand:   constants.KubectlSubCommandDelete,
				CommandToRun: []string{"ns " + nginx.DefaultNamespace, "ns " + cm.DefaultNamespace},
				TimeOut:      params.TaskTimeOut,
			},
		},
	}
	return step
}
