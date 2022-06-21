package task_breaker

import (
	"errors"
	"fmt"
	"strings"

	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"

	"k8s-installer/components/storage"

	"github.com/google/uuid"

	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
)

type ClusterTaskBreakDown struct {
	Operation          schema.Operation
	Cluster            schema.Cluster
	Config             serverConfig.Config
	NodeCollectionList schema.NodeInformationCollection
	SumLicenseLabel    uint16
}

/*
break down input object to task with seq
*/

func (breakdown *ClusterTaskBreakDown) BreakDownTask() (schema.Operation, error) {
	// reduce the target node
	// remove duplicate nodes for container runtime installation and basic preparation
	// because the work node and master may shared by each other
	reducedNodeList := map[string]schema.NodeInformation{}

	fullNodeList := map[string]schema.NodeInformation{}

	masterNodeCollection := map[string]schema.NodeInformation{}

	// now wo keep the hostname and node id mapping
	nodeToHostNameMapping := map[string]string{}

	needUntaintNodeCollection := map[string]schema.NodeInformation{}

	//virtual kubelet collection
	virtualKubeletCollection := map[string]schema.ClusterNode{}

	// also we have make sure master`s hostname has to be unique
	var hosts []string
	hostnameChecker := map[string]byte{}
	for index, master := range breakdown.Cluster.Masters {
		reducedNodeList[master.NodeId] = breakdown.NodeCollectionList[master.NodeId]
		fullNodeList[master.NodeId] = breakdown.NodeCollectionList[master.NodeId]
		// use map to do hostname check
		node := breakdown.NodeCollectionList[master.NodeId]
		hostnameChecker[node.SystemInfo.Node.Hostname] = '1'
		hosts = append(hosts, breakdown.NodeCollectionList[master.NodeId].Ipv4DefaultIp+"\t"+generatorHostname(index, breakdown.Operation.ClusterId, constants.MasterHostnameSuffix))

		// set up masters collection with node information
		masterNodeCollection[master.NodeId] = breakdown.NodeCollectionList[master.NodeId]
	}
	for _, worker := range breakdown.Cluster.Workers {
		reducedNodeList[worker.NodeId] = breakdown.NodeCollectionList[worker.NodeId]
		fullNodeList[worker.NodeId] = breakdown.NodeCollectionList[worker.NodeId]
		if _, exits := masterNodeCollection[worker.NodeId]; exits {
			// set need to untaint list
			// because master node is required to act like work node too
			needUntaintNodeCollection[worker.NodeId] = breakdown.NodeCollectionList[worker.NodeId]
		}
		if worker.UseVirtualKubelet {
			virtualKubeletCollection[worker.NodeId] = worker
		}
	}

	step2 := &schema.Step{
		Id:        "step2-" + breakdown.Operation.Id,
		Name:      "InstallOrRemoveCRI",
		NodeSteps: createCRINodeTask(breakdown.Cluster.ContainerRuntime, reducedNodeList, breakdown.Cluster.Action, breakdown.Config.TaskTimeOut.TaskCRI),
	}

	InitFirstControlPlaneNodeTask, err := createKubeadmInitFirstControlPlaneNodeTask(
		breakdown.Cluster.Masters[0].NodeId, breakdown.Cluster.Action, breakdown.Config.TaskTimeOut.TaskKubeadmInitFirstControlPlane,
		breakdown.Cluster, breakdown.NodeCollectionList)
	if err != nil {
		return breakdown.Operation, err
	}

	step3 := &schema.Step{
		Id:        "step3-" + breakdown.Operation.Id,
		Name:      "InitOrDestroyFirstControlPlane",
		NodeSteps: []schema.NodeStep{InitFirstControlPlaneNodeTask},
	}

	if breakdown.Operation.Step == nil {
		breakdown.Operation.Step = map[int]*schema.Step{}
	}

	if breakdown.Cluster.Action == constants.ActionDelete {
		// gracefully delete all pod
		nodeStepsReclaimPods := []schema.NodeStep{}

		if breakdown.Cluster.KsClusterConf != nil {
			// add ks created namespace to release storage
			breakdown.Cluster.ReclaimNamespaces = append(breakdown.Cluster.ReclaimNamespaces, []string{"kubesphere-logging-system", "kubesphere-monitoring-system", "kubesphere-system"}...)
			if breakdown.Cluster.ClusterRole == constants.ClusterRoleMember {
				runtimeCache := cache.GetCurrentCache()
				var firstMasterNode *schema.NodeInformation
				targetHostCluster, err := runtimeCache.GetCluster(breakdown.Cluster.KsClusterConf.MemberOfCluster)
				if err != nil {
					log.Error(fmt.Sprintf("Failed to get host cluster due to error: %s", err))
					return breakdown.Operation, err
				}
				if targetHostCluster == nil {
					newErr := errors.New(fmt.Sprintf("(ignore) Failed to find host cluster with id: %s skip it", breakdown.Cluster.KsClusterConf.MemberOfCluster))
					log.Warn(newErr)
					// when deleting cluster if host cluster not found ,just warn it and move on:)
				} else {
					// only call host cluster api to remove this member when host cluster is in constants.ClusterStatusRunning status
					if targetHostCluster.Status == constants.ClusterStatusRunning {
						firstMasterNode, err = runtimeCache.GetNodeInformation(targetHostCluster.Masters[0].NodeId)
						if err != nil {
							log.Error(fmt.Sprintf("Failed to get first master node with node id %s of host cluster %s due to error: %s", targetHostCluster.Masters[0].NodeId, targetHostCluster.ClusterId, err))
							return breakdown.Operation, err
						}

						if firstMasterNode == nil {
							newErr := errors.New(fmt.Sprintf("Failed to get first master node with node id %s of host cluster %s", targetHostCluster.Masters[0].NodeId, targetHostCluster.ClusterId))
							return breakdown.Operation, newErr
						}

						step := schema.NodeStep{
							Id:     utils.GenNodeStepID(),
							Name:   "TaskTypeKubectl-" + breakdown.Cluster.Masters[0].NodeId,
							NodeID: firstMasterNode.Id,
							Tasks: map[int]schema.ITask{
								0: schema.TaskKubectl{
									TaskType:     constants.TaskTypeKubectl,
									SubCommand:   constants.KubectlSubCommandDelete,
									CommandToRun: []string{"cluster " + breakdown.Cluster.ClusterId[0:16]},
									TimeOut:      60,
								},
							},
						}
						breakdown.Operation.Step[len(breakdown.Operation.Step)] = &schema.Step{
							Id:                       "removeClusterConf-" + breakdown.Operation.Id,
							Name:                     "removeClusterConf",
							NodeSteps:                []schema.NodeStep{step},
							OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
							OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
						}
					} else {
						log.Warnf("Host cluster with id %s is not in stat %s skip it", targetHostCluster.ClusterId, constants.ClusterStatusRunning)
					}
				}
			}

		}

		CommonRunBatchWithSingleNode(breakdown.Cluster.Masters[0].NodeId, breakdown.Cluster.Action, 60, createReclaimResourceCommand(breakdown.Cluster), &nodeStepsReclaimPods, false, true)
		stepReleasePods := &schema.Step{
			Id:                       "setReleasePods-" + breakdown.Operation.Id,
			Name:                     "Release Pods",
			NodeSteps:                nodeStepsReclaimPods,
			WaitBeforeRun:            0,
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}
		// gracefully release all storage
		// but if you pv reclaim policy is not set delete then we only unbound the device
		nodeStepsReclaimStorage := []schema.NodeStep{}
		CommonRunBatchWithSingleNode(breakdown.Cluster.Masters[0].NodeId, breakdown.Cluster.Action, 60, createReclaimStorageCommand(breakdown.Cluster), &nodeStepsReclaimStorage, false, true)
		stepReleaseStorage := &schema.Step{
			Id:                       "setReleaseStorage-" + breakdown.Operation.Id,
			Name:                     "Release Storage",
			NodeSteps:                nodeStepsReclaimStorage,
			WaitBeforeRun:            30,
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepReleasePods
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepReleaseStorage
		stepWait := CreateWaitStep(60, "release-external-storage")
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = &stepWait
	}

	// only needed when we creating a cluster
	if breakdown.Cluster.Action == constants.ActionCreate {
		// always rename the hostname for better control of cluster for future
		if breakdown.Config.EnableHostnameRename {
			stepRenameMaster := &schema.Step{
				Id:        "step0-" + breakdown.Operation.Id,
				Name:      "RenameMastersHostname",
				NodeSteps: createRenameHostnameNodeTask(masterNodeCollection, nodeToHostNameMapping, breakdown.Cluster.ClusterId, constants.MasterHostnameSuffix, breakdown.Config.TaskTimeOut.TaskRenameHostName, hosts),
			}
			breakdown.Operation.Step[0] = stepRenameMaster
		}

		step1 := &schema.Step{
			Id:        "step1-" + breakdown.Operation.Id,
			Name:      "BasicSetup",
			NodeSteps: createBasicConfigNodeTask(reducedNodeList, breakdown.Cluster.Action, breakdown.Config.TaskTimeOut.TaskBasicConfig),
			// now you can define you own step handler
			// do not use the db too much which will cause a lot db workload
			// unless you have to
			//OnStepDoneOrErrorHandler:OnErrorAbortHandler{},
		}

		breakdown.Operation.Step[len(breakdown.Operation.Step)] = step1

		stepLogSetupDep := &schema.Step{
			Id:                       "stepConfigPromtail-" + breakdown.Operation.Id,
			Name:                     "ConfigPromtail",
			NodeSteps:                TaskConfigPromtail(fullNodeList, breakdown.Cluster.Action),
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepLogSetupDep
	}

	breakdown.Operation.Step[len(breakdown.Operation.Step)] = step2
	breakdown.Operation.Step[len(breakdown.Operation.Step)] = step3

	if len(breakdown.Cluster.Masters) > 1 {
		joinControlPlaneList := map[string]byte{}
		for _, master := range breakdown.Cluster.Masters[1:] {
			joinControlPlaneList[master.NodeId] = '0'
		}
		stepJoinControlPlane := &schema.Step{
			Id: "stepJoinControlPlane-" + breakdown.Operation.Id,
		}
		stepJoinControlPlane.Name = "StepJoinControlPlane"
		stepJoinControlPlane.NodeSteps = createKubeadmJoinControlPlaneNodeTask(joinControlPlaneList, breakdown.Cluster.Action, breakdown.Config.TaskTimeOut.TaskKubeadmJoinControlPlane)
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepJoinControlPlane
	}

	// get kube admin.conf
	if len(breakdown.Cluster.Masters) > 1 {
		createAdminConfDir := []schema.NodeStep{}
		for _, master := range breakdown.Cluster.Masters[1:] {
			createAdminConfDirStepId := "NodeStep-" + uuid.New().String()
			CommonAdminConf(master.NodeId, createAdminConfDirStepId, 5, [][]string{
				{"rm", "-f", "/root/.kube/config"},
				{"mkdir", "-p", "/root/.kube/"},
				{"touch", "/root/.kube/config"},
			}, &createAdminConfDir, true, true)
			stepCreateAdminConfDir := &schema.Step{
				Id:                       "createAdminConfDir" + breakdown.Operation.Id,
				Name:                     "createAdminConfDir",
				NodeSteps:                createAdminConfDir,
				OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
				OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			}
			breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepCreateAdminConfDir
		}

		getAdminConfStepId := "stepGetAdminConf-" + uuid.New().String()
		adminConfContent := []schema.NodeStep{}
		CommonAdminConf(breakdown.Cluster.Masters[0].NodeId, getAdminConfStepId, 5, [][]string{
			{"cat", "/etc/kubernetes/admin.conf"},
		}, &adminConfContent, true, true)
		stepGetAdminConf := &schema.Step{
			Id:                       getAdminConfStepId,
			Name:                     "stepGetAdminConf",
			NodeSteps:                adminConfContent,
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepGetAdminConf

		WriteAdminConfStepId := "NodeStep-" + uuid.New().String()
		nodeStepWriteAdminConf := []schema.NodeStep{}
		stepWriteAdminConf := &schema.Step{
			Id:                             "stepWriteAdminConf-" + breakdown.Operation.Id,
			Name:                           "stepWriteAdminConf",
			NodeSteps:                      nil,
			OnStepTimeOutHandler:           OnTimeOutIgnoreHandler{},
			IgnoreDynamicStepCreationError: true,
			DynamicNodeSteps: func(returnData map[string]string, cluster schema.Cluster, operation schema.Operation) ([]schema.NodeStep, error) {
				if _, found := returnData["0"]; !found {
					return nil, fmt.Errorf("(ignore) Failed to find previously step %s return admin.conf dynamic step failed", getAdminConfStepId+"-0")
				}
				returnDataByte := make(map[string][]byte)
				returnDataByte["/root/.kube/config"] = []byte(returnData["0"])
				for _, master := range breakdown.Cluster.Masters[1:] {
					CommonWriteAdminConf(master.NodeId, WriteAdminConfStepId, 10, &nodeStepWriteAdminConf, returnDataByte)
				}

				return nodeStepWriteAdminConf, nil
			},
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepWriteAdminConf
	}

	// remove master from reduce node list
	// worker node that do not share with master will remain in object
	for _, master := range breakdown.Cluster.Masters {
		delete(reducedNodeList, master.NodeId)
	}

	// add all worker to host in order to set /etc/hosts file of workers
	workerIndex := 0
	for id := range reducedNodeList {
		hosts = append(hosts, breakdown.NodeCollectionList[id].Ipv4DefaultIp+"\t"+generatorHostname(workerIndex, breakdown.Operation.ClusterId, constants.WorkerHostnameSuffix))
		workerIndex += 1
	}

	// if reducedNodeList has remaining object
	// which means these node are not shared with master
	if len(reducedNodeList) > 0 {
		// go rename worker`s hostname first
		if breakdown.Cluster.Action == constants.ActionCreate {
			if breakdown.Config.EnableHostnameRename {
				stepRenameWorker := &schema.Step{
					Id:        "step0-" + breakdown.Operation.Id,
					Name:      "RenameWorkersHostname",
					NodeSteps: createRenameHostnameNodeTask(reducedNodeList, nodeToHostNameMapping, breakdown.Cluster.ClusterId, constants.WorkerHostnameSuffix, breakdown.Config.TaskTimeOut.TaskRenameHostName, hosts),
				}

				breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepRenameWorker
			}

		}

		stepSetupLocalVip := &schema.Step{
			Id:        "stepSetupLocalVip-" + breakdown.Operation.Id,
			Name:      "AddOrRemoveLocalVip",
			NodeSteps: createWorkNodeVipNodeTask(reducedNodeList, masterNodeCollection, breakdown.Cluster.Action, breakdown.Config.TaskTimeOut.TaskVip),
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepSetupLocalVip

		if breakdown.Cluster.Action == constants.ActionCreate {
			stepWaitBeforeJoinWorker := CreateWaitStep(5, "wait-control-plane-to-active")
			breakdown.Operation.Step[len(breakdown.Operation.Step)] = &stepWaitBeforeJoinWorker
		}

		stepJoinWorker := &schema.Step{
			Id:        "stepJoinWorkNode",
			Name:      "JoinOrDestroyWorkNode",
			NodeSteps: createKubeadmJoinWorkerNodeTask(reducedNodeList, breakdown.Cluster.Action, breakdown.Config.TaskTimeOut.TaskKubeadmJoinWorker),
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepJoinWorker
	}

	if breakdown.Cluster.Action == constants.ActionCreate {
		// attempt to disable vxlan offload for centos7
		if breakdown.Cluster.CNI.CNIType == constants.CNITypeCalico && strings.Contains(breakdown.Cluster.CNI.Calico.CalicoMode, "Vxlan") || breakdown.Cluster.CNI.CNIType == constants.CNITypeFlannel {
			nodeSteps := []schema.NodeStep{}
			centos7List := map[string]schema.NodeInformation{}
			for nodeId, nodeInfo := range fullNodeList {
				if nodeInfo.SystemInfo.OS.Vendor == "centos" && nodeInfo.SystemInfo.OS.Version == "7" {
					centos7List[nodeId] = nodeInfo
				}
			}
			if len(centos7List) > 0 {
				stepWaitDevice := CreateWaitStep(30, "Wait for device vxlan.calico to appear")
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

	if len(needUntaintNodeCollection) > 0 && breakdown.Cluster.Action == constants.ActionCreate {

		// if need to untaint has object
		// which means it should treated as a work node as well
		// create task to untaint node
		stepUntaintMaster := &schema.Step{
			Id:        "stepUntaintMaster-" + breakdown.Operation.Id,
			Name:      "stepUntaintMaster",
			NodeSteps: []schema.NodeStep{createUntaintNodeTask(breakdown.Cluster.Masters[0].NodeId, needUntaintNodeCollection, nodeToHostNameMapping, breakdown.Cluster.Action, breakdown.Config.TaskTimeOut.TaskRenameHostName, breakdown.Config.EnableHostnameRename)},
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepUntaintMaster

		if breakdown.Config.ReportToKubeCaas {
			// create a sa kubecaas-admin in kube-system
			// then grant cluster-admin role to it
			createSaNodeStepId := "NodeStep-" + uuid.New().String()
			nodeStepsCreateSaToken := []schema.NodeStep{}
			CommonRunBatchWithSingleNodeWithStepID(breakdown.Cluster.Masters[0].NodeId, createSaNodeStepId, 5, [][]string{
				{"kubectl", "create", "sa", "kubecaas-admin", "-n", "kube-system"},
				{"kubectl", "create", "rolebinding", "kubecaas-admin", "--serviceaccount=kube-system:kubecaas-admin", "--clusterrole=cluster-admin", "--group=system:master"},
				{"kubectl", "-n", "kube-system", "get", "serviceaccount/kubecaas-admin", "-o", "jsonpath={.secrets[0].name}"}, // get sa secret token name
			}, &nodeStepsCreateSaToken, true, false)
			stepCreateKubeCaasAdmin := &schema.Step{
				Id:                       "stepCreateKubeCaasAdmin-" + breakdown.Operation.Id,
				Name:                     "stepCreateKubeCaasAdmin",
				NodeSteps:                nodeStepsCreateSaToken,
				OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
				OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			}
			breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepCreateKubeCaasAdmin
			getSaNodeStepId := "NodeStep-" + uuid.New().String()
			stepGetKubeCaasAdminToken := &schema.Step{
				Id:                             "stepGetKubeCaasAdminToken-" + breakdown.Operation.Id,
				Name:                           "stepGetKubeCaasAdminToken",
				NodeSteps:                      nil,
				OnStepTimeOutHandler:           OnTimeOutIgnoreHandler{},
				IgnoreDynamicStepCreationError: true,
				OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
					OnAllDoneCustomerHandler: func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
						if _, found := returnData[getSaNodeStepId+"-0"]; !found {
							log.Errorf("(ignore) Failed to find previously step %s return key customer handler result in error stat", getSaNodeStepId+"-0")
							return
						}
						cluster.ClusterAdminToken = returnData[getSaNodeStepId+"-0"]
					},
				},
				DynamicNodeSteps: func(returnData map[string]string, cluster schema.Cluster, operation schema.Operation) ([]schema.NodeStep, error) {
					if _, found := returnData[createSaNodeStepId+"-2"]; !found {
						return nil, fmt.Errorf("(ignore) Failed to find previously step %s return key dynamic step create failed", createSaNodeStepId+"-2")
					}
					nodeStepsGetSaToken := []schema.NodeStep{}
					CommonRunBatchWithSingleNodeWithStepID(breakdown.Cluster.Masters[0].NodeId, getSaNodeStepId, 5, [][]string{
						{"kubectl", "-n", "kube-system", "get", "secret", returnData[createSaNodeStepId+"-2"], "-o", "jsonpath={.data.token}"},
					}, &nodeStepsGetSaToken, true, false)
					return nodeStepsGetSaToken, nil
				},
			}
			breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepGetKubeCaasAdminToken
		}
	}

	// set default storage class
	storage.SetDefaultStorageClass(&breakdown.Cluster)

	// handle storage addons
	addOnsStorageRegister := &AddOnsRegister{}

	addOnsStorageRegister.RegisterUp(AddOnStorageNFS{
		Name:        "NFS",
		TaskTimeOut: 5,
	}.SetDataWithPlugin(breakdown.Cluster.Storage.NFS, breakdown.SumLicenseLabel, breakdown.Cluster))

	addOnsStorageRegister.RegisterUp(AddOnStorageCeph{
		Name:        "Ceph",
		TaskTimeOut: 5,
	}.SetDataWithPlugin(breakdown.Cluster.Storage.Ceph, breakdown.SumLicenseLabel, breakdown.Cluster))

	if errs := InstallAddOns(addOnsStorageRegister, &breakdown.Operation, breakdown.Cluster, breakdown.Config, breakdown.Cluster.Action); len(errs) > 0 {
		return schema.Operation{}, errors.New("Failed to breakdown storage addons task please see server logs ")
	}

	// setup external lb step
	if len(breakdown.Cluster.ExternalLB.NodeIds) > 0 {
		stepSetupExternalLB := &schema.Step{
			Id:        "stepSetupExternalLB-" + breakdown.Operation.Id,
			Name:      "StepSetupExternalLB",
			NodeSteps: createExternalLBNodeTask(breakdown.Cluster.ExternalLB, breakdown.Cluster.ExternalLB.NodeIds, breakdown.Cluster.Action, breakdown.Config.TaskTimeOut.TaskVip),
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepSetupExternalLB
	}

	if breakdown.Cluster.Action == constants.ActionDelete {
		// remove all kubelet service systemd file
		nodeSteps := []schema.NodeStep{}
		CommonBatchRun(fullNodeList, breakdown.Cluster.Action, 10, [][]string{
			{"rm", "-f", "/usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf"},
			{"rm", "-f", "/usr/lib/systemd/system/kubelet.service"},
			{"rm", "-rf", "/etc/kubernetes/"},
			{"rm", "-rf", "/var/lib/etcd/"},
			// sometime kubeadm reset will not clean all shit, we have to ensure all related data is deleted by running it again
			{"umount", "/var/lib/kubelet/pods"},
			{"rm", "-rf", "/var/lib/kubelet/"},
			{"rm", "-rf", "/etc/cni/net.d/"},
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

	if breakdown.Cluster.Action == constants.ActionCreate {
		getAdminConfStepId := "stepGetAdminConf-" + uuid.New().String()
		adminConfContent := []schema.NodeStep{}
		CommonAdminConf(breakdown.Cluster.Masters[0].NodeId, getAdminConfStepId, 5, [][]string{
			{"cat", "/etc/kubernetes/admin.conf"},
		}, &adminConfContent, true, true)
		stepGetAdminConf := &schema.Step{
			Id:                       getAdminConfStepId,
			Name:                     "stepGetAdminConf",
			NodeSteps:                adminConfContent,
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}
		breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepGetAdminConf

		if breakdown.Cluster.KsClusterConf != nil {
			if breakdown.Cluster.ClusterRole == constants.ClusterRoleMember {
				runtimeCache := cache.GetCurrentCache()
				var firstMasterNode *schema.NodeInformation
				targetHostCluster, err := runtimeCache.GetCluster(breakdown.Cluster.KsClusterConf.MemberOfCluster)
				if err != nil {
					log.Error(fmt.Sprintf("Failed to get host cluster due to error: %s", err))
					return breakdown.Operation, err
				}
				if targetHostCluster == nil {
					newErr := errors.New(fmt.Sprintf("Failed to find host cluster with id: %s", breakdown.Cluster.KsClusterConf.MemberOfCluster))
					log.Error(newErr)
					return breakdown.Operation, newErr
				}

				firstMasterNode, err = runtimeCache.GetNodeInformation(targetHostCluster.Masters[0].NodeId)
				if err != nil {
					log.Error(fmt.Sprintf("Failed to get first master node with node id %s of host cluster %s due to error: %s", targetHostCluster.Masters[0].NodeId, targetHostCluster.ClusterId, err))
					return breakdown.Operation, err
				}

				if firstMasterNode == nil {
					newErr := errors.New(fmt.Sprintf("Failed to get first master node with node id %s of host cluster %s", targetHostCluster.Masters[0].NodeId, targetHostCluster.ClusterId))
					return breakdown.Operation, newErr
				}

				// copy member .kube/config to host cluster
				WriteAdminConfStepId := "NodeStep-" + uuid.New().String()
				nodeStepWriteAdminConf := []schema.NodeStep{}
				stepWriteAdminConfToHost := &schema.Step{
					Id:                             "stepWriteAdminConfToHost-" + breakdown.Operation.Id,
					Name:                           "stepWriteAdminConfToHost",
					NodeSteps:                      nil,
					OnStepTimeOutHandler:           OnTimeOutIgnoreHandler{},
					IgnoreDynamicStepCreationError: true,
					DynamicNodeSteps: func(returnData map[string]string, cluster schema.Cluster, operation schema.Operation) ([]schema.NodeStep, error) {
						if _, found := returnData["0"]; !found {
							return nil, fmt.Errorf("(ignore) Failed to find previously step %s return admin.conf dynamic step failed", getAdminConfStepId+"-0")
						}
						returnDataByte := make(map[string][]byte)
						returnDataByte[fmt.Sprintf("/root/.kube/config%v", cluster.ClusterId)] = []byte(returnData["0"])

						CommonWriteAdminConf(firstMasterNode.Id, WriteAdminConfStepId, 10, &nodeStepWriteAdminConf, returnDataByte)
						return nodeStepWriteAdminConf, nil
					},
				}
				breakdown.Operation.Step[len(breakdown.Operation.Step)] = stepWriteAdminConfToHost

			}

		}
	}

	addOnsRegister := &AddOnsRegister{}

	// fill default value for DnsServerDeploy
	if breakdown.Cluster.DnsServerDeploy != nil {
		breakdown.Cluster.DnsServerDeploy.CompleteDnsServerDeploy(fmt.Sprintf("%s:%d", breakdown.Cluster.ContainerRuntime.PrivateRegistryAddress, breakdown.Cluster.ContainerRuntime.PrivateRegistryPort))
	}

	addOnsRegister.RegisterUp(AddDnsServerDeploy{
		Name:        "AddDnsServerDeploy",
		TaskTimeOut: 60,
	}.SetDataWithPlugin(breakdown.Cluster.DnsServerDeploy, breakdown.SumLicenseLabel, breakdown.Cluster))

	addOnsRegister.RegisterUp(AddClusterDnsUpstream{
		Name:        "AddDnsServerDeploy",
		TaskTimeOut: 60,
	}.SetDataWithPlugin(breakdown.Cluster.ClusterDnsUpstream, breakdown.SumLicenseLabel, breakdown.Cluster))

	addOnsRegister.RegisterUp(AddOnClusterLB{
		Name:        "ClusterLBAddOn",
		TaskTimeOut: 5,
	}.SetDataWithPlugin(breakdown.Cluster.ClusterLB, breakdown.SumLicenseLabel, breakdown.Cluster))

	addOnsRegister.RegisterUp(AddOnHelm{
		Name:        "stepSetupHelm",
		HelmNodes:   masterNodeCollection,
		TaskTimeOut: 10,
	}.SetDataWithPlugin(breakdown.Cluster.Helm, breakdown.SumLicenseLabel, breakdown.Cluster))

	addOnsRegister.RegisterUp(AddOnHarbor{
		Name:         "HarborSetup",
		TaskTimeOut:  10,
		FullNodeList: fullNodeList,
	}.SetDataWithPlugin(breakdown.Cluster.Harbor, 0, breakdown.Cluster))

	isIngressEnable := false

	if breakdown.Cluster.Ingress != nil {
		isIngressEnable = breakdown.Cluster.Ingress.Enable
		hostnameMapping := map[string]string{}
		if breakdown.Config.EnableHostnameRename {
			hostnameMapping = nodeToHostNameMapping
		} else {
			for _, node := range breakdown.Cluster.Ingress.NodeIds {
				if foundNode, found := breakdown.NodeCollectionList[node.NodeId]; found {
					hostnameMapping[node.NodeId] = foundNode.SystemInfo.Node.Hostname
				} else {
					return breakdown.Operation, errors.New(fmt.Sprintf("Failed to found node id %s in node collection", node.NodeId))
				}
			}
		}
		addOnsRegister.RegisterUp(AddOnIngress{
			IngressNodes:          breakdown.Cluster.Ingress.NodeIds,
			NodeToHostNameMapping: hostnameMapping,
			FirstMasterId:         breakdown.Cluster.Masters[0].NodeId,
			Name:                  "IngressAddOn",
			TaskTimeOut:           20,
		}.SetDataWithPlugin(breakdown.Cluster.Ingress, breakdown.SumLicenseLabel, breakdown.Cluster))
		// if breakdown.Cluster.Action == constants.ActionCreate {
		// 	stepWait := CreateWaitStep(30, "waiting-for-ingress-controller")
		// 	breakdown.Operation.Step[len(breakdown.Operation.Step)] = &stepWait
		// }
	}

	if breakdown.Cluster.CloudProvider != nil {
		addOnsRegister.RegisterUp(AddOnOpenStackCloudProvider{
			FullNodeList: fullNodeList,
			Name:         "OpenStackCloudProviderAddOn",
			TaskTimeOut:  5,
		}.SetDataWithPlugin(breakdown.Cluster.CloudProvider.OpenStack, breakdown.SumLicenseLabel, breakdown.Cluster))
	}

	addOnsRegister.RegisterUp(AddOnPostgresOperator{
		Name:        "PostgresOperatorAddon",
		TaskTimeOut: 60,
	}.SetDataWithPlugin(breakdown.Cluster.PostgresOperator, breakdown.SumLicenseLabel, breakdown.Cluster))

	addOnsRegister.RegisterUp(AddOnMiddlePlatform{
		Name:        "MiddlePlatformAddOn",
		TaskTimeOut: 60,
	}.SetDataWithPlugin(breakdown.Cluster.MiddlePlatform, breakdown.SumLicenseLabel, breakdown.Cluster))

	if breakdown.Cluster.Console != nil {
		breakdown.Cluster.Console.ConsoleTitle = breakdown.Config.ConsoleTitle
		breakdown.Cluster.Console.ConsoleResourcePath = breakdown.Config.ConsoleResourcePath
	}
	addOnsRegister.RegisterUp(AddOnConsole{
		Name:        "ConsoleAddOn",
		TaskTimeOut: 10,
	}.SetDataWithPlugin(breakdown.Cluster.Console, breakdown.SumLicenseLabel, breakdown.Cluster))

	isConsoleEnable := false
	var consoleNamespace string
	if breakdown.Cluster.Console != nil {
		isConsoleEnable = breakdown.Cluster.Console.Enable
		consoleNamespace = breakdown.Cluster.Console.Namespace
	}
	addOnsRegister.RegisterUp(AddOnEFK{
		Name:             "EFKAddOn",
		TaskTimeOut:      60,
		FirstMasterId:    breakdown.Cluster.Masters[0].NodeId,
		IsIngressEnable:  isIngressEnable,
		IsConsoleEneble:  isConsoleEnable,
		ConsoleNamespace: consoleNamespace,
	}.SetDataWithPlugin(breakdown.Cluster.EFK, breakdown.SumLicenseLabel, breakdown.Cluster))

	addOnsRegister.RegisterUp(AddOnGAP{
		Name:          "GAPAddOn",
		TaskTimeOut:   20,
		FirstMasterId: breakdown.Cluster.Masters[0].NodeId,
	}.SetDataWithPlugin(breakdown.Cluster.GAP, breakdown.SumLicenseLabel, breakdown.Cluster))

	addOnsRegister.RegisterUp(AddOnKsInstaller{
		Name:          "KsInstallerAddOn",
		TaskTimeOut:   20,
		FirstMasterId: breakdown.Cluster.Masters[0].NodeId,
		// note we use KsClusterConf for both AddOnKsInstaller and AddOnKsClusterConf on purpose
	}.SetDataWithPlugin(breakdown.Cluster.KsClusterConf, breakdown.SumLicenseLabel, breakdown.Cluster))

	addOnsRegister.RegisterUp(AddOnKsClusterConf{
		Name:          "KsClusterConfAddOn",
		TaskTimeOut:   20,
		FirstMasterId: breakdown.Cluster.Masters[0].NodeId,
	}.SetDataWithPlugin(breakdown.Cluster.KsClusterConf, breakdown.SumLicenseLabel, breakdown.Cluster))

	addOnsRegister.RegisterUp(AddOnMinIO{
		Name:          "MinIOAddOn",
		TaskTimeOut:   20,
		FirstMasterId: breakdown.Cluster.Masters[0].NodeId,
	}.SetDataWithPlugin(breakdown.Cluster.MinIO, breakdown.SumLicenseLabel, breakdown.Cluster))
	addOnsRegister.RegisterUp(AddOnVelero{
		Name:          "VeleroAddOn",
		TaskTimeOut:   100,
		FirstMasterId: breakdown.Cluster.Masters[0].NodeId,
	}.SetDataWithPlugin(breakdown.Cluster.Velero, breakdown.SumLicenseLabel, breakdown.Cluster))
	addOnsRegister.RegisterUp(AddOnAutoRestarter{
		Name:        "AutoRestarterAddOn",
		TaskTimeOut: 20,
	}.SetDataWithPlugin(breakdown.Cluster.AutoRestarter, breakdown.SumLicenseLabel, breakdown.Cluster))

	if errs := InstallAddOnsWithCluster(addOnsRegister, &breakdown.Operation, breakdown.Cluster, breakdown.Config); len(errs) > 0 {
		// do nothing meaning we ignore add-on installation error
	}

	if breakdown.Cluster.MiddlePlatform != nil &&
		breakdown.Cluster.MiddlePlatform.Enable &&
		breakdown.Cluster.Action == constants.ActionCreate {
		step := &schema.Step{
			Id:                       "stepPatchNamespace-" + breakdown.Operation.Id,
			Name:                     "StepPatchNamespace",
			NodeSteps:                setupPatchNamespaceCommand(breakdown.Cluster.Masters[0].NodeId, breakdown.Cluster.Action, 5),
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}

		breakdown.Operation.Step[len(breakdown.Operation.Step)] = step
	}
	return breakdown.Operation, nil
}

func createReclaimResourceCommand(cluster schema.Cluster) [][]string {
	rmDeploy := "kubectl delete deploy --all -n %s"
	rmStatefulSet := "kubectl delete sts --all -n %s"
	rmDaemonSet := "kubectl delete ds --all -n %s"
	result := [][]string{}
	for _, ns := range cluster.ReclaimNamespaces {
		if ns == "kube-system" {
			continue
		}
		result = append(result, strings.Split(fmt.Sprintf(rmDeploy, ns), " "))
		result = append(result, strings.Split(fmt.Sprintf(rmStatefulSet, ns), " "))
		result = append(result, strings.Split(fmt.Sprintf(rmDaemonSet, ns), " "))
	}
	return result
}

func createReclaimStorageCommand(cluster schema.Cluster) [][]string {
	rmPVC := "kubectl delete pvc --all -n %s"
	//rmPV := "kubectl delete pv --all -n %s"
	result := [][]string{}
	for _, ns := range cluster.ReclaimNamespaces {
		if ns == "kube-system" {
			continue
		}
		result = append(result, strings.Split(fmt.Sprintf(rmPVC, ns), " "))
		//result = append(result, strings.Split(fmt.Sprintf(rmPV, ns), " "))
	}
	return result
}

func setupPatchNamespaceCommand(nodeId string, action string, timeOut int) []schema.NodeStep {
	var nodeSteps []schema.NodeStep
	nodes := make(map[string]schema.NodeInformation, 1)
	nodes["master"] = schema.NodeInformation{
		Id: nodeId,
	}
	pathCmd := [][]string{
		{"bash", "-c", `kubectl  get ns | sed '1d' | awk '{cmd="kubectl patch ns "$1 " -p '\'\{'\"metadata\":\{\"labels\":\{\"caas.io/department\":\"system\"\}\}\}'\''";system(cmd)}'`},
		{"bash", "-c", `kubectl  get ns | sed '1d' | awk '{cmd="kubectl patch ns "$1 " -p '\'\{'\"metadata\":\{\"annotations\":\{\"caas.io/creator\":\"admin\"\}\}\}'\''";system(cmd)}'`},
	}
	CommonBatchRun(nodes, "", timeOut, pathCmd, &nodeSteps, false)
	return nodeSteps
}
