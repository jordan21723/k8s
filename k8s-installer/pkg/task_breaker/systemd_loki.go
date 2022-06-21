package task_breaker

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
)

func TaskConfigPromtail(fullNodeList map[string]schema.NodeInformation, action string) []schema.NodeStep {
	var nodeSteps []schema.NodeStep

	for _, node := range fullNodeList {
		nodeSteps = append(nodeSteps, schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "TaskConfigPromtail" + node.Id,
			NodeID: node.Id,
			Tasks: map[int]schema.ITask{
				0: schema.TaskCRI{
					TaskType: constants.TaskConfigPromtail,
					Action:   action,
					TimeOut:  5,
				},
			},
		})
	}
	return nodeSteps
}
