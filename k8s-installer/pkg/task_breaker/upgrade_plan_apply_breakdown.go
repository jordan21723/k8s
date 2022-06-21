package task_breaker

import (
	"errors"
	"fmt"

	config "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
)

type UpgradingTaskBreakDown struct {
	Cluster          schema.Cluster
	Operation        schema.Operation
	NodesCollection  schema.NodeInformationCollection
	TargetK8sVersion string
	ApplyVersion     schema.UpgradableVersion
	Dep              dep.DepMap
	Config           config.Config
}

func (UpgradingTaskBreakDown *UpgradingTaskBreakDown) BreakDownTask() (schema.Operation, error) {

	if UpgradingTaskBreakDown.Operation.Step == nil {
		UpgradingTaskBreakDown.Operation.Step = map[int]*schema.Step{}
	}

	masters := map[string]byte{}

	firstMasterNodeId := UpgradingTaskBreakDown.Cluster.Masters[0].NodeId
	//packageSaveDir := UpgradingTaskBreakDown.Config.ApiServer.ResourceServerFilePath + "/upgrade/" + UpgradingTaskBreakDown.ApplyVersion.Name + "/"
	packageSaveDir := "/usr/share/k8s-installer/kubernetes-" + UpgradingTaskBreakDown.ApplyVersion.Name + "/"
	for _, node := range UpgradingTaskBreakDown.Cluster.Masters {
		if _, found := UpgradingTaskBreakDown.NodesCollection[node.NodeId]; !found {
			return UpgradingTaskBreakDown.Operation, errors.New(fmt.Sprintf("Failed to found node information with id: %s", node.NodeId))
		}
		masters[node.NodeId] = 0
		// first we drain first master
		stepCordonMaster := &schema.Step{
			Id:   "cordon-master-" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "cordons master",
			NodeSteps: []schema.NodeStep{
				createCordonNodeStep(firstMasterNodeId, UpgradingTaskBreakDown.NodesCollection[node.NodeId].SystemInfo.Node.Hostname, false, []string{"--ignore-daemonsets", "--delete-local-data", "--force"}),
			},
		}
		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepCordonMaster

		// we install new package deps
		stepDownloadDeps := &schema.Step{
			Id:   "download-new-package-deps" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "download new package deps",
			NodeSteps: []schema.NodeStep{
				{
					Id:     utils.GenNodeStepID(),
					Name:   "node step download new package deps",
					NodeID: node.NodeId,
					Tasks: map[int]schema.ITask{
						0: schema.TaskCommonDownloadDep{
							TaskType:   constants.TaskDownloadDep,
							TimeOut:    300,
							Dep:        UpgradingTaskBreakDown.Dep,
							K8sVersion: UpgradingTaskBreakDown.ApplyVersion.Name,
							SaveTo:     packageSaveDir,
							Md5:        GenerateDepMd5(UpgradingTaskBreakDown.Dep),
						},
					},
				},
			},
		}

		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepDownloadDeps

		// install new deps
		stepInstallNewDeps := &schema.Step{
			Id:   "install-new-package-deps" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "install new package deps",
			NodeSteps: []schema.NodeStep{
				{
					Id:     utils.GenNodeStepID(),
					Name:   "node step install new package deps",
					NodeID: node.NodeId,
					Tasks: map[int]schema.ITask{
						0: schema.TaskRunCommand{
							TaskType: constants.TaskTypeRunCommand,
							TimeOut:  300,
							Commands: map[int][]string{
								0: updatePackageCommand(UpgradingTaskBreakDown.NodesCollection[node.NodeId], packageSaveDir),
							},
						},
					},
				},
			},
		}

		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepInstallNewDeps

		// ask master to pre load images
		stepPreloadImages := &schema.Step{
			Id:                       "pre-load-image" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name:                     "Pre Load Image",
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			NodeSteps: []schema.NodeStep{
				{
					Id:     utils.GenNodeStepID(),
					Name:   "pre load image",
					NodeID: node.NodeId,
					Tasks: map[int]schema.ITask{
						0: schema.TaskPreLoadImage{
							TaskType: constants.TaskTypePreLoadImage,
							TimeOut:  300,
						},
					},
				},
			},
		}

		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepPreloadImages

		// upgrade control plane to new version
		stepKubeadmUpgradeApply := &schema.Step{
			Id:   "kubeadm-apply-upgrade" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "kubeadm apply upgrade",
			NodeSteps: []schema.NodeStep{
				{
					Id:     utils.GenNodeStepID(),
					Name:   "node step upgrade to new k8s version",
					NodeID: node.NodeId,
					Tasks: map[int]schema.ITask{
						0: schema.TaskRunCommand{
							TaskType: constants.TaskTypeRunCommand,
							TimeOut:  300,
							Commands: map[int][]string{
								0: {"kubeadm", "upgrade", "apply", "-f", UpgradingTaskBreakDown.TargetK8sVersion},
							},
						},
					},
				},
			},
		}

		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepKubeadmUpgradeApply

		// uncordon the masters
		stepUnCordonMaster := &schema.Step{
			Id:   "uncordon-master-" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "uncordon master",
			NodeSteps: []schema.NodeStep{
				createCordonNodeStep(firstMasterNodeId, UpgradingTaskBreakDown.NodesCollection[node.NodeId].SystemInfo.Node.Hostname, true, nil),
			},
		}

		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepUnCordonMaster
	}

	for _, node := range UpgradingTaskBreakDown.Cluster.Workers {

		// skip those workers share with masters
		if _, found := masters[node.NodeId]; found {
			continue
		}

		if _, found := UpgradingTaskBreakDown.NodesCollection[node.NodeId]; !found {
			return UpgradingTaskBreakDown.Operation, errors.New(fmt.Sprintf("Failed to found node information with id: %s", node.NodeId))
		}

		// drain workers
		stepCordonWorker := &schema.Step{
			Id:   "cordon-worker-" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "cordons worker",
			NodeSteps: []schema.NodeStep{
				createCordonNodeStep(firstMasterNodeId, UpgradingTaskBreakDown.NodesCollection[node.NodeId].SystemInfo.Node.Hostname, false, []string{"--ignore-daemonsets", "--delete-local-data", "--force"}),
			},
		}
		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepCordonWorker

		// we install new package deps
		stepDownloadDeps := &schema.Step{
			Id:   "download-new-package-deps" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "download new package deps",
			NodeSteps: []schema.NodeStep{
				{
					Id:     utils.GenNodeStepID(),
					Name:   "node step download new package deps",
					NodeID: node.NodeId,
					Tasks: map[int]schema.ITask{
						0: schema.TaskCommonDownloadDep{
							TaskType:   constants.TaskDownloadDep,
							TimeOut:    300,
							Dep:        UpgradingTaskBreakDown.Dep,
							K8sVersion: UpgradingTaskBreakDown.ApplyVersion.Name,
							SaveTo:     packageSaveDir,
							Md5:        GenerateDepMd5(UpgradingTaskBreakDown.Dep),
						},
					},
				},
			},
		}

		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepDownloadDeps

		// install new deps
		stepInstallNewDeps := &schema.Step{
			Id:   "install-new-package-deps" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "install new package deps",
			NodeSteps: []schema.NodeStep{
				{
					Id:     utils.GenNodeStepID(),
					Name:   "node step install new package deps",
					NodeID: node.NodeId,
					Tasks: map[int]schema.ITask{
						0: schema.TaskRunCommand{
							TaskType: constants.TaskTypeRunCommand,
							TimeOut:  300,
							Commands: map[int][]string{
								0: updatePackageCommand(UpgradingTaskBreakDown.NodesCollection[node.NodeId], packageSaveDir),
							},
						},
					},
				},
			},
		}

		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepInstallNewDeps

		// ask client to pre load images
		stepPreloadImages := &schema.Step{
			Id:   "pre-load-image" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "Pre Load Image",
			// if image load failed locally, wo can simply ignore it and  let cri do the download job
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			NodeSteps: []schema.NodeStep{
				{
					Id:     utils.GenNodeStepID(),
					Name:   "pre load image",
					NodeID: node.NodeId,
					Tasks: map[int]schema.ITask{
						0: schema.TaskPreLoadImage{
							TaskType: constants.TaskTypePreLoadImage,
							TimeOut:  300,
						},
					},
				},
			},
		}

		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepPreloadImages

		// upgrade control plane to new version
		stepKubeadmUpgradeApply := &schema.Step{
			Id:   "kubeadm-apply-upgrade" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "kubeadm apply upgrade",
			NodeSteps: []schema.NodeStep{
				{
					Id:     utils.GenNodeStepID(),
					Name:   "node step upgrade to new k8s version",
					NodeID: node.NodeId,
					Tasks: map[int]schema.ITask{
						0: schema.TaskRunCommand{
							TaskType: constants.TaskTypeRunCommand,
							TimeOut:  300,
							Commands: map[int][]string{
								0: {"kubeadm", "upgrade", "node"},
							},
						},
					},
				},
			},
		}

		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepKubeadmUpgradeApply

		// restart workers kubelet
		stepRestartWorkerKubelet := &schema.Step{
			Id:   "restart-worker-kubelet" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "restart worker kubelet",
			NodeSteps: []schema.NodeStep{
				{
					Id:     utils.GenNodeStepID(),
					Name:   "restart worker kubelet",
					NodeID: node.NodeId,
					Tasks: map[int]schema.ITask{
						0: schema.TaskRunCommand{
							TaskType: constants.TaskTypeRunCommand,
							TimeOut:  30,
							Commands: map[int][]string{
								0: {"systemctl", "daemon-reload"},
								1: {"systemctl", "restart", "kubelet"},
							},
						},
					},
				},
			},
		}

		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepRestartWorkerKubelet

		// uncordon the workers
		stepUnCordonMaster := &schema.Step{
			Id:   "uncordon-master-" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "uncordon master",
			NodeSteps: []schema.NodeStep{
				createCordonNodeStep(firstMasterNodeId, UpgradingTaskBreakDown.NodesCollection[node.NodeId].SystemInfo.Node.Hostname, true, nil),
			},
		}

		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepUnCordonMaster
	}

	for _, node := range UpgradingTaskBreakDown.Cluster.Masters {
		// restart workers kubelet
		stepRestartWorkerKubelet := &schema.Step{
			Id:   "restart-worker-kubelet" + node.NodeId + "-" + UpgradingTaskBreakDown.Operation.Id,
			Name: "restart worker kubelet",
			NodeSteps: []schema.NodeStep{
				{
					Id:     utils.GenNodeStepID(),
					Name:   "restart worker kubelet",
					NodeID: node.NodeId,
					Tasks: map[int]schema.ITask{
						0: schema.TaskRunCommand{
							TaskType: constants.TaskTypeRunCommand,
							TimeOut:  30,
							Commands: map[int][]string{
								0: {"systemctl", "daemon-reload"},
								1: {"systemctl", "restart", "kubelet"},
							},
						},
					},
				},
			},
		}

		UpgradingTaskBreakDown.Operation.Step[len(UpgradingTaskBreakDown.Operation.Step)] = stepRestartWorkerKubelet
	}
	return UpgradingTaskBreakDown.Operation, nil
}

func updatePackageCommand(information schema.NodeInformation, packagePath string) []string {
	if information.SystemInfo.OS.Vendor == constants.OSFamilyCentos {
		return []string{"rpm", "-ivh", "--replacefiles", "--replacepkgs", "--nodeps", packagePath + "*.rpm"}
	} else if information.SystemInfo.OS.Vendor == constants.OSFamilyUbuntu {
		return []string{}
	} else {
		return []string{}
	}
}
