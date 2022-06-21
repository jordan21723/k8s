package task_breaker

import (
	"fmt"
	"reflect"

	cs "k8s-installer/components/console"
	mp "k8s-installer/components/middle_platform"
	"k8s-installer/components/openldap"
	pgo "k8s-installer/components/postgres_operator"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/pkg/util/netutils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
)

type AddOnMiddlePlatform struct {
	Name        string
	TaskTimeOut int
	enable      bool
}

func (middlePlatform AddOnMiddlePlatform) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
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
				URL:        fmt.Sprintf("http://apigateway.%s.svc.%s/capi/iam.io/v1/menus", cluster.MiddlePlatform.Namespace, cluster.ClusterName),
				Method:     "GET",
				NameServer: coreDNSAddr,
			},
		},
	}
	return schema.Step{
		Id:        "stepCheckMiddlePlatform-" + operationId,
		Name:      "StepCheckMiddlePlatform",
		NodeSteps: []schema.NodeStep{task},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int,
				returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				if reply.Stat == constants.StatusSuccessful &&
					reply.ReturnData["status_code"] == "401" {
					cluster.MiddlePlatform.Status = constants.StateReady
					return
				}
				cluster.MiddlePlatform.Status = constants.StateNotReady
			},
			nil,
		},
		OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
			func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation, nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
				cluster.MiddlePlatform.Status = constants.StateNotReady
			},
			nil,
		},
	}, nil
}

func (middlePlatform AddOnMiddlePlatform) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	// when ks is enabled this addon should disable
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() || cluster.ClusterRole == constants.ClusterRoleHost || cluster.ClusterRole == constants.ClusterRoleMember {
		middlePlatform.enable = false
	} else {
		if plugAble.GetLicenseLabel()&sumLicenseLabel == plugAble.GetLicenseLabel() {
			middlePlatform.enable = plugAble.IsEnable()
		} else {
			// license does not contain module name so disable it
			middlePlatform.enable = false
		}
	}
	return middlePlatform
}

func (middlePlatform AddOnMiddlePlatform) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (middlePlatform AddOnMiddlePlatform) GetAddOnName() string {
	return middlePlatform.Name
}

func (middlePlatform AddOnMiddlePlatform) IsEnable() bool {
	return middlePlatform.enable
}

func (middlePlatform AddOnMiddlePlatform) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return middlePlatform.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func (middlePlatform AddOnMiddlePlatform) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if action == constants.ActionCreate {
		// setup middle platform
		if cluster.MiddlePlatform.LDAP != nil && cluster.MiddlePlatform.LDAP.CaaSDeploy != nil {
			// Enable LDAP for middle platform and deploying OpenLDAP on CaaS.
			task, err := setupOpenLDAPTask(cluster.Masters[0].NodeId, joinPrivateRegistry(
				cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort,
			), *cluster.MiddlePlatform.LDAP.CaaSDeploy, 5)
			if err != nil {
				return err
			}
			operation.Step[len(operation.Step)] = &schema.Step{
				Id:        "stepSetupOpenLDAP-" + operation.Id,
				Name:      "StepSetupOpenLDAP",
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
		}

		task, err := setupMiddlePlatformTask(cluster.Masters[0].NodeId, joinPrivateRegistry(
			cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort,
		), *cluster.MiddlePlatform, 5)
		if err != nil {
			return err
		}
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "stepSetupMiddlePlatform-" + operation.Id,
			Name:      "StepSetupMiddlePlatform",
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
	} else if action == constants.ActionDelete {
		// if some task is not container based should do something to remove it when api call is try remove cluster
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:        "uninstallConsole-" + operation.Id,
			Name:      "uninstallConsole",
			NodeSteps: []schema.NodeStep{uninstallMiddlePlatformTask(cluster.Masters[0].NodeId, 60)},
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

func setupOpenLDAPTask(nodeId, registry string, do openldap.DeployOpenLDAP, timeOut int) (schema.NodeStep, error) {
	tmpl, err := do.SetImageRegistry(registry).TemplateRender()
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
					"openldap": []byte(tmpl),
				},
				TimeOut: timeOut,
			},
		},
	}
	return step, nil
}

func setupMiddlePlatformTask(nodeId, registry string, dm mp.DeployMiddlePlatform, timeOut int) (schema.NodeStep, error) {
	dm.SetImageRegistry(registry)
	tmpl, err := dm.TemplateRender()
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
					"middle-platform": []byte(tmpl),
				},
				TimeOut: timeOut,
			},
		},
	}
	return step, nil
}

func uninstallMiddlePlatformTask(firstMasterId string, taskTimeOut int) schema.NodeStep {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + firstMasterId,
		NodeID: firstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandDelete,
				CommandToRun: []string{"ns " + cs.DefaultNamespace, "ClusterRoleBinding kube-resource",
					"ns " + mp.DefaultNamespace, "ClusterRoleBinding pgo-cluster-role",
					"clusterrole pgo-cluster-role", "ns " + pgo.DefaultNamespace,
					"ns " + openldap.DefaultNamespace},
				TimeOut: taskTimeOut,
			},
		},
	}
	return step
}
