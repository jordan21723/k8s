package task_breaker

import (
	"bytes"
	"fmt"
	restarter "k8s-installer/components/auto_restarter"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/pkg/util/netutils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
	"reflect"
)

type AddOnAutoRestarter struct {
	Name        string
	TaskTimeOut int
	enable      bool
}

func (a AddOnAutoRestarter) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	coreDNSAddr, err := netutils.GetCoreDNSAddr(cluster.ControlPlane.ServiceV4CIDR)
	if err != nil {
		return schema.Step{}, err
	}

	nodeId := cluster.Masters[0].NodeId

	task := &schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeCurl" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskCurl{
				TaskType:   constants.TaskTypeCurl,
				TimeOut:    5,
				URL:        fmt.Sprintf("http://auto-restarter.%s.svc.%s/healthz", cluster.AutoRestarter.Namespace, cluster.ClusterName),
				Method:     "GET",
				NameServer: coreDNSAddr,
			},
		},
	}
	return schema.Step{
		Id:        fmt.Sprintf("stepCheck%s-%s", cluster.AutoRestarter.GetName(), operationId),
		Name:      "StepCheck" + cluster.AutoRestarter.GetName(),
		NodeSteps: []schema.NodeStep{*task},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int,
				returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				if reply.Stat == constants.StatusSuccessful &&
					reply.ReturnData["status_code"] == "200" {
					cluster.AutoRestarter.Status = constants.StateReady
					return
				}
				cluster.AutoRestarter.Status = constants.StateNotReady
			},
			nil,
		},
		OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
			func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation, nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
				cluster.AutoRestarter.Status = constants.StateNotReady
			},
			nil,
		},
	}, nil
}

func (params AddOnAutoRestarter) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		params.enable = false
	} else {
		// do not look into license switch
		params.enable = plugAble.IsEnable()
	}
	return params
}

func (a AddOnAutoRestarter) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (a AddOnAutoRestarter) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return a.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func (a AddOnAutoRestarter) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	// Get the first Node ID of cluster.
	nodeId := cluster.Masters[0].NodeId
	if action == constants.ActionCreate {
		task, err := a.taskForSettingUp(nodeId, joinPrivateRegistry(
			cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort,
		), cluster.AutoRestarter)
		if err != nil {
			return err
		}
		operation.Step[len(operation.Step)] = &schema.Step{
			// Id: "stepSetup" + cluster.AutoRestarter.GetName(),
			Id:            fmt.Sprintf("stepSetup%s-%s", cluster.AutoRestarter.GetName(), operation.Id),
			Name:          "StepSetup" + cluster.AutoRestarter.GetName(),
			NodeSteps:     []schema.NodeStep{*task},
			WaitBeforeRun: 20,
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
				OnStepReturnCustomerHandler: func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
					log.Errorf("Node %s in step %s result in error state but we are going to ignore it.", reply.NodeId, step.Name)
				},
			},
			OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
				OnStepTimeoutCustomerHandler: func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation, nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
					log.Errorf("Node %s in step %s result in timeout state but we are going to ignore it.", nodeStepId, step.Name)
				},
			},
		}
	} else if action == constants.ActionDelete {
		a.taskForDeleting(nodeId)
	}
	return nil
}

func (a AddOnAutoRestarter) taskForSettingUp(nodeId, registry string, da *restarter.DeployAutoRestarter) (*schema.NodeStep, error) {
	da.SetImageRegistry(registry)
	tmpl, err := da.TemplateRender()
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBufferString(tmpl)
	step := &schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"auto-restarter": buffer.Bytes(),
				},
				TimeOut: a.TaskTimeOut,
			},
		},
	}
	return step, nil
}

func (a AddOnAutoRestarter) taskForDeleting(nodeId string) *schema.NodeStep {
	step := &schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:     constants.TaskTypeKubectl,
				SubCommand:   constants.KubectlSubCommandDelete,
				CommandToRun: []string{""},
				TimeOut:      a.TaskTimeOut,
			},
		},
	}
	return step
}

func (a AddOnAutoRestarter) GetAddOnName() string {
	return a.Name
}

func (a AddOnAutoRestarter) IsEnable() bool {
	return a.enable
}

func (a *AddOnAutoRestarter) GetLicenseLabel() uint16 {
	return 1
}
