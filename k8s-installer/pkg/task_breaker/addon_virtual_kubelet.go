package task_breaker

import (
	"fmt"
	"reflect"

	"k8s-installer/components/virtual_kubelet"
	vkMock "k8s-installer/components/virtual_kubelet/mock"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/task_breaker/utils"

	//"k8s-installer/pkg/log"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
)

type AddOnVirtualKubelet struct {
	FullNodeList map[string]schema.ClusterNode
	Name         string
	NodeSuffix   string
	TaskTimeOut  int
	Enable       bool
}

func (virtualKubeletAddOn AddOnVirtualKubelet) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// implement you LivenessProbe step to detect you addons state
	return schema.Step{}, nil
}

func (virtualKubeletAddOn AddOnVirtualKubelet) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		virtualKubeletAddOn.Enable = false
	} else {
		virtualKubeletAddOn.Enable = plugAble.IsEnable()
	}
	return virtualKubeletAddOn
}

func (virtualKubeletAddOn AddOnVirtualKubelet) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (virtualKubeletAddOn AddOnVirtualKubelet) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return virtualKubeletAddOn.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func (virtualKubeletAddOn AddOnVirtualKubelet) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {

	if len(virtualKubeletAddOn.FullNodeList) == 0 {
		return nil
	}
	deleteNodeTask, applyRoleBindingTask, deployMockTask, err := createRemoveVKNodeKubeletNodeStep(cluster.Masters[0].NodeId, virtualKubeletAddOn.FullNodeList, cluster, virtualKubeletAddOn.NodeSuffix)
	if err != nil {
		return err
	}

	if action == constants.ActionCreate {

		// delete node that use virtual kubelet first
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "delete-node" + operation.Id,
			Name:      "delete-node",
			NodeSteps: []schema.NodeStep{deleteNodeTask},
		}

		// apply node role binding to ensure vk run correctly
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "apply-role-binding" + operation.Id,
			Name:      "apply-role-binding",
			NodeSteps: []schema.NodeStep{applyRoleBindingTask},
		}

		// then deploy vk
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "stop-normal-kubelet" + operation.Id,
			Name:      "stop-normal-kubelet",
			NodeSteps: deployMockTask,
		}
	} else if action == constants.ActionDelete {
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "stop-normal-kubelet" + operation.Id,
			Name:      "stop-normal-kubelet",
			NodeSteps: deployMockTask,
		}
	}
	return nil
}

func (virtualKubeletAddOn AddOnVirtualKubelet) GetAddOnName() string {
	return virtualKubeletAddOn.Name
}

func (virtualKubeletAddOn AddOnVirtualKubelet) IsEnable() bool {
	return virtualKubeletAddOn.Enable
}

func createVirtualKubeletStep(suffix string, operation *schema.Operation, cluster schema.Cluster, fullNodeList map[string]schema.ClusterNode) error {
	deleteNodeTask, applyRoleBindingTask, deployMockTask, err := createRemoveVKNodeKubeletNodeStep(cluster.Masters[0].NodeId, fullNodeList, cluster, suffix)
	if err != nil {
		return err
	}

	// delete node that use virtual kubelet first
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:        "delete-node" + operation.Id,
		Name:      "delete-node",
		NodeSteps: []schema.NodeStep{deleteNodeTask},
	}

	// apply node role binding to ensure vk run correctly
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:        "apply-role-binding" + operation.Id,
		Name:      "apply-role-binding",
		NodeSteps: []schema.NodeStep{applyRoleBindingTask},
	}

	// then deploy vk
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:        "stop-normal-kubelet" + operation.Id,
		Name:      "stop-normal-kubelet",
		NodeSteps: deployMockTask,
	}

	return nil
}

func createRemoveVKNodeKubeletNodeStep(firstMasterId string, nodes map[string]schema.ClusterNode, cluster schema.Cluster, suffix string) (schema.NodeStep, schema.NodeStep, []schema.NodeStep, error) {
	nodeToDelete := []string{}
	virtualKubeletStep := []schema.NodeStep{}
	var err error
	var configTemplate, systemdTemplate, bindingTemplate string
	index := 0
	for _, node := range nodes {
		hostname := generatorHostname(index, cluster.ClusterId, suffix)
		nodeToDelete = append(nodeToDelete, fmt.Sprintf("node %s", hostname))
		mock := schema.VKProviderMock{
			CpuLimit:    node.VirtualKubelet.VKProviderCaas.CpuLimit,
			MemoryLimit: node.VirtualKubelet.VKProviderCaas.MemoryLimit,
			PodLimit:    node.VirtualKubelet.VKProviderCaas.PodLimit,
		}
		deployVKMock := vkMock.NewDeployMock(hostname, mock)
		configTemplate, err = deployVKMock.TemplateVKLimitRender()
		if err != nil {
			return schema.NodeStep{}, schema.NodeStep{}, nil, err
		}

		systemdTemplate, err = deployVKMock.TemplateVKSystemdRender()
		if err != nil {
			return schema.NodeStep{}, schema.NodeStep{}, nil, err
		}

		bindingTemplate, err = deployVKMock.TemplateVKRoleBindingRender()
		if err != nil {
			return schema.NodeStep{}, schema.NodeStep{}, nil, err
		}

		step := schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskVirtualKubelet-" + node.NodeId,
			NodeID: node.NodeId,
			Tasks: map[int]schema.ITask{
				0: schema.TaskVirtualKubelet{
					TaskType:        constants.TaskTypeVirtualKubelet,
					TimeOut:         10,
					Provider:        node.VirtualKubelet.Provider,
					Config:          []byte(configTemplate),
					SystemdTemplate: []byte(systemdTemplate),
					Md5Dep:          GenerateDepMd5(virtual_kubelet.VKMockDep),
				},
			},
		}
		virtualKubeletStep = append(virtualKubeletStep, step)
		index += 1
	}

	stepDeleteNode := schema.NodeStep{
		Id:               utils.GenNodeStepID(),
		Name:             "TaskKubectl",
		NodeID:           firstMasterId,
		ServerMSGTimeOut: 5,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:     "TaskKubectl",
				TimeOut:      5,
				SubCommand:   constants.KubectlSubCommandDelete,
				CommandToRun: nodeToDelete,
			},
		},
	}

	stepApplyNodeRBAC := schema.NodeStep{
		Id:               utils.GenNodeStepID(),
		Name:             "TaskKubectl",
		NodeID:           firstMasterId,
		ServerMSGTimeOut: 5,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   "TaskKubectl",
				TimeOut:    5,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"nodeRoleBinding":   []byte(vkMock.ClusterRoleBinding2),
					"nodeIdRoleBinding": []byte(bindingTemplate),
				},
			},
		},
	}

	return stepDeleteNode, stepApplyNodeRBAC, virtualKubeletStep, nil
}
