package task_breaker

import (
	"fmt"
	"k8s-installer/pkg/cache"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
	"reflect"
)

type (
	AddDnsServerDeploy struct {
		Name        string
		TaskTimeOut int
		enable      bool
	}
)

func (params AddDnsServerDeploy) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	// when ks is enabled this addon should disable
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		params.enable = false
	} else {
		// do not look into license switch
		params.enable = plugAble.IsEnable()
	}
	return params
}

func (params AddDnsServerDeploy) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (params AddDnsServerDeploy) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return params.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if your addons is none container based app you should do something to remove it
		return nil
	}
}

func (params AddDnsServerDeploy) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if action == constants.ActionCreate {
		runtimeCache := cache.GetCurrentCache()

		for _, master := range cluster.Masters {
			masterInformation, err := runtimeCache.GetNodeInformation(master.NodeId)
			if err != nil {
				return err
			}

			if masterInformation == nil {
				return fmt.Errorf("Failed to find master with node id %s ", master.NodeId)
			}

			var nodeSteps []schema.NodeStep
			copyTo := map[string][]byte{}
			if coreFile, err := cluster.DnsServerDeploy.CorefileTemplateRender(); err != nil {
				return err
			} else {
				copyTo["/etc/coredns/Corefile"] = []byte(coreFile)
			}

			if pluginConfig, err := cluster.DnsServerDeploy.CaasConfigTemplateRender(masterInformation.SystemInfo.Node.Hostname, masterInformation.Ipv4DefaultIp); err != nil {
				return err
			} else {
				copyTo["/etc/coredns/config.yaml"] = []byte(pluginConfig)
			}
			nodeSteps = append(nodeSteps, schema.NodeStep{
				Id:     utils.GenNodeStepID(),
				Name:   "SetupDnsConfigs-" + master.NodeId,
				NodeID: master.NodeId,
				Tasks: map[int]schema.ITask{
					0: schema.TaskCopyTextBaseFile{
						TaskType:  constants.TaskTypeCopyTextFile,
						TimeOut:   3,
						TextFiles: copyTo,
					},
				},
			})
			operation.Step[len(operation.Step)] = &schema.Step{
				Id:                       "setupDnsConfigs-" + operation.Id,
				Name:                     "setupDnsConfigs",
				NodeSteps:                nodeSteps,
				OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
				OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			}
		}

		copyTo := map[string][]byte{}

		if staticTempalte, err := cluster.DnsServerDeploy.DnsStaticPodTemplateRender(); err != nil {
			return err
		} else {
			copyTo["/etc/kubernetes/manifests/caas-dns.yaml"] = []byte(staticTempalte)
		}

		var deployStaticPod []schema.NodeStep
		for _, master := range cluster.Masters {
			deployStaticPod = append(deployStaticPod, schema.NodeStep{
				Id:     utils.GenNodeStepID(),
				Name:   "SetupDnsStaticPod-" + master.NodeId,
				NodeID: master.NodeId,
				Tasks: map[int]schema.ITask{
					0: schema.TaskCopyTextBaseFile{
						TaskType:  constants.TaskTypeCopyTextFile,
						TimeOut:   3,
						TextFiles: copyTo,
					},
				},
			})
		}

		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "SetupDnsStaticPod-" + operation.Id,
			Name:                     "SetupDnsStaticPod",
			NodeSteps:                deployStaticPod,
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}

	} else if action == constants.ActionDelete {
	}
	return nil
}

func (params AddDnsServerDeploy) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// implement you LivenessProbe step to detect you addons state
	return schema.Step{
		Id:                       "stepCheckKubesphere-" + operationId,
		Name:                     "StepCheckKubesphere",
		NodeSteps:                []schema.NodeStep{},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
	}, nil
}

func (params AddDnsServerDeploy) GetAddOnName() string {
	return params.Name
}

func (params AddDnsServerDeploy) IsEnable() bool {
	return params.enable
}
