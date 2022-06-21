package kubesphere

const (
	ksInstallerTemplateV31 = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clusterconfigurations.installer.kubesphere.io
spec:
  group: installer.kubesphere.io
  versions:
  - name: v1alpha1
    served: true
    storage: true
  scope: Namespaced
  names:
    plural: clusterconfigurations
    singular: clusterconfiguration
    kind: ClusterConfiguration
    shortNames:
    - cc

---
apiVersion: v1
kind: Namespace
metadata:
  name: kubesphere-system

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ks-installer
  namespace: kubesphere-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: ks-installer
rules:
- apiGroups:
  - ""
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - apps
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - extensions
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - batch
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - apiregistration.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - apiextensions.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - tenant.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - certificates.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - devops.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - monitoring.coreos.com
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - logging.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - jaegertracing.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - storage.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - admissionregistration.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - policy
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - autoscaling
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - networking.istio.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - config.istio.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - iam.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - notification.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - auditing.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - events.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - core.kubefed.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - installer.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - storage.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - security.istio.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - monitoring.kiali.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - kiali.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - networking.k8s.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - kubeedge.kubesphere.io
  resources:
  - '*'
  verbs:
  - '*'
- apiGroups:
  - types.kubefed.io
  resources:
  - '*'
  verbs:
  - '*'

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: ks-installer
subjects:
- kind: ServiceAccount
  name: ks-installer
  namespace: kubesphere-system
roleRef:
  kind: ClusterRole
  name: ks-installer
  apiGroup: rbac.authorization.k8s.io

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ks-installer
  namespace: kubesphere-system
  labels:
    app: ks-install
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ks-install
  template:
    metadata:
      labels:
        app: ks-install
    spec:
      serviceAccountName: ks-installer
      containers:
      - name: installer
        image: {{with .ImageRegistry}}{{.}}/{{end}}kubesphere/ks-installer:v3.1.0
        imagePullPolicy: "Always"
        resources:
          limits:
            cpu: "1"
            memory: 1Gi
          requests:
            cpu: 20m
            memory: 100Mi
        volumeMounts:
        - mountPath: /etc/localtime
          name: host-time
      volumes:
      - hostPath:
          path: /etc/localtime
          type: ""
        name: host-time
`
	ksClusterConfigurationTemplateV31 = `
---
apiVersion: installer.kubesphere.io/v1alpha1
kind: ClusterConfiguration
metadata:
  name: ks-installer
  namespace: kubesphere-system
  labels:
    version: v3.1.0
spec:
  persistence:
    storageClass: {{.ServerConfig.StorageClass}}        # If there is not a default StorageClass in your cluster, you need to specify an existing StorageClass here.
  authentication:
    jwtSecret: {{.ServerConfig.JwtSecret}}           # Keep the jwtSecret consistent with the host cluster. Retrive the jwtSecret by executing "kubectl -n kubesphere-system get cm kubesphere-config -o yaml | grep -v "apiVersion" | grep jwtSecret" on the host cluster.
  local_registry: {{.ServerConfig.LocalRegistryServer}}        # Add your private registry address if it is needed.
  etcd:
    monitoring: {{.EtcdMonitorConfig.Enabled}}       # Whether to enable etcd monitoring dashboard installation. You have to create a secret for etcd before you enable it.
    endpointIps: {{.EtcdMonitorConfig.ListToString}}  # etcd cluster EndpointIps, it can be a bunch of IPs here.
    port: {{.EtcdMonitorConfig.Port}}              # etcd port
    tlsEnable: {{.EtcdMonitorConfig.TlsEnable}}
  common:
    redis:
      enabled: true
    redisVolumSize: {{.ServerConfig.RedisVolumeSize}} # Redis PVC size.
    openldap:
      enabled: true
    openldapVolumeSize: {{.ServerConfig.OpenldapVolumeSize}}   # openldap PVC size.
    minioVolumeSize: {{.ServerConfig.MinioVolumeSize}} # Minio PVC size.
    monitoring:
      endpoint: {{.MonitorConfig.MonitoringEndpoint}} # Prometheus endpoint to get metrics data
    es:   # Storage backend for logging, events and auditing.
      elasticsearchMasterReplicas: {{.ESConfig.ElasticsearchMasterReplicas}}   # total number of master nodes, it's not allowed to use even number
      elasticsearchDataReplicas: {{.ESConfig.ElasticsearchDataReplicas}}     # total number of data nodes.
      elasticsearchMasterVolumeSize: {{.ESConfig.ElasticsearchMasterVolumeSize}}   # Volume size of Elasticsearch master nodes.
      elasticsearchDataVolumeSize: {{.ESConfig.ElasticsearchDataVolumeSize}}    # Volume size of Elasticsearch data nodes.
      logMaxAge: {{.ESConfig.LogMaxAge}}  # Log retention time in built-in Elasticsearch, it is 7 days by default.
      elkPrefix: {{.ESConfig.ElkPrefix}}  # The string making up index names. The index name will be formatted as ks-<elk_prefix>-log.
      basicAuth:
        enabled: false
        username: ""
        password: ""
      externalElasticsearchUrl: ""
      externalElasticsearchPort: ""
  console:
    enableMultiLogin: {{.ConsoleConfig.EnableMultiLogin}}  # enable/disable multiple sign on, it allows an account can be used by different users at the same time.
    port: {{.ConsoleConfig.Port}}
  alerting:                # (CPU: 0.1 Core, Memory: 100 MiB) Whether to install KubeSphere alerting system. It enables Users to customize alerting policies to send messages to receivers in time with different time intervals and alerting levels to choose from.
    enabled: {{.AlertingConfig.Enabled}}
    thanosruler:
      replicas: {{.AlertingConfig.ThanosRulerReplica}}
    #   resources: {}
  auditing:                # Whether to install KubeSphere audit log system. It provides a security-relevant chronological set of records，recording the sequence of activities happened in platform, initiated by different tenants.
    enabled: {{.AuditingConfig.Enabled}}
  devops:                  # (CPU: 0.47 Core, Memory: 8.6 G) Whether to install KubeSphere DevOps System. It provides out-of-box CI/CD system based on Jenkins, and automated workflow tools including Source-to-Image & Binary-to-Image.
    enabled: {{.DevOpsConfig.Enabled}}
    jenkinsMemoryLim: {{.DevOpsConfig.JenkinsMemoryLimit}}      # Jenkins memory limit.
    jenkinsMemoryReq: {{.DevOpsConfig.JenkinsMemoryRequest}}   # Jenkins memory request.
    jenkinsVolumeSize: {{.DevOpsConfig.JenkinsVolumeSize}}     # Jenkins volume size.
    jenkinsJavaOpts_Xms: {{.DevOpsConfig.JenkinsJavaOptsXms}}  # The following three fields are JVM parameters.
    jenkinsJavaOpts_Xmx: {{.DevOpsConfig.JenkinsJavaOptsXmx}}
    jenkinsJavaOpts_MaxRAM: {{.DevOpsConfig.JenkinsJavaOptsMaxRAM}}
  events:                  # Whether to install KubeSphere events system. It provides a graphical web console for Kubernetes Events exporting, filtering and alerting in multi-tenant Kubernetes clusters.
    enabled: {{.EventsConfig.Enabled}}
    ruler:
      enabled: true
      replicas: {{.EventsConfig.RulerReplicas}}
  logging:                 # (CPU: 57 m, Memory: 2.76 G) Whether to install KubeSphere logging system. Flexible logging functions are provided for log query, collection and management in a unified console. Additional log collectors can be added, such as Elasticsearch, Kafka and Fluentd.
    enabled: {{.LoggingConfig.Enabled}}
    logsidecar:
      enabled: true
      replicas: {{.LoggingConfig.LogSidecarReplicas}}
  metrics_server:                    # (CPU: 56 m, Memory: 44.35 MiB) Whether to install metrics-server. IT enables HPA (Horizontal Pod Autoscaler).
    enabled: {{.MetricsServer.Enabled}}
  monitoring:
    storageClass: {{if .MonitorConfig.StorageClass}}{{.MonitorConfig.StorageClass}}{{else}}""{{end}}            # If there is a independent StorageClass your need for prometheus, you can specify it here. default StorageClass used by default.
    prometheusReplicas: {{.MonitorConfig.PrometheusReplicas}}            # Prometheus replicas are responsible for monitoring different segments of data source and provide high availability as well.
    prometheusMemoryRequest: {{.MonitorConfig.PrometheusMemoryRequest}}   # Prometheus request memory.
    prometheusVolumeSize: {{.MonitorConfig.PrometheusVolumeSize}}       # Prometheus PVC size.
    alertmanagerReplicas: {{.MonitorConfig.AlertManagerReplicas}}          # AlertManager Replicas.
  multicluster:
    clusterRole: {{.MultiClusterConfig.ClusterRole}}  # host | member | none  # You can install a solo cluster, or specify it as the role of host or member cluster.
  network:
    networkpolicy: # Network policies allow network isolation within the same cluster, which means firewalls can be set up between certain instances (Pods).
      # Make sure that the CNI network plugin used by the cluster supports NetworkPolicy. There are a number of CNI network plugins that support NetworkPolicy, including Calico, Cilium, Kube-router, Romana and Weave Net.
      enabled: {{.NetworkConfig.NetworkPolicyEnabled}}
    ippool: # if calico cni is integrated then use the value "calico", "none" means that the ippool function is disabled
      type: {{.NetworkConfig.IPPool}}
    topology: # "weave-scope" means to use "weave-scope" to provide network topology information, "none" means that the topology function is disabled
      type: {{.NetworkConfig.NetworkTopology}}
  openpitrix:
    store:
      enabled: {{.OpenPitrixConfig.Enabled}}
  servicemesh:         # (0.3 Core, 300 MiB) Whether to install KubeSphere Service Mesh (Istio-based). It provides fine-grained traffic management, observability and tracing, and offer visualization for traffic topology.
    enabled: {{.ServiceMeshConfig.Enabled}}     # base component (pilot)
  kubeedge:
    enabled: false
    cloudCore:
      nodeSelector: {"node-role.kubernetes.io/worker": ""}
      tolerations: []
      cloudhubPort: "10000"
      cloudhubQuicPort: "10001"
      cloudhubHttpsPort: "10002"
      cloudstreamPort: "10003"
      tunnelPort: "10004"
      cloudHub:
        advertiseAddress: # At least a public IP Address or an IP which can be accessed by edge nodes must be provided
          - ""            # Causion!: Leave this entry to empty will cause CloudCore to exit abnormally once KubeEdge is enabled.
        nodeLimit: "100"
      service:
        cloudhubNodePort: "30000"
        cloudhubQuicNodePort: "30001"
        cloudhubHttpsNodePort: "30002"
        cloudstreamNodePort: "30003"
        tunnelNodePort: "30004"
    edgeWatcher:
      nodeSelector: {"node-role.kubernetes.io/worker": ""}
      tolerations: []
      edgeWatcherAgent:
        nodeSelector: {"node-role.kubernetes.io/worker": ""}
        tolerations: []
`
)
