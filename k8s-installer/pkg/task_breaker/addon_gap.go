package task_breaker

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"k8s-installer/components/gap"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/pkg/util/netutils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
)

type (
	AddOnGAP struct {
		Name          string
		FirstMasterId string
		TaskTimeOut   int
		enable        bool
	}
)

func (params AddOnGAP) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
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
				URL:        fmt.Sprintf("http://monitor-prometheus-server.%s.svc.%s/-/ready", cluster.GAP.Namespace, cluster.ClusterName),
				Method:     "GET",
				NameServer: coreDNSAddr,
			},
		},
	}
	return schema.Step{
		Id:        "stepCheckGAP-" + operationId,
		Name:      "StepCheckGAP",
		NodeSteps: []schema.NodeStep{task},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int,
				returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				if reply.Stat == constants.StatusSuccessful &&
					reply.ReturnData["status_code"] == "200" {
					cluster.GAP.Status = constants.StateReady
					return
				}
				cluster.GAP.Status = constants.StateNotReady
			},
			nil,
		},
		OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
			func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation, nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
				cluster.GAP.Status = constants.StateNotReady
			},
			nil,
		},
	}, nil
}

func (params AddOnGAP) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	// when ks is enabled this addon should disable
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() || cluster.ClusterRole == constants.ClusterRoleHost || cluster.ClusterRole == constants.ClusterRoleMember {
		params.enable = false
	} else {
		// do not look into license switch
		params.enable = plugAble.IsEnable()
	}
	return params
}

func (params AddOnGAP) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (params AddOnGAP) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return params.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if your addons is none container based app you should do something to remove it
		return nil
	}
}

func (params AddOnGAP) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if action == constants.ActionCreate {
		registry := joinPrivateRegistry(cluster.ContainerRuntime.PrivateRegistryAddress,
			cluster.ContainerRuntime.PrivateRegistryPort)

		params.InstallGap("/tmp/k8s-install-yaml/gap", registry, operation)
	} else if action == constants.ActionDelete {
		params.UninstallGap(params.FirstMasterId, params.TaskTimeOut, operation)
	}
	return nil
}

func (params AddOnGAP) InstallGap(charPath, registry string, operation *schema.Operation) {
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:        "stepCreateGapChart" + operation.Id,
		Name:      "stepCreateGapChart",
		NodeSteps: []schema.NodeStep{setupCreateGapChartFile(params.FirstMasterId, params.TaskTimeOut, gapTextFiles(charPath, registry))},
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
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:        "stepInstallPrometheus" + operation.Id,
		Name:      "stepInstallPrometheus",
		NodeSteps: []schema.NodeStep{setupPrometheus("gap", params.FirstMasterId, params.TaskTimeOut)},
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
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:        "stepInstallGrafana" + operation.Id,
		Name:      "stepInstallGrafana",
		NodeSteps: []schema.NodeStep{setupGrafana("gap", params.FirstMasterId, params.TaskTimeOut)},
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

func gapTextFiles(charPath, registry string) map[string][]byte {
	textFiles := make(map[string][]byte)
	// prom-alertrules.yaml prom-alertsmanager.yaml prom-settings.yaml
	textFiles[filepath.Join(charPath, "prom-alertrules.yaml")] = []byte(strings.ReplaceAll(gap.GAPPrometheusAlertRules, "caas4", registry+"/caas4"))
	textFiles[filepath.Join(charPath, "prom-alertsmanager.yaml")] = []byte(strings.ReplaceAll(gap.GAPPrometheusAlertsManager, "caas4", registry+"/caas4"))
	textFiles[filepath.Join(charPath, "prom-settings.yaml")] = []byte(strings.ReplaceAll(gap.GAPPrometheusSettings, "caas4", registry+"/caas4"))
	// prometheus Chart.yaml
	textFiles[filepath.Join(charPath, "prometheus", "Chart.yaml")] = []byte(strings.ReplaceAll(gap.GAPPrometheusChart, "caas4", registry+"/caas4"))
	// prometheus values.yaml
	textFiles[filepath.Join(charPath, "prometheus", "values.yaml")] = []byte(strings.ReplaceAll(gap.GAPPrometheusValues, "caas4", registry+"/caas4"))
	// prometheus templates
	for key, val := range gap.GAPPrometheusTemplates {
		textFiles[filepath.Join(charPath, "prometheus", "templates", key)] = []byte(strings.ReplaceAll(val, "caas4", registry+"/caas4"))
	}

	// dingtalk-webhook.yaml grafana-dashboards.yaml grafana-settings.yaml
	textFiles[filepath.Join(charPath, "dingtalk-webhook.yaml.yaml")] = []byte(strings.ReplaceAll(gap.GAPGrafanaWebhook, "caas4", registry+"/caas4"))
	textFiles[filepath.Join(charPath, "grafana-dashboards.yaml")] = []byte(strings.ReplaceAll(gap.GAPGrafanaDashboards, "caas4", registry+"/caas4"))
	textFiles[filepath.Join(charPath, "grafana-settings.yaml")] = []byte(strings.ReplaceAll(gap.GAPGrafanaSet, "caas4", registry+"/caas4"))
	// grafana Chart.yaml
	textFiles[filepath.Join(charPath, "grafana", "Chart.yaml")] = []byte(strings.ReplaceAll(gap.GAPGrafanaChart, "caas4", registry+"/caas4"))
	// grafana values.yaml
	textFiles[filepath.Join(charPath, "grafana", "values.yaml")] = []byte(strings.ReplaceAll(gap.GAPGrafanaValues, "caas4", registry+"/caas4"))
	// grafana templates
	for key, val := range gap.GAPGrafanaTemplates {
		textFiles[filepath.Join(charPath, "grafana", "templates", key)] = []byte(strings.ReplaceAll(val, "caas4", registry+"/caas4"))
	}
	return textFiles
}

func setupCreateGapChartFile(firstMasterId string, timeOut int, textFiles map[string][]byte) schema.NodeStep {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskCopyTextBaseFile-" + firstMasterId,
		NodeID: firstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskCopyTextBaseFile{
				TaskType:  constants.TaskTypeCopyTextFile,
				TimeOut:   timeOut,
				TextFiles: textFiles,
			},
		},
	}
	return step
}

func setupPrometheus(namespace, firstMasterId string, timeOut int) schema.NodeStep {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskRunCommand-" + firstMasterId,
		NodeID: firstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType: constants.TaskTypeRunCommand,
				TimeOut:  timeOut,
				Commands: map[int][]string{0: {"helm", "install", "monitor", "/tmp/k8s-install-yaml/gap/prometheus",
					"--namespace=" + namespace, "-f", "/tmp/k8s-install-yaml/gap/prom-settings.yaml",
					"-f", "/tmp/k8s-install-yaml/gap/prom-alertsmanager.yaml",
					"-f", "/tmp/k8s-install-yaml/gap/prom-alertrules.yaml", "--create-namespace"}},
			},
		},
	}
	return step
}

func setupGrafana(namespace, firstMasterId string, timeOut int) schema.NodeStep {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskRunCommand-" + firstMasterId,
		NodeID: firstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType: constants.TaskTypeRunCommand,
				TimeOut:  timeOut,
				Commands: map[int][]string{0: {"helm", "install", "grafana", "/tmp/k8s-install-yaml/gap/grafana",
					"--namespace=" + namespace, "-f", "/tmp/k8s-install-yaml/gap/grafana-settings.yaml",
					"-f", "/tmp/k8s-install-yaml/gap/grafana-dashboards.yaml", "--create-namespace"}},
			},
		},
	}
	return step
}

func (params AddOnGAP) UninstallGap(firstMasterId string, timeOut int, operation *schema.Operation) {
	step := schema.NodeStep{
		Id:     "stepUninstallGap-" + operation.Id,
		Name:   "stepUninstallGap-" + operation.Id,
		NodeID: firstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:     constants.TaskTypeKubectl,
				SubCommand:   constants.KubectlSubCommandDelete,
				CommandToRun: []string{"ns gap"},
				TimeOut:      timeOut,
			},
		},
	}
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:        "stepUninstallGap" + operation.Id,
		Name:      "stepUninstallGap",
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

func (params AddOnGAP) GetAddOnName() string {
	return params.Name
}

func (params AddOnGAP) IsEnable() bool {
	return params.enable
}
