package task_breaker

import (
	nfsDeploy "k8s-installer/components/storage/nfs"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
	"reflect"
)

type AddOnStorageNFS struct {
	Name        string
	TaskTimeOut int
	enable      bool
}

func (nfs AddOnStorageNFS) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// implement you LivenessProbe step to detect you addons state
	return schema.Step{}, nil
}

func (nfs AddOnStorageNFS) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		nfs.enable = false
	} else {
		// do not look into license switch
		nfs.enable = plugAble.IsEnable()
	}
	return nfs
}

func (nfs AddOnStorageNFS) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (nfs AddOnStorageNFS) GetAddOnName() string {
	return nfs.Name
}

func (nfs AddOnStorageNFS) IsEnable() bool {
	return nfs.enable
}

func (nfs AddOnStorageNFS) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return nfs.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func (nfs AddOnStorageNFS) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
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
								"nfs-namespace": []byte(nfsDeploy.NamespaceTemplate),
							},
							TimeOut: nfs.TaskTimeOut,
						},
					},
				},
			},
		}
	} else if action == constants.ActionDelete {

	}

	// during cluster deleting we don`t need to do anything
	// because the cluster will be destroyed
	if cluster.Action != constants.ActionDelete {
		stepDeployNFSProvisioner, err := nfs.createNFSPrepareNodeStep(cluster, action)
		if err != nil {
			return err
		}
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "step-setup-nfs" + operation.Id,
			Name:      "step-setup-nfs",
			NodeSteps: []schema.NodeStep{stepDeployNFSProvisioner},
		}
	}

	return nil
}

func (nfs AddOnStorageNFS) createNFSPrepareNodeStep(cluster schema.Cluster, action string) (schema.NodeStep, error) {
	cluster.Storage.NFS.ImageRegistry = joinPrivateRegistry(cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort)

	deploymentTemplate, err := cluster.Storage.NFS.DeploymentTemplateRender()
	if err != nil {
		return schema.NodeStep{}, err
	}
	storageClassTemplate, errSC := cluster.Storage.NFS.StorageClassTemplateRender()
	if errSC != nil {
		return schema.NodeStep{}, errSC
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
					"nfs-rbac":          []byte(nfsDeploy.RBAC),
					"nfs-provisioner":   []byte(deploymentTemplate),
					"nfs-storage-class": []byte(storageClassTemplate),
				},
				TimeOut: nfs.TaskTimeOut,
			},
		},
	}
	return step, nil
}
