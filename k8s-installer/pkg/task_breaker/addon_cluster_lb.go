package task_breaker

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"k8s-installer/pkg/cache"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/pkg/template"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"

	"github.com/thoas/go-funk"
)

type (
	Traefik struct {
		Version       string
		MasterNodes   []string
		IngressNodes  []string
		KSConsolePort *int
	}
	Keepalived struct {
		NodeID           string
		RouterID         int
		NodeIP           string
		UnicastPeers     []string
		NetworkInterface string
		VIP              string
	}
	LBNodeSet = map[string]schema.NodeInformation // LBNodeSet contains LB nodes
)

func (t Traefik) RenderMainConfig() (string, error) {
	str, err := template.New("traefik config").Render(TraefikConfigTemplate, t)
	if err != nil {
		return "", err
	}
	return strings.Replace(str, "+", "`", -1), nil
}

func (t Traefik) RenderProviderConfig() (string, error) {
	str, err := template.New("traefik providers").Render(TraefikProviderTemplate,
		t)
	if err != nil {
		return "", err
	}
	return strings.Replace(str, "+", "`", -1), nil
}

func (k Keepalived) RenderConfig() (string, error) {
	str, err := template.New("keepalived config").Render(KeepalivedConfigTemplate,
		k)
	if err != nil {
		return "", err
	}
	return strings.Replace(str, "+", "`", -1), nil
}

const (
	TraefikVersion = "v2.2.11"
	TraefikSystemD = `
[Unit]
Description=The traefik load balancer

[Service]
Type=notify
ExecStart=/usr/local/bin/traefik --configfile=/etc/traefik/traefik.toml
ExecReload=/bin/kill -s HUP $MAINPID
TimeoutStopSec=5
KillMode=mixed

[Install]
WantedBy=multi-user.target
`
	TraefikConfigTemplate = `
[entryPoints]
  [entryPoints.http]
  address = ":80"
  [entryPoints.https]
  address = ":443"
  [entryPoints.k8s-api]
  address = ":9443"
{{if .KSConsolePort}}
  [entryPoints.ks-console]
  address = ":{{.KSConsolePort}}"
{{end}}

[log]
  level = "INFO"
  filePath = "/var/log/traefik"

[providers]
  [providers.file]
    directory = "/etc/traefik/conf"
`
	TraefikProviderTemplate = `
[tcp]
  [tcp.routers]
    [tcp.routers.to-http]
      entryPoints = ["http"]
      rule = "HostSNI(+*+)"
      service = "http"
    [tcp.routers.to-https]
      entryPoints = ["https"]
      rule = "HostSNI(+*+)"
      service = "https"
    [tcp.routers.to-k8s-api]
      entryPoints = ["k8s-api"]
      rule = "HostSNI(+*+)"
      service = "k8s-api"
{{if .KSConsolePort}}
    [tcp.routers.to-ks-console]
      entryPoints = ["ks-console"]
      rule = "HostSNI(+*+)"
      service = "ks-console"
{{end}}
    [tcp.services]
      [tcp.services.http.loadBalancer]
{{range .IngressNodes}}
[[tcp.services.http.loadBalancer.servers]]
address = "{{.}}:80"
{{end}}
      [tcp.services.https.loadBalancer]
{{range .IngressNodes}}
[[tcp.services.https.loadBalancer.servers]]
address = "{{.}}:443"
{{end}}
      [tcp.services.k8s-api.loadBalancer]
{{range .MasterNodes}}
[[tcp.services.k8s-api.loadBalancer.servers]]
address = "{{.}}:6443"
{{end}}
{{if .KSConsolePort}}
      [tcp.services.ks-console.loadBalancer]
{{range .MasterNodes}}
[[tcp.services.ks-console.loadBalancer.servers]]
address = "{{.}}:{{.KSConsolePort}}"
{{end}}
{{end}}
`
	KeepalivedConfigTemplate = `
global_defs {
    router_id lb-master-{{.NodeID}}
    script_user root
}

vrrp_script check-haproxy {
    script "/usr/bin/killall -0 traefik"
    interval 5
    weight -60
}

vrrp_instance VI-lb-master {
    state MASTER
    priority 120
    unicast_src_ip {{.NodeIP}}
    unicast_peer {
{{range .UnicastPeers}}
        {{.}}
{{end}}
    }
    dont_track_primary
    interface {{.NetworkInterface}}
    virtual_router_id {{.RouterID}}
    advert_int 3
    track_script {
        check-haproxy
    }
    virtual_ipaddress {
        {{.VIP}}
    }
}
`
	KeepalivedSubDir = "keepalived"
)

type AddOnClusterLB struct {
	Name        string
	TaskTimeOut int
	enable      bool
	// NodeSet     map[string]schema.NodeInformation // cluster lb node set
}

func (a AddOnClusterLB) GetAddOnName() string {
	return a.Name
}

func (a AddOnClusterLB) IsEnable() bool {
	return a.enable
}

func (a AddOnClusterLB) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		a.enable = false
	} else {
		// do not look into license switch
		a.enable = plugAble.IsEnable()
	}
	return a
}

func (a AddOnClusterLB) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (a AddOnClusterLB) LivenessProbe(operationId string, cluster schema.Cluster,
	config serverConfig.Config) (schema.Step, error) {
	// curl -s -o /dev/null -w '%{http_code}' https://${VIP}:9443/livez -k
	nodeStep := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeCurl-" + cluster.Masters[0].NodeId,
		NodeID: cluster.Masters[0].NodeId, // exec probe handler on the first master node
		Tasks: map[int]schema.ITask{
			0: schema.TaskCurl{
				TaskType: constants.TaskTypeCurl,
				TimeOut:  5,
				URL:      fmt.Sprintf("https://%s:9443/livez", cluster.ClusterLB.VIP),
				Method:   "GET",
				SkipTLS:  true,
			},
		},
	}
	return schema.Step{
		Id:        "stepCheckClusterLB-" + operationId,
		Name:      "StepCheckClusterLB",
		NodeSteps: []schema.NodeStep{nodeStep},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int,
				returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				// curl -s -o /dev/null -w '%{http_code}' https://${VIP}:9443/livez -k
				// return 200
				if reply.Stat == constants.StatusSuccessful &&
					reply.ReturnData["status_code"] == "200" {
					cluster.ClusterLB.Status = constants.StateReady
					return
				}
				cluster.ClusterLB.Status = constants.StateNotReady
			},
			nil,
		},
		OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
			func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation,
				nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
				cluster.ClusterLB.Status = constants.StateNotReady
			},
			nil,
		},
	}, nil
}

func (a AddOnClusterLB) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster,
	config serverConfig.Config, action string) error {
	if action == constants.ActionDelete {
		return a.removeAddOn(operation, cluster, action)
	} else {
		return a.deployAddOn(operation, cluster, action, config)
	}
}

func (a AddOnClusterLB) DeployOrRemoveAddonsWithCluster(operation *schema.Operation,
	cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return a.deployAddOn(operation, cluster, cluster.Action, config)
	}
	// during cluster destroying
	return a.removeAddOn(operation, cluster, cluster.Action)
}

func (a AddOnClusterLB) getLBNodesSet(clusterLB *schema.ClusterLB) (LBNodeSet, error) {
	set := LBNodeSet{}
	rc := cache.GetCurrentCache()
	// Put all load balancer node into map.
	for _, node := range clusterLB.Nodes {
		info, err := rc.GetNodeInformation(node.NodeId)
		if err != nil {
			return set, err
		}
		set[node.NodeId] = *info
	}
	return set, nil
}

func (a AddOnClusterLB) deployAddOn(operation *schema.Operation, cluster schema.Cluster,
	action string, config serverConfig.Config) error {
	lbNodeSet, err := a.getLBNodesSet(cluster.ClusterLB)
	if err != nil {
		return err
	}
	// Assemble step to download and install cluster load balancer.
	nodeSteps := a.installClusterLBDeps(lbNodeSet, action)
	step := &schema.Step{
		Id:        "stepInstallClusterLBDependencies-" + operation.Id,
		Name:      "InstallClusterLBDep",
		NodeSteps: nodeSteps,
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			OnStepReturnCustomerHandler: func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				log.Errorf("Node %s in step %s result in error state but we are going to ignore it.", reply.NodeId, step.Name)
			},
		},
		OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
			OnStepTimeoutCustomerHandler: func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation, nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
				log.Errorf("Node %s in step %s result in timeout state but we are going to ignore it.", nodeStepId, step.Name)
			},
		},
	}
	operation.Step[len(operation.Step)] = step
	// Assemble step to configure and start cluster load balancer service.
	nodeSteps, err = a.setupClusterLB(lbNodeSet, action, &cluster, config)
	if err != nil {
		return err
	}
	step = &schema.Step{
		Id:        "stepSetupClusterLBService-" + operation.Id,
		Name:      "SetupClusterLBService",
		NodeSteps: nodeSteps,
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			OnStepReturnCustomerHandler: func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				log.Errorf("Node %s in step %s result in error state but we are going to ignore it.", reply.NodeId, step.Name)
			},
		},
		OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
			OnStepTimeoutCustomerHandler: func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation, nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
				log.Errorf("Node %s in step %s result in timeout state but we are going to ignore it.", nodeStepId, step.Name)
			},
		},
	}
	operation.Step[len(operation.Step)] = step
	// Assemble step to download keepalived rpm.
	keepalivedDepDir := path.Join(constants.DepResourceDir, KeepalivedSubDir)
	nodeSteps = a.installKeepalivedDeps(lbNodeSet, action, keepalivedDepDir)
	step = &schema.Step{
		Id:        "stepSetupKeepalivedService-" + operation.Id,
		Name:      "InstallKeeplivedDep",
		NodeSteps: nodeSteps,
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			OnStepReturnCustomerHandler: func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				log.Errorf("Node %s in step %s result in error state but we are going to ignore it.", reply.NodeId, step.Name)
			},
		},
		OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
			OnStepTimeoutCustomerHandler: func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation, nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
				log.Errorf("Node %s in step %s result in timeout state but we are going to ignore it.", nodeStepId, step.Name)
			},
		},
	}
	operation.Step[len(operation.Step)] = step
	// Assemble step to configure and start keepalived service.
	nodeSteps, err = a.setupKeepalived(lbNodeSet, action, keepalivedDepDir, &cluster,
		config)
	if err != nil {
		return err
	}
	step = &schema.Step{
		Id:        "stepSetupKeepalivedService-" + operation.Id,
		Name:      "SetupKeepalivedService",
		NodeSteps: nodeSteps,
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			OnStepReturnCustomerHandler: func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				log.Errorf("Node %s in step %s result in error state but we are going to ignore it.", reply.NodeId, step.Name)
			},
		},
		OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
			OnStepTimeoutCustomerHandler: func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation, nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
				log.Errorf("Node %s in step %s result in timeout state but we are going to ignore it.", nodeStepId, step.Name)
			},
		},
	}
	operation.Step[len(operation.Step)] = step

	return nil
}

func (a AddOnClusterLB) removeAddOn(operation *schema.Operation, cluster schema.Cluster,
	action string) error {
	lbNodeSet, err := a.getLBNodesSet(cluster.ClusterLB)
	if err != nil {
		return err
	}
	// Assemble step to stop and remove cluster load balancer.
	nodeSteps, err := a.removeClusterLB(lbNodeSet, action)
	if err != nil {
		return nil
	}
	step := &schema.Step{
		Id:        "stepRemoveClusterLBService-" + operation.Id,
		Name:      "RemoveClusterLBService",
		NodeSteps: nodeSteps,
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			OnStepReturnCustomerHandler: func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				log.Errorf("Node %s in step %s result in error state but we are going to ignore it.", reply.NodeId, step.Name)
			},
		},
		OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
			OnStepTimeoutCustomerHandler: func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation, nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
				log.Errorf("Node %s in step %s result in timeout state but we are going to ignore it.", nodeStepId, step.Name)
			},
		},
	}
	operation.Step[len(operation.Step)] = step
	// Assemble step to stop and remove keepalived.
	nodeSteps, err = a.removeKeepalived(lbNodeSet, action)
	if err != nil {
		return err
	}
	step = &schema.Step{
		Id:        "stepRemoveKeepalivedService-" + operation.Id,
		Name:      "RemoveKeepalivedService",
		NodeSteps: nodeSteps,
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{
			OnStepReturnCustomerHandler: func(reply schema.QueueReply, step *schema.Step, taskDoneSignal chan int, returnData map[string]string, cluster *schema.Cluster, operation *schema.Operation, nodeCollection *schema.NodeInformationCollection) {
				log.Errorf("Node %s in step %s result in error state but we are going to ignore it.", reply.NodeId, step.Name)
			},
		},
		OnStepTimeOutHandler: OnTimeOutIgnoreHandler{
			OnStepTimeoutCustomerHandler: func(step *schema.Step, cluster *schema.Cluster, operation *schema.Operation, nodeStepId string, nodeCollection *schema.NodeInformationCollection) {
				log.Errorf("Node %s in step %s result in timeout state but we are going to ignore it.", nodeStepId, step.Name)
			},
		},
	}
	operation.Step[len(operation.Step)] = step

	return nil
}

type (
	TraefikDeps struct {
	}

	KeepaliveDeps struct {
	}
)

func (C TraefikDeps) GetDeps() dep.DepMap {
	return traefikDep
}

func (C KeepaliveDeps) GetDeps() dep.DepMap {
	return keepalivedDep
}

var (
	traefikDep = schema.RegisterDep(dep.DepMap{
		"centos": {
			constants.V1_18_6: {
				"x86_64": {
					"traefik": "traefik",
				},
				"aarch64": {
					"traefik": "traefik",
				},
			},
		},
		"ubuntu": {},
	})

	keepalivedDep = schema.RegisterDep(dep.DepMap{
		"centos": {
			constants.V1_18_6: {
				"x86_64": {
					"perl.rpm":                   "perl-5.16.3-297.el7.x86_64.rpm",
					"perl-Carp.rpm":              "perl-Carp-1.26-244.el7.noarch.rpm",
					"perl-constant.rpm":          "perl-constant-1.27-2.el7.noarch.rpm",
					"perl-Encode.rpm":            "perl-Encode-2.51-7.el7.x86_64.rpm",
					"perl-Exporter.rpm":          "perl-Exporter-5.68-3.el7.noarch.rpm",
					"perl-File-Path.rpm":         "perl-File-Path-2.09-2.el7.noarch.rpm",
					"perl-File-Temp.rpm":         "perl-File-Temp-0.23.01-3.el7.noarch.rpm",
					"perl-Filter.rpm":            "perl-Filter-1.49-3.el7.x86_64.rpm",
					"perl-Getopt-Long.rpm":       "perl-Getopt-Long-2.40-3.el7.noarch.rpm",
					"perl-HTTP-Tiny.rpm":         "perl-HTTP-Tiny-0.033-3.el7.noarch.rpm",
					"perl-macros.rpm":            "perl-macros-5.16.3-297.el7.x86_64.rpm",
					"perl-parent.rpm":            "perl-parent-0.225-244.el7.noarch.rpm",
					"perl-PathTools.rpm":         "perl-PathTools-3.40-5.el7.x86_64.rpm",
					"perl-Pod-Escapes.rpm":       "perl-Pod-Escapes-1.04-297.el7.noarch.rpm",
					"perl-podlators.rpm":         "perl-podlators-2.5.1-3.el7.noarch.rpm",
					"perl-Pod-Perldoc.rpm":       "perl-Pod-Perldoc-3.20-4.el7.noarch.rpm",
					"perl-Pod-Simple.rpm":        "perl-Pod-Simple-3.28-4.el7.noarch.rpm",
					"perl-Pod-Usage.rpm":         "perl-Pod-Usage-1.63-3.el7.noarch.rpm",
					"perl-Scalar-List-Utils.rpm": "perl-Scalar-List-Utils-1.27-248.el7.x86_64.rpm",
					"perl-Socket.rpm":            "perl-Socket-2.010-5.el7.x86_64.rpm",
					"perl-Storable.rpm":          "perl-Storable-2.45-3.el7.x86_64.rpm",
					"perl-Text-ParseWords.rpm":   "perl-Text-ParseWords-3.29-4.el7.noarch.rpm",
					"perl-threads.rpm":           "perl-threads-1.87-4.el7.x86_64.rpm",
					"perl-threads-shared.rpm":    "perl-threads-shared-1.43-6.el7.x86_64.rpm",
					"perl-Time-HiRes.rpm":        "perl-Time-HiRes-1.9725-3.el7.x86_64.rpm",
					"perl-Time-Local.rpm":        "perl-Time-Local-1.2300-2.el7.noarch.rpm",
					"perl-libs.rpm":              "perl-libs-5.16.3-297.el7.x86_64.rpm",
					"psmisc.rpm":                 "psmisc-22.20-16.el7.x86_64.rpm",
					"keepalived":                 "keepalived-1.3.5-16.el7.x86_64.rpm",
					"net-snmp-libs.rpm":          "net-snmp-libs-5.7.2-47.el7.x86_64.rpm",
					"net-snmp-agent-libs.rpm":    "net-snmp-agent-libs-5.7.2-47.el7.x86_64.rpm",
					"lm_sensors-libs.rpm":        "lm_sensors-libs-3.4.0-8.20160601gitf9185e5.el7.x86_64.rpm",
				},
				"aarch64": {
					"perl.rpm":                   "perl-5.16.3-297.el7.aarch64.rpm",
					"perl-Carp.rpm":              "perl-Carp-1.26-244.el7.noarch.rpm",
					"perl-constant.rpm":          "perl-constant-1.27-2.el7.noarch.rpm",
					"perl-Encode.rpm":            "perl-Encode-2.51-7.el7.aarch64.rpm",
					"perl-Exporter.rpm":          "perl-Exporter-5.68-3.el7.noarch.rpm",
					"perl-File-Path.rpm":         "perl-File-Path-2.09-2.el7.noarch.rpm",
					"perl-File-Temp.rpm":         "perl-File-Temp-0.23.01-3.el7.noarch.rpm",
					"perl-Filter.rpm":            "perl-Filter-1.49-3.el7.aarch64.rpm",
					"perl-Getopt-Long.rpm":       "perl-Getopt-Long-2.40-3.el7.noarch.rpm",
					"perl-HTTP-Tiny.rpm":         "perl-HTTP-Tiny-0.033-3.el7.noarch.rpm",
					"perl-macros.rpm":            "perl-macros-5.16.3-297.el7.aarch64.rpm",
					"perl-parent.rpm":            "perl-parent-0.225-244.el7.noarch.rpm",
					"perl-PathTools.rpm":         "perl-PathTools-3.40-5.el7.aarch64.rpm",
					"perl-Pod-Escapes.rpm":       "perl-Pod-Escapes-1.04-297.el7.noarch.rpm",
					"perl-podlators.rpm":         "perl-podlators-2.5.1-3.el7.noarch.rpm",
					"perl-Pod-Perldoc.rpm":       "perl-Pod-Perldoc-3.20-4.el7.noarch.rpm",
					"perl-Pod-Simple.rpm":        "perl-Pod-Simple-3.28-4.el7.noarch.rpm",
					"perl-Pod-Usage.rpm":         "perl-Pod-Usage-1.63-3.el7.noarch.rpm",
					"perl-Scalar-List-Utils.rpm": "perl-Scalar-List-Utils-1.27-248.el7.aarch64.rpm",
					"perl-Socket.rpm":            "perl-Socket-2.010-5.el7.aarch64.rpm",
					"perl-Storable.rpm":          "perl-Storable-2.45-3.el7.aarch64.rpm",
					"perl-Text-ParseWords.rpm":   "perl-Text-ParseWords-3.29-4.el7.noarch.rpm",
					"perl-threads.rpm":           "perl-threads-1.87-4.el7.aarch64.rpm",
					"perl-threads-shared.rpm":    "perl-threads-shared-1.43-6.el7.aarch64.rpm",
					"perl-Time-HiRes.rpm":        "perl-Time-HiRes-1.9725-3.el7.aarch64.rpm",
					"perl-Time-Local.rpm":        "perl-Time-Local-1.2300-2.el7.noarch.rpm",
					"perl-libs.rpm":              "perl-libs-5.16.3-297.el7.aarch64.rpm",
					"psmisc.rpm":                 "psmisc-22.20-16.el7.aarch64.rpm",
					"keepalived":                 "keepalived-1.3.5-16.el7.aarch64.rpm",
					"net-snmp-libs.rpm":          "net-snmp-libs-5.7.2-47.el7.aarch64.rpm",
					"net-snmp-agent-libs.rpm":    "net-snmp-agent-libs-5.7.2-47.el7.aarch64.rpm",
					"lm_sensors-libs.rpm":        "lm_sensors-libs-3.4.0-8.20160601gitf9185e5.el7.aarch64.rpm",
				},
			},
		},
	})
)

func (a AddOnClusterLB) installClusterLBDeps(nodeSet LBNodeSet,
	action string) []schema.NodeStep {
	nodeSteps := []schema.NodeStep{}
	if action == constants.ActionDelete {
		return nodeSteps
	}
	// Download traefik binary to /usr/local/bin
	CommonDownloadDep(nodeSet, schema.TaskCommonDownloadDep{
		TaskType:   constants.TaskDownloadDep,
		Action:     action,
		TimeOut:    300,
		Dep:        traefikDep,
		SaveTo:     "/usr/local/bin",
		K8sVersion: constants.V1_18_6,
		Md5:        GenerateDepMd5(traefikDep),
	}, &nodeSteps)
	return nodeSteps
}

func (a AddOnClusterLB) installKeepalivedDeps(nodeSet LBNodeSet, action, depDir string) []schema.NodeStep {
	nodeSteps := []schema.NodeStep{}
	if action == constants.ActionDelete {
		return nodeSteps
	}
	// Download keepalived rpm to /tmp/keepalived
	CommonDownloadDep(nodeSet, schema.TaskCommonDownloadDep{
		TaskType:   constants.TaskDownloadDep,
		Action:     action,
		TimeOut:    300,
		Dep:        keepalivedDep,
		SaveTo:     depDir,
		K8sVersion: constants.V1_18_6,
		Md5:        GenerateDepMd5(traefikDep),
	}, &nodeSteps)
	return nodeSteps
}

func (a AddOnClusterLB) setupClusterLB(nodeSet LBNodeSet,
	action string, cluster *schema.Cluster, config serverConfig.Config) ([]schema.NodeStep, error) {
	nodeSteps := []schema.NodeStep{}
	// Save systemd unit file of traefik to /usr/lib/systemd/system
	CommonConfig(nodeSet,
		"traefik.service",
		TraefikSystemD,
		schema.TaskCommonDownload{
			TaskType:         constants.TaskDownload,
			Action:           action,
			TimeOut:          300,
			FromDir:          "",
			K8sVersion:       constants.V1_18_6,
			FileList:         []string{},
			SaveTo:           "/usr/lib/systemd/system",
			IsUseDefaultPath: false,
		}, &nodeSteps, config, nil)
	masters, ingres, err := a.getProxiedMembers(cluster)
	if err != nil {
		return nodeSteps, err
	}
	traefik := &Traefik{
		Version:      TraefikVersion,
		MasterNodes:  masters,
		IngressNodes: ingres,
	}
	if cluster.KsClusterConf != nil &&
		cluster.KsClusterConf.MultiClusterConfig.ClusterRole == constants.ClusterRoleHost &&
		cluster.KsClusterConf.ConsoleConfig.Port != 0 {
		traefik.KSConsolePort = &cluster.KsClusterConf.ConsoleConfig.Port
	}
	content, err := traefik.RenderMainConfig()
	if err != nil {
		return nodeSteps, err
	}
	// Save main config file of traefik to /etc/traefik
	CommonConfig(nodeSet,
		"traefik.toml",
		content,
		schema.TaskCommonDownload{
			TaskType:         constants.TaskDownload,
			Action:           action,
			TimeOut:          300,
			FromDir:          "",
			K8sVersion:       constants.V1_18_6,
			FileList:         []string{},
			SaveTo:           "/etc/traefik",
			IsUseDefaultPath: false,
		}, &nodeSteps, config, nil)
	content, err = traefik.RenderProviderConfig()
	if err != nil {
		return nodeSteps, err
	}
	// Save provider config file of traefik to /etc/traefik/conf
	CommonConfig(nodeSet,
		"providers.toml",
		content,
		schema.TaskCommonDownload{
			TaskType:         constants.TaskDownload,
			Action:           action,
			TimeOut:          300,
			FromDir:          "",
			K8sVersion:       constants.V1_18_6,
			FileList:         []string{},
			SaveTo:           "/etc/traefik/conf",
			IsUseDefaultPath: false,
		}, &nodeSteps, config, nil)
	// Change mode and start traefik service
	CommonBatchRun(nodeSet, action, 300, [][]string{
		{"/bin/sh", "-c", `while [ ! -f /usr/lib/systemd/system/traefik.service ]; do sleep 1; done`},
		{"/bin/sh", "-c", `while [ ! -f /etc/traefik/traefik.toml ]; do sleep 1; done`},
		{"/bin/sh", "-c", `while [ ! -f /etc/traefik/conf/providers.toml ]; do sleep 1; done`},
		{"/bin/sh", "-c", `while [ ! -f /usr/local/bin/traefik ]; do sleep 1; done`},
		{"chmod", "+x", "/usr/local/bin/traefik"},
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", "traefik.service"},
		{"systemctl", "restart", "traefik.service"},
	}, &nodeSteps, false)
	return nodeSteps, nil
}

/**
 * Translate node IDs to node IPs
 */
func (a AddOnClusterLB) translate(ids []string, nodeSet LBNodeSet) []string {
	ips := make([]string, len(ids))
	for i, id := range ids {
		ips[i] = nodeSet[id].Ipv4DefaultIp
	}
	return ips
}

func (a AddOnClusterLB) setupKeepalived(nodeSet LBNodeSet, action, depDir string,
	cluster *schema.Cluster, config serverConfig.Config) ([]schema.NodeStep, error) {
	nodeSteps := []schema.NodeStep{}
	allNodes := a.translate(cluster.ClusterLB.GetAllNodeIDs(), nodeSet)

	rc := cache.GetCurrentCache()
	clusterSet, err := rc.GetClusterCollection()
	if err != nil {
		return nodeSteps, err
	}

	for id, node := range nodeSet {
		keepalived := &Keepalived{
			RouterID:         len(clusterSet) + 1, // no duplicated
			NodeID:           id,
			NodeIP:           node.Ipv4DefaultIp, // current node IP
			UnicastPeers:     funk.SubtractString(allNodes, []string{node.Ipv4DefaultIp}),
			NetworkInterface: nodeSet[id].DefaultMetworkInterface, // NIC name of LB node
			VIP:              cluster.ClusterLB.VIP,
		}
		content, err := keepalived.RenderConfig()
		if err != nil {
			return nodeSteps, err
		}
		// Save config file of keepalived tp /etc/keepalived
		CommonConfig(map[string]schema.NodeInformation{id: nodeSet[id]},
			"keepalived.conf",
			content,
			schema.TaskCommonDownload{
				TaskType:         constants.TaskDownload,
				Action:           action,
				TimeOut:          300,
				FromDir:          "",
				K8sVersion:       constants.V1_18_6,
				FileList:         []string{},
				SaveTo:           "/etc/keepalived",
				IsUseDefaultPath: false,
			}, &nodeSteps, config, nil)
	}

	// Link rpms
	CommonLink(nodeSet, action, 300, keepalivedDep, depDir, depDir, &nodeSteps)
	// Install and start keepalived
	cmds := [][]string{}
	for pkg := range keepalivedDep["centos"][constants.V1_18_6]["x86_64"] {
		cmds = append(cmds, []string{"rpm", "-ivh", "--replacefiles", "--replacepkgs",
			"--nodeps", path.Join(depDir, pkg)})
	}
	CommonBatchRun(nodeSet, action, 300, append(cmds,
		[]string{"/bin/sh", "-c", `while [ ! -f /etc/keepalived/keepalived.conf ]; do sleep 1; done`},
		[]string{"systemctl", "enable", "keepalived.service"},
		[]string{"systemctl", "restart", "keepalived.service"},
	), &nodeSteps, false)
	return nodeSteps, nil
}

func (a AddOnClusterLB) removeClusterLB(nodeSet LBNodeSet,
	action string) ([]schema.NodeStep, error) {
	nodeSteps := []schema.NodeStep{}
	CommonBatchRun(nodeSet, action, 300, [][]string{
		{"/bin/sh", "-c", `if [ -f /usr/lib/systemd/system/traefik.service ];then systemctl stop traefik.service; fi`},
		{"/bin/sh", "-c", `if [ -f /usr/lib/systemd/system/traefik.service ];then systemctl disable traefik.service; fi`},
		{"rm", "-f", "/usr/lib/systemd/system/traefik.service"},
		{"rm", "-rf", "/etc/traefik"},
		{"rm", "-f", "/usr/local/bin/traefik"},
	}, &nodeSteps, false)
	return nodeSteps, nil
}

func (a AddOnClusterLB) removeKeepalived(nodeSet LBNodeSet,
	action string) ([]schema.NodeStep, error) {
	nodeSteps := []schema.NodeStep{}
	CommonBatchRun(nodeSet, action, 300, [][]string{
		{"/bin/sh", "-c", "if [ -f /usr/lib/systemd/system/keepalived.service ];then systemctl stop keepalived.service; fi"},
		{"/bin/sh", "-c", `if [ -f /usr/lib/systemd/system/keepalived.service ];then systemctl disable keepalived.service; fi`},
		{"rm", "-rf", "/etc/keepalived"},
	}, &nodeSteps, false)
	return nodeSteps, nil
}

func (a AddOnClusterLB) getProxiedMembers(cluster *schema.Cluster) (masters []string,
	ingres []string, err error) {
	rc := cache.GetCurrentCache()
	// Get master nodes in cluster
	for _, m := range cluster.Masters {
		node, err := rc.GetNodeInformation(m.NodeId)
		if err != nil {
			return masters, ingres, err
		}
		masters = append(masters, node.Ipv4DefaultIp)
	}
	// Get ingress node in cluster
	if cluster.Ingress != nil {
		for _, i := range cluster.Ingress.NodeIds {
			node, err := rc.GetNodeInformation(i.NodeId)
			if err != nil {
				return masters, ingres, err
			}
			ingres = append(ingres, node.Ipv4DefaultIp)
		}
	}

	return
}
