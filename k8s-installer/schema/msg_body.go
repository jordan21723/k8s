package schema

import (
	"k8s-installer/pkg/dep"
)

type QueueBody struct {
	OperationId string  `json:"operation_id"`
	TaskType    string  `json:"task_type"`
	Cluster     Cluster `json:"clusters"`
	NodeStepId  string  `json:"node_step_id"`
	// this is the data in msg_task in byte
	// parse it with json unmarsh()
	TaskData          []byte            `json:"task_data"`
	Task              ITask             `json:"-"`
	ResourceServerURL string            `json:"resource_server_url"`
	StepReturnData    map[string]string `json:"step_return_data"`
}

type QueueReply struct {
	OperationId string            `json:"operation_id"`
	Stat        string            `json:"stat"`
	Message     string            `json:"message"`
	ReturnData  map[string]string `json:"return_data"`
	NodeId      string            `json:"node-id"`
	NodeStepId  string            `json:"node_step_id"`
}

/*
task related
*/
type ITask interface {
	GetAction() string
	GetTaskType() string
	GetTaskTimeOut() int
}

type TaskPrintJoin struct {
	TaskType string `json:"task_type"`
	Action   string `json:"action"`
	TimeOut  int    `json:"time_out"`
}

func (t TaskPrintJoin) GetAction() string {
	return t.Action
}

func (t TaskPrintJoin) GetTaskType() string {
	return t.TaskType
}

func (t TaskPrintJoin) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskVirtualKubelet struct {
	TaskType string `json:"task_type"`
	Action   string `json:"action"`
	TimeOut  int    `json:"time_out"`
	Provider string `json:"provider"`
	Config   []byte `json:"config"`
	/*NodeRoleBindingTemplate   []byte `json:"node_role_binding_template"`
	NodeIDRoleBindingTemplate []byte `json:"node_id_role_binding_template"`*/
	SystemdTemplate []byte `json:"systemd_template"`
	Hostname        string `json:"hostname"`
	Md5Dep          dep.DepMap
}

func (t TaskVirtualKubelet) GetAction() string {
	return t.Action
}

func (t TaskVirtualKubelet) GetTaskType() string {
	return t.TaskType
}

func (t TaskVirtualKubelet) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskCopyTextBaseFile struct {
	TaskType  string            `json:"task_type"`
	Action    string            `json:"action"`
	TimeOut   int               `json:"time_out"`
	TextFiles map[string][]byte `json:"text_files"`
}

func (t TaskCopyTextBaseFile) GetAction() string {
	return t.Action
}

func (t TaskCopyTextBaseFile) GetTaskType() string {
	return t.TaskType
}

func (t TaskCopyTextBaseFile) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskRunCommand struct {
	TaskType      string           `json:"task_type"`
	Action        string           `json:"action"`
	TimeOut       int              `json:"time_out"`
	Commands      map[int][]string `json:"commands"`
	RequireResult bool             `json:"require_result"`
	IgnoreError   bool             `json:"ignore_error"`
	CommandRunId  string           `json:"command_run_id"`
}

func (t TaskRunCommand) GetAction() string {
	return t.Action
}

func (t TaskRunCommand) GetTaskType() string {
	return t.TaskType
}

func (t TaskRunCommand) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskRenameHostName struct {
	TaskType string `json:"task_type"`
	Action   string `json:"action"`
	Data     string `json:"data"`
	TimeOut  int    `json:"time_out"`
	Hostname string `json:"hostname"`
	Hosts    string `json:"hosts"`
}

func (t TaskRenameHostName) GetAction() string {
	return t.Action
}

func (t TaskRenameHostName) GetTaskType() string {
	return t.TaskType
}

func (t TaskRenameHostName) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskSetHosts struct {
	TaskType string              `json:"task_type"`
	Action   string              `json:"action"`
	TimeOut  int                 `json:"time_out"`
	Hosts    map[string][]string `json:"hosts"`
}

func (t TaskSetHosts) GetAction() string {
	return t.Action
}

func (t TaskSetHosts) GetTaskType() string {
	return t.TaskType
}

func (t TaskSetHosts) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskKubectl struct {
	TaskType        string            `json:"task_type"`
	Action          string            `json:"action"`
	TimeOut         int               `json:"time_out"`
	SubCommand      string            `json:"sub_command"`
	YamlTemplate    map[string][]byte `json:"yaml_template"`
	CommandToRun    []string          `json:"command_to_run"`
	CanIgnoreError  bool              `json:"can_ignore_error"`
	RequestResponse bool              `json:"request_response"`
	// when you set WaitBeforeExecutor meaning yaml template will be excutor after certain seconds
	// and result will be ignore same effect like set CanIgnoreError = true
	WaitBeforeExecutor int `json:"wait_before_executor"`
}

func (t TaskKubectl) GetAction() string {
	return t.Action
}

func (t TaskKubectl) GetTaskType() string {
	return t.TaskType
}

func (t TaskKubectl) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskKubeadm struct {
	TaskType      string       `json:"task_type"`
	Action        string       `json:"action"`
	ControlPlane  ControlPlane `json:"control_plane"`
	KubeadmTask   string       `json:"kubeadm-task"`
	TimeOut       int          `json:"time_out"`
	CNITemplate   []byte       `json:"cni_template"`
	KubeadmConfig []byte       `json:"kubeadm_config"`
	Md5Dep        dep.DepMap
}

func (t TaskKubeadm) GetAction() string {
	return t.Action
}

func (t TaskKubeadm) GetTaskType() string {
	return t.TaskType
}

func (t TaskKubeadm) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskLoki struct {
}

type TaskCRI struct {
	TaskType   string           `json:"task_type"`
	CRIType    ContainerRuntime `json:"container_runtime"`
	Action     string           `json:"action"`
	TimeOut    int              `json:"time_out"`
	K8sVersion string           `json:"k8s_version"`
	LogSize    int32            `json:"log_size"`
	LogMaxFile int32            `json:"log_max_file"`
	Md5Dep     dep.DepMap
}

func (t TaskCRI) GetAction() string {
	return t.Action
}

func (t TaskCRI) GetTaskType() string {
	return t.TaskType
}

func (t TaskCRI) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskLoadBalance struct {
	TaskType    string     `json:"task_type"`
	ProxyType   string     `json:"proxy_type"`
	Vips        ExternalLB `json:"external_lb"`
	Action      string     `json:"action"`
	ProxyConfig []byte     `json:"proxy_config"`
	TimeOut     int        `json:"time_out"`
	Reinstall   bool       `json:"reinstall"`
	Md5Dep      dep.DepMap
}

type ProxySection struct {
	Name        string   `json:"proxy_section_name"`
	SectionType string   `json:"section_type"`
	SectionData []string `json:"section_data"`
}

func (t TaskLoadBalance) GetAction() string {
	return t.Action
}

func (t TaskLoadBalance) GetTaskType() string {
	return t.TaskType
}

func (t TaskLoadBalance) GetTaskTimeOut() int {
	return t.TimeOut
}

/* type TaskClusterLoadBalancer struct {
	TaskType       string   `json:"task_type"`
	MasterNodeIPs  []string `json:"master_node_ips"`
	IngressNodeIPs []string `json:"ingress_node_ips"`
	Action         string   `json:"action"`
	TimeOut        int      `json:"timeout"`
}

func (t TaskClusterLoadBalancer) GetAction() string {
	return t.Action
}

func (t TaskClusterLoadBalancer) GetTaskType() string {
	return t.TaskType
}

func (t TaskClusterLoadBalancer) GetTaskTimeOut() int {
	return t.TimeOut
} */

type TaskBasicConfig struct {
	TaskType string `json:"task_type"`
	Role     int    `json:"role"`
	Action   string `json:"action"`
	TimeOut  int    `json:"time_out"`
}

func (t TaskBasicConfig) GetAction() string {
	return t.Action
}

func (t TaskBasicConfig) GetTaskType() string {
	return t.TaskType
}

func (t TaskBasicConfig) GetTaskTimeOut() int {
	return t.TimeOut
}

var Deps []dep.DepMap

func RegisterDep(dep dep.DepMap) dep.DepMap {
	Deps = append(Deps, dep)
	return dep
}

type TaskLink struct {
	TaskType string
	Action   string
	TimeOut  int
	From     dep.DepMap
	SaveTo   string
	LinkTo   string
}

func (t TaskLink) GetAction() string {
	return t.Action
}

func (t TaskLink) GetTaskType() string {
	return t.TaskType
}

func (t TaskLink) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskCurl struct {
	TaskType   string            `json:"task_type"`
	Action     string            `json:"action"`
	TimeOut    int               `json:"time_out"`
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers,omitempty"`
	SkipTLS    bool              `json:"skip_tls"`
	Body       interface{}       `json:"body,omitempty"`
	NameServer string            `json:"name_server"`
}

func (t TaskCurl) GetAction() string {
	return t.Action
}

func (t TaskCurl) GetTaskType() string {
	return t.TaskType
}

func (t TaskCurl) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskCommonDownloadDep struct {
	TaskType string `json:"task_type"`
	Action   string `json:"action"`
	TimeOut  int    `json:"time_out"`

	Dep        dep.DepMap `json:"dep"`
	SaveTo     string     `json:"save_to"`
	K8sVersion string     `json:"k8s_version"`
	Md5        dep.DepMap
}

func (t TaskCommonDownloadDep) GetAction() string {
	return t.Action
}

func (t TaskCommonDownloadDep) GetTaskType() string {
	return t.TaskType
}

func (t TaskCommonDownloadDep) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskCommonDownload struct {
	TaskType string `json:"task_type"`
	Action   string `json:"action"`
	TimeOut  int    `json:"time_out"`

	FromDir          string   `json:"from_dir"`
	K8sVersion       string   `json:"k8s_version"`
	FileList         []string `json:"dep"`
	SaveTo           string   `json:"save_to"`
	IsUseDefaultPath bool     `json:"is_use_default_path"`
}

func (t TaskCommonDownload) GetAction() string {
	return t.Action
}

func (t TaskCommonDownload) GetTaskType() string {
	return t.TaskType
}

func (t TaskCommonDownload) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskGenerateKSClusterConfig struct {
	TaskType  string `json:"task_type"`
	Action    string `json:"action"`
	TimeOut   int    `json:"time_out"`
	IpAddress string `json:"ip_address"`
}

func (t TaskGenerateKSClusterConfig) GetAction() string {
	return t.Action
}

func (t TaskGenerateKSClusterConfig) GetTaskType() string {
	return t.TaskType
}

func (t TaskGenerateKSClusterConfig) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskConfigPromtail struct {
	TaskType string `json:"task_type"`
	Action   string `json:"action"`
	TimeOut  int    `json:"time_out"`
}

func (t TaskConfigPromtail) GetAction() string {
	return t.Action
}

func (t TaskConfigPromtail) GetTaskType() string {
	return t.TaskType
}

func (t TaskConfigPromtail) GetTaskTimeOut() int {
	return t.TimeOut
}

type TaskPreLoadImage struct {
	TaskType     string `json:"task_type"`
	Action       string `json:"action"`
	TimeOut      int    `json:"time_out"`
	ImageDirPath string `json:"image_dir_path"`
}

func (t TaskPreLoadImage) GetAction() string {
	return t.Action
}

func (t TaskPreLoadImage) GetTaskType() string {
	return t.TaskType
}

func (t TaskPreLoadImage) GetTaskTimeOut() int {
	return t.TimeOut
}
