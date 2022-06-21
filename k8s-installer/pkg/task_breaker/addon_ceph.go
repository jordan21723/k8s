package task_breaker

import (
	"k8s-installer/components/storage/ceph"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
	"reflect"
)

type AddOnStorageCeph struct {
	Name        string
	TaskTimeOut int
	enable      bool
}

func (c AddOnStorageCeph) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// implement you LivenessProbe step to detect you addons state
	return schema.Step{}, nil
}

func (c AddOnStorageCeph) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		c.enable = false
	} else {
		// do not look into license switch
		c.enable = plugAble.IsEnable()
	}
	return c
}

func (c AddOnStorageCeph) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (c AddOnStorageCeph) GetAddOnName() string {
	return c.Name
}

func (c AddOnStorageCeph) IsEnable() bool {
	return c.enable
}

func (c AddOnStorageCeph) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return c.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func (c AddOnStorageCeph) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if action == constants.ActionCreate {
		// create namespace first
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:   "step-create-namespace" + operation.Id,
			Name: "step-create-namespaces",
			NodeSteps: []schema.NodeStep{
				{
					Id:     utils.GenNodeStepID(),
					Name:   "TaskTypeKubectl-" + cluster.Masters[0].NodeId,
					NodeID: cluster.Masters[0].NodeId,
					Tasks: map[int]schema.ITask{
						0: schema.TaskKubectl{
							TaskType:   constants.TaskTypeKubectl,
							SubCommand: constants.KubectlSubCommandCreateOrApply,
							YamlTemplate: map[string][]byte{
								"ceph-namespace": []byte(ceph.TemplateNamespace),
							},
							TimeOut: c.TaskTimeOut,
						},
					},
				},
			},
		}
	} else if action == constants.ActionDelete {
		// TODO
	}

	// during cluster deleting we don`t need to do anything
	// because the cluster will be destroyed
	if cluster.Action != constants.ActionDelete {
		stepDeployCephProvisioner, err := c.createCephPrepareNodeStep(cluster, action)
		if err != nil {
			return err
		}
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "step-setup-ceph" + operation.Id,
			Name:      "step-setup-ceph",
			NodeSteps: []schema.NodeStep{stepDeployCephProvisioner},
		}
	}

	return nil
}

func (c AddOnStorageCeph) createCephPrepareNodeStep(cluster schema.Cluster, action string) (schema.NodeStep, error) {
	cluster.Storage.Ceph.ImageRegistry = joinPrivateRegistry(cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort)

	deploymentTemplate, err := cluster.Storage.Ceph.DeploymentTemplateRender()
	if err != nil {
		return schema.NodeStep{}, err
	}

	// handler kubectl sub command either apply -f or delete -f
	command := ""
	if action == constants.ActionCreate {
		command = constants.KubectlSubCommandCreateOrApply
	} else {
		command = constants.KubectlSubCommandDelete
	}

	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + cluster.Masters[0].NodeId,
		NodeID: cluster.Masters[0].NodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: command,
				YamlTemplate: map[string][]byte{
					"ceph": []byte(deploymentTemplate),
				},
				TimeOut: c.TaskTimeOut,
			},
		},
	}
	return step, nil
}
