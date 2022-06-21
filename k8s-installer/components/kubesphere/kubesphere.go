package kubesphere

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
	"k8s-installer/schema/plugable"

	"github.com/google/uuid"
)

const (
	DefaultConsolePort = 30880
	DefaultClusterRole = "none"

	// Temporary fixation params

	DefaultEtcdMonitoring  = false
	DefaultEtcdEndpointIps = "localhost"
	DefaultEtcdTlsEnable   = false
	DefaultEtcdPort        = 2379

	DefaultJwtSecret          = "sSWrnRTvXr40TzF9VVuH2a3DleRf4uQBuonrESzQNTIg"
	DefaultMinioVolumeSize    = "20Gi"
	DefaultOpenldapVolumeSize = "5Gi"
	DefaultRedisVolumSize     = "5Gi"

	DefaultElasticsearchMasterReplica    = 1
	DefaultElasticsearchDataReplica      = 2
	DefaultElasticsearchMasterVolumeSize = "10Gi"
	DefaultElasticsearchDataVolumeSize   = "50Gi"
	DefaultLogMaxAge                     = 7
	DefaultElkPrefix                     = "logstash"

	DefaultJenkinsMemoryLim      = "2Gi"
	DefaultJenkinsMemoryReq      = "1500Mi"
	DefaultJenkinsVolumeSize     = "8Gi"
	DefaultJenkinsJavaOptsXms    = "512m"
	DefaultJenkinsJavaOptsXmx    = "512m"
	DefaultJenkinsJavaOptsMaxRAM = "2g"

	DefaultPrometheusReplica       = 2
	DefaultAlertManagerReplica     = 2
	DefaultPrometheusMemoryRequest = "400Mi"
	DefaultPrometheusVolumeSize    = "50Gi"
	DefaultPrometheusEndpoint      = "http://prometheus-operated.kubesphere-monitoring-system.svc:9090"

	DefaultAlertThanosRuleReplica = 2

	DefaultEventRulerReplica = 2

	DefaultLoggingSidecarReplica = 2

	KsClusterTypeProduction  = "production"
	KsClusterTypeTesting     = "testing"
	KsClusterTypeDemo        = "demo"
	KsClusterTypeDevelopment = "development"
)

type KSClusterConfig struct {
	Enabled            bool                          `json:"enabled" description:"enable"`
	Status             string                        `json:"status,omitempty" description:"Status"`
	Dependencies       map[string]plugable.IPlugAble `json:"-"`
	Name               string                        `json:"-"`
	Namespace          string                        `json:"-"`
	IsManagedByKubefed bool                          `json:"is_managed_by_kubefed"`
	MemberOfCluster    string                        `json:"member_of_cluster,omitempty" description:"id of host cluster"`
	ServerConfig       *KSServerConfig               `json:"server,omitempty"`
	ESConfig           *KSElasticSearchConfig        `json:"es,omitempty" description:"Storage backend for logging, events and auditing."`
	MonitorConfig      *KSMonitorConfig              `json:"monitor,omitempty" description:"monitoring params"`
	EtcdMonitorConfig  *KSEtcdMonitorConfig          `json:"etcd_monitor,omitempty" description:"etcd params"`
	ConsoleConfig      *KSConsoleConfig              `json:"console,omitempty" description:"console params"`
	AlertingConfig     *KSAlertingConfig             `json:"alerting,omitempty" description:"(CPU: 0.3 Core, Memory: 300 MiB) Whether to install KubeSphere alerting system. It enables Users to customize alerting policies to send messages to receivers in time with different time intervals and alerting levels to choose from."`
	AuditingConfig     *GeneralEnabled               `json:"auditing,omitempty" description:"Whether to install KubeSphere audit log system. It provides a security-relevant chronological set of records，recording the sequence of activities happened in platform, initiated by different tenants."`
	DevOpsConfig       *KSDevOpsConfig               `json:"devops,omitempty" description:"(CPU: 0.47 Core, Memory: 8.6 G) Whether to install KubeSphere DevOps System. It provides out-of-box CI/CD system based on Jenkins, and automated workflow tools including Source-to-Image & Binary-to-Image."`
	EventsConfig       *KSEventsConfig               `json:"events,omitempty" description:"Whether to install KubeSphere events system. It provides a graphical web console for Kubernetes Events exporting, filtering and alerting in multi-tenant Kubernetes clusters."`
	LoggingConfig      *KSLoggingConfig              `json:"logging,omitempty" description:"(CPU: 57 m, Memory: 2.76 G) Whether to install KubeSphere logging system. Flexible logging functions are provided for log query, collection and management in a unified console. Additional log collectors can be added, such as Elasticsearch, Kafka and Fluentd."`
	MetricsServer      *GeneralEnabled               `json:"metrics_server,omitempty" description:"(CPU: 56 m, Memory: 44.35 MiB) Whether to install metrics-server. IT enables HPA (Horizontal Pod Autoscaler)."`
	MultiClusterConfig *KSMultiClusterConfig         `json:"multicluster" description:"multicluster params"`
	NetworkConfig      *KSNetworkConfig              `json:"network,omitempty"`
	OpenPitrixConfig   *GeneralEnabled               `json:"openpitrix" description:"(2 Core, 3.6 G) Whether to install KubeSphere Application Store. It provides an application store for Helm-based applications, and offer application lifecycle management."`
	ServiceMeshConfig  *GeneralEnabled               `json:"servicemesh,omitempty" description:"(0.3 Core, 300 MiB) Whether to install KubeSphere Service Mesh (Istio-based). It provides fine-grained traffic management, observability and tracing, and offer visualization for traffic topology."`
}

type KSServerConfig struct {
	LocalRegistryServer string `json:"local_registry_server,omitempty" validate:"omitempty" description:"registry server addr. example: 127.0.0.1:5000"`
	StorageClass        string `json:"storage_class" description:"If there is not a default StorageClass in your cluster, you need to specify an existing StorageClass here."`
	JwtSecret           string `json:"jwt_secret" description:"Keep the jwtSecret consistent with the host cluster. Retrive the jwtSecret by executing \"kubectl -n kubesphere-system get cm kubesphere-config -o yaml | grep -v \"apiVersion\" | grep jwtSecret\" on the host cluster."`
	MinioVolumeSize     string `json:"minio_volume_size,omitempty" description:"Minio PVC size."`
	OpenldapVolumeSize  string `json:"openldap_volume_size,omitempty" description:"openldap PVC size."`
	RedisVolumeSize     string `json:"redis_volume_size,omitempty" description:"Redis PVC size."`
}

func (k *KSServerConfig) Completed() {
	k.JwtSecret = util.StringDefaultIfNotSet(k.JwtSecret, DefaultJwtSecret)
	k.MinioVolumeSize = util.StringDefaultIfNotSet(k.MinioVolumeSize, DefaultMinioVolumeSize)
	k.OpenldapVolumeSize = util.StringDefaultIfNotSet(k.OpenldapVolumeSize, DefaultOpenldapVolumeSize)
	k.RedisVolumeSize = util.StringDefaultIfNotSet(k.RedisVolumeSize, DefaultRedisVolumSize)
}

type KSAlertingConfig struct {
	Enabled            bool `json:"enabled" description:"enable"`
	ThanosRulerReplica int  `json:"ruler_replica"`
}

func (k *KSAlertingConfig) Completed() {
	k.ThanosRulerReplica = util.IntDefaultIfZero(k.ThanosRulerReplica, DefaultAlertThanosRuleReplica)
}

type KSNetworkConfig struct {
	NetworkPolicyEnabled bool   `json:"np_enabled" description:"Network policies allow network isolation within the same cluster, which means firewalls can be set up between certain instances (Pods). Make sure that the CNI network plugin used by the cluster supports NetworkPolicy. There are a number of CNI network plugins that support NetworkPolicy, including Calico, Cilium, Kube-router, Romana and Weave Net."`
	IPPool               string `json:"ip_pool,omitempty" description:"if calico cni is integrated then use the value calico, none means that the ippool function is disabled"`
	NetworkTopology      string `json:"topology,omitempty" description:"only support weave-scope"`
}

func (k *KSNetworkConfig) Completed() {
	k.IPPool = util.StringDefaultIfNotSet(k.IPPool, "none")
	k.NetworkTopology = util.StringDefaultIfNotSet(k.NetworkTopology, "none")
}

type GeneralEnabled struct {
	Enabled bool `json:"enabled" description:"enable"`
}

type KSEtcdMonitorConfig struct {
	Enabled     bool     `json:"enabled" description:"Whether to enable etcd monitoring dashboard installation. You have to create a secret for etcd before you enable it."`
	EndpointIps []string `json:"endpoint_ips" description:"etcd cluster EndpointIps, it can be a bunch of IPs here."`
	Port        int      `json:"port" description:"etcd port"`
	TlsEnable   bool     `json:"tls_enable" description:"tls enable"`
}

func (k *KSEtcdMonitorConfig) Completed() {
	if len(k.EndpointIps) == 0 {
		k.EndpointIps = append(k.EndpointIps, DefaultEtcdEndpointIps)
	}
	k.Port = util.IntDefaultIfZero(k.Port, DefaultEtcdPort)
}

func (k *KSEtcdMonitorConfig) ListToString() string {
	str := ""
	for i, v := range k.EndpointIps {
		if i == len(k.EndpointIps)-1 {
			str = str + v
			break
		}
		str = str + "," + v
	}
	return str
}

type KSElasticSearchConfig struct {
	ElasticsearchMasterReplicas   int    `json:"elasticsearch_master_replicas,omitempty" validate:"omitempty,gte=1" description:"total number of master nodes, it's not allowed to use even number"`
	ElasticsearchDataReplicas     int    `json:"elasticsearch_data_replicas,omitempty" validate:"omitempty,gte=1" description:"total number of data nodes."`
	ElasticsearchMasterVolumeSize string `json:"elasticsearch_master_volume_size" description:"Volume size of Elasticsearch master nodes."`
	ElasticsearchDataVolumeSize   string `json:"elasticsearch_data_volume_size" description:"Volume size of Elasticsearch data nodes."`
	LogMaxAge                     int    `json:"log_max_age,omitempty" validate:"omitempty,gte=1" description:"Log retention time in built-in Elasticsearch, it is 7 days by default."`
	ElkPrefix                     string `json:"elk_prefix,omitempty" validate:"omitempty" enum:"logstash" description:"The string making up index names. The index name will be formatted as ks-<elk_prefix>-log."`
}

func (k *KSElasticSearchConfig) Completed() {
	k.ElasticsearchDataVolumeSize = util.StringDefaultIfNotSet(k.ElasticsearchDataVolumeSize, DefaultElasticsearchDataVolumeSize)
	k.ElasticsearchMasterVolumeSize = util.StringDefaultIfNotSet(k.ElasticsearchMasterVolumeSize, DefaultElasticsearchMasterVolumeSize)
	k.ElasticsearchDataReplicas = util.IntDefaultIfZero(k.ElasticsearchDataReplicas, DefaultElasticsearchDataReplica)
	k.ElasticsearchMasterReplicas = util.IntDefaultIfZero(k.ElasticsearchMasterReplicas, DefaultElasticsearchMasterReplica)
	k.LogMaxAge = util.IntDefaultIfZero(k.LogMaxAge, DefaultLogMaxAge)
	k.ElkPrefix = util.StringDefaultIfNotSet(k.ElkPrefix, DefaultElkPrefix)
}

type KSConsoleConfig struct {
	EnableMultiLogin bool `json:"enable_multi_login" description:"enable/disable multiple sing on, it allows an account can be used by different users at the same time."`
	Port             int  `json:"port" validate:"omitempty,gte=30000,lte=31000" description:"port"`
}

func (k *KSConsoleConfig) Completed() {
	k.Port = util.IntDefaultIfZero(k.Port, DefaultConsolePort)
}

type KSDevOpsConfig struct {
	Enabled               bool   `json:"enabled" description:"enable"`
	JenkinsMemoryLimit    string `json:"jenkins_memory_limit,omitempty" description:"Jenkins memory limit."`
	JenkinsMemoryRequest  string `json:"jenkins_memory_request,omitempty" description:"Jenkins memory request."`
	JenkinsVolumeSize     string `json:"jenkins_volume_size,omitempty" description:"Jenkins volume size."`
	JenkinsJavaOptsXms    string `json:"jenkins_java_opts_xms,omitempty" description:"The following three fields are JVM parameters."`
	JenkinsJavaOptsXmx    string `json:"jenkins_java_opts_xmx,omitempty" description:"jenkinsJavaOpts_Xmx"`
	JenkinsJavaOptsMaxRAM string `json:"jenkins_java_opts_max_ram,omitempty" description:"jenkinsJavaOpts_MaxRAM"`
}

func (k *KSDevOpsConfig) Completed() {
	k.JenkinsMemoryLimit = util.StringDefaultIfNotSet(k.JenkinsMemoryLimit, DefaultJenkinsMemoryLim)
	k.JenkinsMemoryRequest = util.StringDefaultIfNotSet(k.JenkinsMemoryRequest, DefaultJenkinsMemoryReq)
	k.JenkinsVolumeSize = util.StringDefaultIfNotSet(k.JenkinsVolumeSize, DefaultJenkinsVolumeSize)
	k.JenkinsJavaOptsXms = util.StringDefaultIfNotSet(k.JenkinsJavaOptsXms, DefaultJenkinsJavaOptsXms)
	k.JenkinsJavaOptsXmx = util.StringDefaultIfNotSet(k.JenkinsJavaOptsXmx, DefaultJenkinsJavaOptsXmx)
	k.JenkinsJavaOptsMaxRAM = util.StringDefaultIfNotSet(k.JenkinsJavaOptsMaxRAM, DefaultJenkinsJavaOptsMaxRAM)
}

type KSEventsConfig struct {
	Enabled       bool `json:"enabled" description:"enable"`
	RulerReplicas int  `json:"ruler_replicas" validate:"omitempty,gte=1" description:"replicas"`
}

func (k *KSEventsConfig) Completed() {
	k.RulerReplicas = util.IntDefaultIfZero(k.RulerReplicas, DefaultEventRulerReplica)
}

type KSLoggingConfig struct {
	Enabled            bool `json:"enabled" description:"enable"`
	LogSidecarReplicas int  `json:"logsidecar_replicas,omitempty" validate:"omitempty,gte=1" description:"logsidecar replicas"`
}

func (k *KSLoggingConfig) Completed() {
	k.LogSidecarReplicas = util.IntDefaultIfZero(k.LogSidecarReplicas, DefaultLoggingSidecarReplica)
}

type KSMonitorConfig struct {
	StorageClass            string `json:"storage_class"`
	PrometheusReplicas      int    `json:"prometheus_replicas,omitempty" validate:"omitempty,gte=1" description:"Prometheus replicas are responsible for monitoring different segments of data source and provide high availability as well."`
	PrometheusMemoryRequest string `json:"prometheus_memory_request,omitempty" description:"Prometheus request memory."`
	PrometheusVolumeSize    string `json:"prometheus_volume_size,omitempty" description:"Prometheus PVC size."`
	AlertManagerReplicas    int    `json:"alertmanager_replicas,omitempty" validate:"omitempty,gte=1" description:"AlertManager Replicas."`
	MonitoringEndpoint      string `json:"monitor_endpoint" description:"prometheus endpoint address"`
}

func (k *KSMonitorConfig) Completed() {
	k.AlertManagerReplicas = util.IntDefaultIfZero(k.AlertManagerReplicas, DefaultAlertManagerReplica)
	k.PrometheusReplicas = util.IntDefaultIfZero(k.PrometheusReplicas, DefaultPrometheusReplica)
	k.PrometheusMemoryRequest = util.StringDefaultIfNotSet(k.PrometheusMemoryRequest, DefaultPrometheusMemoryRequest)
	k.PrometheusVolumeSize = util.StringDefaultIfNotSet(k.PrometheusVolumeSize, DefaultPrometheusVolumeSize)
	k.MonitoringEndpoint = util.StringDefaultIfNotSet(k.MonitoringEndpoint, DefaultPrometheusEndpoint)
}

type KSMultiClusterConfig struct {
	ClusterRole string `json:"cluster_role" validate:"omitempty,oneof=host member none" enum:"host|member|none" description:"host | member | none  # You can install a solo cluster, or specify it as the role of host or member cluster."`
	ClusterType string `json:"cluster_type" validate:"omitempty,oneof=production testing demo development" enum:"production|testing|demo|development" description:"choose from production|testing|demo|development"`
}

func (k *KSMultiClusterConfig) Completed() {
	k.ClusterRole = util.StringDefaultIfNotSet(k.ClusterRole, "host")
	k.ClusterType = util.StringDefaultIfNotSet(k.ClusterType, KsClusterTypeTesting)
}

//func NewDeployKsClusterConf(registry string) KSClusterConfig {
//	return KSClusterConfig{}
//}

func (k *KSClusterConfig) GetNamespace() string {
	return k.Namespace
}

func (k *KSClusterConfig) IsEnable() bool {
	return k.Enabled
}

func (k *KSClusterConfig) GetName() string {
	return k.Name
}

func (k *KSClusterConfig) GetStatus() string {
	return k.Status
}

func (k *KSClusterConfig) SetDependencies(deps map[string]plugable.IPlugAble) {
	k.Dependencies = deps
}

func (k *KSClusterConfig) GetDependencies() map[string]plugable.IPlugAble {
	return k.Dependencies
}

func (k *KSClusterConfig) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(k)
}

func (k *KSClusterConfig) GetLicenseLabel() uint16 {
	return constants.LLKsClusterConf
}

func (k *KSClusterConfig) SetLocalImageRegistry(registry string) {
	k.ServerConfig.LocalRegistryServer = registry
}

func (k *KSClusterConfig) TemplateRender() (string, error) {
	k.CompleteDeploy()
	return template.New("ks-clusterconf").Render(ksClusterConfigurationTemplateV31, k)
}

func (k *KSClusterConfig) CompleteDeploy() *KSClusterConfig {
	if k.ServerConfig == nil {
		k.ServerConfig = &KSServerConfig{}
	}
	k.ServerConfig.Completed()

	if k.ESConfig == nil {
		k.ESConfig = &KSElasticSearchConfig{}
	}
	k.ESConfig.Completed()

	if k.MonitorConfig == nil {
		k.MonitorConfig = &KSMonitorConfig{}
	}
	k.MonitorConfig.Completed()

	if k.EtcdMonitorConfig == nil {
		k.EtcdMonitorConfig = &KSEtcdMonitorConfig{}
	}
	k.EtcdMonitorConfig.Completed()

	if k.ConsoleConfig == nil {
		k.ConsoleConfig = &KSConsoleConfig{}
	}
	k.ConsoleConfig.Completed()

	if k.AlertingConfig == nil {
		k.AlertingConfig = &KSAlertingConfig{}
	}
	k.AlertingConfig.Completed()

	if k.AuditingConfig == nil {
		k.AuditingConfig = &GeneralEnabled{}
	}

	if k.DevOpsConfig == nil {
		k.DevOpsConfig = &KSDevOpsConfig{}
	}
	k.DevOpsConfig.Completed()

	if k.EventsConfig == nil {
		k.EventsConfig = &KSEventsConfig{}
	}
	k.EventsConfig.Completed()

	if k.LoggingConfig == nil {
		k.LoggingConfig = &KSLoggingConfig{}
	}
	k.LoggingConfig.Completed()

	if k.MetricsServer == nil {
		k.MetricsServer = &GeneralEnabled{}
	}

	if k.MultiClusterConfig == nil {
		k.MultiClusterConfig = &KSMultiClusterConfig{}
	}
	k.MultiClusterConfig.Completed()

	if k.NetworkConfig == nil {
		k.NetworkConfig = &KSNetworkConfig{}
	}
	k.NetworkConfig.Completed()

	if k.OpenPitrixConfig == nil {
		k.OpenPitrixConfig = &GeneralEnabled{}
	}

	if k.ServiceMeshConfig == nil {
		k.ServiceMeshConfig = &GeneralEnabled{}
	}
	return k
}

type DeployKsInstaller struct {
	Enable       bool                          `json:"enabled,omitempty" description:"enable"`
	Status       string                        `json:"status,omitempty" description:"Status"`
	Dependencies map[string]plugable.IPlugAble `json:"-"`

	Name          string `json:"name,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	ImageRegistry string `json:"image_registry,omitempty"`
	HostClusterID string `json:"host_cluster_id,omitempty"`
}

func NewDeployKsInstaller() DeployKsInstaller {
	k := DeployKsInstaller{}
	k.CompleteDeploy()
	return k
}

func (k *DeployKsInstaller) TemplateRender() (string, error) {
	k.CompleteDeploy()
	return template.New("ks-installer").Render(ksInstallerTemplateV31, k)
}

func (k DeployKsInstaller) CompleteDeploy() *DeployKsInstaller {
	k.Enable = true
	k.HostClusterID = "host-cluster" + uuid.New().String()
	return &k
}

func (k *DeployKsInstaller) SetImageRegistry(registry string) {
	k.ImageRegistry = registry
}

func (k *DeployKsInstaller) GetNamespace() string {
	return k.Namespace
}

func (k *DeployKsInstaller) IsEnable() bool {
	return true
}

func (k *DeployKsInstaller) GetName() string {
	return k.Name
}

func (k *DeployKsInstaller) GetStatus() string {
	return k.Status
}

func (k *DeployKsInstaller) SetDependencies(deps map[string]plugable.IPlugAble) {
	k.Dependencies = deps
}

func (k *DeployKsInstaller) GetDependencies() map[string]plugable.IPlugAble {
	return k.Dependencies
}

func (k *DeployKsInstaller) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(k)
}

func (k *DeployKsInstaller) GetLicenseLabel() uint16 {
	return constants.LLKsInstaller
}

type GlobalRole struct{}

var globalRole GlobalRole

func (k *GlobalRole) TemplateRenderManage() (string, error) {
	return template.New("role-template-manage-ki").Render(GlobalRoleManage, k)
}

func (k *GlobalRole) TemplateRenderView() (string, error) {
	return template.New("role-template-view-ki").Render(GlobalRoleView, k)
}

type HarborController struct {
	ImageRegistry string `json:"image_registry,omitempty"`
	Replicas      int    `json:"replicas"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Endpoint      string `json:"endpoint"`
}

func (h *HarborController) TemplateRender() (string, error) {
	return template.New("harbor-controller-manager").Render(harborControllerTemplate, h)
}
