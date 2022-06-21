package task_breaker

import (
	"fmt"
	"k8s-installer/components/velero"
	"k8s-installer/pkg/cache"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/google/uuid"
)

type AddOnVelero struct {
	Name          string
	FirstMasterId string
	TaskTimeOut   int
	enable        bool
}

func (params AddOnVelero) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// TODU
	return schema.Step{}, nil
}

func (params AddOnVelero) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		params.enable = false
	} else {
		// do not look into license switch
		params.enable = plugAble.IsEnable()
	}
	return params
}

func (params AddOnVelero) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (params AddOnVelero) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return params.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func (params AddOnVelero) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if cluster.Velero == nil {
		return nil
	}

	if action == constants.ActionCreate {
		masterInfo, err := cache.GetCurrentCache().GetNodeInformation(cluster.Masters[0].NodeId)
		if err != nil {
			return err
		}
		params.installTask(masterInfo.Ipv4DefaultIp, cluster, operation)
		return nil
	} else if action == constants.ActionDelete {
		// TODO
		// uninstallMinIOTask
	}
	return nil
}

func (params AddOnVelero) installTask(masterIp string, cluster schema.Cluster, operation *schema.Operation) {
	stepsDown := []schema.NodeStep{}
	stepsChmod := []schema.NodeStep{}
	for _, v := range cluster.Masters {
		stepsDown = append(stepsDown, downVeleroFile(v.NodeId, cluster.Action, 30))
		stepsChmod = append(stepsChmod, chmodVeleroFile(v.NodeId, cluster.Action, 30))
	}
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:                       "setupInstallVelero-Down" + operation.Id,
		Name:                     "setupInstallVelero-Down",
		NodeSteps:                stepsDown,
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
	}
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:                       "setupInstallVelero-Chmod" + operation.Id,
		Name:                     "setupInstallVelero-Chmod",
		NodeSteps:                stepsChmod,
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
	}

	pr := joinPrivateRegistry(
		cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort)

	secretText := `[default]
aws_access_key_id = minio
aws_secret_access_key = minio123`

	installCmd := []byte(fmt.Sprintf("%s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s %s",
		"velero", "install", "--provider", "aws", "--plugins", pr+"/caas4/velero-plugin-for-aws:v1.0.0",
		"--bucket", "velero", "--image", pr+"/caas4/velero:v1.6.0", "--secret-file", "/usr/share/k8s-installer/velero/credentials-velero",
		"--use-volume-snapshots=true", "--snapshot-location-config", "region=default", "--use-restic", "--backup-location-config",
		"region=minio,s3ForcePathStyle=true,s3Url=http://minio.minio.svc:9000,publicUrl=http://"+masterIp+":30069"))

	operation.Step[len(operation.Step)] = &schema.Step{
		Id:   "stepCreateSecretText" + operation.Id,
		Name: "stepCreateSecretText",
		NodeSteps: setupCreateVeleroSecrte(
			cluster.Masters[0].NodeId, params.TaskTimeOut,
			map[string][]byte{"/usr/share/k8s-installer/velero/credentials-velero": []byte(secretText)}),
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
	}

	operation.Step[len(operation.Step)] = &schema.Step{
		Id:   "stepCreateChartAndInstall" + operation.Id,
		Name: "stepCreateChartAndInstall",
		NodeSteps: []schema.NodeStep{
			writeFile(cluster.Masters[0].NodeId, params.TaskTimeOut,
				map[string][]byte{"/usr/share/k8s-installer/velero/install.sh": installCmd}),
		},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
	}

	operation.Step[len(operation.Step)] = &schema.Step{
		Id:   "stepCreateChartAndInstall" + operation.Id,
		Name: "stepCreateChartAndInstall",
		NodeSteps: []schema.NodeStep{
			installVelero(cluster.Masters[0].NodeId, params.TaskTimeOut),
		},
		WaitBeforeRun:            2,
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
	}
}

func setupCreateVeleroSecrte(firstMasterId string, timeOut int, textFiles map[string][]byte) []schema.NodeStep {
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
	return []schema.NodeStep{step}
}

func (params AddOnVelero) UninstallTask(firstMasterId string, timeOut int, operation *schema.Operation) {
	// TODO
}

func (params AddOnVelero) GetAddOnName() string {
	return params.Name
}

func (params AddOnVelero) IsEnable() bool {
	return params.enable
}

func createVeleroChartFile(firstMasterId string, timeOut int, textFiles map[string][]byte) schema.NodeStep {
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

func veleroTextFiles(charPath, registry string) map[string][]byte {
	textFiles := make(map[string][]byte)
	// chart.yaml
	textFiles[filepath.Join(charPath, "Chart.yaml")] = []byte(velero.Chart)
	// values.yaml
	textFiles[filepath.Join(charPath, "values.yaml")] = []byte(strings.ReplaceAll(velero.Values, "velero/velero", registry+"/caas4"))
	// crds yaml
	for key, val := range velero.Crds {
		textFiles[filepath.Join(charPath, "crds", key)] = []byte(val)
	}
	// templetes yaml
	for key, val := range velero.Templates {
		textFiles[filepath.Join(charPath, "templates", key)] = []byte(strings.ReplaceAll(val, "caas4", registry+"/caas4"))
	}
	return textFiles
}

func downVeleroFile(masterId string, action string, timeOut int) schema.NodeStep {
	return schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeDownVeleroFile-" + masterId,
		NodeID: masterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskCommonDownloadDep{
				TaskType: constants.TaskDownloadDep,
				Action:   action,
				TimeOut:  timeOut,

				Dep:        velero.VeleroDep,
				SaveTo:     "/usr/bin/",
				K8sVersion: constants.V1_18_6,
				Md5:        GenerateDepMd5(velero.VeleroDep),
			},
		},
	}
}

func chmodVeleroFile(masterId string, action string, timeOut int) schema.NodeStep {
	return schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeChmodVeleroFile-" + masterId,
		NodeID: masterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType: constants.TaskTypeRunCommand,
				Action:   action,
				TimeOut:  timeOut,
				Commands: map[int][]string{
					0: {"chmod", "777", "/usr/bin/velero"}},
			},
		},
	}
}

func setupVeleroMoveTo(masterId string, src, dest, action string, timeOut int) schema.NodeStep {
	return schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeVeleroMoveTo-" + masterId,
		NodeID: masterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType: constants.TaskTypeRunCommand,
				Action:   action,
				TimeOut:  timeOut,
				Commands: map[int][]string{
					0: {"mv", "-f", src, dest},
					1: {"chmod", "777", filepath.Join(dest, "velero")},
				},
			},
		},
	}
}

func writeFile(nodeId string, timeOut int, textFiles map[string][]byte) schema.NodeStep {
	return schema.NodeStep{
		Id:     "TaskTypeCopyTextFile-" + uuid.New().String(),
		Name:   "TaskTypeCopyTextFile",
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskCopyTextBaseFile{
				TaskType:  constants.TaskTypeCopyTextFile,
				TimeOut:   timeOut,
				TextFiles: textFiles,
			},
		},
	}
}

func installVelero(firstMasterId string, timeOut int) schema.NodeStep {
	return schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskRunCommand-" + firstMasterId,
		NodeID: firstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskRunCommand{
				TaskType: constants.TaskTypeRunCommand,
				TimeOut:  timeOut,
				Commands: map[int][]string{
					0: {
						"bash", "/usr/share/k8s-installer/velero/install.sh",
					}},
			},
		},
	}
}
