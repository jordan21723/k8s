package schema

import (
	"k8s-installer/components/dns"
	"k8s-installer/components/minio"
	"k8s-installer/components/velero"
	"time"

	restarter "k8s-installer/components/auto_restarter"
	cs "k8s-installer/components/console"
	"k8s-installer/components/efk"
	"k8s-installer/components/gap"
	"k8s-installer/components/helm"
	"k8s-installer/components/kubesphere"
	mp "k8s-installer/components/middle_platform"
	pgo "k8s-installer/components/postgres_operator"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/schema/plugable"
)

type Addons struct {
	ClusterLB        *ClusterLB                     `json:"cluster_lb,omitempty"`
	PostgresOperator *pgo.DeployPostgresOperator    `json:"pgo,omitempty"`
	MiddlePlatform   *mp.DeployMiddlePlatform       `json:"middle_platform,omitempty"`
	EFK              *efk.DeployEFK                 `json:"efk,omitempty"`
	CloudProvider    *CloudProvider                 `json:"cloud_providers,omitempty"`
	Console          *cs.DeployConsole              `json:"console,omitempty"`
	Helm             *helm.Helm                     `json:"helm,omitempty"`
	GAP              *gap.DeployGAP                 `json:"gap,omitempty"`
	MinIo            *minio.DeployMinIO             `json:"minio,omitempty"`
	Velero           *velero.DeployVelero           `json:"velero,omitempty"`
	AutoRestarter    *restarter.DeployAutoRestarter `json:"auto_restarter,omitempty"`
}

type Cluster struct {
	DnsServerDeploy      *dns.Server                    `json:"dns_server_deploy,omitempty" description:"dns server deploy"`
	ClusterDnsUpstream   *dns.UpStreamServer            `json:"cluster_dns_upstream,omitempty" description:"cluster dns upstream"`
	ClusterId            string                         `json:"cluster_id,omitempty" validate:"omitempty,hostname_rfc1123" description:"do not input auto generator"`
	ClusterName          string                         `json:"cluster_name" validate:"required,fqdn" description:"do not input auto generator"`
	ControlPlane         ControlPlane                   `json:"control_plane" validate:"required"`
	ExternalLB           ExternalLB                     `json:"external_lb,omitempty" optional:"true" description:"Deprecated, will be remove in future"`
	ClusterLB            *ClusterLB                     `json:"cluster_lb,omitempty" description:"External lb object"`
	ContainerRuntime     ContainerRuntime               `json:"container_runtime" validate:"required" description:"container runtime section"`
	CNI                  CNI                            `json:"cni" validate:"required" description:"container network interface section"`
	Masters              []ClusterNode                  `json:"masters" validate:"required,gt=0,dive" description:"Master nodes list"`
	Workers              []ClusterNode                  `json:"workers,omitempty" description:"Worker nodes list"`
	Ingress              *Ingress                       `json:"ingresses,omitempty" description:"Ingress nodes list"`
	CloudProvider        *CloudProvider                 `json:"cloud_providers,omitempty" description:"Cloud provider config"`
	ClusterOperationIDs  []string                       `json:"cluster_operation_ids,omitempty" description:"auto generated,do not input and all input will be ignore"`
	CurrentOperation     map[string]byte                `json:"current_operation,omitempty" description:"auto generated,do not input and all input will be ignore"`
	Action               string                         `json:"action,omitempty" description:"auto generated, do not input and all input will be ignore"`
	Status               string                         `json:"cluster_status,omitempty" description:"auto generated,do not input and all input will be ignore"`
	Created              string                         `json:"date_created,omitempty" description:"auto generated,do not input and all input will be ignore"`
	Modified             string                         `json:"date_modified,omitempty" description:"auto generated,do not input and all input will be ignore"`
	MiddlePlatform       *mp.DeployMiddlePlatform       `json:"middle_platform,omitempty" optional:"true" description:"middle platform section"`
	PostgresOperator     *pgo.DeployPostgresOperator    `json:"pgo,omitempty" description:"postgres operator section"`
	Console              *cs.DeployConsole              `json:"console,omitempty" description:"console section"`
	EFK                  *efk.DeployEFK                 `json:"efk,omitempty" description:"postgres EFK section"`
	GAP                  *gap.DeployGAP                 `json:"gap,omitempty" description:"postgres GAP section"`
	Helm                 *helm.Helm                     `json:"helm,omitempty" description:"postgres helm section"`
	Description          string                         `json:"description,omitempty" optional:"true" description:"cluster description"`
	Storage              Storage                        `json:"storage,omitempty" validate:"required" description:"storage section required"`
	ReclaimNamespaces    []string                       `json:"reclaim_namespaces,omitempty"`
	AdditionalVersionDep dep.DepMap                     `json:"additional_version_dep,omitempty"`
	ClusterAdminToken    string                         `json:"cluster_admin_token,omitempty" description:"cluster admin sa token,api input will be ignored"`
	Region               string                         `json:"region,omitempty" description:"region id"`
	AddonsProxyIp        string                         `json:"addons_proxy_ip,omitempty" description:"addons float ip"`
	ClusterRole          string                         `json:"cluster_role" description:"cluster role: host、member"`
	KsInstaller          kubesphere.DeployKsInstaller   `json:"-"`
	KsClusterConf        *kubesphere.KSClusterConfig    `json:"ks_cluster_conf,omitempty" description:"Kubesphere cluster conf"`
	MinIO                *minio.DeployMinIO             `json:"-"`
	Velero               *velero.DeployVelero           `json:"-"`
	IsProtected          bool                           `json:"is_protected,omitempty" description:"whether this cluster can or cannot be deleted"`
	Harbor               *Harbor                        `json:"harbor,omitempty" description:"harbor interfacing"`
	BackupRegularName    string                         `json:"backup_regular_name,omitempty" description:"regular backup name"`
	BackupRegularEnable  bool                           `json:"backup_regular_enable,omitempty" description:"regular backup enable"`
	AutoRestarter        *restarter.DeployAutoRestarter `json:"auto_restarter,omitempty" description:"auto restarter controller"`
	ClusterInstaller     string                         `json:"cluster_installer" description:"kubeadm or rancher, if value input not one of kubeadm or rancher then value will set to kubeadm"`
	Mock                 bool                           `json:"mock,omitempty" description:"mock means only during cluster install or destroy setup only change data in db and do not actually install or destroy cluster"`
	Rancher              RancherRequest                 `json:"rancher,omitempty" description:"parameters required to operate rancher"`
}

type ClusterApi struct {
	DnsServerDeploy     *dns.Server                    `json:"dns_server_deploy,omitempty" description:"dns server deploy"`
	ClusterDnsUpstream  *dns.UpStreamServer            `json:"cluster_dns_upstream,omitempty" description:"cluster dns upstream"`
	ClusterId           string                         `json:"cluster_id,omitempty" validate:"omitempty,hostname_rfc1123" description:"do not input auto generator"`
	ClusterName         string                         `json:"cluster_name" validate:"required,fqdn" description:"do not input auto generator"`
	ControlPlane        ControlPlane                   `json:"control_plane" validate:"required"`
	ExternalLB          ExternalLB                     `json:"external_lb,omitempty" optional:"true" description:"Deprecated, will be remove in future"`
	ClusterLB           *ClusterLB                     `json:"cluster_lb,omitempty" description:"External lb object"`
	ContainerRuntime    ContainerRuntime               `json:"container_runtime" validate:"required" description:"container runtime section"`
	CNI                 CNI                            `json:"cni" validate:"required" description:"container network interface section"`
	Masters             []ClusterNode                  `json:"masters" validate:"required,gt=0,dive" description:"Master nodes list"`
	Workers             []ClusterNode                  `json:"workers,omitempty" description:"Worker nodes list"`
	Ingress             *Ingress                       `json:"ingresses,omitempty" description:"Ingress nodes list"`
	CloudProvider       *CloudProvider                 `json:"cloud_providers,omitempty" description:"Cloud provider config"`
	ClusterOperationIDs []string                       `json:"cluster_operation_ids,omitempty" description:"auto generated,do not input and all input will be ignore"`
	CurrentOperation    map[string]byte                `json:"current_operation,omitempty" description:"auto generated,do not input and all input will be ignore"`
	Action              string                         `json:"action,omitempty" description:"auto generated, do not input and all input will be ignore"`
	Status              string                         `json:"cluster_status,omitempty" description:"auto generated,do not input and all input will be ignore"`
	Created             string                         `json:"date_created,omitempty" description:"auto generated,do not input and all input will be ignore"`
	Modified            string                         `json:"date_modified,omitempty" description:"auto generated,do not input and all input will be ignore"`
	MiddlePlatform      *mp.DeployMiddlePlatform       `json:"middle_platform,omitempty" optional:"true" description:"middle platform section"`
	PostgresOperator    *pgo.DeployPostgresOperator    `json:"pgo,omitempty" description:"postgres operator section"`
	Console             *cs.DeployConsole              `json:"console,omitempty" description:"console section"`
	EFK                 *efk.DeployEFK                 `json:"efk,omitempty" description:"postgres EFK section"`
	GAP                 *gap.DeployGAP                 `json:"gap,omitempty" description:"postgres GAP section"`
	Helm                *helm.Helm                     `json:"helm,omitempty" description:"postgres helm section"`
	Description         string                         `json:"description,omitempty" optional:"true" description:"cluster description"`
	Storage             Storage                        `json:"storage,omitempty" validate:"required" description:"storage section required"`
	ReclaimNamespaces   []string                       `json:"reclaim_namespaces,omitempty"`
	ClusterAdminToken   string                         `json:"cluster_admin_token,omitempty" description:"cluster admin sa token,api input will be ignored"`
	Region              string                         `json:"region,omitempty" description:"region id"`
	AddonsProxyIp       string                         `json:"addons_proxy_ip,omitempty" description:"addons float ip"`
	ClusterRole         string                         `json:"cluster_role" description:"cluster role: host、member"`
	KsInstaller         kubesphere.DeployKsInstaller   `json:"-"`
	KsClusterConf       *kubesphere.KSClusterConfig    `json:"ks_cluster_conf,omitempty" description:"Kubesphere cluster conf"`
	MinIO               *minio.DeployMinIO             `json:"-"`
	Velero              *velero.DeployVelero           `json:"-"`
	IsProtected         bool                           `json:"is_protected,omitempty" description:"whether this cluster can or cannot be deleted"`
	Harbor              *Harbor                        `json:"harbor,omitempty" description:"harbor interfacing"`
	BackupRegularName   string                         `json:"backup_regular_name,omitempty" description:"regular backup name"`
	BackupRegularEnable bool                           `json:"backup_regular_enable,omitempty" description:"regular backup enable"`
	AutoRestarter       *restarter.DeployAutoRestarter `json:"auto_restarter,omitempty" description:"auto restarter controller"`
	ClusterInstaller    string                         `json:"cluster_installer" description:"kubeadm or rancher, if value input not one of kubeadm or rancher then value will set to kubeadm"`
	Mock                bool                           `json:"mock,omitempty" description:"mock means only during cluster install or destroy setup only change data in db and do not actually install or destroy cluster"`
	Rancher             RancherRequest                 `json:"rancher,omitempty" description:"parameters required to operate rancher"`
}

type ClusterTemplatePost struct {
	ContainerRuntimes []ContainerRuntime          `json:"container_runtime"`
	CloudProvider     *CloudProvider              `json:"cloud_providers,omitempty"`
	MiddlePlatform    *mp.DeployMiddlePlatform    `json:"middle_platform,omitempty"`
	PostgresOperator  *pgo.DeployPostgresOperator `json:"pgo,omitempty"`
	Console           *cs.DeployConsole           `json:"console,omitempty"`
	EFK               *efk.DeployEFK              `json:"efk,omitempty"`
	KSHost            string                      `json:"ks_host,omitempty"`
	KSClusterConf     *kubesphere.KSClusterConfig `json:"ks_cluster_conf,omitempty" description:"Kubesphere cluster conf"`
	DNSUpStreamServer *dns.UpStreamServer         `json:"dns_up_stream_server" description:"Dns Up stream server"`
	//Rancher           RancherRequest              `json:"rancher,omitempty" description:"parameters required to operate rancher"`
}

type ClusterTemplate struct {
	ContainerRuntimes  []ContainerRuntime          `json:"container_runtime"`
	CloudProvider      *CloudProvider              `json:"cloud_providers,omitempty"`
	MiddlePlatform     *mp.DeployMiddlePlatform    `json:"middle_platform,omitempty"`
	PostgresOperator   *pgo.DeployPostgresOperator `json:"pgo,omitempty"`
	Console            *cs.DeployConsole           `json:"console,omitempty"`
	EFK                *efk.DeployEFK              `json:"efk,omitempty"`
	GAP                *gap.DeployGAP              `json:"gap,omitempty"`
	KSHost             string                      `json:"ks_host,omitempty"`
	KSClusterConf      *kubesphere.KSClusterConfig `json:"ks_cluster_conf,omitempty" description:"Kubesphere cluster conf"`
	ClusterDnsUpstream *dns.UpStreamServer         `json:"cluster_dns_upstream,omitempty" description:"cluster dns upstream"`
	//Rancher            RancherRequest              `json:"rancher,omitempty" description:"parameters required to operate rancher"`
}

type ControlPlane struct {
	ServiceV4CIDR       string          `json:"service_v4_cidr" validate:"required,cidrv4" description:"ipv4 service cidr"`
	ServiceV6CIDR       string          `json:"service_v6_cidr,omitempty" description:"ipv6 service cidr"`
	FeatureGates        map[string]bool `json:"feature_gates,omitempty" description:"auto generated,do not input and all input will be ignore"`
	EnableIPVS          bool            `json:"enable_ipvs,omitempty" description:"enable ipvs or not, default=false"`
	KubernetesVersion   string          `json:"kubernetes_version,omitempty" description:"auto generated,do not input and all input will be ignore"`
	AllowVirtualKubelet bool            `json:"allow_virtual_kubelet,omitempty" description:"Deprecated, will be remove in future"`
}

type ExternalLB struct {
	ClusterVipV4 string `json:"cluster_vip_ipv4" description:"Deprecated, will be remove in future"`
	//ClusterVipV6              string        `json:"cluster_vip_ipv6,omitempty" description:"unexternal vip(ipv6)"`
	//ClusterVipPort            string        `json:"cluster_vip_port" description:"external vip(ipv6) default:9443"`
	//ClusterVipInterface       string        `json:"cluster_vip_interface"`
	//ReinstallIfAlreadyInstall bool          `json:"lb_reinstall_if_already_install,omitempty"`
	NodeIds []ClusterNode `json:"lb-nodes" description:"Deprecated, will be remove in future"`
}

type ClusterNode struct {
	NodeId            string         `json:"node-id" validate:"required"`
	UseVirtualKubelet bool           `json:"use_virtual_kubelet,omitempty" description:"Deprecated, will be remove in future"`
	VirtualKubelet    VirtualKubelet `json:"virtual_kubelet,omitempty" description:"Deprecated, will be remove in future"`
}

type VirtualKubelet struct {
	Provider           string             `json:"provider,omitempty" validate:"omitempty,oneof=mock alibabacloud azure aws" description:"Deprecated, will be remove in future"`
	VKProviderCaas     VKProviderCaas     `json:"vk_provider_Caas,omitempty" description:"Deprecated, will be remove in future"`
	VKProviderMock     VKProviderMock     `json:"vk_provider_mock,omitempty" description:"Deprecated, will be remove in future"`
	VKProviderAliCloud VKProviderAliCloud `json:"vk_provider_ali_cloud,omitempty" description:"Deprecated, will be remove in future"`
	VKProviderAWS      VKProviderAWS      `json:"vk_provider_aws,omitempty" description:"Deprecated, will be remove in future"`
	VKProviderAzure    VKProviderAzure    `json:"vk_provider_azure,omitempty" description:"Deprecated, will be remove in future"`
}

type VKProviderCaas struct {
	CaasApiUrl    string `json:"caas_api_url,omitempty" validate:"omitempty,url" description:"Deprecated, will be remove in future"`
	CaasUsername  string `json:"caas_username,omitempty" validate:"omitempty" description:"Deprecated, will be remove in future"`
	CaasPassword  string `json:"caas_password,omitempty" validate:"omitempty" description:"Deprecated, will be remove in future"`
	CaasClusterId string `json:"caas_cluster_id,omitempty" validate:"omitempty" description:"Deprecated, will be remove in future"`
	CpuLimit      string `json:"cpu_limit,omitempty" validate:"omitempty" description:"Deprecated, will be remove in future"`
	MemoryLimit   string `json:"memory_limit,omitempty" validate:"omitempty" description:"Deprecated, will be remove in future"`
	PodLimit      string `json:"pod_limit,omitempty" validate:"omitempty" description:"Deprecated, will be remove in future"`
}

type VKProviderAliCloud struct {
}

type VKProviderAzure struct {
}

type VKProviderAWS struct {
}

/*
mock is only for demo purpose
do not use it on production env
*/
type VKProviderMock struct {
	CpuLimit    string `json:"cpu_limit,omitempty" validate:"omitempty,required"`
	MemoryLimit string `json:"memory_limit,omitempty" validate:"omitempty,required"`
	PodLimit    string `json:"pod_limit,omitempty" validate:"omitempty,required"`
}

type Ingress struct {
	//IngressType  string                        `json:"ingress-type,omitempty"`
	NodeIds []ClusterNode `json:"ingress-nodes" validate:"required"`
	//HostNetwork  bool                          `json:"ingress-use-host-network,omitempty"`
	Enable       bool                          `json:"enable,omitempty" description:"true = install ingress with post data,false = do not install ingress only save post data in db only,this should be set to true when user set ingress node on dashboard ui"`
	Dependencies map[string]plugable.IPlugAble `json:"-"`
	Status       string                        `json:"status,omitempty"`
}

type CloudProvider struct {
	OpenStack    *CloudProviderOpenStack    `json:"openstack"`
	AlibabaCloud *CloudProviderAlibabaCloud `json:"alibaba_cloud"`
	Azure        *CloudProviderAzure        `json:"azure"`
	AWS          *CloudProviderAWS          `json:"aws"`
	MetaData     map[string]string          `json:"cloud_provider_metadata"`
}

type CloudProviderOpenStack struct {
	Enable       bool                          `json:"enable"`
	Username     string                        `json:"username"`
	Password     string                        `json:"password"`
	AuthURL      string                        `json:"auth_url"`
	ProjectId    string                        `json:"project_id"`
	DomainId     string                        `json:"domain_id"`
	Region       string                        `json:"region"`
	CaCert       string                        `json:"ca_cert,omitempty"`
	Cinder       CinderConfig                  `json:"cinder"`
	Dependencies map[string]plugable.IPlugAble `json:"-"`
	Status       string                        `json:"status,omitempty"`
}

type CloudProviderAzure struct {
	Enable bool `json:"enable,omitempty"`
}

type CloudProviderAlibabaCloud struct {
	Enable bool `json:"enable,omitempty"`
}

type CloudProviderAWS struct {
	Enable bool `json:"enable,omitempty"`
}

type CinderConfig struct {
	BackendType      string `json:"backend_type"`
	AvailabilityZone string `json:"availability_zone"`
	StorageClassName string `json:"storage_class_name" validate:"omitempty,k8s_storage_class"`
	ReclaimPolicy    string `json:"reclaim_policy,omitempty" validate:"omitempty,oneof=Delete Retain"`
	IsDefaultSc      bool   `json:"-"`
}

type CNI struct {
	CniVersion string `json:"cni_version,omitempty" description:"auto generated,do not input and all input will be ignore"`
	CNIType    string `json:"cni_type" enum:"calico" validate:"required,oneof=calico" description:"choose box, only support calico so far"`
	PodV4CIDR  string `json:"pod_v4_cidr" validate:"required" description:"pod cidr v4"`
	PodV6CIDR  string `json:"pod_v6_cidr,omitempty" description:"pod cidr v6"`
	Calico     Calico `json:"calico,omitempty" description:"calico config section"`
	MTU        int    `json:"mtu,omitempty" description:"mtu"`
	//Flannel    Flannel `json:"flannel,omitempty" description:"not used config section"`
	//Multus     Multus  `json:"multus,omitempty"`
}

type Calico struct {
	//Ipv4LookupMethod string `json:"ipv4_lookup_method,omitempty" description:"auto generated,do not input and all input will be ignore"`
	IPAutoDetection  string `json:"ip_autodetection,omitempty" description:"default value is 'first-found'', also can be: 'can-reach=DESTINATION', 'interface=INTERFACE-REGEX'', 'skip-interface=INTERFACE-REGEX', detail see doc https://docs.projectcalico.org/reference/node/configuration#ip-setting"`
	IP6AutoDetection string `json:"ip6_autodetection,omitempty" description:"default value is 'first-found'', also can be: 'can-reach=DESTINATION', 'interface=INTERFACE-REGEX'', 'skip-interface=INTERFACE-REGEX', detail see doc https://docs.projectcalico.org/reference/node/configuration#ip-setting"`
	//Ipv6LookupMethod string `json:"ipv6_lookup_method,omitempty" description:"auto generated,do not input and all input will be ignore"`
	//Mode             string `json:"enable_ipip,omitempty" enum:"Always|CrossSubnet|Never" validate:"omitempty,oneof=Always CrossSubnet Never"`
	CalicoMode      string `json:"calico_mode" enum:"BGP|Overlay-IPIP-All|Overlay-IPIP-Cross-Subnet|Overlay-Vxlan-All|Overlay-Vxlan-Cross-Subnet|overlay" validate:"omitempty,oneof=BGP Overlay-IPIP-All Overlay-IPIP-Cross-Subnet Overlay-Vxlan-All Overlay-Vxlan-Cross-Subnet overlay"`
	EnableDualStack bool   `json:"enable_dual_stack,omitempty" description:"enable dual stack or not, default=false"`
}

type Flannel struct {
	TunnelInterfaceIP string `json:"tunnel_interface_ip,omitempty"`
}

type SriovCni struct {
}

// no support yet
type Multus struct {
	Calico   Calico   `json:"calico,omitempty"`
	SriovCni SriovCni `json:"sriov_cni,omitempty"`
}

type ContainerRuntime struct {
	CRIType                string `json:"container_runtime_type" enum:"docker|containerd" validate:"required,oneof=docker containerd" description:"container runtime type either docker or containerd"`
	PrivateRegistryAddress string `json:"private_registry_address,omitempty" validate:"omitempty,ipv4|fqdn" description:"private registry address e.g. 10.0.0.20"`
	PrivateRegistryPort    int    `json:"private_registry_port,omitempty" validate:"omitempty,max=65535,min=1" description:"private registry port e.g. 4000"`
	//PrivateRegistryCert       string `json:"private_registry_cert,omitempty" validate:"omitempty" description:"not used, do not input will be ignore"`
	//PrivateRegistryKey        string `json:"private_registry_key,omitempty" validate:"omitempty" description:"not used, do not input will be ignore"`
	//PrivateRegistryCA         string `json:"private_registry_ca,omitempty" validate:"omitempty" description:"not used, do not input will be ignore"`
	ReinstallIfAlreadyInstall bool `json:"-"`
}

func (osCloudProvider *CloudProviderOpenStack) IsEnable() bool {
	return osCloudProvider.Enable
}

func (osCloudProvider *CloudProviderOpenStack) GetName() string {
	return "OpenStack Cloud Provider"
}

func (osCloudProvider *CloudProviderOpenStack) GetStatus() string {
	return osCloudProvider.Status
}

func (osCloudProvider *CloudProviderOpenStack) SetDependencies(deps map[string]plugable.IPlugAble) {
	osCloudProvider.Dependencies = deps
}
func (osCloudProvider *CloudProviderOpenStack) GetDependencies() map[string]plugable.IPlugAble {
	return osCloudProvider.Dependencies
}

func (osCloudProvider *CloudProviderOpenStack) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(osCloudProvider)
}

func (osCloudProvider *CloudProviderOpenStack) GetLicenseLabel() uint16 {
	return constants.LLOpenStack
}

func (ingress *Ingress) IsEnable() bool {
	return ingress.Enable
}

func (ingress *Ingress) GetName() string {
	return "Ingress"
}

func (ingress *Ingress) GetStatus() string {
	return ingress.Status
}

func (ingress *Ingress) SetDependencies(deps map[string]plugable.IPlugAble) {
	ingress.Dependencies = deps
}
func (ingress *Ingress) GetDependencies() map[string]plugable.IPlugAble {
	return ingress.Dependencies
}

func (ingress *Ingress) CheckDependencies() (error, []string) {
	return plugable.CommonCheckDependency(ingress)
}

func (ingress *Ingress) GetLicenseLabel() uint16 {
	return constants.LLIngress
}

type SystemInfo struct {
	Company       string    `json:"company,omitempty" description:"do not input auto generator, license company info"`
	Expired       time.Time `json:"end,omitempty" description:"do not input auto generator, license expired date"`
	CPU           string    `json:"cpu,omitempty" description:"do not input auto generator, license cpu limit info"`
	Node          string    `json:"node,omitempty" description:"do not input auto generator, license node limit info"`
	Product       string    `json:"product,omitempty" description:"do not input auto generator, license production info"`
	Version       string    `json:"version,omitempty" description:"do not input auto generator, license version info"`
	MacAddress    string    `json:"mac_address,omitempty" description:"do not input auto generator, license mac address info"`
	LicenseValid  bool      `json:"licenseValid" description:"do not input auto generator, license is valid"`
	SystemVersion string    `json:"systemVersion" description:"do not input auto generator, license system version info"`
	SystemProduct string    `json:"systemProduct" description:"do not input auto generator, license system product info"`
	ErrorDetail   string    `json:"error,omitempty" description:"do not input auto generator, license check failed root cause"`
	License       string    `json:"license" description:"do not input auto generator, license string"`
	Modules       string    `json:"modules" description:"do not input auto generator, such as-> modules:web-console"`
}

type LicenseInfo struct {
	License string `json:"license"`
}

type Backup struct {
	Action     string   `json:"-"`
	Output     string   `json:"-"`
	BackupName string   `json:"backup_name" description:"backup name"`
	Args       []string `json:"arg,omitempty" description:"backup args"`
}

type BackupRegular struct {
	Action            string   `json:"-"`
	Output            string   `json:"-"`
	CronjobTime       string   `json:"cronjob_time,omitempty"`
	BackupRegularName string   `json:"backup_regular_name,omitempty" description:"backup regular name"`
	Registry          string   `json:"-"`
	Args              []string `json:"arg,omitempty" description:"backup regular args"`
}

type Restore struct {
	Action      string   `json:"-"`
	RestoreName string   `json:"restore_name" description:"restore name"`
	BackupName  string   `json:"backup_name" description:"from backup name to restore"`
	Args        []string `json:"arg,omitempty" description:"restore args"`
}

type BackupList struct {
	Kind        string         `json:"kind" description:"BackupList kind"`
	ClusterName string         `json:"cluster_name" description:"cluster name"`
	Region      string         `json:"region" description:"region"`
	Items       []BackupDetail `json:"items" description:"backup detail list"`
}

type BackupDetail struct {
	Kind     string               `json:"kind" description:"Backup kind"`
	Metadata BackupDetailMetadata `json:"metadata" description:"Backup Metadata"`
	Spec     BackupDetailSpec     `json:"spec" description:"Backup Spec"`
	Status   BackupDetailStatus   `json:"status" description:"Backup Status"`
}

type BackupDetailMetadata struct {
	Name string `json:"name" description:"backup name"`
}

type BackupDetailSpec struct {
	TTL string `json:"ttl" description:"backup effective duration"`
}

type BackupDetailStatus struct {
	Phase               string `json:"phase" description:"backup status"`
	Expiration          string `json:"expiration" description:"expiration time"`
	StartTimestamp      string `json:"startTimestamp" description:"start backup time"`
	CompletionTimestamp string `json:"completionTimestamp" description:"backup complete time"`
}

type RestoreList struct {
	Kind  string          `json:"kind" description:"RestoreList kind"`
	Items []RestoreDetail `json:"items" description:"restore detail list"`
}

type RestoreDetail struct {
	Kind     string                `json:"kind" description:"restore kind"`
	Metadata RestoreDetailMetadata `json:"metadata" description:"restore Metadata"`
	Spec     RestoreDetailSpec     `json:"spec" description:"restore Spec"`
	Status   RestoreDetailStatus   `json:"status" description:"restore Status"`
}

type RestoreDetailMetadata struct {
	Name string `json:"name" description:"restore name"`
}

type RestoreDetailSpec struct {
	BackupName string `json:"backupName" description:"restore from backup"`
}

type RestoreDetailStatus struct {
	Phase               string `json:"phase" description:"restore status"`
	StartTimestamp      string `json:"startTimestamp" description:"restore start time"`
	CompletionTimestamp string `json:"completionTimestamp" description:"complete restore time"`
}

type BackupRegularlyDetail struct {
	Kind     string                  `json:"kind"`
	Metadata BackupRegularlyMetadata `json:"metadata"`
	Spec     BackupRegularlySpec     `json:"spec"`
	Status   BackupRegularlyStatus   `json:"status"`
}

type BackupRegularlyMetadata struct {
	Name              string `json:"name"`
	CreationTimestamp string `json:"creationTimestamp"`
}

type BackupRegularlySpec struct {
	Schedule string `json:"schedule"`
}

type BackupRegularlyStatus struct {
	LastScheduleTime string `json:"lastScheduleTime"`
}

type RancherRequest struct {
	RancherAddr          string `json:"rancher_addr" description:"rancher addr, example: https://172.20.163.120"`
	RancherToken         string `json:"rancher_token" description:"rancher api token"`
	ManagedByClusterName string `json:"managed_by_cluster_name,omitempty" description:"rancher managed by cluster name"`
}

// rancher api response json
type RancherResource struct {
	ResourceType string                `json:"resourceType"`
	Data         []RancherResourceData `json:"data"`
}

// rancher api response json
type RancherResourceData struct {
	Id    string                  `json:"id"`
	Name  string                  `json:"name"`
	Links RancherResourceDataLink `json:"links"`
}

// rancher api response json
type RancherResourceDataLink struct {
	ClusterRegistrationTokens string `json:"clusterRegistrationTokens"`
	Nodes                     string `json:"nodes"`
	Remove                    string `json:"remove"`
}

// rancher api response json
type RancherClusterRegistration struct {
	ResourceType string                           `json:"resourceType"`
	Data         []RancherClusterRegistrationData `json:"data"`
}

// rancher api response json
type RancherClusterRegistrationData struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	NodeCommand string `json:"nodeCommand"`
}

// rancher api response json
type RancherNodeInfo struct {
	ResourceType string            `json:"resourceType"`
	Data         []RancherNodeData `json:"data"`
}

// rancher api response json
type RancherNodeData struct {
	Id        string `json:"id" `
	Name      string `json:"name"`
	IPAddress string `json:"ipAddress"`
	State     string `json:"state"`
}
