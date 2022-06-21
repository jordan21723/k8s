package task_breaker

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"

	"k8s-installer/components/helm"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
)

type (
	AddOnHelm struct {
		Name        string
		HelmNodes   map[string]schema.NodeInformation
		Version     int
		TaskTimeOut int
		enable      bool
	}
)

func (params AddOnHelm) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// implement you LivenessProbe step to detect you addons state
	return schema.Step{}, nil
}

func (params AddOnHelm) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	// when ks is enabled this addon should disable
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() || cluster.ClusterRole == constants.ClusterRoleHost || cluster.ClusterRole == constants.ClusterRoleMember {
		params.enable = false
	} else {
		// do not look into license switch
		params.enable = plugAble.IsEnable()
	}
	return params
}

func (params AddOnHelm) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (params AddOnHelm) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return params.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if your addons is none container based app you should do something to remove it
		return nil
	}
}

func (params AddOnHelm) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if len(params.HelmNodes) == 0 {
		log.Error("helm operate: You need to select at least one master node to perform this action")
		return errors.New("helm operate: You need to select at least one master node to perform this action")
	}
	if action == constants.ActionCreate {
		for _, nodeInfo := range params.HelmNodes {
			params.setupInstallHelm(&nodeInfo, operation, cluster.Action)
		}
	} else if action == constants.ActionDelete {
		for _, val := range params.HelmNodes {
			params.setupUnInstall(val.Id, cluster.Action, operation)
		}
	}

	return nil
}

func (params AddOnHelm) setupInstallHelm(nodeInfo *schema.NodeInformation, operation *schema.Operation, action string) {
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:                       "setupInstallHelm-Down" + operation.Id,
		Name:                     "setupInstallHelm-Down",
		NodeSteps:                []schema.NodeStep{setupDownHelmFile(nodeInfo.Id, action, params.TaskTimeOut)},
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
	}
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:                       "setupInstallHelm-Chmod" + operation.Id,
		Name:                     "setupInstallHelm-Chmod",
		NodeSteps:                []schema.NodeStep{setupHelmChmod(nodeInfo.Id, action, params.TaskTimeOut)},
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
	}
}

func setupDownHelmFile(nodeId, action string, timeOut int) schema.NodeStep {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeDownHelmFile-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskCommonDownloadDep{
				TaskType: constants.TaskDownloadDep,
				Action:   action,
				TimeOut:  timeOut,

				Dep:        helm.HelmDep,
				SaveTo:     "/usr/bin/",
				K8sVersion: constants.V1_18_6,
				Md5:        GenerateDepMd5(helm.HelmDep),
			},
		},
	}
	return step
}

func getHelmSrcPath(nodeInfo *schema.NodeInformation) string {
	if nodeInfo == nil {
		log.Error("getHelmSrcPath: NodeInformation is empty")
		return ""
	}
	// /tmp/k8s-installer/resource/1.18.6/centos/7/x86_64/package/helm
	return filepath.Join("/usr/share/k8s-installer/helm/helm")
}

// example: CentOS Linux 7 (Core)  ->  centos
func osName(osFullName string) string {
	if strings.Contains(strings.ToLower(osFullName), "centos") {
		return "centos"
	}
	return osFullName
}

func setupHelmMoveTo(nodeId, src, dest, action string, timeOut int) schema.NodeStep {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeHelmMoveTo-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType: constants.TaskTypeRunCommand,
				Action:   action,
				TimeOut:  timeOut,
				Commands: map[int][]string{
					0: {"mv", "-f", src, dest}},
			},
		},
	}
	return step
}

func setupHelmChmod(nodeId, action string, timeOut int) schema.NodeStep {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeHelmChmod-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType: constants.TaskTypeRunCommand,
				Action:   action,
				TimeOut:  timeOut,
				Commands: map[int][]string{
					0: {"chmod", "777", "/usr/bin/helm"}},
			},
		},
	}
	return step
}

func (params AddOnHelm) setupUnInstall(nodeId, action string, operation *schema.Operation) {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeUninstallHelm-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType: constants.TaskTypeRunCommand,
				Action:   action,
				TimeOut:  params.TaskTimeOut,
				Commands: map[int][]string{
					0: {"rm", "-rf", "/usr/local/bin/helm"}},
			},
		},
	}
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:        "setupUnInstallHelm" + operation.Id,
		Name:      "setupUnInstallHelm",
		NodeSteps: []schema.NodeStep{step},
	}
}

func (params AddOnHelm) IsEnable() bool {
	return params.enable
}

func (params AddOnHelm) GetAddOnName() string {
	return params.Name
}
