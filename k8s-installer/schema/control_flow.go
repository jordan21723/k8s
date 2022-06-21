package schema

import (
	"k8s-installer/pkg/config/client"
	"k8s-installer/pkg/config/server"
	"k8s-installer/pkg/coredns"
)

// data model for global cache
type OperationCollection map[string]Operation
type ClusterCollection map[string]Cluster
type NodeInformationCollection map[string]NodeInformation
type ServerRuntimeConfigCollection map[string]server.Config
type ClientRuntimeConfigCollection map[string]client.Config
type ClusterNodeRelationShipCollection map[string]map[string]byte
type UserCollection map[string]User
type RoleCollection map[string]Role
type UpgradeVersionCollection map[string]UpgradableVersion
type UpgradePlanCollection map[string]UpgradePlan
type RegionCollection map[string]Region
type TopLevelDomainCollection map[string]coredns.TopLevelDomain
type SubDomainCollection map[string]coredns.DNSDomain

/*
Operation contains multiple steps
*/

type Operation struct {
	Id                string            `json:"operation_id"`
	Name              string            `json:"operation_name"`
	Status            string            `json:"status"`
	CurrentStep       int               `json:"current_step"`
	OperationLog      string            `json:"operation_log"`
	Step              map[int]*Step     `json:"-"`
	Logs              []string          `json:"logs"`
	ClusterId         string            `json:"cluster_id"`
	OperationType     string            `json:"operation_type"`
	PreStepReturnData map[string]string `json:"pre_step_return_data"`
	RequestParameter  RequestParameter  `json:"request_parameter"`
	Created           string            `json:"date_created"`
	LastRun           string            `json:"date_modified"`
	RunByUser         string            `json:"run_by_user"`
	// host stands for which bastion handle this operation
	Host string `json:"host"`
}

type RequestParameter struct {
	BodyParameters    interface{}            `json:"body_parameters"`
	QueryParameters   map[string]interface{} `json:"query_parameters"`
	RuntimeParameters map[string]interface{} `json:"runtime_parameters"`
}

/*
Step contains multiple node Step
*/

type IStepDoneOrErrorHandler interface {
	OnStepDoneOrErrorHandler(reply QueueReply, step *Step, taskDoneSignal chan int, returnData map[string]string, cluster *Cluster, operation *Operation, nodeCollection *NodeInformationCollection, saveClusterToDB func(cluster Cluster) error, saveOperationToDB func(operation Operation) error)
}

type IStepTimeOutHandler interface {
	OnStepTimeOutHandler(step *Step, abortWhenCount int, taskTimeoutSignal chan string, taskDoneSignal chan int, cluster *Cluster, operation *Operation, nodeCollection *NodeInformationCollection, saveClusterToDB func(cluster Cluster) error, saveOperationToDB func(operation Operation) error, nodeStepId string)
}

//type OnStepDoneOrErrorHandler func(reply QueueReply, step *Step, taskDoneSignal chan int, returnData map[string]string, cluster Cluster, operation Operation,saveClusterToDB func(cluster Cluster) error, saveOperationToDB func(operation Operation) error)
//type OnStepTimeOutHandler func(step *Step, abortWhenCount int, taskTimeoutSignal chan string, cluster Cluster, operation Operation ,saveClusterToDB func(cluster Cluster) error, saveOperationToDB func(operation Operation) error)

type Step struct {
	Id                       string                  `json:"step_id"`
	Name                     string                  `json:"step_name"`
	OnSuccessNodes           []ClusterNode           `json:"on_success_nodes"`
	OnTimeoutNodes           []ClusterNode           `json:"on_timeout_nodes"`
	OnFailedNodes            []ClusterNode           `json:"on_failed_nodes"`
	UnReachableNodes         []ClusterNode           `json:"on_failed_nodes"`
	NodeSteps                []NodeStep              `json:"-"`
	WaitBeforeRun            int                     `json:"-"`
	OnStepDoneOrErrorHandler IStepDoneOrErrorHandler `json:"-"`
	OnStepTimeOutHandler     IStepTimeOutHandler     `json:"-"`
	// this is how we create a node steps base on previously step`s return data
	DynamicNodeSteps               func(returnData map[string]string, cluster Cluster, operation Operation) ([]NodeStep, error) `json:"-"`
	IgnoreDynamicStepCreationError bool                                                                                         `json:"-"`
}

type NodeStep struct {
	Id               string        `json:"node_step_id"`
	Name             string        `json:"node_step_name"`
	NodeID           string        `json:"node_step_node_id"`
	ServerMSGTimeOut int           `json:"-"`
	Tasks            map[int]ITask `json:"-"`
}
