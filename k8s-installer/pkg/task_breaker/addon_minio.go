package task_breaker

import (
	"k8s-installer/components/minio"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
	"reflect"
)

type AddOnMinIO struct {
	Name          string
	FirstMasterId string
	TaskTimeOut   int
	enable        bool
}

func (params AddOnMinIO) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// TODO
	return schema.Step{}, nil
}

func (params AddOnMinIO) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		params.enable = false
	} else {
		// do not look into license switch
		params.enable = plugAble.IsEnable()
	}
	return params
}

func (params AddOnMinIO) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (params AddOnMinIO) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return params.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func (params AddOnMinIO) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if cluster.MinIO == nil {
		return nil
	}

	if action == constants.ActionCreate {
		// setup minio
		masterIdList := make([]string, len(cluster.Masters))
		for i, val := range cluster.Masters {
			masterIdList[i] = val.NodeId
		}
		if cluster.MinIO.ImageRegistry == "" {
			cluster.MinIO.ImageRegistry = joinPrivateRegistry(cluster.ContainerRuntime.PrivateRegistryAddress,
				cluster.ContainerRuntime.PrivateRegistryPort)
		}
		return params.installTask(cluster.ClusterId, config.BackupPath, masterIdList, 20, cluster.MinIO, operation)
	} else if action == constants.ActionDelete {
		// TODO
		// uninstallMinIOTask
	}
	return nil
}

func (params AddOnMinIO) GetAddOnName() string {
	return params.Name
}

func (params AddOnMinIO) IsEnable() bool {
	return params.enable
}

func (params AddOnMinIO) installTask(clusterId, backupPath string, nodeId []string, timeOut int, minIO *minio.DeployMinIO, operation *schema.Operation) error {
	if len(nodeId) == 0 {
		log.Warn("zero master node install minIO")
		return nil
	}
	// 1. mkdir velero file dir
	step1 := make([]schema.NodeStep, len(nodeId))
	for i, val := range nodeId {
		step1[i] = createDir(val, minIO.BackupPath, timeOut)
	}
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:                       "setupInstallMinIO-CreateDir" + operation.Id,
		Name:                     "setupInstallMinIO-CreateDir",
		NodeSteps:                step1,
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
	}

	// 2. install minio
	step2, err := installMinIO(clusterId, backupPath, nodeId[0], timeOut, minIO)
	if err != nil {
		return err
	}
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:                       "setupInstallMinIO-ApplyDeploy" + operation.Id,
		Name:                     "setupInstallMinIO-ApplyDeploy",
		NodeSteps:                []schema.NodeStep{step2},
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
	}

	return nil
}

func uninstallTask(nodeId string, timeOut int, cleanData bool, minIO *minio.DeployMinIO, operation *schema.Operation) error {
	// TODO
	return nil
}

func createDir(masterNodeId, dirPath string, timeOut int) schema.NodeStep {
	return schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskRunCommand-" + masterNodeId,
		NodeID: masterNodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType: constants.TaskTypeRunCommand,
				Commands: map[int][]string{
					0: {
						"mkdir", "-pv", dirPath,
					},
				},
				TimeOut: timeOut,
			},
		},
	}
}

func installMinIO(clusterId, backupPath, masterNodeId string, timeOut int, minIO *minio.DeployMinIO) (schema.NodeStep, error) {
	mTmpl, err := minIO.TemplateRender(clusterId, backupPath)
	if err != nil {
		return schema.NodeStep{}, err
	}
	return schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskRunCommand-" + masterNodeId,
		NodeID: masterNodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"minio-deployment": []byte(mTmpl),
				},
				TimeOut: timeOut,
			},
		},
	}, nil
}
