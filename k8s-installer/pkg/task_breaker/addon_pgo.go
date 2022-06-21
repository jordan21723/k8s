package task_breaker

import (
	"fmt"
	"reflect"

	pgo "k8s-installer/components/postgres_operator"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/pkg/util/netutils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
)

type AddOnPostgresOperator struct {
	Name        string
	TaskTimeOut int
	enable      bool
}

func (a AddOnPostgresOperator) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
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
				TaskType: constants.TaskTypeCurl,
				TimeOut:  5,
				URL: fmt.Sprintf("http://postgres-operator.%s.svc.%s:8000/healthz",
					cluster.PostgresOperator.Namespace, cluster.ClusterName),
				Method:     "GET",
				NameServer: coreDNSAddr,
			},
		},
	}
	return schema.Step{
		Id:        "stepCheckPostgresOperator-" + operationId,
		Name:      "StepCheckPostgresOperator",
		NodeSteps: []schema.NodeStep{task},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				if reply.Stat == constants.StatusSuccessful &&
					reply.ReturnData["status_code"] == "200" {
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

func (a AddOnPostgresOperator) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
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

func (a AddOnPostgresOperator) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (a AddOnPostgresOperator) GetAddOnName() string {
	return a.Name
}

func (a AddOnPostgresOperator) IsEnable() bool {
	return a.enable
}

func (a AddOnPostgresOperator) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if action == constants.ActionCreate {
		task, err := a.setupPostgresOperatorTask(cluster.Masters[0].NodeId, joinPrivateRegistry(
			cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort,
		), *cluster.PostgresOperator)
		if err != nil {
			return err
		}
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "stepSetupPostgresOperator-" + operation.Id,
			Name:      "StepSetupPostgresOperator",
			NodeSteps: []schema.NodeStep{task},
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
	} else {
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "stepSetupPostgresOperator-" + operation.Id,
			Name:      "StepSetupPostgresOperator",
			NodeSteps: []schema.NodeStep{uninstallPgoTask(cluster.Masters[0].NodeId, a.TaskTimeOut)},
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
	}
	return nil
}

func (a AddOnPostgresOperator) setupPostgresOperatorTask(nodeId, registry string, dp pgo.DeployPostgresOperator) (schema.NodeStep, error) {
	dp.SetImageRegistry(registry)
	tmpl, err := dp.TemplateRender()
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
					"pgo": []byte(tmpl),
				},
				TimeOut: a.TaskTimeOut,
			},
		},
	}
	return step, nil
}

func (a AddOnPostgresOperator) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return a.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func uninstallPgoTask(firstMasterId string, taskTimeOut int) schema.NodeStep {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + firstMasterId,
		NodeID: firstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandDelete,
				CommandToRun: []string{"ClusterRoleBinding pgo-cluster-role",
					"clusterrole pgo-cluster-role", "ns " + pgo.DefaultNamespace},
				TimeOut: taskTimeOut,
			},
		},
	}
	return step
}
