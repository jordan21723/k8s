package schema

import "github.com/zcalusic/sysinfo"

type NodeInformation struct {
	Id                      string           `json:"node_id"`
	Status                  string           `json:"status"`
	IsDisabled              bool             `json:"is_disabled" description:"is node been disabled"`
	Role                    int              `json:"role" description:"1 = master, 2 = worker , 4 = Ingress , 8 = ExternalLB"`
	Supported               bool             `json:"supported" description:"is node operation system is support"`
	Ipv4DefaultIp           string           `json:"node_ipv4_default_ip" description:"node ipv4 default gateway interface ip"`
	Ipv4DefaultGw           string           `json:"node_ipv4_default_gw" description:"node ipv4 default gateway ip"`
	NodeIPV4AddressList     []string         `json:"node_ipv4_address_list" description:"all ipv4 address of the node"`
	NodeIPV6AddressList     []string         `json:"node_ipv6_address_list" description:"all ipv6 address of the node"`
	SystemInfo              sysinfo.SysInfo  `json:"system_info" description:"system information detail from lib github.com/zcalusic/sysinfo"`
	BelongsToCluster        string           `json:"belongs_to_cluster" description:"node belong to cluster"`
	ContainerRuntime        ContainerRuntime `json:"container_runtime_type" description:"container runtime of node"`
	ProxyIpv4CIDR           string           `json:"proxy_ipv_4_cidr" description:"proxy ip address of node, only use when bastion not able to reach client ip but client can reach bastion ip"`
	KubeNodeStat            string           `json:"kube_node_stat" description:"stat of node in kubernetes cluster,only is valid when the node in k8s cluster"`
	DefaultMetworkInterface string           `json:"default_network_interface" description:"default gw network interface name"`
	AgentStatus             string           `json:"agent_status" description:"binary agent status"`
	IssueList               []string         `json:"issue_list" description:"issue list of the node such as mount point is not ready"`
	LastReportInDate        string           `json:"last_report_in_date"  description:"node last report in date"`
	Region                  *Region          `json:"region,omitempty"  description:"node region status"`
	PortStatus              string           `json:"port_status,omitempty" description:"do not input, port status only show in node detail api"`
	ClusterInstaller        string           `json:"cluster_installer"`
}

func (n *NodeInformation) DeepCopyRegion() *Region {
	if n.Region == nil {
		return nil
	}
	r := *n.Region
	return &r
}

type SSHCredential struct {
	Username  []byte `json:"username" validate:"required" description:"ssh username"`
	Password  []byte `json:"password" validate:"required" description:"ssh password"`
	IpAddress string `json:"ip" validate:"required" description:"node ip address"`
	Port      int    `json:"port" validate:"required" description:"node ssh port"`
}

type SSHRSAkey struct {
	PublicKey string `json:"key" description:"public key"`
}
