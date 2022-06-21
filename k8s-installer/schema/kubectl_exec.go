package schema

type KubectlExec struct {
	ClusterId  string            `json:"cluster_id" validate:"required"  description:"cluster id"`
	Namespace  string            `json:"namespace" validate:"required" description:"namespace"`
	Labels     map[string]string `json:"labels" validate:"required" description:"labels"`
	Containers []string          `json:"containers" validate:"required" description:"containers name"`
	Command    string            `json:"command" validate:"required" description:"command to run"`
}

type KubectlExecResult struct {
	PodName         string            `json:"pod_name" description:"pod name"`
	WithError       bool              `json:"with_error" description:"does it result in error stat"`
	ErrorMessage    string            `json:"error_message" description:"error message"`
	ContainerResult map[string]string `json:"container_result" description:"container command result"`
}