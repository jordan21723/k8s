package client

import (
	"encoding/json"
	"fmt"

	osSysInfo "k8s-installer/node/os"
	"k8s-installer/node/os/family"
	clientOSFamily "k8s-installer/node/os/family"
	centos "k8s-installer/node/os/family/centos"
	centosVersion "k8s-installer/node/os/family/centos/version"
	"k8s-installer/node/os/family/ubuntu"
	ubuntuVersion "k8s-installer/node/os/family/ubuntu/version"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/schema"

	natsLib "github.com/nats-io/nats.go"
)

/*
client handle incoming message
if error occurred simple wait server message sending timeout
else we send back msg to notify server that this node`s job is done
*/

func MessageHandler(msg *natsLib.Msg) {
	go messageHandler(msg)
}

func messageHandler(msg *natsLib.Msg) {
	log.Debugf("Got incoming msg subject %s", msg.Subject)
	log.Debugf("Got incoming msg content %s", string(msg.Data))
	dataReceived, err := ParseQueueBody(msg.Data)
	if err != nil {
		log.Errorf("handler message error %s", err.Error())
		// reply error
		reply := family.CreateReplyBody("unable to parse", cache.NodeId, constants.StatusError, err.Error(), "", map[string]string{})
		family.ReplyMsg(reply, msg)
		return
	}
	log.Debugf("Start processing node step with id %s", dataReceived.NodeStepId)
	nodeOSFamilyClient, errGetClientOSFamily := GetOSFamily(dataReceived.OperationId)
	if errGetClientOSFamily != nil {
		log.Errorf("Error occurred with method getOSFamily as following %s", errGetClientOSFamily.Error())
		reply := family.CreateReplyBody(dataReceived.OperationId, cache.NodeId, constants.StatusError, errGetClientOSFamily.Error(), dataReceived.NodeStepId, map[string]string{})
		family.ReplyMsg(reply, msg)
		return
	}
	// runtimeCache := cache.GetCurrentCache()
	// config := runtimeCache.GetClientRuntimeConfig(cache.NodeId)
	var clientOperationErr error
	stat := constants.StatusSuccessful
	message := ""
	returnData := map[string]string{}
	switch dataReceived.TaskType {
	case constants.TaskTypeRenameHostname:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().RenameHostname(dataReceived.Task.(schema.TaskRenameHostName), dataReceived.Cluster)
	case constants.TaskTypeBasic:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().BasicNodeSetup(dataReceived.Task.(schema.TaskBasicConfig), dataReceived.Cluster)
	case constants.TaskTypeCRI:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().InstallOrRemoveContainerRuntime(dataReceived.Task.(schema.TaskCRI), dataReceived.Cluster, dataReceived.ResourceServerURL, dataReceived.Task.(schema.TaskCRI).Md5Dep)
	case constants.TaskTypeWorkNodeVip:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().InstallOrRemoveLoadBalance(dataReceived.Task.(schema.TaskLoadBalance), dataReceived.Cluster, dataReceived.ResourceServerURL, dataReceived.Task.(schema.TaskLoadBalance).Md5Dep)
	case constants.TaskTypeCurl:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().GoCurl(dataReceived.Task.(schema.TaskCurl))
	case constants.TaskTypeKubectl:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().RunKubectl(dataReceived.Task.(schema.TaskKubectl), dataReceived.StepReturnData, dataReceived.Cluster)
	case constants.TaskTypeRunCommand:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().RunCommand(dataReceived.Task.(schema.TaskRunCommand), dataReceived.StepReturnData, dataReceived.Cluster)
	case constants.TaskTypeRunAsyncCommand:
		go func() {
			nodeOSFamilyClient.GetOSVersion().AsyncRunCommand(dataReceived.OperationId, cache.NodeId, dataReceived.NodeStepId, msg, dataReceived.Task.(schema.TaskRunCommand), dataReceived.StepReturnData, dataReceived.Cluster)
		}()
		return
	case constants.TaskTypeCopyTextFile:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().CopyTextFile(dataReceived.Task.(schema.TaskCopyTextBaseFile), dataReceived.StepReturnData, dataReceived.Cluster)
	case constants.TaskTypeVirtualKubelet:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().InstallOrDestroyVirtualKubelet(dataReceived.Task.(schema.TaskVirtualKubelet), dataReceived.ResourceServerURL, dataReceived.Cluster, dataReceived.Task.(schema.TaskVirtualKubelet).Md5Dep)
	case constants.TaskPrintJoin:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().PrintJoinString(dataReceived.Task.(schema.TaskPrintJoin))
	case constants.TaskTypeKubeadm:
		kubeadmTask := dataReceived.Task.(schema.TaskKubeadm)
		switch kubeadmTask.KubeadmTask {
		case constants.KubeadmTaskInitFirstControlPlane:
			returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().InitOrDestroyFirstControlPlane(kubeadmTask, dataReceived.ResourceServerURL, dataReceived.Cluster, dataReceived.Task.(schema.TaskKubeadm).Md5Dep)
		case constants.KubeadmTaskJoinControlPlane:
			returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().JoinOrDestroyControlPlane(kubeadmTask, dataReceived.ResourceServerURL, dataReceived.StepReturnData, dataReceived.Cluster, dataReceived.Task.(schema.TaskKubeadm).Md5Dep)
		case constants.KubeadmTaskJoinWorker:
			returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().JoinOrDestroyWorkNode(kubeadmTask, dataReceived.ResourceServerURL, dataReceived.StepReturnData, dataReceived.Cluster, dataReceived.Task.(schema.TaskKubeadm).Md5Dep)
		default:
			clientOperationErr = fmt.Errorf("Kubeadm task type %s is not a valid task type !!!", kubeadmTask.KubeadmTask)
			message = fmt.Sprintf("Kubeadm task type %s is not a valid task type !!!", kubeadmTask.KubeadmTask)
			stat = constants.StatusError
		}
	case constants.TaskLink:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().CommonLink(dataReceived.Task.(schema.TaskLink).From, dataReceived.Task.(schema.TaskLink).SaveTo, dataReceived.Task.(schema.TaskLink).LinkTo)
	case constants.TaskDownloadDep:
		go func() {
			nodeOSFamilyClient.GetOSVersion().CommonDownloadDep(dataReceived.OperationId, cache.NodeId, dataReceived.NodeStepId, msg, dataReceived.ResourceServerURL, dataReceived.Task.(schema.TaskCommonDownloadDep).Dep, dataReceived.Task.(schema.TaskCommonDownloadDep).SaveTo, dataReceived.Task.(schema.TaskCommonDownloadDep).K8sVersion, dataReceived.Task.(schema.TaskCommonDownloadDep).Md5)
		}()
		return
	case constants.TaskDownload:
		go func() {
			nodeOSFamilyClient.GetOSVersion().CommonDownload(dataReceived.OperationId, cache.NodeId, dataReceived.NodeStepId, msg, dataReceived.ResourceServerURL, dataReceived.Task.(schema.TaskCommonDownload).FromDir, dataReceived.Task.(schema.TaskCommonDownload).K8sVersion, dataReceived.Task.(schema.TaskCommonDownload).FileList, dataReceived.Task.(schema.TaskCommonDownload).SaveTo, dataReceived.Task.(schema.TaskCommonDownload).IsUseDefaultPath)
		}()
		return
	case constants.TaskGenerateKSClusterConfig:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().GenerateKSClusterConfig(dataReceived.Cluster, dataReceived.Task.(schema.TaskGenerateKSClusterConfig).IpAddress)
	case constants.TaskConfigPromtail:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().ConfigPromtail(dataReceived.Cluster)
	case constants.TaskTypePreLoadImage:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().PreLoadImage(dataReceived.Task.(schema.TaskPreLoadImage), dataReceived.Cluster.ContainerRuntime.CRIType)
	case constants.TaskTypeSetHost:
		returnData, clientOperationErr = nodeOSFamilyClient.GetOSVersion().TaskSetHost(dataReceived.Task.(schema.TaskSetHosts))
	default:
		clientOperationErr = fmt.Errorf("Task type %s is not a valid task type !!!", dataReceived.TaskType)
		message = fmt.Sprintf("Task type %s is not a valid task type !!!", dataReceived.TaskType)
		stat = constants.StatusError
	}

	if clientOperationErr != nil {
		message = fmt.Sprintf("Operation %s excution failed with node task %s on node %s due to error %s",
			dataReceived.OperationId,
			dataReceived.TaskType,
			cache.NodeId,
			clientOperationErr.Error())
		log.Errorf(message)
		stat = constants.StatusError
	}
	log.Debugf("Complete processing NodeStep with id %s , sending back message to notify server!!!", dataReceived.NodeStepId)
	reply := family.CreateReplyBody(dataReceived.OperationId, cache.NodeId, stat, message, dataReceived.NodeStepId, returnData)
	family.ReplyMsg(reply, msg)
}

func GetOSFamily(operationId string) (clientOSFamily.IOSFamily, error) {
	if osInfo, err := osSysInfo.GetAllSystemInformation(); err != nil {
		log.Errorf("Failed to get node with id %s system information due to err %s", operationId, err.Error())
		return nil, err
	} else {
		runtimeCache := cache.GetCurrentCache()
		cfg := runtimeCache.GetClientRuntimeConfig(cache.NodeId)
		switch osInfo.OS.Vendor {
		case constants.OSFamilyCentos:
			log.Debug("Method getOSFamily go with centos")
			switch osInfo.OS.Version {
			case "7":
				log.Debugf("Method getOSFamily go with centos with version 7")
				return centos.Centos{
					Version: centosVersion.V7{
						Config: cfg,
					},
				}, nil
			default:
				log.Debugf("Centos with version %s is not a support version", osInfo.OS.Version)
				return nil, nil
			}
		case constants.OSFamilyUbuntu:
			return ubuntu.Ubuntu{
				Version: ubuntuVersion.V1804{
					Config: cfg,
				},
			}, nil
		default:
			log.Debugf("Os Family %s is not a support os type", osInfo.OS.Vendor)
			return nil, nil
		}
	}
}

func ParseQueueBody(data []byte) (schema.QueueBody, error) {
	dataReceived := schema.QueueBody{}
	if err := json.Unmarshal(data, &dataReceived); err != nil {
		log.Debugf("Failed to parse msg to struct QueueBody due to error %s", err.Error())
		return schema.QueueBody{}, err
	}
	var err error
	if len(dataReceived.TaskData) > 0 {
		switch dataReceived.TaskType {
		case constants.TaskTypeBasic:
			basicTask := schema.TaskBasicConfig{}
			if err = json.Unmarshal(dataReceived.TaskData, &basicTask); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskBasicConfig due to error %s", err.Error())
			} else {
				dataReceived.Task = basicTask
			}
		case constants.TaskTypeKubectl:
			kubeletTask := schema.TaskKubectl{}
			if err = json.Unmarshal(dataReceived.TaskData, &kubeletTask); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskKubectl due to error %s", err.Error())
			} else {
				dataReceived.Task = kubeletTask
			}
		case constants.TaskTypeCRI:
			taskCRI := schema.TaskCRI{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskCRI); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskCRI due to error %s", err.Error())
			} else {
				dataReceived.Task = taskCRI
			}
		case constants.TaskTypeKubeadm:
			taskKubeadm := schema.TaskKubeadm{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskKubeadm); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskKubeadm due to error %s", err.Error())
			} else {
				dataReceived.Task = taskKubeadm
			}
		case constants.TaskTypeWorkNodeVip:
			taskVip := schema.TaskLoadBalance{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskVip); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskLoadBalance due to error %s", err.Error())
			} else {
				dataReceived.Task = taskVip
			}
		case constants.TaskTypeCurl:
			taskCurl := schema.TaskCurl{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskCurl); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskCurl due to error %s", err.Error())
			} else {
				dataReceived.Task = taskCurl
			}
		case constants.TaskTypeRenameHostname:
			taskRenameHostname := schema.TaskRenameHostName{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskRenameHostname); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskRenameHostName due to error %s", err.Error())
			} else {
				dataReceived.Task = taskRenameHostname
			}
		case constants.TaskTypeRunCommand:
			taskRunCommand := schema.TaskRunCommand{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskRunCommand); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskRunCommand due to error %s", err.Error())
			} else {
				dataReceived.Task = taskRunCommand
			}
		case constants.TaskTypeRunAsyncCommand:
			taskRunCommand := schema.TaskRunCommand{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskRunCommand); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskRunCommand due to error %s", err.Error())
			} else {
				dataReceived.Task = taskRunCommand
			}
		case constants.TaskTypeCopyTextFile:
			taskCopyTextBaseFile := schema.TaskCopyTextBaseFile{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskCopyTextBaseFile); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskCopyTextBaseFile due to error %s", err.Error())
			} else {
				dataReceived.Task = taskCopyTextBaseFile
			}
		case constants.TaskTypeVirtualKubelet:
			taskVirtualKubelet := schema.TaskVirtualKubelet{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskVirtualKubelet); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskVirtualKubelet due to error %s", err.Error())
			} else {
				dataReceived.Task = taskVirtualKubelet
			}
		case constants.TaskPrintJoin:
			taskPrintJoin := schema.TaskPrintJoin{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskPrintJoin); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskPrintJoin due to error %s", err.Error())
			} else {
				dataReceived.Task = taskPrintJoin
			}
		case constants.TaskLink:
			taskLink := schema.TaskLink{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskLink); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct taskLink due to error %s", err.Error())
			} else {
				dataReceived.Task = taskLink
			}
		case constants.TaskDownloadDep:
			taskCommonDownloadDep := schema.TaskCommonDownloadDep{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskCommonDownloadDep); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct taskLink due to error %s", err.Error())
			} else {
				dataReceived.Task = taskCommonDownloadDep
			}
		case constants.TaskDownload:
			taskCommonDownload := schema.TaskCommonDownload{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskCommonDownload); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct taskLink due to error %s", err.Error())
			} else {
				dataReceived.Task = taskCommonDownload
			}
		case constants.TaskGenerateKSClusterConfig:
			taskGenerateKSClusterConfig := schema.TaskGenerateKSClusterConfig{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskGenerateKSClusterConfig); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct taskLink due to error %s", err.Error())
			} else {
				dataReceived.Task = taskGenerateKSClusterConfig
			}
		case constants.TaskConfigPromtail:
			taskConfigPromtail := schema.TaskConfigPromtail{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskConfigPromtail); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct taskLink due to error %s", err.Error())
			} else {
				dataReceived.Task = taskConfigPromtail
			}
		case constants.TaskTypePreLoadImage:
			taskPreLoadImage := schema.TaskPreLoadImage{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskPreLoadImage); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct taskPreLoadImage due to error %s", err.Error())
			} else {
				dataReceived.Task = taskPreLoadImage
			}
		case constants.TaskTypeSetHost:
			taskSetHosts := schema.TaskSetHosts{}
			if err = json.Unmarshal(dataReceived.TaskData, &taskSetHosts); err != nil {
				log.Debugf("Failed to parse msg.TaskData to struct TaskSetHosts due to error %s", err.Error())
			} else {
				dataReceived.Task = taskSetHosts
			}
		default:
			err = fmt.Errorf("Task type %s is not valid task type", dataReceived.TaskType)
		}
	}
	return dataReceived, err
}
