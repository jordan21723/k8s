package task_breaker

import (
	"encoding/base64"
	"fmt"
	"k8s-installer/components/cloud_provider/csi/cinder"
	"k8s-installer/pkg/cache"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/pkg/util"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
	"reflect"
	"strings"
)

type AddOnOpenStackCloudProvider struct {
	FullNodeList map[string]schema.NodeInformation
	Name         string
	TaskTimeOut  int
	enable       bool
}

func (openstackCloudProvider AddOnOpenStackCloudProvider) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// implement you LivenessProbe step to detect you addons state
	return schema.Step{}, nil
}

func (openstackCloudProvider AddOnOpenStackCloudProvider) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		openstackCloudProvider.enable = false
	} else {
		// do not look into license switch
		openstackCloudProvider.enable = plugAble.IsEnable()
	}
	return openstackCloudProvider
}

func (openstackCloudProvider AddOnOpenStackCloudProvider) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	if nodeAction == constants.ActionCreate {

		if cluster.CloudProvider.OpenStack.Enable {
			usingTls := strings.HasPrefix(cluster.CloudProvider.OpenStack.AuthURL, "https")

			runtimeCache := cache.GetCurrentCache()

			fullNodesCollection, err := runtimeCache.GetNodeInformationCollection()
			addNodesCollection := map[string]schema.NodeInformation{}

			if err != nil {
				return fmt.Errorf("Openstack Cloud Provider: Failed to all nodes list due to error: %s ", err.Error())
			}

			for _, node := range nodesToAddOrRemove {
				if nodeFound, found := fullNodesCollection[node.NodeId]; !found {
					return fmt.Errorf("Openstack Cloud Provider: Failed to find node with id: %s ", node.NodeId)
				} else {
					addNodesCollection[node.NodeId] = nodeFound
				}
			}

			stepCopyCloudConfigStep := &schema.Step{
				Id:        "step-copy-cloud-config" + operation.Id,
				Name:      "copy-cloud-config",
				NodeSteps: createCopyOpenStackCloudConfigNodeStep(addNodesCollection, cluster, usingTls),
			}
			operation.Step[len(operation.Step)] = stepCopyCloudConfigStep

			// run update-ca command
			if usingTls {
				stepAcceptCA := &schema.Step{
					Id:        "step-openstack-accept-ca" + operation.Id,
					Name:      "runCommand",
					NodeSteps: createRunUpdateCANodeStep(addNodesCollection, cluster, true),
				}
				operation.Step[len(operation.Step)] = stepAcceptCA
			}
		}
	}
	return nil
}

func (openstackCloudProvider AddOnOpenStackCloudProvider) GetAddOnName() string {
	return openstackCloudProvider.Name
}

func (openstackCloudProvider AddOnOpenStackCloudProvider) IsEnable() bool {
	return openstackCloudProvider.enable
}

func (openstackCloudProvider AddOnOpenStackCloudProvider) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return openstackCloudProvider.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func (openstackCloudProvider AddOnOpenStackCloudProvider) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if action == constants.ActionCreate {
		privateRegistry := util.StringDefaultIfNotSet(joinPrivateRegistry(cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort), "docker.io")

		if cluster.CloudProvider.OpenStack.Enable {
			usingTls := strings.HasPrefix(cluster.CloudProvider.OpenStack.AuthURL, "https")

			stepCopyCloudConfigStep := &schema.Step{
				Id:        "step-copy-cloud-config" + operation.Id,
				Name:      "copy-cloud-config",
				NodeSteps: createCopyOpenStackCloudConfigNodeStep(openstackCloudProvider.FullNodeList, cluster, usingTls),
			}
			operation.Step[len(operation.Step)] = stepCopyCloudConfigStep

			// run update-ca command
			if usingTls {
				stepAcceptCA := &schema.Step{
					Id:        "step-openstack-accept-ca" + operation.Id,
					Name:      "runCommand",
					NodeSteps: createRunUpdateCANodeStep(openstackCloudProvider.FullNodeList, cluster, true),
				}
				operation.Step[len(operation.Step)] = stepAcceptCA
			}

			openstackNodeStep, err := createOpenStackNodeStep(cluster, cluster.Masters[0].NodeId, privateRegistry, config.TaskTimeOut.TaskCRI)
			if err != nil {
				log.Errorf("Failed to render cinder csi node step due to error: %s", err.Error())
				log.Error("Skip cinder csi node step set up !!!")
				return err
			} else {
				stepOpenStack := &schema.Step{
					Id:        "step-openstack-" + operation.Id,
					Name:      "kubectl",
					NodeSteps: []schema.NodeStep{openstackNodeStep},
				}
				operation.Step[len(operation.Step)] = stepOpenStack
			}
		}
	} else if action == constants.ActionDelete {
	}

	return nil
}

func createOpenStackCloudConfig(cluster schema.Cluster, usingTls bool) (map[string][]byte, error) {
	fileToCopy := map[string][]byte{}
	cinderDeploy := cinder.NewDeployCinder("", *cluster.CloudProvider.OpenStack)
	cloudConfig, err := cinderDeploy.TemplateRender("cloudConfig", cinder.TemplateCloudConfig)
	if err != nil {
		log.Errorf("Failed to parse to cloudConfig due to error %s", err.Error())
		return nil, err
	}
	fileToCopy["/etc/kubernetes/cloud.conf"] = []byte(cloudConfig)

	// if keystone use tls we should copy full chain to each node
	if usingTls {
		fileToCopy["/etc/pki/ca-trust/source/anchors/openstack.crt"] = []byte(cluster.CloudProvider.OpenStack.CaCert)
	}
	return fileToCopy, nil
}

func createCopyOpenStackCloudConfigNodeStep(fullNodeList map[string]schema.NodeInformation, cluster schema.Cluster, usingTls bool) []schema.NodeStep {
	var nodeSteps []schema.NodeStep
	tasks, err := createOpenStackCloudConfig(cluster, usingTls)
	if err != nil {
		return []schema.NodeStep{}
	}
	for nodeId := range fullNodeList {
		nodeSteps = append(nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskTypeCopyTextFile-" + nodeId,
			NodeID: nodeId,
			Tasks: map[int]schema.ITask{
				0: schema.TaskCopyTextBaseFile{
					TaskType:  constants.TaskTypeCopyTextFile,
					TimeOut:   3,
					TextFiles: tasks,
				},
			},
		})
	}
	return nodeSteps
}

func createRunUpdateCANodeStep(fullNodeList map[string]schema.NodeInformation, cluster schema.Cluster, usingTls bool) []schema.NodeStep {
	var nodeSteps []schema.NodeStep
	for nodeId := range fullNodeList {
		nodeSteps = append(nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskTypeCRI-" + nodeId,
			NodeID: nodeId,
			Tasks: map[int]schema.ITask{
				0: schema.TaskRunCommand{
					TaskType: constants.TaskTypeRunCommand,
					TimeOut:  10,
					Commands: map[int][]string{
						0: {"update-ca-trust"},
					},
				},
			},
		})
	}
	return nodeSteps
}

func createOpenStackNodeStep(cluster schema.Cluster, firstMasterNodeId, registry string, timeOut int) (schema.NodeStep, error) {
	var controllerPluginTemplate, nodePluginTemplate, storageClassTemplate, csiSecretTemplate, cloudConfigTemplate string
	var err error
	cinderDeploy := cinder.NewDeployCinder(registry, *cluster.CloudProvider.OpenStack)
	controllerPluginTemplate, err = cinderDeploy.TemplateRender("controllerPluginTemplate", cinder.TemplateControlPlane)
	if err != nil {
		return schema.NodeStep{}, err
	}
	nodePluginTemplate, err = cinderDeploy.TemplateRender("nodePluginTemplate", cinder.TemplateNodePlugin)
	if err != nil {
		return schema.NodeStep{}, err
	}
	storageClassTemplate, err = cinderDeploy.TemplateRender("storageClassTemplate", cinder.TemplateStorageClass)
	if err != nil {
		return schema.NodeStep{}, err
	}
	cloudConfigTemplate, err = cinderDeploy.TemplateRender("cloudConfig", cinder.TemplateCloudConfig)
	if err != nil {
		return schema.NodeStep{}, err
	}

	result := base64.URLEncoding.EncodeToString([]byte(cloudConfigTemplate))
	// we need base64 encode it for secret
	csiSecretTemplate = fmt.Sprintf(cinder.TemplateCSISecret, result)

	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + firstMasterNodeId,
		NodeID: firstMasterNodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"cinder-csi-controllerplugin-rbac": []byte(cinder.ControllerPluginRBAC),
					"cinder-csi-nodeplugin-rbac":       []byte(cinder.NodePluginRBAC),
					"csi-cinder-driver":                []byte(cinder.CsiDriver),
					"csi-secret-cinderplugin":          []byte(csiSecretTemplate),
					"cinder-csi-controllerplugin":      []byte(controllerPluginTemplate),
					"cinder-csi-nodeplugin":            []byte(nodePluginTemplate),
					"cinder-storageclass":              []byte(storageClassTemplate),
				},
				TimeOut: timeOut,
			},
		},
	}

	return step, nil
}
