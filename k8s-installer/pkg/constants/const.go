package constants

import (
	"time"
)

const ClusterInstallerRancher = "rancher"
const ClusterInstallerKubeadm = "kubeadm"

const ClusterRoleHost = "host"
const ClusterRoleMember = "member"

const MasterHostnameSuffix = "-master-"
const WorkerHostnameSuffix = "-worker-"

const DnsNotificationModeNow = "now"
const DnsNotificationModeNever = "never"

const DnsRecordTypeA = "A"
const DnsRecordTypeAAAA = "AAAA"
const DnsRecordTypeCNAME = "CNAME"

const ActionCreate = "create"
const ActionDelete = "delete"
const ActionBackup = "backup"
const ActionBackupRegular = "backup_regular"
const ActionRestore = "restore"
const ActionDisabled = "disabled"
const ActionEnabled = "enabled"

const OperationTypeClusterSetupOrDestroy = "OperationTypeClusterSetupOrDestroy"
const OperationTypeAddOrRemoveNodesToCluster = "OperationTypeAddOrRemoveNodesToCluster"
const OperationTypeClusterBackupOrRestore = "OperationTypeClusterBackupOrRestore"
const OperationTypeManageAddonsToCluster = "OperationTypeManageAddonsToCluster"
const OperationTypeSingleTask = "OperationTypeSingleTask"
const OperationTypeUpgradeCluster = "OperationTypeUpgradeCluster"

const ProtocolTcp = "tcp"

const (
	TaskPrepareOfflineResource = "TaskPrepareOfflineResource"
	TaskTypeCRI                = "TaskCRI"
	TaskTypeBasic              = "TaskBasicConfig"
	TaskTypeWorkNodeVip        = "TaskLoadBalance"
	TaskTypeKubeadm            = "TaskKubeadm"
	TaskTypeKubectl            = "TaskKubectl"
	TaskTypeRenameHostname     = "TaskRenameHostname"
	TaskTypeRunCommand         = "TaskRunCommand"
	TaskTypeRunAsyncCommand    = "TaskTypeRunAsyncCommand"
	TaskTypeVirtualKubelet     = "TaskVirtualKubelet"
	TaskTypeCopyTextFile       = "TaskCopyTextFile"
	TaskPrintJoin              = "TaskPrintJoin"
	TaskTypeCurl               = "TaskCurl"
	TaskTypeSetHost            = "TaskTypeSetHost"
	TaskTypePreLoadImage       = "TaskPreLoadImage"
	// TaskTypeClusterLoadBalancer = "TaskClusterLoadBalancer"
)

const CRITypeDocker = "docker"
const CRITypeContainerd = "containerd"

const KubeadmTaskInitFirstControlPlane = "InitOrDestroyFirstControlPlane"
const KubeadmTaskJoinControlPlane = "JoinOrDestroyControlPlane"
const KubeadmTaskJoinWorker = "JoinWorker"

const ProxyTypeHaproxy = "haproxy"
const ProxyHaproxySectionListen = "listen"

const CNITypeCalico = "calico"
const CNITypeFlannel = "flannel"
const CalicoV4DetectMethodAuto = "first-found"
const CalicoV6DetectMethodAuto = "first-found"

const StatusInstalling = "installing"
const StatusDeleting = "deleting"
const StatusSuccessful = "successful"
const StatusError = "error"
const StatusAddingNode = "adding-nodes"
const StatusDeletingNode = "deleting-nodes"
const StatusDeployAddons = "deploy-addons"
const StatusProcessing = "processing"
const StatusUpgrading = "upgrading"

const NodeRoleMaster = 1
const NodeRoleWorker = 2
const NodeRoleIngress = 4
const NodeRoleExternalLB = 8

const OSFamilyCentos = "centos"
const OSFamilyUbuntu = "ubuntu"

const CpuArchX86 = "x86_64"
const CpuAarch64 = "aarch64"

const ReturnDataKeyJoinControlPlaneCMD = "joinControlPlaneCMD"
const ReturnDataKeyJoinWorkerCMD = "joinWorkerCMD"
const ReturnDataKeyKubectl = "kubectlResult"
const ReturnDataKeyJoinRancherNodeCMD = "joinRancherNode"

const ClusterStatusRunning = "cluster-running"
const ClusterStatusDestroyed = "cluster-destroyed"
const ClusterStatusBackingUp = "cluster-backing-up"
const ClusterStatusRestoreing = "cluster-restoreing"

const KubectlSubCommandTaint = "taint"
const KubectlSubCommandDelete = "delete"
const KubectlSubCommandCreateOrApply = "apply"
const KubectlSubCommandGet = "get"

const OfflineResourcePath = "/tmp/k8s-installer/resource"
const DepResourceDir = "/usr/share/k8s-installer"

const TaskConfig = "TaskConfig"
const TaskLink = "TaskLink"

const TaskDownloadDep = "TaskDownloadDep"
const TaskDownload = "TaskDownload"
const TaskGenerateKSClusterConfig = "TaskGenerateKSClusterConfig"
const TaskConfigPromtail = "TaskConfigPromtail"

const Available = "available"
const UnAvailable = "unAvailable"

type Size int64

const StorageMinSize Size = 5
const StorageMaxSize Size = 50

const (
	B  Size = 1
	KB      = 1024 * B
	MB      = 1024 * KB
	GB      = 1024 * MB
)

const (
	BackSourceReasonNone          = 0
	BackSourceReasonMd5NotMatch   = 2
	BackSourceReasonDownloadError = 3
	BackSourceReasonNoSpace       = 4
	ForceNotBackSourceAddition    = 1000

	DataExpireTime         = 3 * time.Minute
	ServerAliveTime        = 5 * time.Minute
	DefaultDownloadTimeout = 5 * time.Minute

	DefaultLocalLimit = 20 * MB
	DefaultMinRate    = 64 * KB
)

const (
	StateReady    = "Ready"
	StateNotReady = "NotReady"
)

const (
	SystemInfoProduct = "caas"
	SystemInfoVersion = "4.2.0"
)

const (
	EFKDefaultUser     = "admin"
	EFKDefaultPassword = "admin"
)

const Publickey = "AQAH4SYXUTDCENKLLJFWABCS4SKNYOHJIACL75OZEOETLBEKNAJ4GFPLGATFFNAGNMRE572MQIFOCRS66NM3HQCZBQT74L2UROWST5YQMQKVI3YOBHSRDCEB4SWWG6A3B6ZLZEBF6R3YQVZYCZ2S3EON4WGA===="
