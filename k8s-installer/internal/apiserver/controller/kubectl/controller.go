package kubectl

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/google/uuid"
	natClient "github.com/nats-io/nats.go"
	"k8s-installer/internal/apiserver/controller/common"
	"k8s-installer/internal/apiserver/utils"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	mqHelper "k8s-installer/pkg/server/message_queue"
	"k8s-installer/schema"
	k8sCore "k8s.io/api/core/v1"
	"net/http"
	"strconv"
	"strings"
)

func KubectlExec(request *restful.Request, response *restful.Response) {

	kubectlEntity := schema.KubectlExec{}
	if err := request.ReadEntity(&kubectlEntity); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	if err := utils.Validate(kubectlEntity); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	runtimeCache := cache.GetCurrentCache()
	cluster, err := runtimeCache.GetClusterWithShortId(kubectlEntity.ClusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	if cluster == nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Unable to find cluster with id %s", kubectlEntity.ClusterId))
		return
	}

	node, errGetNodeInfo := runtimeCache.GetNodeInformation(cluster.Masters[0].NodeId)
	if errGetNodeInfo != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to query node %s due to error %s", cluster.Masters[0].NodeId, errGetNodeInfo.Error()))
		return
	}
	if node == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find node with id %s", cluster.Masters[0].NodeId))
		return
	}

	// get pod with labels
	commandGetRelatedPods := "kubectl get pods -n %s %s -o json"

	// append all labels
	allLabels := ""

	for key, val := range kubectlEntity.Labels {
		if len(allLabels) == 0 {
			allLabels += "-l "
			allLabels += fmt.Sprintf("%s=%s", key, val)
		} else {
			allLabels += fmt.Sprintf(",%s=%s", key, val)
		}
	}

	commandGetRelatedPods = fmt.Sprintf(commandGetRelatedPods, kubectlEntity.Namespace, allLabels)
	task := schema.TaskRunCommand{
		TaskType: constants.TaskTypeRunCommand,
		TimeOut:  3,
		Commands: map[int][]string{
			0: strings.Split(commandGetRelatedPods, " "),
		},
		RequireResult: true,
		IgnoreError:   false,
	}

	data, err := mqHelper.CreateMsgBody("kubeclt-"+uuid.New().String(), "", task, schema.Cluster{}, map[string]string{})
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to create msg body to get port information due to error: %s", err.Error()))
		return
	}

	config := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
	var podList k8sCore.PodList
	// sending message to get pod list
	err = common.SendingSingleCommand(config.MessageQueue, *node, []byte(data), func(msg *natClient.Msg) error {
		var reply schema.QueueReply
		if err := json.Unmarshal(msg.Data, &reply); err != nil {
			return err
		}
		if reply.Stat == constants.StatusError {
			return errors.New(reply.Message)
		}
		if _, exits := reply.ReturnData["0"]; !exits {
			return errors.New("Return data does not contain key 0 ")
		}
		if err := json.Unmarshal([]byte(reply.ReturnData["0"]), &podList); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get related pod list due to error: %s", err.Error()))
		return
	}

	if len(podList.Items) == 0 {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to get related pod list with labels '%s' in namespace '%s' of cluster '%s'", allLabels, kubectlEntity.Namespace, cluster.ClusterId))
		return
	}

	// loop the pod
	var result []schema.KubectlExecResult

	command := "kubectl exec %s -c %s -n %s -- %s"

	for _, pod := range podList.Items {
		if pod.Status.Phase != k8sCore.PodRunning {
			result = append(result, ErrorKubectlExecResult(pod.Name, "Pod is not in running stat"))
		} else {
			commandsToRun := map[int][]string{}
			for index, container := range kubectlEntity.Containers {
				commandsToRun[index] = strings.Split(fmt.Sprintf(command, pod.Name, container, kubectlEntity.Namespace, kubectlEntity.Command), " ")
			}
			task := schema.TaskRunCommand{
				TaskType:      constants.TaskTypeRunCommand,
				TimeOut:       3,
				Commands:      commandsToRun,
				RequireResult: true,
				IgnoreError:   false,
			}
			data, err := mqHelper.CreateMsgBody("kubeclt-"+uuid.New().String(), "", task, schema.Cluster{}, map[string]string{})
			if err != nil {
				utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to create msg body to get port information due to error: %s", err.Error()))
				return
			}
			err = common.SendingSingleCommand(config.MessageQueue, *node, []byte(data), func(msg *natClient.Msg) error {
				var reply schema.QueueReply
				if err := json.Unmarshal(msg.Data, &reply); err != nil {
					result = append(result, ErrorKubectlExecResult(pod.Name, err.Error()))
					return err
				}
				if reply.Stat == constants.StatusError {
					result = append(result, ErrorKubectlExecResult(pod.Name, reply.Message))
					return errors.New(reply.Message)
				}
				containerResult := map[string]string{}
				for index, container := range kubectlEntity.Containers {
					key := strconv.Itoa(index)
					if _, ok := reply.ReturnData[key]; ok {
						containerResult[container] = reply.ReturnData[key]
					} else {
						containerResult[container] = "Empty reply from node"
					}
				}
				result = append(result, schema.KubectlExecResult{
					PodName:         pod.Name,
					WithError:       false,
					ErrorMessage:    "",
					ContainerResult: containerResult,
				})
				return nil
			})
			if err != nil {
				log.Warnf("Failed to kubectl exec command for pod %s due to error: %s, skip it", pod.Name, err.Error())
			}
		}

	}

	response.WriteAsJson(result)
}

func ErrorKubectlExecResult(pod, errMsg string) schema.KubectlExecResult {
	return schema.KubectlExecResult{
		PodName:         pod,
		WithError:       true,
		ErrorMessage:    errMsg,
		ContainerResult: nil,
	}
}
