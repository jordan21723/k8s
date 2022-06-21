package node

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/message_queue/nats"
	"k8s-installer/pkg/restplus"
	mqHelper "k8s-installer/pkg/server/message_queue"
	"k8s-installer/pkg/util/sshutils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	natClient "github.com/nats-io/nats.go"

	"k8s-installer/internal/apiserver/utils"
	"k8s-installer/pkg/constants"

	"k8s-installer/pkg/cache"
	"k8s-installer/schema"

	"github.com/emicklei/go-restful"
)

func ListNode(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	nodeMap, err := runtimeCache.GetNodeInformationCollection()
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}

	regionId := request.QueryParameter("region_id")
	nodeList := []schema.NodeInformation{}

	if regionId == "" {
		for _, node := range nodeMap {
			nodeList = append(nodeList, node)
		}
	} else {
		for _, node := range nodeMap {
			if node.Region != nil {
				if node.Region.ID == regionId {
					nodeList = append(nodeList, node)
				}
			}
		}
	}

	response.WriteAsJson(nodeList)
}

func ListUsableNode(request *restful.Request, response *restful.Response) {
	rawPageSize := request.QueryParameter("page_size")
	if rawPageSize == "" {
		utils.ResponseError(response, http.StatusBadRequest, "page_size is required")
		return
	}

	rawPageIndex := request.QueryParameter("page_index")
	if rawPageIndex == "" {
		utils.ResponseError(response, http.StatusBadRequest, "page_index is required")
		return
	}

	regionId := request.QueryParameter("region_id")

	pageSize, errParsePageSize := strconv.ParseInt(rawPageSize, 10, 64)
	if errParsePageSize != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("unable to convert page_size with value %s to int64", rawPageSize))
		return
	}

	pageIndex, errParsePageIndex := strconv.ParseInt(rawPageIndex, 10, 64)
	if errParsePageIndex != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("unable to convert page_size with value %s to int64", rawPageIndex))
		return
	}

	clusterInstaller := request.QueryParameter("cluster_installer")

	if clusterInstaller != constants.ClusterInstallerRancher {
		clusterInstaller = constants.ClusterInstallerKubeadm
	}

	runtimeCache := cache.GetCurrentCache()
	nodeMap, totalPage, err := runtimeCache.GetAvailableNodesMap(pageIndex, pageSize, func(information schema.NodeInformation) *schema.NodeInformation {
		if information.Role == 0 && information.Status == constants.StateReady && !information.IsDisabled && (regionId == "" || information.Region.ID == regionId) && information.ClusterInstaller == clusterInstaller {
			return &information
		} else {
			return nil
		}
	})
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Internal server error %s", err.Error()))
		return
	}
	nodeList := []schema.NodeInformation{}
	for _, node := range nodeMap {
		nodeList = append(nodeList, node)
	}
	response.WriteAsJson(struct {
		TotalPage int64                    `json:"total_page"`
		Nodes     []schema.NodeInformation `json:"nodes"`
	}{
		TotalPage: totalPage,
		Nodes:     nodeList,
	})
}

func GetNode(request *restful.Request, response *restful.Response) {
	nodeId := request.PathParameter("node-id")
	if strings.TrimSpace(nodeId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "node-id"))
		return
	}
	runtimeCache := cache.GetCurrentCache()
	node, errGetNodeInfo := runtimeCache.GetNodeInformation(nodeId)
	if errGetNodeInfo != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to query node %s due to error %s", nodeId, errGetNodeInfo.Error()))
		return
	}
	if node == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find node with id %s", nodeId))
		return
	}

	task := schema.TaskRunCommand{
		TaskType: constants.TaskTypeRunCommand,
		TimeOut:  3,
		Commands: map[int][]string{
			0: {"ss", "-lntlp"},
		},
		RequireResult: true,
		IgnoreError:   false,
	}

	data, err := mqHelper.CreateMsgBody("kubeclt-"+uuid.New().String(), "", task, schema.Cluster{}, map[string]string{})
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to create msg body to get port information due to error: %s", err.Error()))
		return
	}

	doneOrErrorSignal := make(chan struct {
		Error   error
		Message string
	})
	config := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
	nats.SendingMessageWithReply(mqHelper.GeneratorNodeSubscribe(*node, config.MessageQueue.SubjectSuffix),
		node.Id,
		node.Id,
		"getNodePort",
		config.MessageQueue,
		3*time.Second,
		[]byte(data),
		func(msg *natClient.Msg) error {
			var reply schema.QueueReply
			if err := json.Unmarshal(msg.Data, &reply); err != nil {
				return err
			}
			node.PortStatus = reply.ReturnData["0"]
			return nil
		},
		func(chan struct {
			Error   error
			Message string
		}) func(nodeId string, nodeStepId string) {
			return func(nodeId string, nodeStepId string) {
				doneOrErrorSignal <- struct {
					Error   error
					Message string
				}{Error: errors.New("Message queue timeout "), Message: fmt.Sprintf("Time out when try to reach node %s", nodeId)}
			}
		}(doneOrErrorSignal))

	response.WriteAsJson(node)
}

func UpdateNodeRole(request *restful.Request, response *restful.Response) {
	nodeId := request.PathParameter("node-id")
	if strings.TrimSpace(nodeId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "node-id"))
		return
	}
	runtimeCache := cache.GetCurrentCache()
	node, errGetNodeInfo := runtimeCache.GetNodeInformation(nodeId)
	if errGetNodeInfo != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to query node %s due to error %s", nodeId, errGetNodeInfo.Error()))
		return
	}
	if node == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find node with id %s", nodeId))
		return
	}

	nodePost := struct {
		Role int `json:"role" description:"1 = master, 2 = worker , 4 = Ingress , 8 = ExternalLB"`
	}{}

	if err := request.ReadEntity(&nodePost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	node.Role = nodePost.Role

	if err := runtimeCache.SaveOrUpdateNodeInformation(nodeId, *node); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to update node %s due to error %s", nodeId, err.Error()))
		return
	}

	response.WriteAsJson(nodePost)

}

func EnableOrDisableNode(request *restful.Request, response *restful.Response) {
	nodeId := request.PathParameter("node-id")
	if strings.TrimSpace(nodeId) == "" {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find path parameter: %s", "node-id"),
		})
		return
	}
	runtimeCache := cache.GetCurrentCache()
	node, errGetNodeInfo := runtimeCache.GetNodeInformation(nodeId)
	if errGetNodeInfo != nil {
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, &schema.HttpErrorResult{
			ErrorCode: http.StatusInternalServerError,
			Message:   fmt.Sprintf("Failed to query node %s due to error %s", nodeId, errGetNodeInfo.Error()),
		})
		return
	}
	if node == nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find node with id %s", nodeId),
		})
		return
	}

	if node.BelongsToCluster != "" {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to change node status %s because this node is belong cluster %s", nodeId, node.BelongsToCluster))
		return
	}

	if node.Role != 0 {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to delete node %s because this node is not clean. The role still set to %d", nodeId, node.Role))
		return
	}

	node.IsDisabled = request.SelectedRoutePath() == "/node/v1/disable/{node-id}"

	if err := runtimeCache.SaveOrUpdateNodeInformation(nodeId, *node); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to disable node %s due to error %s", nodeId, err.Error()))
		return
	}

	response.WriteAsJson(node)

}

func DeleteNode(request *restful.Request, response *restful.Response) {
	nodeId := request.PathParameter("node-id")
	if strings.TrimSpace(nodeId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "node-id"))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	node, errGetNodeInfo := runtimeCache.GetNodeInformation(nodeId)
	if errGetNodeInfo != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to query node %s due to error %s", nodeId, errGetNodeInfo.Error()))
		return
	}
	if node == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find node with id %s", nodeId))
		return
	}

	if node.BelongsToCluster != "" {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to change node status %s because this node is belong cluster %s", nodeId, node.BelongsToCluster))
		return
	}

	if node.Role != 0 {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to delete node %s because this node is not clean. The role still set to %d", nodeId, node.Role))
		return
	}

	if !node.IsDisabled {
		utils.ResponseError(response, http.StatusPreconditionFailed, "For safety reason please disable the node before you remove the node")
		return
	}

	if err := runtimeCache.DeleteNodeInformation(nodeId); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to remove node %s due to error %s", nodeId, err.Error()))
		return
	}
}

func EnableNode(request *restful.Request, response *restful.Response) {
	updateNodeStatus(request, response, false)
}

func DisableNode(request *restful.Request, response *restful.Response) {
	updateNodeStatus(request, response, true)
}

func updateNodeStatus(request *restful.Request, response *restful.Response, disable bool) {
	nodeId := request.PathParameter("node-id")
	if strings.TrimSpace(nodeId) == "" {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find path parameter: %s", "node-id"),
		})
		return
	}
	runtimeCache := cache.GetCurrentCache()
	node, errGetNodeInfo := runtimeCache.GetNodeInformation(nodeId)
	if errGetNodeInfo != nil {
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, &schema.HttpErrorResult{
			ErrorCode: http.StatusInternalServerError,
			Message:   fmt.Sprintf("Failed to query node %s due to error %s", nodeId, errGetNodeInfo.Error()),
		})
		return
	}
	if node == nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find node with id %s", nodeId),
		})
		return
	}

	if node.BelongsToCluster != "" {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to change node status %s because this node is belong cluster %s", nodeId, node.BelongsToCluster))
		return
	}

	if node.Role != 0 {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to delete node %s because this node is not clean. The role still set to %d", nodeId, node.Role))
		return
	}

	node.IsDisabled = disable

	if err := runtimeCache.SaveOrUpdateNodeInformation(nodeId, *node); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to disable node %s due to error %s", nodeId, err.Error()))
		return
	}

	response.WriteAsJson(node)
}

var (
	upGrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024 * 1024 * 10,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func NodeSSH(request *restful.Request, response *restful.Response) {
	rows := restplus.GetIntValueWithDefault(request, "rows", 35)
	cols := restplus.GetIntValueWithDefault(request, "cols", 150)
	msg := request.QueryParameter("msg")
	credential, err := decodedMsgToSSH(msg)
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}
	wsConn, err := upGrader.Upgrade(response.ResponseWriter, request.Request, nil)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
	defer wsConn.Close()

	wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))
	wsConn.SetPongHandler(func(appData string) error {
		log.Debug("in pong handler...")
		wsConn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	quitChan := make(chan struct{}, 3)
	gracefulExitChan := make(chan struct{})
	go func() {
		defer log.Debug("websocket send ping return...")
		for {
			select {
			case <-ticker.C:
				//wsConn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-quitChan:
				return
			}
		}
	}()
	wsConn.SetCloseHandler(func(code int, text string) error {
		log.Debug("in close handler...")
		message := websocket.FormatCloseMessage(code, "")
		wsConn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
		quitChan <- struct{}{}
		return nil
	})

	sshClient, err := sshutils.NewSSHClient(credential)
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("connec to node failed due to error: %s", err.Error()))
		return
	}
	defer sshClient.Close()

	sshConn, err := sshutils.NewLoginSSHWSSession(cols, rows, true, sshClient, wsConn)
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("connec to node ssh sesstion failed due to error: %s", err.Error()))
		return
	}
	defer sshConn.Close()

	sshConn.Start(quitChan)
	go sshConn.Wait(quitChan, gracefulExitChan)
	select {
	case <-quitChan:
		return
	case <-gracefulExitChan:
		time.Sleep(1 * time.Second)
		message := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		if err := wsConn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second)); err != nil {
			log.Error("close websocket error due to: %s", err.Error())
		}
		return
	}
	//<-quitChan
}

func decodedMsgToSSH(msg string) (*schema.SSHCredential, error) {
	c := &schema.SSHCredential{}
	decoded, err := base64.StdEncoding.DecodeString(msg)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(decoded, c)
	if err != nil {
		return nil, err
	}
	err = utils.Validate(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func GetSSHRSAKey(request *restful.Request, response *restful.Response) {
	response.WriteHeaderAndEntity(http.StatusOK, schema.SSHRSAkey{PublicKey: constants.PublicKey})
}
