package task_breaker

import (
	"errors"
	"fmt"
	"reflect"

	ks "k8s-installer/components/kubesphere"
	"k8s-installer/pkg/cache"
	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
)

type AddOnKsInstaller struct {
	Name          string
	TaskTimeOut   int
	FirstMasterId string
	enable        bool
}

func (ks AddOnKsInstaller) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// implement you LivenessProbe step to detect you addons state
	return schema.Step{
		Id:                       "stepCheckKubesphere-" + operationId,
		Name:                     "StepCheckKubesphere",
		NodeSteps:                []schema.NodeStep{},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
	}, nil
}

func (ks AddOnKsInstaller) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		ks.enable = false
	} else {
		// do not look into license switch
		ks.enable = plugAble.IsEnable()
	}
	return ks
}

func (ks AddOnKsInstaller) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (ks AddOnKsInstaller) GetAddOnName() string {
	return ks.Name
}

func (ks AddOnKsInstaller) IsEnable() bool {
	return ks.enable
}

func (ks AddOnKsInstaller) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return ks.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func (ks AddOnKsInstaller) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if cluster.KsClusterConf == nil {
		return nil
	}

	if action == constants.ActionCreate {
		// setup ks
		task, err := installKsInstallerTask(cluster.Masters[0].NodeId, cluster.KsClusterConf.ServerConfig.StorageClass, joinPrivateRegistry(
			cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort,
		), cluster.KsInstaller, 5)
		if err != nil {
			log.Error(err)
			return err
		}

		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "stepKsInstaller-" + operation.Id,
			Name:                     "stepKsInstaller",
			NodeSteps:                []schema.NodeStep{task},
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}

		// stop process member cluster step if the cluster is a host cluster
		if cluster.KsClusterConf.MultiClusterConfig.ClusterRole == constants.ClusterRoleHost {
			return nil
		}

		runtimeCache := cache.GetCurrentCache()
		targetHostCluster, err := runtimeCache.GetCluster(cluster.KsClusterConf.MemberOfCluster)
		if err != nil {
			return err
		}

		// we do not need to check cluster exists because we did it on the api validation level

		var memberIpOrVip string

		hostClusterFirstMasterNode, err := runtimeCache.GetNodeInformation(targetHostCluster.Masters[0].NodeId)

		// unable to found first master of host cluster in db cannot continue to join host
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to found first master node with id: %s of cluster %s due to error: %s, which is not exptected on building step", targetHostCluster.Masters[0].NodeId, cluster.KsClusterConf.MemberOfCluster, err.Error()))
		}

		// unable to found first master of host cluster in db cannot continue to join host
		if hostClusterFirstMasterNode == nil {
			return errors.New(fmt.Sprintf("Failed to found first master node with id: %s of cluster %s which is not exptected on building step", targetHostCluster.Masters[0].NodeId, cluster.KsClusterConf.MemberOfCluster))
		}

		// check host cluster using a external lb
		if cluster.ClusterLB != nil {
			memberIpOrVip = cluster.ClusterLB.VIP
		} else {
			if postClusterFirstMaster, err := runtimeCache.GetNodeInformation(cluster.Masters[0].NodeId); err != nil {
				return errors.New(fmt.Sprintf("Failed to found first master node of post cluster with id: %s due to error: %s", cluster.Masters[0].NodeId, err.Error()))
			} else if postClusterFirstMaster == nil {
				return errors.New(fmt.Sprintf("Failed to found first master node of post cluster with id: %s which is not exptected on building step", cluster.Masters[0].NodeId))
			} else {
				memberIpOrVip = postClusterFirstMaster.Ipv4DefaultIp
			}

		}

		task = schema.NodeStep{
			Id:     utils.GenNodeStepID(),
			Name:   "n" + hostClusterFirstMasterNode.Id,
			NodeID: hostClusterFirstMasterNode.Id,
			Tasks: map[int]schema.ITask{
				0: schema.TaskGenerateKSClusterConfig{
					TaskType:  constants.TaskGenerateKSClusterConfig,
					TimeOut:   200,
					IpAddress: memberIpOrVip,
				},
			},
		}

		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "stepKsConnector-" + operation.Id,
			Name:                     "stepKsConnector",
			NodeSteps:                []schema.NodeStep{task},
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			WaitBeforeRun:            480,
		}

	} else if action == constants.ActionDelete {
		runtimeCache := cache.GetCurrentCache()
		// ensure host cluster is still running
		// or api call might hang
		targetHostCluster, err := runtimeCache.GetCluster(cluster.KsClusterConf.MemberOfCluster)
		if err != nil {
			return err
		}

		if targetHostCluster == nil || targetHostCluster.Status != constants.ClusterStatusRunning {
			return nil
		}
		// if some task is not container based should do something to remove it when api call is try remove cluster
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "removeKsInstaller-" + operation.Id,
			Name:                     "removeKsInstaller",
			NodeSteps:                []schema.NodeStep{removeKsInstallerTask(cluster.Masters[0].NodeId, 60)},
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}
	}
	return nil
}

func installKsInstallerTask(nodeId, storageClass, registry string, ksi ks.DeployKsInstaller, timeOut int) (schema.NodeStep, error) {
	ksi.SetImageRegistry(registry)
	ksitmpl, err := ksi.TemplateRender()
	if err != nil {
		return schema.NodeStep{}, err
	}
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"ks-installer": []byte(ksitmpl),
				},
				TimeOut: timeOut,
			},
		},
	}
	return step, nil
}

func removeKsInstallerTask(firstMasterId string, taskTimeOut int) schema.NodeStep {
	// TODO: remove kubesphere
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + firstMasterId,
		NodeID: firstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:     constants.TaskTypeKubectl,
				SubCommand:   constants.KubectlSubCommandDelete,
				CommandToRun: []string{},
				TimeOut:      taskTimeOut,
			},
		},
	}
	return step
}

type AddOnKsClusterConf struct {
	Name          string
	TaskTimeOut   int
	FirstMasterId string
	enable        bool
}

func (ks AddOnKsClusterConf) LivenessProbe(operationId string, cluster schema.Cluster, config serverConfig.Config) (schema.Step, error) {
	// implement you LivenessProbe step to detect you addons state
	return schema.Step{
		Id:                       "stepCheckKubesphere-" + operationId,
		Name:                     "StepCheckKubesphere",
		NodeSteps:                []schema.NodeStep{},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
	}, nil
}

func (ks AddOnKsClusterConf) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		ks.enable = false
	} else {
		// do not look into license switch
		ks.enable = plugAble.IsEnable()
	}
	return ks
}

func (ks AddOnKsClusterConf) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	return nil
}

func (ks AddOnKsClusterConf) GetAddOnName() string {
	return ks.Name
}

func (ks AddOnKsClusterConf) IsEnable() bool {
	return ks.enable
}

func (ks AddOnKsClusterConf) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	if cluster.Action == constants.ActionCreate {
		return ks.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil
}

func (ks AddOnKsClusterConf) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config, action string) error {
	if action == constants.ActionCreate && cluster.KsClusterConf != nil {
		// setup Kubesphere
		// Enable Kubesphere
		registry := cluster.KsClusterConf.ServerConfig.LocalRegistryServer
		if registry == "" {
			registry = joinPrivateRegistry(
				cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort)
		}
		task, err := installKsClusterConfTask(cluster.Masters[0].NodeId, registry, *cluster.KsClusterConf, 10)
		if err != nil {
			return err
		}

		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "stepKsClusterConf-" + operation.Id,
			Name:                     "stepKsClusterConf",
			NodeSteps:                []schema.NodeStep{task},
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}

		if cluster.KsClusterConf.MetricsServer.Enabled {
			nodeSteps, err := patchMetricServer(cluster.Masters[0].NodeId, action)
			if err != nil {
				return err
			}
			operation.Step[len(operation.Step)] = &schema.Step{
				Id:                       "StepPatchMetricServer-" + operation.Id,
				Name:                     "PatchMetricServer",
				NodeSteps:                nodeSteps,
				WaitBeforeRun:            180,
				OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
				OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			}
		}

		waitStep := CreateWaitStep(120, "Wait-for-ks-object-to-appear")
		operation.Step[len(operation.Step)] = &waitStep

		step, err := createGlobalRole(cluster.Masters[0].NodeId, 10)
		if err != nil {
			return err
		}

		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "StepCreateGlobalRole-" + operation.Id,
			Name:                     "CreateGlobalRole",
			NodeSteps:                []schema.NodeStep{step},
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			WaitBeforeRun:            120,
		}

		patchSteps, err := patchKSGlobalRole(cluster.Masters[0].NodeId, action)
		if err != nil {
			return err
		}
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "StepPatchGlobalRole-" + operation.Id,
			Name:                     "PatchGlobalRole",
			NodeSteps:                patchSteps,
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}

		harborKsConsoleSet := ""

		// when cluster is host and harbor config is not nil, deploy harbor controller
		if cluster.Harbor != nil && cluster.Harbor.Enable {
			if cluster.KsClusterConf.MultiClusterConfig.ClusterRole == constants.KSRoleHost {
				nodePortSvcSteps, err := createKSLDAPNodePortService(cluster.Masters[0].NodeId, action)
				if err != nil {
					return err
				}
				operation.Step[len(operation.Step)] = &schema.Step{
					Id:                       "StepCreateKSLDAPSVC-" + operation.Id,
					Name:                     "CreateKSLDAPSVC",
					NodeSteps:                nodePortSvcSteps,
					OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
					OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
				}
			}

			kubeHarborTask, err := buildHarborControllerTask(cluster.Masters[0].NodeId, cluster.Harbor, registry, len(cluster.Masters), 10)
			if err != nil {
				return err
			}

			kubeHarborNodeSteps := []schema.NodeStep{kubeHarborTask}

			harborKsConsoleSet = "      enableHarborController: true"
			operation.Step[len(operation.Step)] = &schema.Step{
				Id:                       "stepKubeHarbor-" + operation.Id,
				Name:                     "stepKubeHarbor",
				NodeSteps:                kubeHarborNodeSteps,
				OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
				OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			}
		}

		nodeStep := []schema.NodeStep{}

		err = patchKSConsoleConfig(cluster.Masters[0].NodeId, action, harborKsConsoleSet, &nodeStep)
		if err != nil {
			return err
		}

		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "StepPatchKsConsoleConfig-" + operation.Id,
			Name:                     "StepPatchKsConsoleConfig",
			NodeSteps:                nodeStep,
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
		}

		alterTask := customAlterRuleTask(cluster.Masters[0].NodeId, 5, []byte(CustomAlterRule))
		operation.Step[len(operation.Step)] = &schema.Step{
			Id:                       "StepPatchKsAlerRule-" + operation.Id,
			Name:                     "StepPatchKsAlerRule(please wait: 200 secs)",
			NodeSteps:                []schema.NodeStep{alterTask},
			OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
			OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			WaitBeforeRun:            400,
		}
	} else if action == constants.ActionDelete {
		// nothing to do when cluster is being destroy
		// because all components is container based
	}

	return nil
}

func installKsClusterConfTask(nodeId, registry string, ksc ks.KSClusterConfig, timeOut int) (schema.NodeStep, error) {
	ksc.SetLocalImageRegistry(registry)
	ksctmpl, err := ksc.TemplateRender()
	if err != nil {
		return schema.NodeStep{}, err
	}
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"ks-clusterconf": []byte(ksctmpl),
				},
				TimeOut: timeOut,
			},
		},
	}
	return step, nil
}

func removeKsClusterConfTask(firstMasterId string, taskTimeOut int) schema.NodeStep {
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + firstMasterId,
		NodeID: firstMasterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:     constants.TaskTypeKubectl,
				SubCommand:   constants.KubectlSubCommandDelete,
				CommandToRun: []string{},
				TimeOut:      taskTimeOut,
			},
		},
	}
	return step
}

func customAlterRuleTask(masterId string, taskTimeOut int, temp []byte) schema.NodeStep {
	step := schema.NodeStep{
		Id:     "CustomAlterRuleTask" + utils.GenNodeStepID(),
		Name:   "CustomAlterRuleTask-" + masterId,
		NodeID: masterId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"ks-caas-alter-rule": temp,
				},
				TimeOut: taskTimeOut,
			},
		},
	}

	return step
}

func patchMetricServer(nodeId,
	action string) ([]schema.NodeStep, error) {
	nodeSteps := []schema.NodeStep{}
	CommonRunBatchWithSingleNode(nodeId, action, 300, [][]string{
		{"bash", "-c", `kubectl scale deploy metrics-server -n kube-system --replicas=0`},
		{"bash", "-c", `kubectl patch deployment metrics-server --namespace kube-system --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/command", "value":["/metrics-server","--cert-dir=/tmp","--logtostderr","--secure-port=8443","--cert-dir=/tmp","--logtostderr","--secure-port=8443","--kubelet-insecure-tls","--kubelet-preferred-address-types=InternalIP"]}]'`},
		{"bash", "-c", `kubectl scale deploy metrics-server -n kube-system --replicas=1`},
	}, &nodeSteps, false, false)
	return nodeSteps, nil
}

func createGlobalRole(nodeId string, timeOut int) (schema.NodeStep, error) {
	gr := &ks.GlobalRole{}
	manage, err := gr.TemplateRenderManage()
	if err != nil {
		return schema.NodeStep{}, err
	}
	view, err := gr.TemplateRenderView()
	if err != nil {
		return schema.NodeStep{}, err
	}
	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"role-template-manage-ki": []byte(manage),
					"role-template-view-ki":   []byte(view),
				},
				TimeOut: timeOut,
			},
		},
	}
	return step, nil
}

func patchKSGlobalRole(nodeId,
	action string) ([]schema.NodeStep, error) {
	nodeSteps := []schema.NodeStep{}
	CommonRunBatchWithSingleNode(nodeId, action, 300, [][]string{
		{"bash", "-c", `kubectl annotate --overwrite globalrole  platform-regular iam.kubesphere.io/aggregation-roles='["role-template-view-app-templates","role-template-view-ki"]'`},
		{"bash", "-c", `kubectl annotate --overwrite globalrole  platform-admin iam.kubesphere.io/aggregation-roles='["role-template-manage-clusters","role-template-view-clusters","role-template-view-roles","role-template-manage-roles","role-template-view-roles","role-template-view-workspaces","role-template-manage-workspaces","role-template-manage-users","role-template-view-roles","role-template-view-users","role-template-manage-app-templates","role-template-view-app-templates","role-template-manage-platform-settings","role-template-manage-ki"]'`},
		{"bash", "-c", `kubectl annotate --overwrite globalrole  users-manager iam.kubesphere.io/aggregation-roles='["role-template-view-users","role-template-manage-users","role-template-view-roles","role-template-manage-roles","role-template-view-ki"]'`},
		{"bash", "-c", `kubectl annotate --overwrite globalrole  users-manager iam.kubesphere.io/aggregation-roles='["role-template-view-workspaces","role-template-manage-workspaces","role-template-view-users","role-template-view-ki"]'`},
	}, &nodeSteps, false, false)
	return nodeSteps, nil
}

const (
	ksLDAPService = `cat >> /usr/share/k8s-installer/ldap-nodeport.yaml << EOF
apiVersion: v1
kind: Service
metadata:
  name: openldap-nodeport
  namespace: kubesphere-system
spec:
  externalTrafficPolicy: Cluster
  ports:
  - name: ldap
    nodePort: 30881
    port: 389
    protocol: TCP
    targetPort: 389
  selector:
    app.kubernetes.io/instance: ks-openldap
    app.kubernetes.io/name: openldap-ha
  sessionAffinity: None
  type: NodePort
EOF`

	ksConsoleConfig = `cat >> /usr/share/k8s-installer/console-config.yaml << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: ks-console-config
  namespace: kubesphere-system
data:
  local_config.yaml: |
    server:
      http:
        hostname: localhost
        port: 8000
        static:
          production:
            /public: server/public
            /assets: dist/assets
            /dist: dist
      redis:
        port: 6379
        host: redis.kubesphere-system.svc
      redisTimeout: 5000
      sessionTimeout: 7200000
    client:
      version:
        kubesphere: v3.0.0
        kubernetes: v1.18.6
        openpitrix: v0.3.5
      enableKubeConfig: true
      k8s-installer-address: %s
%s
EOF`
)

func createKSLDAPNodePortService(nodeId,
	action string) ([]schema.NodeStep, error) {
	nodeSteps := []schema.NodeStep{}
	CommonRunBatchWithSingleNode(nodeId, action, 300, [][]string{
		{"bash", "-c", ksLDAPService},
		{"bash", "-c", `kubectl apply -f /usr/share/k8s-installer/ldap-nodeport.yaml`},
	}, &nodeSteps, false, false)
	return nodeSteps, nil
}

func buildHarborControllerTask(nodeId string, conf *schema.Harbor, registry string, replicas int, timeOut int) (schema.NodeStep, error) {

	h := &ks.HarborController{
		ImageRegistry: registry,
		Replicas:      replicas,
		Username:      conf.UserName,
		Password:      conf.Password,
	}

	if conf.EnableTls {
		h.Endpoint = fmt.Sprintf("https://%s:%d", conf.Ip, conf.Port)
	} else {
		h.Endpoint = fmt.Sprintf("http://%s:%d", conf.Ip, conf.Port)
	}

	harborControllerTmpl, err := h.TemplateRender()
	if err != nil {
		return schema.NodeStep{}, err
	}

	step := schema.NodeStep{
		Id:     utils.GenNodeStepID(),
		Name:   "TaskTypeKubectl-" + nodeId,
		NodeID: nodeId,
		Tasks: map[int]schema.ITask{
			0: schema.TaskKubectl{
				TaskType:   constants.TaskTypeKubectl,
				SubCommand: constants.KubectlSubCommandCreateOrApply,
				YamlTemplate: map[string][]byte{
					"kube-harbor": []byte(harborControllerTmpl),
				},
				TimeOut: timeOut,
			},
		},
	}
	return step, nil
}

func patchKSConsoleConfig(nodeId, action, harborConfig string, nodeSteps *[]schema.NodeStep) error {
	runtimeContext := cache.GetCurrentCache()
	template := fmt.Sprintf(ksConsoleConfig, runtimeContext.GetServerRuntimeConfig(cache.NodeId).ApiServer.ApiVipAddress, harborConfig)
	CommonRunBatchWithSingleNode(nodeId, action, 300, [][]string{
		{"bash", "-c", template},
		{"bash", "-c", `kubectl apply -f /usr/share/k8s-installer/console-config.yaml`},
		{"bash", "-c", `kubectl rollout restart deployments/ks-console -n kubesphere-system; exit 0`},
	}, nodeSteps, false, false)
	return nil
}

const CustomAlterRule = `
---
apiVersion: v1
items:
- apiVersion: monitoring.coreos.com/v1
  kind: PrometheusRule
  metadata:
    labels:
      prometheus: k8s
      role: alert-rules
    name: prometheus-k8s-caas-rules
    namespace: kubesphere-monitoring-system
  spec:
    groups:
    - name: caas.custom.rules
      rules:
      - alert: NodeLeftCluster
        annotations:
          message: 'Node {{ $labels.node }} left cluster'
        expr: kubelet_node_name{} offset 20m unless kubelet_node_name{}
        for: 30m
        labels:
          severity: warning
      - alert: 	NodeJoinCluster
        annotations:
          message: 'Node {{ $labels.node }} join cluster'
        expr: kubelet_node_name{} unless kubelet_node_name{} offset 20m
        for: 30m
        labels:
          severity: warning
      - alert: PodRestartCount
        annotations:
          message: 'Pod {{ $labels.namespace }}/{{ $labels.pod }} restart count'
        expr: sum(increase(kube_pod_container_status_restarts_total{} [2m]))by(namespace,pod)>0
        labels:
          severity: none
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
`
