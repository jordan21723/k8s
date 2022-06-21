package task_breaker

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/util"
	"net/http"
	"strings"

	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
)

type ClusterNodeTaskBreakDown struct {
	Operation          schema.Operation
	Cluster            schema.Cluster
	Config             serverConfig.Config
	NodesToAddOrRemove []schema.ClusterNode
	NodeCollectionList schema.NodeInformationCollection
	Action             string
	SumLicenseLabel    uint16
}

func (breakdown *ClusterNodeTaskBreakDown) BreakDownTask() (schema.Operation, error) {
	if len(breakdown.NodesToAddOrRemove) == 0 || len(breakdown.NodeCollectionList) == 0 {
		log.Warn("Nothing to do due to node collection or no nodes need to add or remove to or from cluster")
		return breakdown.Operation, nil
	}

	reducedNodeList := map[string]schema.NodeInformation{}
	masterNodeCollection := map[string]schema.NodeInformation{}
	fullNodeList := map[string]schema.NodeInformation{}
	var hosts []string
	virtualKubeletCollection := map[string]schema.ClusterNode{}
	lbCollection := map[string]schema.NodeInformation{}
	hostnameChecker := map[string]byte{}
	nodesToDeleteHostname := []string{}
	cordonsCmd := map[int][]string{}
	workListNoReduces := []schema.NodeInformation{}

	for _, worker := range breakdown.Cluster.Workers {
		if _, found := breakdown.NodeCollectionList[worker.NodeId]; found {
			workListNoReduces = append(workListNoReduces, breakdown.NodeCollectionList[worker.NodeId])
			reducedNodeList[worker.NodeId] = breakdown.NodeCollectionList[worker.NodeId]
			fullNodeList[worker.NodeId] = breakdown.NodeCollectionList[worker.NodeId]
		} else {
			return breakdown.Operation, errors.New(fmt.Sprintf("Unable to find worker node information with node id: %s", worker.NodeId))
		}
	}

	for index, master := range breakdown.Cluster.Masters {
		if _, found := breakdown.NodeCollectionList[master.NodeId]; !found {
			return breakdown.Operation, errors.New(fmt.Sprintf("Unable to find master node information with node id: %s", master.NodeId))
		}
		// use map to do hostname check
		node := breakdown.NodeCollectionList[master.NodeId]
		hostnameChecker[node.SystemInfo.Node.Hostname] = '1'
		hosts = append(hosts, breakdown.NodeCollectionList[master.NodeId].Ipv4DefaultIp+"\t"+generatorHostname(index, breakdown.Operation.ClusterId, constants.MasterHostnameSuffix))
		masterNodeCollection[master.NodeId] = breakdown.NodeCollectionList[master.NodeId]
		fullNodeList[master.NodeId] = breakdown.NodeCollectionList[master.NodeId]
		delete(reducedNodeList, master.NodeId)
	}

	workerMapping := map[string]schema.NodeInformation{}

	for index, worker := range breakdown.NodesToAddOrRemove {
		if _, found := breakdown.NodeCollectionList[worker.NodeId]; !found {
			return breakdown.Operation, errors.New(fmt.Sprintf("Unable to find worker node information with node id: %s", worker.NodeId))
		}

		if breakdown.NodeCollectionList[worker.NodeId].Role&constants.NodeRoleWorker == constants.NodeRoleWorker {
			// if the node is a worker node
			workerMapping[worker.NodeId] = breakdown.NodeCollectionList[worker.NodeId]
		}

		if worker.UseVirtualKubelet {
			virtualKubeletCollection[worker.NodeId] = worker
		}

		if breakdown.Action == constants.ActionDelete {
			// found nodes whose is also a lb
			if breakdown.NodeCollectionList[worker.NodeId].Role&constants.NodeRoleExternalLB == constants.NodeRoleExternalLB {
				lbCollection[worker.NodeId] = breakdown.NodeCollectionList[worker.NodeId]
			}

			if breakdown.NodeCollectionList[worker.NodeId].Role&constants.NodeRoleWorker == constants.NodeRoleWorker {
				nodesToDeleteHostname = append(nodesToDeleteHostname, breakdown.NodeCollectionList[worker.NodeId].SystemInfo.Node.Hostname)
				cordonsCmd[index] = []string{"kubectl", "drain", breakdown.NodeCollectionList[worker.NodeId].SystemInfo.Node.Hostname, "--ignore-daemonsets", "--delete-local-data"}
			}
		}
	}

	breakdown.Operation.Step = map[int]*schema.Step{}

	if breakdown.Action == constants.ActionCreate {
		stepPrintJoinString := &schema.Step{
			Id:        "step0-" + breakdown.Operation.Id,
			Name:      "GetJoinString",
			NodeSteps: []schema.NodeStep{createJoinStringNodeStep(breakdown.Cluster.Masters[0].NodeId)},
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepPrintJoinString

		if breakdown.Config.EnableHostnameRename {
			renameWorkerStep, errRenameWorkerStep := createRenameWorkerHostnameNodeTaskWithNodeID(workListNoReduces, breakdown.NodesToAddOrRemove, hosts, breakdown.Cluster.ClusterId, constants.WorkerHostnameSuffix, len(reducedNodeList), 5)

			if errRenameWorkerStep != nil {
				return breakdown.Operation, errRenameWorkerStep
			}

			stepRenameWorker := &schema.Step{
				Id:        "step1-" + breakdown.Operation.Id,
				Name:      "RenameWorkerHostname",
				NodeSteps: renameWorkerStep,
			}

			breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepRenameWorker
		}

	} else if breakdown.Action == constants.ActionDelete {
		if len(nodesToDeleteHostname) > 0 {
			// run kubectl drain node xxxx
			stepCordonNodes := &schema.Step{
				Id:   "cordon-nodes-" + breakdown.Operation.Id,
				Name: "cordon nodes",
				NodeSteps: []schema.NodeStep{
					{
						Id:     utils.GenNodeStepID(),
						Name:   "cordon nodes",
						NodeID: breakdown.Cluster.Masters[0].NodeId,
						Tasks: map[int]schema.ITask{
							0: schema.TaskRunCommand{
								TaskType:    constants.TaskTypeRunCommand,
								TimeOut:     len(nodesToDeleteHostname) * 60,
								Commands:    cordonsCmd,
								IgnoreError: true,
							},
						},
					},
				},
			}
			breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepCordonNodes

			// run kubectl delete node xxx
			stepKubectlDeleteNode := &schema.Step{
				Id:        "step0-" + breakdown.Operation.Id,
				Name:      "KubectlDeleteNode",
				NodeSteps: []schema.NodeStep{createKubectlDeleteNodesNodeTask(nodesToDeleteHostname, breakdown.Cluster.Masters[0].NodeId)},
			}
			breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepKubectlDeleteNode
		}

		if len(lbCollection) > 0 {
			nodeStepsRemoveClusterLB := []schema.NodeStep{}
			CommonBatchRun(lbCollection, breakdown.Cluster.Action, 10, [][]string{
				// remove keepalive
				{"systemctl", "stop", "keepalived.service"},
				{"systemctl", "disable", "keepalived.service"},
				{"rm", "-f", "/etc/keepalived/keepalived.conf"},
				// remove traefik
				{"systemctl", "stop", "traefik.service"},
				{"systemctl", "disable", "traefik.service"},
				{"rm", "-f", "/usr/lib/systemd/system/traefik.service"},
				{"rm", "-rf", "/etc/traefik"},
				{"rm", "-f", "/usr/local/bin/traefik"},
			}, &nodeStepsRemoveClusterLB, false)

			stepRemoveLb := &schema.Step{
				Id:                       "step-" + breakdown.Operation.Id,
				Name:                     "RemoveExternalLB",
				NodeSteps:                nodeStepsRemoveClusterLB,
				OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
				OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			}
			breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepRemoveLb
		}

	}

	if len(workerMapping) > 0 {
		stepStepUpCRI := &schema.Step{
			Id:        "step2-" + breakdown.Operation.Id,
			Name:      "InstallOrRemoveCRI",
			NodeSteps: createCRINodeTask(breakdown.Cluster.ContainerRuntime, workerMapping, breakdown.Action, breakdown.Config.TaskTimeOut.TaskCRI),
		}

		if breakdown.Action == constants.ActionDelete {
			// only ignore error when deleting
			// in case the node failed to join workers previously during setup cri
			stepStepUpCRI.OnStepDoneOrErrorHandler = OnErrorIgnoreHandler{}
		}

		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepStepUpCRI

		stepSetUpLocalHaproxy := &schema.Step{
			Id:        "Step3-" + breakdown.Operation.Id,
			Name:      "AddOrRemoveLocalVip",
			NodeSteps: createWorkNodeVipNodeTask(workerMapping, masterNodeCollection, breakdown.Action, breakdown.Config.TaskTimeOut.TaskVip),
		}

		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepSetUpLocalHaproxy

		stepJoinWorker := &schema.Step{
			Id:        "step4-" + breakdown.Operation.Id,
			Name:      "JoinOrDestroyWorkNode",
			NodeSteps: createKubeadmJoinWorkerNodeTask(workerMapping, breakdown.Action, breakdown.Config.TaskTimeOut.TaskKubeadmJoinWorker),
		}

		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepJoinWorker

		if breakdown.Action == constants.ActionCreate {
			if breakdown.Cluster.CNI.CNIType == constants.CNITypeCalico && strings.Contains(breakdown.Cluster.CNI.Calico.CalicoMode, "Vxlan") || breakdown.Cluster.CNI.CNIType == constants.CNITypeFlannel {
				nodeSteps := []schema.NodeStep{}
				centos7List := map[string]schema.NodeInformation{}
				for nodeId, nodeInfo := range workerMapping {
					if nodeInfo.SystemInfo.OS.Vendor == "centos" && nodeInfo.SystemInfo.OS.Version == "7" {
						centos7List[nodeId] = nodeInfo
					}
				}
				if len(centos7List) > 0 {
					stepWaitDevice := CreateWaitStep(15, "device vxlan.calico to appear")
					breakdown.Operation.Step[len(breakdown.Operation.Step)] = &stepWaitDevice
					CommonBatchRun(centos7List, breakdown.Cluster.Action, 10, [][]string{
						{"ethtool", "--offload", "vxlan.calico", "rx", "off", "tx", "off"},
					}, &nodeSteps, false)
					stepDisableOffload := &schema.Step{
						Id:                       "stepDisableOffload-" + breakdown.Operation.Id,
						Name:                     "stepDisableOffload",
						WaitBeforeRun:            0,
						NodeSteps:                nodeSteps,
						OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
						OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
					}
					breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepDisableOffload
				}
			}

		}
	}

	if breakdown.Action == constants.ActionDelete && len(workerMapping) > 0 {
		// remove all kubelet service systemd file
		nodeSteps := []schema.NodeStep{}
		CommonBatchRun(workerMapping, breakdown.Cluster.Action, 10, [][]string{
			{"rm", "-f", "/usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf"},
			{"rm", "-f", "/usr/lib/systemd/system/kubelet.service"},
			{"rm", "-rf", "/etc/kubernetes/"},
			{"rm", "-rf", "/var/lib/etcd/"},
			// sometime kubeadm reset will not clean all shit, we have to ensure all related data is deleted by running it again
			{"kubeadm", "reset", "-f"},
			{"umount", "/var/lib/kubelet/pods"},
			{"rm", "-rf", "/var/lib/kubelet/"},
		}, &nodeSteps, false)
		stepRemoveKubeletSystemd := &schema.Step{
			Id:                       "StepRemoveKubeletSystemd-" + breakdown.Operation.Id,
			Name:                     "StepRemoveKubeletSystemd",
			WaitBeforeRun:            0,
			NodeSteps:                nodeSteps,
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepRemoveKubeletSystemd
	}

	addOnsRegister := &AddOnsRegister{}

	addOnsRegister.RegisterUp(AddOnHarbor{
		Name:         "HarborSetup",
		TaskTimeOut:  10,
		FullNodeList: fullNodeList,
	}.SetDataWithPlugin(breakdown.Cluster.Harbor, 0, breakdown.Cluster))

	if breakdown.Cluster.CloudProvider != nil {
		addOnsRegister.RegisterUp(AddOnOpenStackCloudProvider{
			FullNodeList: fullNodeList,
			Name:         "OpenStackCloudProviderAddOn",
			TaskTimeOut:  5,
		}.SetDataWithPlugin(breakdown.Cluster.CloudProvider.OpenStack, breakdown.SumLicenseLabel, breakdown.Cluster))
	}

	if errs := InstallAddOnsWithNodeAddOrRemove(addOnsRegister, &breakdown.Operation, breakdown.Cluster, breakdown.Config, breakdown.Action, breakdown.NodesToAddOrRemove); len(errs) > 0 {
		// do nothing meaning we ignore add-on installation error
	}

	return breakdown.Operation, nil
}

func (breakdown *ClusterNodeTaskBreakDown) BreakDownRancherTask() (schema.Operation, error) {
	if len(breakdown.NodesToAddOrRemove) == 0 || len(breakdown.NodeCollectionList) == 0 {
		log.Warn("Nothing to do due to node collection or no nodes need to add or remove to or from cluster")
		return breakdown.Operation, nil
	}
	breakdown.Operation.Step = map[int]*schema.Step{}

	return rancherBreakDownTask(breakdown)

}

func rancherBreakDownTask(breakdown *ClusterNodeTaskBreakDown) (schema.Operation, error) {
	if breakdown.Cluster.Rancher.RancherAddr == "" {
		return breakdown.Operation, errors.New("rancher addr is nil")
	}
	if breakdown.Cluster.Rancher.RancherToken == "" {
		return breakdown.Operation, errors.New("rancher api token is nil")
	}
	if breakdown.Cluster.Rancher.ManagedByClusterName == "" {
		return breakdown.Operation, errors.New("The rancher cluster name being manipulated cannot be empty")
	}

	if breakdown.Action == constants.ActionCreate {
		joinNodeStep := []schema.NodeStep{}
		for _, node := range breakdown.NodesToAddOrRemove {
			nodeCmd, err := getRancherNodeCommand(breakdown.Cluster.Rancher.RancherAddr,
				breakdown.Cluster.Rancher.RancherToken, breakdown.Cluster.Rancher.ManagedByClusterName, []string{"worker"})
			if err != nil {
				log.Warn("rancher operation error: ", err.Error())
				return breakdown.Operation, err
			}
			cmdArray := []string{}
			if nodeCmd != "" {
				cmdArray = strings.Split(nodeCmd, " ")
				if len(cmdArray) > 1 {
					cmdArray = cmdArray[1:]
				} else {
					return breakdown.Operation, errors.New("Failed to get the correct command")
				}
			}
			CommonRunBatchWithSingleNode(node.NodeId, breakdown.Action, 10, [][]string{cmdArray}, &joinNodeStep, true, false)
		}
		stepJoinNode := &schema.Step{
			Id:        "step0-" + breakdown.Operation.Id,
			Name:      "JoinRancherNode",
			NodeSteps: joinNodeStep,
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepJoinNode
	}

	if breakdown.Action == constants.ActionDelete {
		nodeSteps := []schema.NodeStep{}
		for _, node := range breakdown.NodesToAddOrRemove {
			err := removeRancherNode(breakdown.Cluster.Rancher.RancherAddr, breakdown.Cluster.Rancher.RancherToken,
				breakdown.Cluster.Rancher.ManagedByClusterName, node.NodeId)
			if err != nil {
				return breakdown.Operation, err
			}
			nodeSteps = append(nodeSteps, clearRancherNodeEnvTask(node.NodeId)...)
		}
		stepJoinNode := &schema.Step{
			Id:        "step0-" + breakdown.Operation.Id,
			Name:      "RemoveRancherNode",
			NodeSteps: nodeSteps,
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepJoinNode
	}

	return breakdown.Operation, nil
}

// rancherAddr: https://172.25.0.150:443
// rancherToken: token-5xbf4:xh7hdxzdvt772hhdf87fts8wxvb59t96jqj79428pr5695jf58zdzr
// rancherClusterName: c-txyu
// nodeId: ki-client node id
// nodeRoles: controlPlane/worker/etcd
func getRancherNodeCommand(rancherAddr, rancherToken, rancherClusterName string, nodeRoles []string) (string, error) {
	if rancherClusterName == "local" {
		return "", errors.New("rancher local cluster can't join node to cluster ")
	}
	if len(nodeRoles) == 0 {
		return "", errors.New("nodes must have a role")
	}
	for _, v := range nodeRoles {
		switch v {
		case "controlPlane":
			continue
		case "worker":
			continue
		case "etcd":
			continue
		default:
			return "", errors.New(fmt.Sprintf("node role '%s' is error", v))
		}
	}

	clusterUrl := fmt.Sprintf("%s%s", rancherAddr, `/v3/cluster`)
	header := map[string]string{
		"Authorization": fmt.Sprintf("%s %s", "Basic", base64.StdEncoding.EncodeToString([]byte(rancherToken))),
	}
	resp, code, err := util.CommonRequest(clusterUrl, http.MethodGet, "", nil, header, true, true, 0)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK {
		return "", errors.New(fmt.Sprintf("request code: %d . responese: %s", code, string(resp)))
	}
	clusterResp := schema.RancherResource{}
	err = json.Unmarshal(resp, &clusterResp)
	if err != nil {
		return "", err
	}

	clusterRegistrationLink := ""
	for _, data := range clusterResp.Data {
		if data.Name == rancherClusterName {
			clusterRegistrationLink = data.Links.ClusterRegistrationTokens
			break
		}
	}

	resp, code, err = util.CommonRequest(clusterRegistrationLink, http.MethodGet, "", nil, header, true, true, 0)
	if err != nil {
		return "", err
	}
	if code != http.StatusOK {
		return "", errors.New(fmt.Sprintf("request code: %d . responese: %s", code, string(resp)))
	}
	clusterRegistrationResp := schema.RancherClusterRegistration{}
	err = json.Unmarshal(resp, &clusterRegistrationResp)
	if err != nil {
		return "", err
	}
	for _, data := range clusterRegistrationResp.Data {
		if data.NodeCommand != "" {
			nodeCommand := data.NodeCommand
			for _, role := range nodeRoles {
				nodeCommand = fmt.Sprintf("%s --%s", nodeCommand, role)
			}
			return nodeCommand, nil
		}
	}
	return "", nil
}

// curl --request DELETE -u token-zj8nl:57ps97n6xlqgzrg9w2b849ppxqpnvwkzdkpj4fsdflx79xzzrtbjvf -k https://172.20.163.120:32508/v3/nodes/local:machine-x5rxg
func removeRancherNode(rancherAddr, rancherToken, rancherClusterName, nodeId string) error {
	if rancherClusterName == "local" {
		return errors.New("rancher local cluster can't remove node to cluster ")
	}
	nodeResp, err := listRancherNode(rancherAddr, rancherToken, rancherClusterName)
	if err != nil {
		return err
	}

	cli := cache.GetCurrentCache()
	nodeInfo, err := cli.GetNodeInformation(nodeId)
	if err != nil {
		return err
	}

	delLink := ""
	for _, v := range nodeResp.Data {
		if v.IPAddress == nodeInfo.Ipv4DefaultIp {
			delLink = fmt.Sprintf("%s/v3/nodes/%s", rancherAddr, v.Id)
			break
		}
	}
	if delLink == "" {
		log.Warn("not find this node from rancher cluster")
		return nil
	}
	header := map[string]string{
		"Authorization": fmt.Sprintf("%s %s", "Basic", base64.StdEncoding.EncodeToString([]byte(rancherToken))),
	}
	resp, code, err := util.CommonRequest(delLink, http.MethodDelete, "", nil, header, true, true, 0)
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return errors.New(fmt.Sprintf("request code: %d . responese: %s", code, string(resp)))
	}
	return nil
}

func listRancherNode(rancherAddr, rancherToken, rancherClusterName string) (*schema.RancherNodeInfo, error) {
	clusterUrl := fmt.Sprintf("%s%s", rancherAddr, `/v3/cluster`)
	header := map[string]string{
		"Authorization": fmt.Sprintf("%s %s", "Basic", base64.StdEncoding.EncodeToString([]byte(rancherToken))),
	}
	resp, code, err := util.CommonRequest(clusterUrl, http.MethodGet, "", nil, header, true, true, 0)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("request code: %d . responese: %s", code, string(resp)))
	}
	clusterResp := schema.RancherResource{}
	err = json.Unmarshal(resp, &clusterResp)

	if err != nil {
		return nil, err
	}

	listNodeLink := ""
	for _, data := range clusterResp.Data {
		if data.Name == rancherClusterName {
			listNodeLink = data.Links.Nodes
		}
	}
	resp, code, err = util.CommonRequest(listNodeLink, http.MethodGet, "", nil, header, true, true, 0)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("request code: %d . responese: %s", code, string(resp)))
	}

	nodeResp := schema.RancherNodeInfo{}
	err = json.Unmarshal(resp, &nodeResp)

	return &nodeResp, err
}

func clearRancherNodeEnvTask(nodeId string) []schema.NodeStep {
	nodeSteps := []schema.NodeStep{}
	CommonRunBatchWithSingleNode(nodeId, "", 20, [][]string{
		{"bash", "-c", `docker stop $(docker ps -a -q) && docker rm $(docker ps -a -q)`},
		{"bash", "-c", `ip link del flannel.4096`},
		{"bash", "-c", `ip link del flannel.1`},
		{"bash", "-c", `ip link del cni0`},
		{"bash", "-c", `ip link del kube-ipvs0`},
	}, &nodeSteps, false, true)
	return nodeSteps
}
