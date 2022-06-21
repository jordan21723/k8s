package task_breaker

import (
	"fmt"
	"reflect"

	"k8s-installer/components/efk"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util/netutils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
)

type (
	AddOnEFK struct {
		Name             string
		Namespace        string
		storageClassName string
		FirstMasterId    string
		WaitBeforeRun    int
		TaskTimeOut      int
		enable           bool
		IsIngressEnable  bool
		IsConsoleEneble  bool
		ConsoleNamespace string
	}
)

func (params AddOnEFK) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
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
				URL:        fmt.Sprintf("http://elasticsearch-logging.%s.svc.%s:9200/_cluster/health", cluster.EFK.Namespace, cluster.ClusterName),
				Method:     "GET",
				NameServer: coreDNSAddr,
			},
		},
	}
	return schema.Step{
		Id:        "stepCheckEFK-" + operationId,
		Name:      "StepCheckEFK",
		NodeSteps: []schema.NodeStep{task},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int,
				returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				if reply.Stat == constants.StatusSuccessful &&
					reply.ReturnData["status_code"] == "200" {
					cluster.EFK.Status = constants.StateReady
					return
				}
				cluster.EFK.Status = constants.StateNotReady
			},
			nil,
		},
		OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
			func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation, nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
				cluster.EFK.Status = constants.StateNotReady
			},
			nil,
		},
	}, nil
}

func (params AddOnEFK) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	// when ks is enabled this addon should disable
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() || cluster.ClusterRole == constants.ClusterRoleHost || cluster.ClusterRole == constants.ClusterRoleMember {
		params.enable = false
	} else {
		// do not look into license switch
		params.enable = plugAble.IsEnable()
		efkDeploy := plugAble.(*efk.DeployEFK)
		if efkDeploy != nil {
			params.storageClassName = efkDeploy.StorageClassName
			params.enable = efkDeploy.Enable
		}
	}
	return params
}

func (params AddOnEFK) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (params AddOnEFK) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return params.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if your addons is none container based app you should do something to remove it
		return nil
	}
}

func (params AddOnEFK) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if action == constants.ActionCreate {
		registry := joinPrivateRegistry(cluster.ContainerRuntime.PrivateRegistryAddress,
			cluster.ContainerRuntime.PrivateRegistryPort)
		return params.InstallEFK(registry, params.storageClassName, operation, *cluster.EFK)
	} else if action == constants.ActionDelete {
		params.UninstallEFK(operation)
	}
	return nil
}

func (params AddOnEFK) InstallEFK(registry, storageClassName string, operation *schema.Operation, de efk.DeployEFK) error {
	// efkdeploy := efk.NewDeployEFK(registry, storageClassName, 2)
	de.SetImageRegistry(registry)
	tmpl, err := de.TemplateRender(efk.EFKBase)
	if err != nil {
		return err
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
					"efk-base": []byte(tmpl),
				},
				TimeOut: params.TaskTimeOut,
			},
		},
	}
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:        "stepInstallEFKBase" + operation.Id,
		Name:      "stepInstallEFKBase",
		NodeSteps: []schema.NodeStep{step},
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

	tmpl, err = de.TemplateRender(efk.EFKDynamicPV)
	if err != nil {
		return err
	}
	step = schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + params.FirstMasterId,
		NodeID: params.FirstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"efk-pv": []byte(tmpl),
				},
				TimeOut: params.TaskTimeOut,
			},
		},
	}

	operation.Step[len(operation.Step)] = &schema.Step{
		Id:        "stepInstallEFKPv" + operation.Id,
		Name:      "stepInstallEFKPv",
		NodeSteps: []schema.NodeStep{step},
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

	if params.IsIngressEnable {
		nodeStep := []schema.NodeStep{}
		if de.User == "" {
			de.User = constants.EFKDefaultUser
		}

		if de.Password == "" {
			de.Password = constants.EFKDefaultPassword
		}

		CommonRunBatchWithSingleNode(params.FirstMasterId, "", 30, [][]string{{"htpasswd", "-b", "-c", "/tmp/auth", de.User, de.Password}, {"kubectl", "create", "secret", "generic", "kibana-basic", "--from-file=/tmp/auth", "-n", de.Namespace}}, &nodeStep, false, false)

		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "stepInstallEFKIngressAuth" + operation.Id,
			Name:      "stepInstallEFKIngressAuth",
			NodeSteps: nodeStep,
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

		if !params.IsConsoleEneble {
			params.ConsoleNamespace = params.Namespace
		}

		params.Namespace = de.Namespace
		tmpl, err = template.New("efk").Render(efk.EFKIngress, params)
		if err != nil {
			return err
		}
		step = schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskTypeKubectl-" + params.FirstMasterId,
			NodeID: params.FirstMasterId,
			Tasks: map[int]schema.ITask{
				0: schema.TaskKubectl{
					TaskType:   constants.TaskTypeKubectl,
					SubCommand: constants.KubectlSubCommandCreateOrApply,
					YamlTemplate: map[string][]byte{
						"efk-ingress": []byte(tmpl),
					},
					TimeOut: params.TaskTimeOut,
				},
			},
		}

		operation.Step[len(operation.Step)] = &schema.Step{
			Id:            "stepInstallEFKIngress" + operation.Id,
			Name:          "stepInstallEFKIngress",
			NodeSteps:     []schema.NodeStep{step},
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
	} else {
		tmpl, err = de.TemplateRender(efk.EFKNodePort)
		if err != nil {
			return err
		}
		step = schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskTypeKubectl-" + params.FirstMasterId,
			NodeID: params.FirstMasterId,
			Tasks: map[int]schema.ITask{
				0: schema.TaskKubectl{
					TaskType:   constants.TaskTypeKubectl,
					SubCommand: constants.KubectlSubCommandCreateOrApply,
					YamlTemplate: map[string][]byte{
						"efk-nodeport": []byte(tmpl),
					},
					TimeOut: params.TaskTimeOut,
				},
			},
		}

		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "stepInstallEFKNodeport" + operation.Id,
			Name:      "stepInstallEFKNodeport",
			NodeSteps: []schema.NodeStep{step},
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

func (params AddOnEFK) UninstallEFK(operation *schema.Operation) {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + params.FirstMasterId,
		NodeID: params.FirstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:     constants.TaskTypeKubectl,
				SubCommand:   constants.KubectlSubCommandDelete,
				CommandToRun: []string{"deploy --all -n efk", "sts --all -n efk", "pvc --all -n efk", "ns efk"},
				TimeOut:      params.TaskTimeOut,
			},
		},
	}
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:        "stepUninstallEFKDeploy" + operation.Id,
		Name:      "stepUninstallEFKDeploy",
		NodeSteps: []schema.NodeStep{step},
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

func (params AddOnEFK) GetAddOnName() string {
	return params.Name
}

func (params AddOnEFK) IsEnable() bool {
	return params.enable
}
