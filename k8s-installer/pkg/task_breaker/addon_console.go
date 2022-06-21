package task_breaker

import (
	"fmt"
	"reflect"

	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"

	"k8s-installer/schema/plugable"

	cs "k8s-installer/components/console"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/util/netutils"
	"k8s-installer/schema"
)

type AddOnConsole struct {
	Name        string
	TaskTimeOut int
	enable      bool
}

func (a AddOnConsole) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// implement you LivenessProbe step to detect you addons state
	coreDNSAddr, err := netutils.GetCoreDNSAddr(cluster.ControlPlane.ServiceV4CIDR)
	if err != nil {
		return schema.Step{}, err
	}

	task := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeCurl" + cluster.Masters[0].NodeId,
		NodeID: cluster.Masters[0].NodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskCurl{
				TaskType:   constants.TaskTypeCurl,
				TimeOut:    5,
				URL:        fmt.Sprintf("http://console.%s.svc.%s/", cluster.Console.Namespace, cluster.ClusterName),
				Method:     "GET",
				NameServer: coreDNSAddr,
			},
		},
	}
	return schema.Step{
		Id:        fmt.Sprintf("stepCheck%s-%s", cluster.Console.GetName(), operationId),
		Name:      "StepCheck" + cluster.Console.GetName(),
		NodeSteps: []schema.NodeStep{task},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int,
				returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				if reply.Stat == constants.StatusSuccessful &&
					// redirect
					reply.ReturnData["status_code"] == "302" {
					cluster.PostgresOperator.Status = constants.StateReady
					return
				}
				cluster.PostgresOperator.Status = constants.StateNotReady
			},
			nil,
		},
		OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
			func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation, nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
				cluster.PostgresOperator.Status = constants.StateNotReady
			},
			nil,
		},
	}, nil
}

func (a AddOnConsole) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	// when ks is enabled this addon should disable
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() || cluster.ClusterRole == constants.ClusterRoleHost || cluster.ClusterRole == constants.ClusterRoleMember {
		a.enable = false
	} else {
		if plugAble.GetLicenseLabel()&sumLicenseLabel == plugAble.GetLicenseLabel() {
			a.enable = plugAble.IsEnable()
		} else {
			// license does not contain module name so disable it
			a.enable = false
		}

	}
	return a
}

func (a AddOnConsole) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (a AddOnConsole) GetAddOnName() string {
	return a.Name
}

func (a AddOnConsole) IsEnable() bool {
	return a.enable
}

func (a AddOnConsole) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if action == constants.ActionCreate {
		cluster.Console.MiddlePlatformNamespace = cluster.MiddlePlatform.GetNamespace()
		if cluster.GAP != nil {
			cluster.Console.PrometheusNamespace = cluster.GAP.GetNamespace()
		}
		task, err := a.setupConsoleTask(cluster.Masters[0].NodeId, joinPrivateRegistry(
			cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort,
		), *cluster.Console)
		if err != nil {
			return err
		}
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:            "stepSetupConsole-" + operation.Id,
			Name:          "StepSetupConsole",
			NodeSteps:     []schema.NodeStep{task},
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
		a.uninstallConsoleTask(cluster.Masters[0].NodeId)
	}
	return nil
}

func (a AddOnConsole) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return a.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func (a AddOnConsole) setupConsoleTask(nodeId, registry string, dc cs.DeployConsole) (schema.NodeStep, error) {
	dc.SetImageRegistry(registry)
	tmpl, err := dc.TemplateRender()
	if err != nil {
		return schema.NodeStep{}, err
	}
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"console": []byte(tmpl),
				},
				TimeOut: a.TaskTimeOut,
			},
		},
	}
	return step, nil
}

func (a AddOnConsole) uninstallConsoleTask(firstMasterId string) schema.NodeStep {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + firstMasterId,
		NodeID: firstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:     constants.TaskTypeKubectl,
				SubCommand:   constants.KubectlSubCommandDelete,
				CommandToRun: []string{"ns " + cs.DefaultNamespace},
				TimeOut:      a.TaskTimeOut,
			},
		},
	}
	return step
}
