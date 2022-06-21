package task_breaker

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"

	"k8s-installer/components/CNI/Calico"
	"k8s-installer/node/container_runtime/docker"
	"k8s-installer/node/k8s"
	"k8s-installer/node/loadbalancer"
	"k8s-installer/pkg/cache"
	serverConfig "k8s-installer/pkg/config/server"
	depMap "k8s-installer/pkg/dep"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"

	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/util/fileutils"
	"k8s-installer/schema"
)

/*
empty Step only for hold main operation control flow to wait certain seconds
*/
func CreateWaitStep(secToWait int, stepName string) schema.Step {
	return schema.Step{
		Name:          fmt.Sprintf("StepWait %d secs for %s", secToWait, stepName),
		WaitBeforeRun: secToWait,
	}
}

func deployOrRemoveAddons(addons IAddOns, operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	action := ""
	if addons.IsEnable() {
		action = constants.ActionCreate
	} else {
		action = constants.ActionDelete
	}
	if err := addons.DeployOrRemoveAddOn(operation, cluster, config, action); err != nil {
		log.Errorf("Failed to deploy addons %s", addons.GetAddOnName())
		return err
	}
	return nil
}

func CommonConfig(nodeList map[string]schema.NodeInformation, fileName string, content string, taskCommonDownload schema.TaskCommonDownload,
	nodeSteps *[]schema.NodeStep, config serverConfig.Config, format func(node schema.NodeInformation) string) {

	for _, node := range nodeList {
		tmpDir := path.Join(config.ApiServer.ResourceServerFilePath, node.Id)
		log.Debugf("Create tmpdir %v", tmpDir)
		if errCreateDirIfNotExists := util.CreateDirIfNotExists(tmpDir); errCreateDirIfNotExists != nil {
			log.Errorf("Failed to create dir %s due to error: %v", tmpDir, errCreateDirIfNotExists.Error())
		}
		tmpFile := path.Join(tmpDir, fileName)
		if format != nil {
			content = format(node)
		}
		log.Debugf("Create tmpfile %v", tmpFile)
		fileutils.DeleteFile(tmpFile)
		fileutils.CreateTestFileWithMD5(tmpFile, content)

		fileList := []string{strings.Replace(tmpFile, config.ApiServer.ResourceServerFilePath, "", 1)}

		taskCommonDownload.FromDir = node.Id
		taskCommonDownload.FileList = fileList

		*nodeSteps = append(*nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskCommonConfig-" + fileName + node.Id,
			NodeID: node.Id,
			Tasks: map[int]schema.ITask{
				0: taskCommonDownload,
			},
		})
	}
}

func CommonRun(nodeList map[string]schema.NodeInformation, action string,
	timeOut int, cmd []string, nodeSteps *[]schema.NodeStep) {

	for _, node := range nodeList {
		*nodeSteps = append(*nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   fmt.Sprintf("TaskRunCommand- %v", cmd),
			NodeID: node.Id,
			Tasks: map[int]schema.ITask{
				0: schema.TaskRunCommand{
					TaskType: constants.TaskTypeRunCommand,
					TimeOut:  timeOut,
					Commands: map[int][]string{0: cmd},
				},
			},
		})
	}
}

func CommonRunWithSingleNode(nodeId, stepName string, action string,
	timeOut int, cmd []string, requireResult bool) schema.NodeStep {
	return schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   stepName,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType:      constants.TaskTypeRunCommand,
				TimeOut:       timeOut,
				Commands:      map[int][]string{0: cmd},
				RequireResult: requireResult,
			},
		},
	}
}

func CommonRunBatchWithSingleNode(nodeId string, action string,
	timeOut int, cmds [][]string, nodeSteps *[]schema.NodeStep, requireResult bool, ignoreError bool) {
	commandsToRun := map[int][]string{}
	for index, cmd := range cmds {
		commandsToRun[index] = cmd
	}
	*nodeSteps = append(*nodeSteps, schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskRunCommand batch command",
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType:      constants.TaskTypeRunCommand,
				TimeOut:       timeOut,
				Commands:      commandsToRun,
				RequireResult: requireResult,
				IgnoreError:   ignoreError,
			},
		},
	})
}

func CommonRunBatchWithSingleNodeWithStepID(nodeId string, stepId string,
	timeOut int, cmds [][]string, nodeSteps *[]schema.NodeStep, requireResult bool, ignoreError bool) {
	commandsToRun := map[int][]string{}
	for index, cmd := range cmds {
		commandsToRun[index] = cmd
	}
	*nodeSteps = append(*nodeSteps, schema.NodeStep{
		Id:     stepId,
		Name:   "TaskRunCommand batch command",
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType:      constants.TaskTypeRunCommand,
				TimeOut:       timeOut,
				Commands:      commandsToRun,
				RequireResult: requireResult,
				IgnoreError:   ignoreError,
				CommandRunId:  stepId,
			},
		},
	})
}

func CommonBatchRun(nodeList map[string]schema.NodeInformation, action string,
	timeOut int, cmds [][]string, nodeSteps *[]schema.NodeStep, async bool) {

	var taskType string
	if async {
		taskType = constants.TaskTypeRunAsyncCommand
	} else {
		taskType = constants.TaskTypeRunCommand
	}
	commandsToRun := map[int][]string{}
	for index, cmd := range cmds {
		commandsToRun[index] = cmd
	}
	for _, node := range nodeList {
		*nodeSteps = append(*nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskRunCommand batch command",
			NodeID: node.Id,
			Tasks: map[int]schema.ITask{
				0: schema.TaskRunCommand{
					TaskType: taskType,
					TimeOut:  timeOut,
					Commands: commandsToRun,
				},
			},
		})
	}
}

func CommonLink(nodeList map[string]schema.NodeInformation, action string,
	timeOut int, from depMap.DepMap,
	saveTo string, linkTo string,
	nodeSteps *[]schema.NodeStep) {

	for _, node := range nodeList {
		*nodeSteps = append(*nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskLink-" + linkTo + node.Id,
			NodeID: node.Id,
			Tasks: map[int]schema.ITask{
				0: schema.TaskLink{
					TaskType: constants.TaskLink,
					Action:   action,
					TimeOut:  timeOut,
					From:     from,
					SaveTo:   saveTo,
					LinkTo:   linkTo,
				},
			},
		})
	}
}

func CommonDownloadDep(nodeList map[string]schema.NodeInformation, taskCommonDownloadDep schema.TaskCommonDownloadDep,
	nodeSteps *[]schema.NodeStep) {

	taskCommonDownloadDep.Md5 = GenerateDepMd5(taskCommonDownloadDep.Dep)

	for _, node := range nodeList {
		*nodeSteps = append(*nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskCommonDownload-" + reflect.TypeOf(taskCommonDownloadDep).String() + node.Id,
			NodeID: node.Id,
			Tasks: map[int]schema.ITask{
				0: taskCommonDownloadDep,
			},
		})
	}
}

func CommonDownload(nodeList map[string]schema.NodeInformation, taskCommonDownload schema.TaskCommonDownload,
	nodeSteps *[]schema.NodeStep) {

	for _, node := range nodeList {
		*nodeSteps = append(*nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskCommonDownloadDep-" + reflect.TypeOf(taskCommonDownload).String() + node.Id,
			NodeID: node.Id,
			Tasks: map[int]schema.ITask{
				0: taskCommonDownload,
			},
		})
	}
}

func CommonAdminConf(nodeId string, stepId string,
	timeOut int, cmds [][]string, nodeSteps *[]schema.NodeStep, requireResult bool, ignoreError bool) {
	commandsToRun := map[int][]string{}
	for index, cmd := range cmds {
		commandsToRun[index] = cmd
	}
	*nodeSteps = append(*nodeSteps, schema.NodeStep{
		Id:     stepId,
		Name:   "TaskRunCommand batch command",
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType:      constants.TaskTypeRunCommand,
				TimeOut:       timeOut,
				Commands:      commandsToRun,
				RequireResult: requireResult,
				IgnoreError:   ignoreError,
			},
		},
	})
}

func CommonWriteAdminConf(nodeId string, stepId string,
	timeOut int, nodeSteps *[]schema.NodeStep, textFiles map[string][]byte) {
	*nodeSteps = append(*nodeSteps, schema.NodeStep{
		Id:     stepId,
		Name:   "TaskTypeCopyTextFile-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskCopyTextBaseFile{
				TaskType:  constants.TaskTypeCopyTextFile,
				TimeOut:   timeOut,
				TextFiles: textFiles,
			},
		},
	})
}

func joinPrivateRegistry(host string, port int) string {
	if host != "" && port != 0 {
		return fmt.Sprintf("%s:%d", host, port)
	}
	return ""
}

func InstallAddOns(register *AddOnsRegister, operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) []error {
	errs := []error{}
	var livenessProbes []schema.Step
	// create a wait step ensure all addons can properly setup
	livenessProbes = append(livenessProbes, CreateWaitStep(3*60, "active-addons"))
	for _, addon := range register.AddOns {
		if addon.IsEnable() {
			if err := addon.DeployOrRemoveAddOn(operation, cluster, config, action); err != nil {
				errs = append(errs, err)
				log.Errorf("Failed to install add-on %s", addon.GetAddOnName())
			} else {
				if lpStep, err := addon.LivenessProbe(operation.Id, cluster, config); err != nil {
					errs = append(errs, err)
					log.Errorf("Failed to setup add-on %s's liveness probe due to error: %s", addon.GetAddOnName(), err.Error())
				} else {
					livenessProbes = append(livenessProbes, lpStep)
				}

			}
		}
	}
	return errs
}

func InstallAddOnsWithCluster(register *AddOnsRegister, operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) []error {
	errs := []error{}
	var livenessProbes []schema.Step
	// create a wait step ensure all addons can properly setup
	livenessProbes = append(livenessProbes, CreateWaitStep(3*60, "active-addons"))

	for _, addon := range register.AddOns {
		if addon.IsEnable() {
			if err := addon.DeployOrRemoveAddonsWithCluster(operation, cluster, config); err != nil {
				errs = append(errs, err)
				log.Errorf("Failed to breakdown task for add-on %s due to error %s", addon.GetAddOnName(), err.Error())
			} else {
				if lpStep, err := addon.LivenessProbe(operation.Id, cluster, config); err != nil {
					errs = append(errs, err)
					log.Errorf("Failed to setup add-on %s's liveness probe due to error: %s", addon.GetAddOnName(), err.Error())
				} else {
					livenessProbes = append(livenessProbes, lpStep)
				}

			}
		}
	}
	// add liveness probe if cluster is being create
	// and there is livenessProbes need to be perform beside wait step
	if cluster.Action != constants.ActionDelete && len(livenessProbes) > 1 {
		addStepsToOperation(operation, livenessProbes)
	}
	return errs
}

func InstallAddOnsWithNodeAddOrRemove(register *AddOnsRegister, operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) []error {
	errs := []error{}
	for _, addon := range register.AddOns {
		if addon.IsEnable() {
			if err := addon.DeployOrRemoveAddonsWithNodeAddOrRemove(operation, cluster, config, nodeAction, nodesToAddOrRemove); err != nil {
				errs = append(errs, err)
				log.Errorf("Failed to breakdown task for add-on %s due to error %s", addon.GetAddOnName(), err.Error())
			}
		}
	}
	return errs
}

func addStepsToOperation(operation *schema.Operation, steps []schema.Step) {
	for index := range steps {
		operation.Step[len(operation.Step)] = &steps[index]
	}
}

func createJoinStringNodeStep(firstMasterId string) schema.NodeStep {
	nodeStep := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskPrintJoin",
		NodeID: firstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskPrintJoin{
				TaskType: constants.TaskPrintJoin,
				TimeOut:  5,
			},
		},
	}
	return nodeStep
}

func createUntaintNodeTask(firstMasterNodeId string, needToUntaintNodeList map[string]schema.NodeInformation, nodeToHostnameMapping map[string]string, action string, timeOut int, enableHostnameRename bool) schema.NodeStep {
	var commandToRun []string
	for _, node := range needToUntaintNodeList {
		if enableHostnameRename {
			commandToRun = append(commandToRun, fmt.Sprintf("node %s node-role.kubernetes.io/master:NoSchedule-", nodeToHostnameMapping[node.Id]))
		} else {
			commandToRun = append(commandToRun, fmt.Sprintf("node %s node-role.kubernetes.io/master:NoSchedule-", node.SystemInfo.Node.Hostname))
		}

	}
	nodeStep := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-untaint",
		NodeID: firstMasterNodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:     constants.TaskTypeKubectl,
				Action:       action,
				TimeOut:      timeOut,
				SubCommand:   constants.KubectlSubCommandTaint,
				CommandToRun: commandToRun,
			},
		},
	}
	return nodeStep
}

func createCRINodeTask(CRIConfig schema.ContainerRuntime, nodeList map[string]schema.NodeInformation, action string, timeOut int) []schema.NodeStep {

	rtc := cache.GetCurrentCache()
	config := rtc.GetServerRuntimeConfig(cache.NodeId)

	var nodeSteps []schema.NodeStep

	d := docker.Deps{}

	// container runtime task will always result in one single task so we can hard code it
	for key := range nodeList {
		nodeSteps = append(nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskTypeCRI-" + key,
			NodeID: key,
			Tasks: map[int]schema.ITask{
				0: schema.TaskCRI{
					TaskType:   constants.TaskTypeCRI,
					CRIType:    CRIConfig,
					Action:     action,
					TimeOut:    timeOut,
					LogMaxFile: config.MaxContainerLogFile,
					LogSize:    config.MaxContainerLogFileSize,
					Md5Dep:     GenerateDepMd5(d.GetDeps()),
				},
			},
		})
	}
	return nodeSteps
}

func createBasicConfigNodeTask(nodeList map[string]schema.NodeInformation, action string, timeOut int) []schema.NodeStep {
	var nodeSteps []schema.NodeStep
	for key := range nodeList {
		nodeSteps = append(nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskTypeBasic-" + key,
			NodeID: key,
			Tasks: map[int]schema.ITask{
				0: schema.TaskBasicConfig{
					TaskType: constants.TaskTypeBasic,
					Role:     0,
					Action:   action,
					TimeOut:  timeOut,
				},
			},
		})
	}
	return nodeSteps
}

func createRenameWorkerHostnameNodeTask(nodes []schema.ClusterNode, hosts []string, clusterId, suffix string, workerNodeCount, timeOut int) []schema.NodeStep {
	var nodeSteps []schema.NodeStep
	index := workerNodeCount
	for _, node := range nodes {
		nodeHostName := generatorHostname(index, clusterId, suffix)
		nodeSteps = append(nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskRenameHostName-" + node.NodeId,
			NodeID: node.NodeId,
			Tasks: map[int]schema.ITask{
				0: schema.TaskRenameHostName{
					TaskType: constants.TaskTypeRenameHostname,
					Action:   "",
					TimeOut:  timeOut,
					Hostname: nodeHostName,                // set new hostname to [cluster name]-[master]-[index]
					Hosts:    strings.Join(hosts, "\n\r"), // /etc/hosts for all master
				},
			},
		})
		index += 1
	}
	return nodeSteps
}

func createRenameWorkerHostnameNodeTaskWithNodeID(workers []schema.NodeInformation, nodes []schema.ClusterNode, hosts []string, clusterId, suffix string, workerNodeCount, timeOut int) ([]schema.NodeStep, error) {
	index := 0
	var noShareWorker []schema.NodeInformation
	for _, worker := range workers {
		if strings.Contains(worker.SystemInfo.Node.Hostname, "worker") {
			noShareWorker = append(noShareWorker, worker)
		}
	}
	if len(noShareWorker) > 0 {
		tempHostname := noShareWorker[len(noShareWorker)-1].SystemInfo.Node.Hostname
		tempIndex := tempHostname[strings.LastIndex(tempHostname, "-")+1:]
		if value, err := strconv.Atoi(tempIndex); err != nil {
			return nil, err
		} else {
			index = value + 1
		}
	}
	var nodeSteps []schema.NodeStep
	for _, node := range nodes {
		nodeHostName := generatorHostname(index, clusterId, suffix)
		nodeSteps = append(nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskRenameHostName-" + node.NodeId,
			NodeID: node.NodeId,
			Tasks: map[int]schema.ITask{
				0: schema.TaskRenameHostName{
					TaskType: constants.TaskTypeRenameHostname,
					Action:   "",
					TimeOut:  timeOut,
					Hostname: nodeHostName,                // set new hostname to [cluster name]-[master]-[index]
					Hosts:    strings.Join(hosts, "\n\r"), // /etc/hosts for all master
				},
			},
		})
		index += 1
	}
	return nodeSteps, nil
}

func createRenameHostnameNodeTask(masters map[string]schema.NodeInformation, nodeToHostNameMapping map[string]string, clusterId, suffix string, timeOut int, hosts []string) []schema.NodeStep {
	var nodeSteps []schema.NodeStep
	index := 0
	for nodeId := range masters {
		nodeHostName := generatorHostname(index, clusterId, suffix)
		nodeToHostNameMapping[nodeId] = nodeHostName
		nodeSteps = append(nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskRenameHostName-" + nodeId,
			NodeID: nodeId,
			Tasks: map[int]schema.ITask{
				0: schema.TaskRenameHostName{
					TaskType: constants.TaskTypeRenameHostname,
					Action:   "",
					TimeOut:  timeOut,
					Hostname: nodeHostName,                // set new hostname to [cluster name]-[master]-[index]
					Hosts:    strings.Join(hosts, "\n\r"), // /etc/hosts for all master
				},
			},
		})
		index += 1
	}
	return nodeSteps
}

func createWorkNodeVipNodeTask(nodeList, masters map[string]schema.NodeInformation, action string, timeOut int) []schema.NodeStep {
	var nodeSteps []schema.NodeStep
	// we always use haproxy as local proxy for kubernetes api server
	haproxy := loadbalancer.Haproxy{}
	for key := range nodeList {
		nodeSteps = append(nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskTypeWorkNodeVip-" + key,
			NodeID: key,
			Tasks: map[int]schema.ITask{
				0: schema.TaskLoadBalance{
					TaskType:    constants.TaskTypeWorkNodeVip,
					ProxyType:   constants.ProxyTypeHaproxy,
					Action:      action,
					TimeOut:     timeOut,
					ProxyConfig: []byte(haproxy.CreateAPIServerConfig("127.0.0.1", "6443", "roundrobin", masters)),
				},
			},
		})
	}
	return nodeSteps
}

func createKubeadmInitFirstControlPlaneNodeTask(nodeId, action string, timeOut int, cluster schema.Cluster, collection schema.NodeInformationCollection) (schema.NodeStep, error) {
	var tmpl string
	var err error
	switch cluster.CNI.CNIType {
	case constants.CNITypeCalico:
		deploy := Calico.NewCalicoDeploy(joinPrivateRegistry(cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort), cluster.CNI)
		tmpl, err = deploy.TemplateRender()
		if err != nil {
			return schema.NodeStep{}, err
		}
	default:
		return schema.NodeStep{}, errors.New(fmt.Sprintf("CNI type %s not supported yet", constants.CNITypeCalico))
	}
	if kubeadmConfig, err := k8s.CreateKubeadmConfigFromCluster(cluster, collection); err != nil {
		return schema.NodeStep{}, err
	} else {
		return schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskTypeKubeadm-" + nodeId,
			NodeID: nodeId,
			Tasks: map[int]schema.ITask{
				0: schema.TaskKubeadm{
					TaskType:      constants.TaskTypeKubeadm,
					Action:        action,
					ControlPlane:  cluster.ControlPlane,
					KubeadmTask:   constants.KubeadmTaskInitFirstControlPlane,
					TimeOut:       timeOut,
					KubeadmConfig: []byte(kubeadmConfig),
					CNITemplate:   []byte(tmpl),
					Md5Dep:        GenerateDepMd5(k8s.KubeDepMapping),
				},
			},
		}, nil
	}
}

func createKubeadmJoinControlPlaneNodeTask(nodeList map[string]byte, action string, timeOut int) []schema.NodeStep {
	var nodeSteps []schema.NodeStep
	for key := range nodeList {
		nodeSteps = append(nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskTypeKubeadm-" + key,
			NodeID: key,
			Tasks: map[int]schema.ITask{
				0: schema.TaskKubeadm{
					TaskType:    constants.TaskTypeKubeadm,
					Action:      action,
					KubeadmTask: constants.KubeadmTaskJoinControlPlane,
					TimeOut:     timeOut,
					Md5Dep:      GenerateDepMd5(k8s.KubeDepMapping),
				},
			},
		})
	}
	return nodeSteps
}

func createKubeadmJoinWorkerNodeTask(nodeList map[string]schema.NodeInformation, action string, timeOut int) []schema.NodeStep {
	var nodeSteps []schema.NodeStep
	for key := range nodeList {
		nodeSteps = append(nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskTypeKubeadm-" + key,
			NodeID: key,
			Tasks: map[int]schema.ITask{
				0: schema.TaskKubeadm{
					TaskType:    constants.TaskTypeKubeadm,
					Action:      action,
					KubeadmTask: constants.KubeadmTaskJoinWorker,
					TimeOut:     timeOut,
					Md5Dep:      GenerateDepMd5(k8s.KubeDepMapping),
				},
			},
		})
	}
	return nodeSteps
}

func createExternalLBNodeTask(lb schema.ExternalLB, nodeList []schema.ClusterNode, action string, timeOut int) []schema.NodeStep {
	var nodeSteps []schema.NodeStep
	for _, val := range nodeList {
		nodeSteps = append(nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskTypeWorkNodeVip-" + val.NodeId,
			NodeID: val.NodeId,
			Tasks: map[int]schema.ITask{
				0: schema.TaskLoadBalance{
					TaskType:  constants.TaskTypeWorkNodeVip,
					ProxyType: constants.ProxyTypeHaproxy,
					Action:    action,
					Vips:      lb,
					TimeOut:   timeOut,
				},
			},
		})
	}
	return nodeSteps
}

func createKubectlDeleteNodesNodeTask(nodesHostName []string, firstMasterId string) schema.NodeStep {
	nodeToDelete := []string{}
	for _, nodeHostName := range nodesHostName {
		nodeToDelete = append(nodeToDelete, fmt.Sprintf("node %s", nodeHostName))
	}
	stepDeleteNode := schema.NodeStep{
		Id:               utils.GenNodeStepID(),
		Name:             "TaskKubectl",
		NodeID:           firstMasterId,
		ServerMSGTimeOut: 5,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType: "TaskKubectl",
				TimeOut:  5,
				// in some case add node failed k8s cluster didn't add the node to cluster
				// so we might want to remove the node from cluster,but the not is not appears in the k8s cluster yet
				// which will leads to unable to do clean up job on the node
				CanIgnoreError: true,
				SubCommand:     constants.KubectlSubCommandDelete,
				CommandToRun:   nodeToDelete,
			},
		},
	}
	return stepDeleteNode
}

func generatorHostname(index int, clusterName, suffix string) string {
	return clusterName + suffix + strconv.Itoa(index)
}

func generatorHostnameWithNodeID(nodeId, clusterName, suffix string) string {
	return clusterName + suffix + nodeId
}

func createFullNodeList(cluster schema.Cluster, nodeCollectionList schema.NodeInformationCollection) map[string]schema.NodeInformation {
	fullNodeList := map[string]schema.NodeInformation{}
	for _, master := range cluster.Masters {
		fullNodeList[master.NodeId] = nodeCollectionList[master.NodeId]
	}
	for _, master := range cluster.Workers {
		fullNodeList[master.NodeId] = nodeCollectionList[master.NodeId]
	}
	return fullNodeList
}

func createCordonStep(nodeId, nodeHostname string, unCordon bool, additionalParas []string) *schema.Step {
	step := &schema.Step{
		Id:   utils.GenNodeStepID(),
		Name: "TaskRunCommand-" + nodeId,
		NodeSteps: []schema.NodeStep{
			createCordonNodeStep(nodeId, nodeHostname, unCordon, additionalParas),
		},
	}
	return step
}

func createCordonNodeStep(nodeId, nodeHostname string, unCordon bool, additionalParas []string) schema.NodeStep {
	cordon := "drain"
	if unCordon {
		cordon = "uncordon"
	}
	commands := []string{"kubectl", cordon, nodeHostname}
	if len(additionalParas) > 0 {
		commands = append(commands, additionalParas...)
	}
	return schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "Cordon Node",
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType: "TaskRunCommand",
				TimeOut:  360,
				Commands: map[int][]string{
					0: commands,
				},
			},
		},
	}
}

func createLableNodeTask(firstMasterNodeId string, node schema.NodeInformation, action string, timeOut int) schema.NodeStep {
	nodeStep := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-lable-node",
		NodeID: firstMasterNodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType: constants.TaskTypeRunCommand,
				TimeOut:  timeOut,
				Commands: map[int][]string{
					0: {"/usr/bin/kubectl", "label", "node", node.SystemInfo.Node.Hostname, "node-role.kubernetes.io/worker=node"}},
				IgnoreError: true,
			},
		},
	}
	return nodeStep
}

func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func GenerateDepMd5(dep depMap.DepMap) depMap.DepMap {

	runtimeCache := cache.GetCurrentCache()
	config := runtimeCache.GetServerRuntimeConfig(cache.NodeId)

	md5Dep := make(depMap.DepMap)
	deepCopy(&md5Dep, dep)

	for osk, osv := range dep {
		for kvk, kvv := range osv {
			for arch, archv := range kvv {
				for k, f := range archv {
					fp := path.Join(config.ApiServer.ResourceServerFilePath, kvk, osk, "7", arch, "package", f)
					md5 := fileutils.Md5Sum(fp)
					md5Dep[osk][kvk][arch][k] = md5
				}
			}
		}
	}
	log.Debugf("md5Dep: %v", md5Dep)
	return md5Dep
}

func CheckAddonsDependencies(cluster schema.Cluster) (err error, errorList []string) {
	if cluster.Console != nil {
		if err, msg := cluster.Console.CheckDependencies(); err != nil {
			errorList = append(errorList, msg...)
		}
	}

	if cluster.MiddlePlatform != nil {
		if err, msg := cluster.MiddlePlatform.CheckDependencies(); err != nil {
			errorList = append(errorList, msg...)
		}
	}

	if cluster.GAP != nil {
		if err, msg := cluster.GAP.CheckDependencies(); err != nil {
			errorList = append(errorList, msg...)
		}
	}

	if len(errorList) > 0 {
		err = errors.New(fmt.Sprintf("Dependency check failed due to error %s", strings.Join(errorList, "\n")))
	}
	return err, errorList
}

func BackupCronJobCreate(masterId string, v *schema.BackupRegular, timeout int) (*schema.NodeStep, error) {
	tmpl := `
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  namespace: velero
  name: {{.BackupRegularName}}
spec:
  failedJobsHistoryLimit: 3
  successfulJobsHistoryLimit: 3
  schedule: {{.CronjobTime}}
  jobTemplate:
    spec:
      backoffLimit: 0
      template:
        spec:
          serviceAccountName: velero
          restartPolicy: Never
          containers:
          - name: busybox
            image: {{.Registry}}/busybox:1.31.1
            imagePullPolicy: IfNotPresent
            command:
            - /bin/sh
            - -c
            - velero {{.Action}} backup {{.BackupRegularName}}$(date +%Y%m%d%H%M) --include-cluster-resources --exclude-namespaces minio,velero --wait
            volumeMounts:
            - mountPath: /usr/bin
              name: bin
          volumes:
          - name: bin
            hostPath:
              path: /usr/bin
`
	yamlTemplate, err := template.New("backup-cronjob").Render(tmpl, v)
	if err != nil {
		return nil, err
	}
	return &schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + masterId,
		NodeID: masterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"backup-cronjob": []byte(yamlTemplate),
				},
				TimeOut: timeout,
			},
		},
	}, nil
}
