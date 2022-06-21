package kubesphere

const ksInstallerTemp = `
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
        image: {{with .ImageRegistry}}{{.}}/{{end}}kubesphere/ks-installer:v3.0.0
        imagePullPolicy: "Always"
        volumeMounts:
        - mountPath: /etc/localtime
          name: host-time
      volumes:
      - hostPath:
          path: /etc/localtime
          type: ""
        name: host-time
`

const ksClusterConfTemp = `
apiVersion: installer.kubesphere.io/v1alpha1
kind: ClusterConfiguration
metadata:
  name: ks-installer
  namespace: kubesphere-system
  labels:
    version: v3.0.0
spec:
  local_registry: {{.RegistryServer}}
  persistence:
    storageClass: {{.Persistence.StorageClass}} # If there is not a default StorageClass in your cluster, you need to specify an existing StorageClass here.
  authentication:
    jwtSecret: {{.Authentication.JwtSecret}} # Keep the jwtSecret consistent with the host cluster. Retrive the jwtSecret by executing "kubectl -n kubesphere-system get cm kubesphere-config -o yaml | grep -v "apiVersion" | grep jwtSecret" on the host cluster.
  etcd:
    monitoring: {{.Etcd.Monitoring}} # Whether to enable etcd monitoring dashboard installation. You have to create a secret for etcd before you enable it.
    endpointIps: {{.Etcd.ListToString}} # etcd cluster EndpointIps, it can be a bunch of IPs here.
    port: {{.Etcd.Port}} # etcd port
    tlsEnable: {{.Etcd.TlsEnable}}
  common:
    mysqlVolumeSize: {{.Common.MysqlVolumeSize}} # MySQL PVC size.
    minioVolumeSize: {{.Common.MinioVolumeSize}} # Minio PVC size.
    etcdVolumeSize: {{.Common.EtcdVolumeSize}}  # etcd PVC size.
    openldapVolumeSize: {{.Common.OpenldapVolumeSize}}  # openldap PVC size.
    redisVolumSize: {{.Common.RedisVolumSize}}  # Redis PVC size.
    es: # Storage backend for logging, events and auditing.
      elasticsearchMasterReplicas: {{.Common.ES.ElasticsearchMasterReplicas}}   # total number of master nodes, it's not allowed to use even number
      elasticsearchDataReplicas: {{.Common.ES.ElasticsearchDataReplicas}}     # total number of data nodes.
      elasticsearchMasterVolumeSize: {{.Common.ES.ElasticsearchMasterVolumeSize}} # Volume size of Elasticsearch master nodes.
      elasticsearchDataVolumeSize: {{.Common.ES.ElasticsearchDataVolumeSize}} # Volume size of Elasticsearch data nodes.
      logMaxAge: {{.Common.ES.LogMaxAge}} # Log retention time in built-in Elasticsearch, it is 7 days by default.
      elkPrefix: {{.Common.ES.ElkPrefix}} # The string making up index names. The index name will be formatted as ks-<elk_prefix>-log.
  console:
    enableMultiLogin: {{.Console.EnableMultiLogin}} # enable/disable multiple sing on, it allows an account can be used by different users at the same time.
    port: {{.Console.Port}}
  alerting: # (CPU: 0.3 Core, Memory: 300 MiB) Whether to install KubeSphere alerting system. It enables Users to customize alerting policies to send messages to receivers in time with different time intervals and alerting levels to choose from.
    enabled: {{.Alerting.Enabled}}
  auditing: # Whether to install KubeSphere audit log system. It provides a security-relevant chronological set of records，recording the sequence of activities happened in platform, initiated by different tenants.
    enabled: {{.Auditing.Enabled}}
  devops: # (CPU: 0.47 Core, Memory: 8.6 G) Whether to install KubeSphere DevOps System. It provides out-of-box CI/CD system based on Jenkins, and automated workflow tools including Source-to-Image & Binary-to-Image.
    enabled: {{.DevOps.Enabled}}
    jenkinsMemoryLim: {{.DevOps.JenkinsMemoryLim}} # Jenkins memory limit.
    jenkinsMemoryReq: {{.DevOps.JenkinsMemoryReq}} # Jenkins memory request.
    jenkinsVolumeSize: {{.DevOps.JenkinsVolumeSize}} # Jenkins volume size.
    jenkinsJavaOpts_Xms: {{.DevOps.JenkinsJavaOptsXms}} # The following three fields are JVM parameters.
    jenkinsJavaOpts_Xmx: {{.DevOps.JenkinsJavaOptsXmx}}
    jenkinsJavaOpts_MaxRAM: {{.DevOps.JenkinsJavaOptsMaxRAM}}
  events: # Whether to install KubeSphere events system. It provides a graphical web console for Kubernetes Events exporting, filtering and alerting in multi-tenant Kubernetes clusters.
    enabled: {{.Events.Enabled}}
    ruler:
      enabled: {{.Events.Ruler.Enabled}}
      replicas: {{.Events.Ruler.Replicas}}
  logging: # (CPU: 57 m, Memory: 2.76 G) Whether to install KubeSphere logging system. Flexible logging functions are provided for log query, collection and management in a unified console. Additional log collectors can be added, such as Elasticsearch, Kafka and Fluentd.
    enabled: {{.Logging.Enabled}}
    logsidecarReplicas: {{.Logging.LogsidecarReplicas}}
  metrics_server: # (CPU: 56 m, Memory: 44.35 MiB) Whether to install metrics-server. IT enables HPA (Horizontal Pod Autoscaler).
    enabled: {{.MetricsServer.Enabled}}
  monitoring:
    prometheusReplicas: {{.Monitoring.PrometheusReplicas}}            # Prometheus replicas are responsible for monitoring different segments of data source and provide high availability as well.
    prometheusMemoryRequest: {{.Monitoring.PrometheusMemoryRequest}} # Prometheus request memory.
    prometheusVolumeSize: {{.Monitoring.PrometheusVolumeSize}} # Prometheus PVC size.
    alertmanagerReplicas: {{.Monitoring.AlertmanagerReplicas}}          # AlertManager Replicas.
  multicluster:
    clusterRole: {{.Multicluster.ClusterRole}}  # host | member | none  # You can install a solo cluster, or specify it as the role of host or member cluster.
  networkpolicy: # Network policies allow network isolation within the same cluster, which means firewalls can be set up between certain instances (Pods).
    # Make sure that the CNI network plugin used by the cluster supports NetworkPolicy. There are a number of CNI network plugins that support NetworkPolicy, including Calico, Cilium, Kube-router, Romana and Weave Net.
    enabled: {{.NetworkPolicy.Enabled}}
  notification: # Email Notification support for the legacy alerting system, should be enabled/disabled together with the above alerting option.
    enabled: {{.Notification.Enabled}}
  openpitrix: # (2 Core, 3.6 G) Whether to install KubeSphere Application Store. It provides an application store for Helm-based applications, and offer application lifecycle management.
    enabled: {{.OpenPitrix.Enabled}}
  servicemesh: # (0.3 Core, 300 MiB) Whether to install KubeSphere Service Mesh (Istio-based). It provides fine-grained traffic management, observability and tracing, and offer visualization for traffic topology.
    enabled: {{.ServiceMesh.Enabled}}
`

// deployment/role-template-manage-ki.yaml
const GlobalRoleManage = `
apiVersion: iam.kubesphere.io/v1alpha2
kind: GlobalRole
metadata:
  annotations:
    iam.kubesphere.io/dependencies: '["role-template-view-ki"]'
    iam.kubesphere.io/module: Clusters Management
    iam.kubesphere.io/role-template-rules: '{"ki": "manage"}'
    kubesphere.io/alias-name: Ki Management
  labels:
    iam.kubesphere.io/role-template: "true"
    kubefed.io/managed: "false"
  name: role-template-manage-ki
rules:
- apiGroups:
  - "core"
  resources:
  - '*'
  verbs:
  - '*'
`

// deployment/role-template-view-ki.yaml
const GlobalRoleView = `
apiVersion: iam.kubesphere.io/v1alpha2
kind: GlobalRole
metadata:
  annotations:
    iam.kubesphere.io/module: Clusters Management
    iam.kubesphere.io/role-template-rules: '{"ki": "view"}'
    kubesphere.io/alias-name: Ki View
  labels:
    iam.kubesphere.io/role-template: "true"
    kubefed.io/managed: "false"
  name: role-template-view-ki
rules:
- apiGroups:
  - "core"
  resources:
  - '*'
  verbs:
  - get
  - list
  - watch
- nonResourceURLs:
  - '*'
  verbs:
  - GET
`

const harborControllerTemplate = `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-harbor
  namespace: kubesphere-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kube-harbor
rules:
- apiGroups:
  - ""
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
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: kube-harbor
subjects:
- kind: ServiceAccount
  name: kube-harbor
  namespace: kubesphere-system
roleRef:
  kind: ClusterRole
  name: kube-harbor
  apiGroup: rbac.authorization.k8s.io
---
kind: ConfigMap
apiVersion: v1
metadata:
  name: harbor-config
  namespace: kubesphere-system
data:
  config.yaml: |
    harbor:
      endpoint: {{.Endpoint}}
      username: {{.Username}}
      password: {{.Password}}

---
kind: Deployment
apiVersion: apps/v1
metadata:
  name: harbor-controller-manager
  namespace: kubesphere-system
  labels:
    app: harbor-controller-manager
    tier: backend
    version: v0.0.1
spec:
  replicas: {{.Replicas}}
  selector:
    matchLabels:
      app: harbor-controller-manager
      tier: backend
      version: v0.0.1
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: harbor-controller-manager
        tier: backend
        version: v0.0.1
    spec:
      volumes:
        - name: config
          configMap:
            name: harbor-config
            defaultMode: 420
        - name: host-time
          hostPath:
            path: /etc/localtime
            type: ''
      containers:
        - name: harbor-controller-manager
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/harbor-controller-manager:v0.0.1
          command:
            - /app/controller-manager
            - '--logtostderr=true'
            - '--leader-elect=true'
          ports:
            - name: api
              containerPort: 8080
              protocol: TCP
            - name: webhook
              containerPort: 8443
              protocol: TCP
            - name: probe
              containerPort: 10300
              protocol: TCP
          resources:
            limits:
              cpu: '1'
              memory: 1000Mi
            requests:
              cpu: 30m
              memory: 50Mi
          volumeMounts:
            - name: config
              mountPath: /etc/kube-harbor/
            - name: host-time
              mountPath: /etc/localtime
          livenessProbe:
            httpGet:
              path: /readyz
              port: 10300
              scheme: HTTP
            initialDelaySeconds: 15
            timeoutSeconds: 1
            periodSeconds: 10
            successThreshold: 1
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /healthz
              port: 10300
              scheme: HTTP
            initialDelaySeconds: 5
            timeoutSeconds: 1
            periodSeconds: 10
            successThreshold: 1
            failureThreshold: 3
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          imagePullPolicy: IfNotPresent
      restartPolicy: Always
      terminationGracePeriodSeconds: 30
      dnsPolicy: ClusterFirst
      serviceAccountName: kube-harbor
      serviceAccount: kube-harbor
      securityContext: {}
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              preference:
                matchExpressions:
                  - key: node-role.kubernetes.io/master
                    operator: In
                    values:
                      - ''
      schedulerName: default-scheduler
      tolerations:
        - key: node-role.kubernetes.io/master
          effect: NoSchedule
        - key: CriticalAddonsOnly
          operator: Exists
        - key: node.kubernetes.io/not-ready
          operator: Exists
          effect: NoExecute
          tolerationSeconds: 60
        - key: node.kubernetes.io/unreachable
          operator: Exists
          effect: NoExecute
          tolerationSeconds: 60
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 25%
      maxSurge: 0
  revisionHistoryLimit: 10
  progressDeadlineSeconds: 600
---
kind: Service
apiVersion: v1
metadata:
  name: harbor-controller-manager
  namespace: kubesphere-system
  labels:
    app: harbor-controller-manager
    tier: backend
    version: v0.0.1
spec:
  ports:
    - protocol: TCP
      port: 443
      targetPort: 8443
  selector:
    app: harbor-controller-manager
    tier: backend
    version: v0.0.1
  type: ClusterIP
  sessionAffinity: None
`
