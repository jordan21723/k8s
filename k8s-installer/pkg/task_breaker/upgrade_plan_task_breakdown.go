package task_breaker

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
)

type MakeUpgradePlanTaskBreakDown struct {
	Cluster          schema.Cluster
	Operation        schema.Operation
	DepMapping       dep.DepMap
	TargetK8sVersion string
}

func (MakeUpgradePlanTaskBreakDown *MakeUpgradePlanTaskBreakDown) BreakDownTask() (schema.Operation, error) {
	if MakeUpgradePlanTaskBreakDown.Operation.Step == nil {
		MakeUpgradePlanTaskBreakDown.Operation.Step = map[int]*schema.Step{}
	}
	nodeId := MakeUpgradePlanTaskBreakDown.Cluster.Masters[0].NodeId
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeDownHelmFile-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskCommonDownloadDep{
				TaskType: constants.TaskDownloadDep,
				TimeOut:  10,

				Dep:        MakeUpgradePlanTaskBreakDown.DepMapping,
				SaveTo:     "/tmp",
				K8sVersion: MakeUpgradePlanTaskBreakDown.TargetK8sVersion,
				Md5:        GenerateDepMd5(MakeUpgradePlanTaskBreakDown.DepMapping),
			},
		},
	}

	MakeUpgradePlanTaskBreakDown.Operation.Step[0] = &schema.Step{
		Id:        "downLoadNewVersionKubeadm" + MakeUpgradePlanTaskBreakDown.Operation.Id,
		Name:      "Download new version kubeadm",
		NodeSteps: []schema.NodeStep{step},
	}
	return MakeUpgradePlanTaskBreakDown.Operation, nil
}

type UpgradeApplyTaskBreakDown struct {
	Cluster          schema.Cluster
	Operation        schema.Operation
	TargetK8sVersion string
}

func (UpgradeApplyTaskBreakDown *UpgradeApplyTaskBreakDown) BreakDownTask() (schema.Operation, error) {
	masterMap := map[string]byte{}
	var reducedWorker []schema.ClusterNode
	for _, master := range UpgradeApplyTaskBreakDown.Cluster.Masters {
		masterMap[master.NodeId] = 0
	}

	// remove worker that duplicate with master
	for _, worker := range UpgradeApplyTaskBreakDown.Cluster.Workers {
		if _, exists := masterMap[worker.NodeId]; !exists {
			reducedWorker = append(reducedWorker, worker)
		}
	}

	return UpgradeApplyTaskBreakDown.Operation, nil
}
