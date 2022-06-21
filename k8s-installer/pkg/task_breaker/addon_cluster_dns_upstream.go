package task_breaker

import (
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
	"reflect"
	"strings"
)

type (
	AddClusterDnsUpstream struct {
		Name         string
		TaskTimeOut  int
		enable       bool
		FullNodeList map[string]schema.NodeInformation
	}
)

func (params AddClusterDnsUpstream) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	// when ks is enabled this addon should disable
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		params.enable = false
	} else {
		// do not look into license switch
		params.enable = plugAble.IsEnable()
	}
	return params
}

func (params AddClusterDnsUpstream) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (params AddClusterDnsUpstream) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return params.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if your addons is none container based app you should do something to remove it
		return nil
	}
}

func (params AddClusterDnsUpstream) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if action == constants.ActionCreate {
		pluginConfig, err := cluster.ClusterDnsUpstream.ClusterDnsTemplateRender(cluster.ClusterName)
		if err != nil {
			return err
		}
		copyTo := map[string][]byte{}
		copyTo["/etc/coredns/PatchClusterDnsUpstream.yaml"] = []byte(pluginConfig)

		nodeStep := schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "CopyClusterDNSPatchConfig-" + cluster.Masters[0].NodeId,
			NodeID: cluster.Masters[0].NodeId,
			Tasks: map[int]schema.ITask{
				0: schema.TaskCopyTextBaseFile{
					TaskType:  constants.TaskTypeCopyTextFile,
					TimeOut:   3,
					TextFiles: copyTo,
				},
			},
		}

		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "CopyClusterDNSPatchConfig-" + operation.Id,
			Name:                     "CopyClusterDNSPatchConfig",
			NodeSteps:                []schema.NodeStep{nodeStep},
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}

		applyClusterCorednsConfigCmdd := "kubectl apply -f /etc/coredns/PatchClusterDnsUpstream.yaml"
		applyClusterCorednsConfigStep := CommonRunWithSingleNode(cluster.Masters[0].NodeId, "ShrinkCorednsReplicasToZero", cluster.Action, 10, strings.Split(applyClusterCorednsConfigCmdd, " "), false)
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "applyClusterCorednsConfig-" + operation.Id,
			Name:                     "applyClusterCorednsConfig",
			NodeSteps:                []schema.NodeStep{applyClusterCorednsConfigStep},
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}

		ShrinkClusterDnsInstCmd := "kubectl scale deploy coredns -n kube-system --replicas=0"
		ShrinkClusterDnsInstStep := CommonRunWithSingleNode(cluster.Masters[0].NodeId, "ShrinkCorednsReplicasToZero", cluster.Action, 10, strings.Split(ShrinkClusterDnsInstCmd, " "), false)

		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "ShrinkCorednsReplicasToZero-" + operation.Id,
			Name:                     "ShrinkCorednsReplicasToZero",
			NodeSteps:                []schema.NodeStep{ShrinkClusterDnsInstStep},
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}

		ResumeClusterDnsInstCmd := "kubectl scale deploy coredns -n kube-system --replicas=2"
		ResumeClusterDnsInstStep := CommonRunWithSingleNode(cluster.Masters[0].NodeId, "ShrinkCorednsReplicasToZero", cluster.Action, 10, strings.Split(ResumeClusterDnsInstCmd, " "), false)

		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "ResumeClusterDnsInst-" + operation.Id,
			Name:                     "ResumeClusterDnsInst",
			NodeSteps:                []schema.NodeStep{ResumeClusterDnsInstStep},
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			WaitBeforeRun:            10,
		}
	} else if action == constants.ActionDelete {
	}
	return nil
}

func (params AddClusterDnsUpstream) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// implement you LivenessProbe step to detect you addons state
	return schema.Step{
		Id:                       "stepCheckKubesphere-" + operationId,
		Name:                     "StepCheckKubesphere",
		NodeSteps:                []schema.NodeStep{},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
	}, nil
}

func (params AddClusterDnsUpstream) GetAddOnName() string {
	return params.Name
}

func (params AddClusterDnsUpstream) IsEnable() bool {
	return params.enable
}
