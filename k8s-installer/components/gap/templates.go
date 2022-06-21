package gap

const GAPPrometheusAlertRules = `
serverFiles:
  alerts:
    groups:
    - name: k8s_alert_rules
      rules:
      # ALERT when container memory usage exceed 90%
      - alert: container_mem_over_90
        expr: (sum(container_memory_working_set_bytes{image!="",name=~"^k8s_.*", pod_name!=""}) by (pod_name)) / (sum (container_spec_memory_limit_bytes{image!="",name=~"^k8s_.*", pod_name!=""}) by (pod_name)) > 0.9 and (sum(container_memory_working_set_bytes{image!="",name=~"^k8s_.*", pod_name!=""}) by (pod_name)) / (sum (container_spec_memory_limit_bytes{image!="",name=~"^k8s_.*", pod_name!=""}) by (pod_name)) < 2
        for: 2m
        annotations:
          summary: "{{ $labels.pod_name }}'s memory usage alert"
          description: "Memory Usage of Pod {{ $labels.pod_name }} on {{ $labels.kubernetes_io_hostname }} has exceeded 90% for more than 2 minutes."

      # ALERT when node is down
      - alert: node_down
        expr: up == 0
        for: 60s
        annotations:
          summary: "Node {{ $labels.kubernetes_io_hostname }} is down"
          description: "Node {{ $labels.kubernetes_io_hostname }} is down"`

const GAPPrometheusAlertsManager = `
alertmanagerFiles:
  alertmanager.yml:
    global:
      smtp_smarthost: 'smtp.163.com:25'
      smtp_from: 'xxxx@163.com'
      smtp_auth_username: 'xxxx@163.com'
      smtp_auth_password: '*********'
      smtp_require_tls: false

    route:
      group_by: ['alertname', 'pod_name']
      group_wait: 10s
      group_interval: 5m
      #receiver: AlertMail
      receiver: dingtalk
      repeat_interval: 3h

    receivers:
    - name: 'AlertMail'
      email_configs:
      - to: 'xxxx@163.com'
    - name: dingtalk
      webhook_configs:
      - send_resolved: false
        # 需要运行插件 dingtalk-webhook.yaml，详情阅读 docs/guide/prometheus.md
        url: http://webhook-dingtalk.monitoring.svc.cluster.local:8060/dingtalk/webhook1/send
`

const GAPPrometheusSettings = `
alertmanager:
  persistentVolume:
    enabled: false
  service:
    type: NodePort
    nodePort: 31001

nodeExporter:
  service:
    clusterIP: ""

server:
  persistentVolume:
    enabled: false
  service:
    type: NodePort
    nodePort: 31000

pushgateway:
  enabled: false

kubeStateMetrics:
  image:
    repository: caas4/kube-state-metrics`

const GAPPrometheusChart = `
name: prometheus
version: 7.1.4
appVersion: 2.4.3
description: Prometheus is a monitoring system and time series database.
home: https://prometheus.io/
icon: https://raw.githubusercontent.com/prometheus/prometheus.github.io/master/assets/prometheus_logo-cb55bb5c346.png
sources:
  - https://github.com/prometheus/alertmanager
  - https://github.com/prometheus/prometheus
  - https://github.com/prometheus/pushgateway
  - https://github.com/prometheus/node_exporter
  - https://github.com/kubernetes/kube-state-metrics
maintainers:
  - name: mgoodness
    email: mgoodness@gmail.com
  - name: gianrubio
    email: gianrubio@gmail.com
engine: gotpl
tillerVersion: ">=2.8.0"`

const GAPPrometheusValues = `
rbac:
  create: true

## Define serviceAccount names for components. Defaults to component's fully qualified name.
##
serviceAccounts:
  alertmanager:
    create: true
    name:
  kubeStateMetrics:
    create: true
    name:
  nodeExporter:
    create: true
    name:
  pushgateway:
    create: true
    name:
  server:
    create: true
    name:

alertmanager:
  ## If false, alertmanager will not be installed
  ##
  enabled: true

  ## alertmanager container name
  ##
  name: alertmanager

  ## alertmanager container image
  ##
  image:
    repository: caas4/alertmanager
    tag: v0.15.2
    pullPolicy: IfNotPresent

  ## alertmanager priorityClassName
  ##
  priorityClassName: ""

  ## Additional alertmanager container arguments
  ##
  extraArgs: {}

  ## The URL prefix at which the container can be accessed. Useful in the case the '-web.external-url' includes a slug
  ## so that the various internal URLs are still able to access as they are in the default case.
  ## (Optional)
  prefixURL: ""

  ## External URL which can access alertmanager
  ## Maybe same with Ingress host name
  baseURL: "/"

  ## Additional alertmanager container environment variable
  ## For instance to add a http_proxy
  ##
  extraEnv: {}

  ## ConfigMap override where fullname is {{.Release.Name}}-{{.Values.alertmanager.configMapOverrideName}}
  ## Defining configMapOverrideName will cause templates/alertmanager-configmap.yaml
  ## to NOT generate a ConfigMap resource
  ##
  configMapOverrideName: ""

  ingress:
    ## If true, alertmanager Ingress will be created
    ##
    enabled: false

    ## alertmanager Ingress annotations
    ##
    annotations: {}
    #   kubernetes.io/ingress.class: nginx
    #   kubernetes.io/tls-acme: 'true'

    ## alertmanager Ingress additional labels
    ##
    extraLabels: {}

    ## alertmanager Ingress hostnames with optional path
    ## Must be provided if Ingress is enabled
    ##
    hosts: []
    #   - alertmanager.domain.com
    #   - domain.com/alertmanager

    ## alertmanager Ingress TLS configuration
    ## Secrets must be manually created in the namespace
    ##
    tls: []
    #   - secretName: prometheus-alerts-tls
    #     hosts:
    #       - alertmanager.domain.com

  ## Alertmanager Deployment Strategy type
  # strategy:
  #   type: Recreate

  ## Node tolerations for alertmanager scheduling to nodes with taints
  ## Ref: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
  ##
  tolerations: []
    # - key: "key"
    #   operator: "Equal|Exists"
    #   value: "value"
    #   effect: "NoSchedule|PreferNoSchedule|NoExecute(1.6 only)"

  ## Node labels for alertmanager pod assignment
  ## Ref: https://kubernetes.io/docs/user-guide/node-selection/
  ##
  nodeSelector: {}

  ## Pod affinity
  ##
  affinity: {}

  ## Use an alternate scheduler, e.g. "stork".
  ## ref: https://kubernetes.io/docs/tasks/administer-cluster/configure-multiple-schedulers/
  ##
  # schedulerName:

  persistentVolume:
    ## If true, alertmanager will create/use a Persistent Volume Claim
    ## If false, use emptyDir
    ##
    enabled: false

    ## alertmanager data Persistent Volume access modes
    ## Must match those of existing PV or dynamic provisioner
    ## Ref: http://kubernetes.io/docs/user-guide/persistent-volumes/
    ##
    accessModes:
      - ReadWriteOnce

    ## alertmanager data Persistent Volume Claim annotations
    ##
    annotations: {}

    ## alertmanager data Persistent Volume existing claim name
    ## Requires alertmanager.persistentVolume.enabled: true
    ## If defined, PVC must be created manually before volume will be bound
    existingClaim: ""

    ## alertmanager data Persistent Volume mount root path
    ##
    mountPath: /data

    ## alertmanager data Persistent Volume size
    ##
    size: 2Gi

    ## alertmanager data Persistent Volume Storage Class
    ## If defined, storageClassName: <storageClass>
    ## If set to "-", storageClassName: "", which disables dynamic provisioning
    ## If undefined (the default) or set to null, no storageClassName spec is
    ##   set, choosing the default provisioner.  (gp2 on AWS, standard on
    ##   GKE, AWS & OpenStack)
    ##
    # storageClass: "-"

    ## Subdirectory of alertmanager data Persistent Volume to mount
    ## Useful if the volume's root directory is not empty
    ##
    subPath: ""

  ## Annotations to be added to alertmanager pods
  ##
  podAnnotations: {}

  replicaCount: 1

  ## alertmanager resource requests and limits
  ## Ref: http://kubernetes.io/docs/user-guide/compute-resources/
  ##
  resources: {}
    # limits:
    #   cpu: 10m
    #   memory: 32Mi
    # requests:
    #   cpu: 10m
    #   memory: 32Mi

  ## Security context to be added to alertmanager pods
  ##
  securityContext: {}

  service:
    annotations: {}
    labels: {}
    clusterIP: ""

    ## Enabling peer mesh service end points for enabling the HA alert manager
    ## Ref: https://github.com/prometheus/alertmanager/blob/master/README.md
    # enableMeshPeer : true

    ## List of IP addresses at which the alertmanager service is available
    ## Ref: https://kubernetes.io/docs/user-guide/services/#external-ips
    ##
    externalIPs: []

    loadBalancerIP: ""
    loadBalancerSourceRanges: []
    servicePort: 80
    # nodePort: 30000
    type: ClusterIP

## Monitors ConfigMap changes and POSTs to a URL
## Ref: https://github.com/jimmidyson/configmap-reload
##
configmapReload:
  ## configmap-reload container name
  ##
  name: configmap-reload

  ## configmap-reload container image
  ##
  image:
    repository: caas4/configmap-reload
    tag: v0.2.2
    pullPolicy: IfNotPresent

  ## Additional configmap-reload container arguments
  ##
  extraArgs: {}

  ## Additional configmap-reload mounts
  ##
  extraConfigmapMounts: []
    # - name: prometheus-alerts
    #   mountPath: /etc/alerts.d
    #   configMap: prometheus-alerts
    #   readOnly: true


  ## configmap-reload resource requests and limits
  ## Ref: http://kubernetes.io/docs/user-guide/compute-resources/
  ##
  resources: {}

initChownData:
  ## If false, data ownership will not be reset at startup
  ## This allows the prometheus-server to be run with an arbitrary user
  ##
  enabled: true

  ## initChownData container name
  ##
  name: init-chown-data

  ## initChownData container image
  ##
  image:
    repository: caas4/busybox
    tag: latest
    pullPolicy: IfNotPresent

  ## initChownData resource requests and limits
  ## Ref: http://kubernetes.io/docs/user-guide/compute-resources/
  ##
  resources: {}

kubeStateMetrics:
  ## If false, kube-state-metrics will not be installed
  ##
  enabled: true

  ## kube-state-metrics container name
  ##
  name: kube-state-metrics

  ## kube-state-metrics container image
  ##
  image:
    repository: caas4/kube-state-metrics
    tag: v1.4.0
    pullPolicy: IfNotPresent

  ## kube-state-metrics priorityClassName
  ##
  priorityClassName: ""

  ## kube-state-metrics container arguments
  ##
  args: {}

  ## Node tolerations for kube-state-metrics scheduling to nodes with taints
  ## Ref: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
  ##
  tolerations: []
    # - key: "key"
    #   operator: "Equal|Exists"
    #   value: "value"
    #   effect: "NoSchedule|PreferNoSchedule|NoExecute(1.6 only)"

  ## Node labels for kube-state-metrics pod assignment
  ## Ref: https://kubernetes.io/docs/user-guide/node-selection/
  ##
  nodeSelector: {}

  ## Annotations to be added to kube-state-metrics pods
  ##
  podAnnotations: {}

  pod:
    labels: {}

  replicaCount: 1

  ## kube-state-metrics resource requests and limits
  ## Ref: http://kubernetes.io/docs/user-guide/compute-resources/
  ##
  resources: {}
    # limits:
    #   cpu: 10m
    #   memory: 16Mi
    # requests:
    #   cpu: 10m
    #   memory: 16Mi

  ## Security context to be added to kube-state-metrics pods
  ##
  securityContext: {}

  service:
    annotations:
      prometheus.io/scrape: "true"
    labels: {}

    # Exposed as a headless service:
    # https://kubernetes.io/docs/concepts/services-networking/service/#headless-services
    clusterIP: None

    ## List of IP addresses at which the kube-state-metrics service is available
    ## Ref: https://kubernetes.io/docs/user-guide/services/#external-ips
    ##
    externalIPs: []

    loadBalancerIP: ""
    loadBalancerSourceRanges: []
    servicePort: 80
    type: ClusterIP

nodeExporter:
  ## If false, node-exporter will not be installed
  ##
  enabled: true

  ## If true, node-exporter pods share the host network namespace
  ##
  hostNetwork: true

  ## If true, node-exporter pods share the host PID namespace
  ##
  hostPID: true

  ## node-exporter container name
  ##
  name: node-exporter

  ## node-exporter container image
  ##
  image:
    repository: caas4/node-exporter
    tag: v0.15.2
    pullPolicy: IfNotPresent

  ## node-exporter priorityClassName
  ##
  priorityClassName: ""

  ## Custom Update Strategy
  ##
  updateStrategy:
    type: OnDelete

  ## Additional node-exporter container arguments
  ##
  extraArgs: {}

  ## Additional node-exporter hostPath mounts
  ##
  extraHostPathMounts: []
    # - name: textfile-dir
    #   mountPath: /srv/txt_collector
    #   hostPath: /var/lib/node-exporter
    #   readOnly: true

  extraConfigmapMounts: []
    # - name: certs-configmap
    #   mountPath: /prometheus
    #   configMap: certs-configmap
    #   readOnly: true

  ## Node tolerations for node-exporter scheduling to nodes with taints
  ## Ref: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
  ##
  tolerations: []
    # - key: "key"
    #   operator: "Equal|Exists"
    #   value: "value"
    #   effect: "NoSchedule|PreferNoSchedule|NoExecute(1.6 only)"

  ## Node labels for node-exporter pod assignment
  ## Ref: https://kubernetes.io/docs/user-guide/node-selection/
  ##
  nodeSelector: {}

  ## Annotations to be added to node-exporter pods
  ##
  podAnnotations: {}

  ## Labels to be added to node-exporter pods
  ##
  pod:
    labels: {}

  ## node-exporter resource limits & requests
  ## Ref: https://kubernetes.io/docs/user-guide/compute-resources/
  ##
  resources: {}
    # limits:
    #   cpu: 200m
    #   memory: 50Mi
    # requests:
    #   cpu: 100m
    #   memory: 30Mi

  ## Security context to be added to node-exporter pods
  ##
  securityContext: {}
    # runAsUser: 0

  service:
    annotations:
      prometheus.io/scrape: "true"
    labels: {}

    # Exposed as a headless service:
    # https://kubernetes.io/docs/concepts/services-networking/service/#headless-services
    clusterIP: None

    ## List of IP addresses at which the node-exporter service is available
    ## Ref: https://kubernetes.io/docs/user-guide/services/#external-ips
    ##
    externalIPs: []

    #hostPort: 9100
    loadBalancerIP: ""
    loadBalancerSourceRanges: []
    servicePort: 9100
    type: NodePort
    nodePort: 31000

server:
  ## Prometheus server container name
  ##
  name: server

  ## Prometheus server container image
  ##
  image:
    repository: caas4/prometheus
    tag: v2.4.3
    pullPolicy: IfNotPresent

  ## prometheus server priorityClassName
  ##
  priorityClassName: ""

  ## The URL prefix at which the container can be accessed. Useful in the case the '-web.external-url' includes a slug
  ## so that the various internal URLs are still able to access as they are in the default case.
  ## (Optional)
  prefixURL: ""

  ## External URL which can access alertmanager
  ## Maybe same with Ingress host name
  baseURL: ""

  ## This flag controls access to the administrative HTTP API which includes functionality such as deleting time
  ## series. This is disabled by default.
  enableAdminApi: false

  global:
    ## How frequently to scrape targets by default
    ##
    scrape_interval: 1m
    ## How long until a scrape request times out
    ##
    scrape_timeout: 10s
    ## How frequently to evaluate rules
    ##
    evaluation_interval: 1m

  ## Additional Prometheus server container arguments
  ##
  extraArgs: {}

  ## Additional Prometheus server hostPath mounts
  ##
  extraHostPathMounts: []
    # - name: certs-dir
    #   mountPath: /etc/kubernetes/certs
    #   hostPath: /etc/kubernetes/certs
    #   readOnly: true

  extraConfigmapMounts: []
    # - name: certs-configmap
    #   mountPath: /prometheus
    #   configMap: certs-configmap
    #   readOnly: true

  ## Additional Prometheus server Secret mounts
  # Defines additional mounts with secrets. Secrets must be manually created in the namespace.
  extraSecretMounts: []
    # - name: secret-files
    #   mountPath: /etc/secrets
    #   secretName: prom-secret-files
    #   readOnly: true

  ## ConfigMap override where fullname is {{.Release.Name}}-{{.Values.server.configMapOverrideName}}
  ## Defining configMapOverrideName will cause templates/server-configmap.yaml
  ## to NOT generate a ConfigMap resource
  ##
  configMapOverrideName: ""

  ingress:
    ## If true, Prometheus server Ingress will be created
    ##
    enabled: false

    ## Prometheus server Ingress annotations
    ##
    annotations: {}
    #   kubernetes.io/ingress.class: nginx
    #   kubernetes.io/tls-acme: 'true'

    ## Prometheus server Ingress additional labels
    ##
    extraLabels: {}

    ## Prometheus server Ingress hostnames with optional path
    ## Must be provided if Ingress is enabled
    ##
    hosts: []
    #   - prometheus.domain.com
    #   - domain.com/prometheus

    ## Prometheus server Ingress TLS configuration
    ## Secrets must be manually created in the namespace
    ##
    tls: []
    #   - secretName: prometheus-server-tls
    #     hosts:
    #       - prometheus.domain.com

  ## Server Deployment Strategy type
  # strategy:
  #   type: Recreate

  ## Node tolerations for server scheduling to nodes with taints
  ## Ref: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
  ##
  tolerations: []
    # - key: "key"
    #   operator: "Equal|Exists"
    #   value: "value"
    #   effect: "NoSchedule|PreferNoSchedule|NoExecute(1.6 only)"

  ## Node labels for Prometheus server pod assignment
  ## Ref: https://kubernetes.io/docs/user-guide/node-selection/
  ##
  nodeSelector: {}

  ## Pod affinity
  ##
  affinity: {}

  ## Use an alternate scheduler, e.g. "stork".
  ## ref: https://kubernetes.io/docs/tasks/administer-cluster/configure-multiple-schedulers/
  ##
  # schedulerName:

  persistentVolume:
    ## If true, Prometheus server will create/use a Persistent Volume Claim
    ## If false, use emptyDir
    ##
    enabled: false

    ## Prometheus server data Persistent Volume access modes
    ## Must match those of existing PV or dynamic provisioner
    ## Ref: http://kubernetes.io/docs/user-guide/persistent-volumes/
    ##
    accessModes:
      - ReadWriteOnce

    ## Prometheus server data Persistent Volume annotations
    ##
    annotations: {}

    ## Prometheus server data Persistent Volume existing claim name
    ## Requires server.persistentVolume.enabled: true
    ## If defined, PVC must be created manually before volume will be bound
    existingClaim: ""

    ## Prometheus server data Persistent Volume mount root path
    ##
    mountPath: /data

    ## Prometheus server data Persistent Volume size
    ##
    size: 8Gi

    ## Prometheus server data Persistent Volume Storage Class
    ## If defined, storageClassName: <storageClass>
    ## If set to "-", storageClassName: "", which disables dynamic provisioning
    ## If undefined (the default) or set to null, no storageClassName spec is
    ##   set, choosing the default provisioner.  (gp2 on AWS, standard on
    ##   GKE, AWS & OpenStack)
    ##
    # storageClass: "-"

    ## Subdirectory of Prometheus server data Persistent Volume to mount
    ## Useful if the volume's root directory is not empty
    ##
    subPath: ""

  ## Annotations to be added to Prometheus server pods
  ##
  podAnnotations: {}
    # iam.amazonaws.com/role: prometheus

  replicaCount: 1

  ## Prometheus server resource requests and limits
  ## Ref: http://kubernetes.io/docs/user-guide/compute-resources/
  ##
  resources: {}
    # limits:
    #   cpu: 500m
    #   memory: 512Mi
    # requests:
    #   cpu: 500m
    #   memory: 512Mi

  ## Security context to be added to server pods
  ##
  securityContext: {}

  service:
    annotations: {}
    labels: {}
    clusterIP: ""

    ## List of IP addresses at which the Prometheus server service is available
    ## Ref: https://kubernetes.io/docs/user-guide/services/#external-ips
    ##
    externalIPs: []

    loadBalancerIP: ""
    loadBalancerSourceRanges: []
    servicePort: 80
    type: ClusterIP

  ## Prometheus server pod termination grace period
  ##
  terminationGracePeriodSeconds: 300

  ## Prometheus data retention period (i.e 360h)
  ##
  retention: ""

pushgateway:
  ## If false, pushgateway will not be installed
  ##
  enabled: false

  ## pushgateway container name
  ##
  name: pushgateway

  ## pushgateway container image
  ##
  image:
    repository: caas4/pushgateway
    tag: v0.5.2
    pullPolicy: IfNotPresent

  ## pushgateway priorityClassName
  ##
  priorityClassName: ""

  ## Additional pushgateway container arguments
  ##
  extraArgs: {}

  ingress:
    ## If true, pushgateway Ingress will be created
    ##
    enabled: false

    ## pushgateway Ingress annotations
    ##
    annotations: {}
    #   kubernetes.io/ingress.class: nginx
    #   kubernetes.io/tls-acme: 'true'

    ## pushgateway Ingress hostnames with optional path
    ## Must be provided if Ingress is enabled
    ##
    hosts: []
    #   - pushgateway.domain.com
    #   - domain.com/pushgateway

    ## pushgateway Ingress TLS configuration
    ## Secrets must be manually created in the namespace
    ##
    tls: []
    #   - secretName: prometheus-alerts-tls
    #     hosts:
    #       - pushgateway.domain.com

  ## Node tolerations for pushgateway scheduling to nodes with taints
  ## Ref: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
  ##
  tolerations: []
    # - key: "key"
    #   operator: "Equal|Exists"
    #   value: "value"
    #   effect: "NoSchedule|PreferNoSchedule|NoExecute(1.6 only)"

  ## Node labels for pushgateway pod assignment
  ## Ref: https://kubernetes.io/docs/user-guide/node-selection/
  ##
  nodeSelector: {}

  ## Annotations to be added to pushgateway pods
  ##
  podAnnotations: {}

  replicaCount: 1

  ## pushgateway resource requests and limits
  ## Ref: http://kubernetes.io/docs/user-guide/compute-resources/
  ##
  resources: {}
    # limits:
    #   cpu: 10m
    #   memory: 32Mi
    # requests:
    #   cpu: 10m
    #   memory: 32Mi

  ## Security context to be added to push-gateway pods
  ##
  securityContext: {}

  service:
    annotations:
      prometheus.io/probe: pushgateway
    labels: {}
    clusterIP: ""

    ## List of IP addresses at which the pushgateway service is available
    ## Ref: https://kubernetes.io/docs/user-guide/services/#external-ips
    ##
    externalIPs: []

    loadBalancerIP: ""
    loadBalancerSourceRanges: []
    servicePort: 9091
    type: NodePort
    nodePort: 31001

## alertmanager ConfigMap entries
##
alertmanagerFiles:
  alertmanager.yml:
    global: {}
      # slack_api_url: ''

    receivers:
      - name: default-receiver
        # slack_configs:
        #  - channel: '@you'
        #    send_resolved: true

    route:
      group_wait: 10s
      group_interval: 5m
      receiver: default-receiver
      repeat_interval: 3h

## Prometheus server ConfigMap entries
##
serverFiles:
  alerts: {}
  rules: {}

  prometheus.yml:
    rule_files:
      - /etc/config/rules
      - /etc/config/alerts

    scrape_configs:
      - job_name: prometheus
        static_configs:
          - targets:
            - localhost:9090

      # A scrape configuration for running Prometheus on a Kubernetes cluster.
      # This uses separate scrape configs for cluster components (i.e. API server, node)
      # and services to allow each to use different authentication configs.
      #
      # Kubernetes labels will be added as Prometheus labels on metrics via the

      # Scrape config for API servers.
      #
      # Kubernetes exposes API servers as endpoints to the default/kubernetes
      # the endpoints associated with the default/kubernetes service using the
      # well as HA API server deployments.
      - job_name: 'kubernetes-apiservers'

        kubernetes_sd_configs:
          - role: endpoints

        # Default to scraping over https. If required, just disable this or change to
        scheme: https

        # This TLS & bearer token file config is used to connect to the actual scrape
        # endpoints for cluster components. This is separate to discovery auth
        # configuration because discovery & scraping are two separate concerns in
        # Prometheus. The discovery auth config is automatic if Prometheus runs inside
        # the cluster. Otherwise, more config options have to be provided within the
        # <kubernetes_sd_config>.
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
          # If your node certificates are self-signed or use a different CA to the
          # master CA, then disable certificate verification below. Note that
          # certificate verification is an integral part of a secure infrastructure
          # so this should only be disabled in a controlled environment. You can
          # disable certificate verification by uncommenting the line below.
          #
          insecure_skip_verify: true
        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

        # Keep only the default/kubernetes service endpoints for the https port. This
        # will add targets for each API server which Kubernetes adds an endpoint to
        # the default/kubernetes service.
        relabel_configs:
          - source_labels: [__meta_kubernetes_namespace, __meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
            action: keep
            regex: default;kubernetes;https

      - job_name: 'kubernetes-nodes'

        # Default to scraping over https. If required, just disable this or change to
        scheme: https

        # This TLS & bearer token file config is used to connect to the actual scrape
        # endpoints for cluster components. This is separate to discovery auth
        # configuration because discovery & scraping are two separate concerns in
        # Prometheus. The discovery auth config is automatic if Prometheus runs inside
        # the cluster. Otherwise, more config options have to be provided within the
        # <kubernetes_sd_config>.
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
          # If your node certificates are self-signed or use a different CA to the
          # master CA, then disable certificate verification below. Note that
          # certificate verification is an integral part of a secure infrastructure
          # so this should only be disabled in a controlled environment. You can
          # disable certificate verification by uncommenting the line below.
          #
          insecure_skip_verify: true
        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

        kubernetes_sd_configs:
          - role: node

        relabel_configs:
          - action: labelmap
            regex: __meta_kubernetes_node_label_(.+)
          - target_label: __address__
            replacement: kubernetes.default.svc:443
          - source_labels: [__meta_kubernetes_node_name]
            regex: (.+)
            target_label: __metrics_path__
            replacement: /api/v1/nodes/${1}/proxy/metrics


      - job_name: 'kubernetes-nodes-cadvisor'

        # Default to scraping over https. If required, just disable this or change to
        scheme: https

        # This TLS & bearer token file config is used to connect to the actual scrape
        # endpoints for cluster components. This is separate to discovery auth
        # configuration because discovery & scraping are two separate concerns in
        # Prometheus. The discovery auth config is automatic if Prometheus runs inside
        # the cluster. Otherwise, more config options have to be provided within the
        # <kubernetes_sd_config>.
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
          # If your node certificates are self-signed or use a different CA to the
          # master CA, then disable certificate verification below. Note that
          # certificate verification is an integral part of a secure infrastructure
          # so this should only be disabled in a controlled environment. You can
          # disable certificate verification by uncommenting the line below.
          #
          insecure_skip_verify: true
        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token

        kubernetes_sd_configs:
          - role: node

        # This configuration will work only on kubelet 1.7.3+
        # As the scrape endpoints for cAdvisor have changed
        # if you are using older version you need to change the replacement to
        # replacement: /api/v1/nodes/${1}:4194/proxy/metrics
        # more info here https://github.com/coreos/prometheus-operator/issues/633
        relabel_configs:
          - action: labelmap
            regex: __meta_kubernetes_node_label_(.+)
          - target_label: __address__
            replacement: kubernetes.default.svc:443
          - source_labels: [__meta_kubernetes_node_name]
            regex: (.+)
            target_label: __metrics_path__
            replacement: /api/v1/nodes/${1}/proxy/metrics/cadvisor

      # Scrape config for service endpoints.
      #
      # The relabeling allows the actual service scrape endpoint to be configured
      # via the following annotations:
      #
      # service then set this appropriately.
      - job_name: 'kubernetes-service-endpoints'

        kubernetes_sd_configs:
          - role: endpoints

        relabel_configs:
          - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scrape]
            action: keep
            regex: true
          - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_scheme]
            action: replace
            target_label: __scheme__
            regex: (https?)
          - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_path]
            action: replace
            target_label: __metrics_path__
            regex: (.+)
          - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
            action: replace
            target_label: __address__
            regex: ([^:]+)(?::\d+)?;(\d+)
            replacement: $1:$2
          - action: labelmap
            regex: __meta_kubernetes_service_label_(.+)
          - source_labels: [__meta_kubernetes_namespace]
            action: replace
            target_label: kubernetes_namespace
          - source_labels: [__meta_kubernetes_service_name]
            action: replace
            target_label: kubernetes_name

      - job_name: 'prometheus-pushgateway'
        honor_labels: true

        kubernetes_sd_configs:
          - role: service

        relabel_configs:
          - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_probe]
            action: keep
            regex: pushgateway

      # Example scrape config for probing services via the Blackbox Exporter.
      #
      # The relabeling allows the actual service scrape endpoint to be configured
      # via the following annotations:
      #
      - job_name: 'kubernetes-services'

        metrics_path: /probe
        params:
          module: [http_2xx]

        kubernetes_sd_configs:
          - role: service

        relabel_configs:
          - source_labels: [__meta_kubernetes_service_annotation_prometheus_io_probe]
            action: keep
            regex: true
          - source_labels: [__address__]
            target_label: __param_target
          - target_label: __address__
            replacement: blackbox
          - source_labels: [__param_target]
            target_label: instance
          - action: labelmap
            regex: __meta_kubernetes_service_label_(.+)
          - source_labels: [__meta_kubernetes_namespace]
            target_label: kubernetes_namespace
          - source_labels: [__meta_kubernetes_service_name]
            target_label: kubernetes_name

      # Example scrape config for pods
      #
      # The relabeling allows the actual pod scrape endpoint to be configured via the
      # following annotations:
      - job_name: 'kubernetes-pods'

        kubernetes_sd_configs:
          - role: pod

        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
            action: keep
            regex: true
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
            action: replace
            target_label: __metrics_path__
            regex: (.+)
          - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
            action: replace
            regex: ([^:]+)(?::\d+)?;(\d+)
            replacement: $1:$2
            target_label: __address__
          - action: labelmap
            regex: __meta_kubernetes_pod_label_(.+)
          - source_labels: [__meta_kubernetes_namespace]
            action: replace
            target_label: kubernetes_namespace
          - source_labels: [__meta_kubernetes_pod_name]
            action: replace
            target_label: kubernetes_pod_name

networkPolicy:
  ## Enable creation of NetworkPolicy resources.
  ##
  enabled: false`

var GAPPrometheusTemplates = map[string]string{
	"alertmanager-configmap.yaml":                alertmanager_configmap,
	"alertmanager-deployment.yaml":               alertmanager_deployment,
	"alertmanager-ingress.yaml":                  alertmanager_ingress,
	"alertmanager-networkpolicy.yaml":            alertmanager_networkpolicy,
	"alertmanager-pvc.yaml":                      alertmanager_pvc,
	"alertmanager-serviceaccount.yaml":           alertmanager_serviceaccount,
	"alertmanager-service.yaml":                  alertmanager_service,
	"_helpers.tpl":                               prometheus_helpers,
	"kube-state-metrics-clusterrolebinding.yaml": kube_state_metrics_clusterrolebinding,
	"kube-stat-metrics_clusterrole.yaml":         kube_state_metrics_clusterrole,
	"kube-state-metrics-deployment.yaml":         kube_state_metrics_deployment,
	"kube-state-metrics-networkpolicy.yaml":      kube_state_metrics_networkpolicy,
	"kube-state-metrics-serviceaccount.yaml":     kube_state_metrics_serviceaccount,
	"kube-state-metrics-svc.yaml":                kube_state_metrics_svc,
	"node-exporter-daemonset.yaml":               node_exporter_daemonset,
	"node-exporter-serviceaccount.yaml":          node_exporter_serviceaccount,
	"node-exporter-service.yaml":                 node_exporter_service,
	"pushgateway-deployment.yaml":                pushgateway_deployment,
	"pushgateway-ingress.yaml":                   pushgateway_ingress,
	"pushgateway-serviceaccount.yaml":            pushgateway_serviceaccount,
	"pushgateway-service.yaml":                   pushgateway_service,
	"server-clusterrolebinding.yaml":             server_clusterrolebinding,
	"server-clusterrole.yaml":                    server_clusterrole,
	"server-configmap.yaml":                      server_configmap,
	"server-deployment.yaml":                     server_deployment,
	"server-ingress.yaml":                        server_ingress,
	"server-networkpolicy.yaml":                  server_networkpolicy,
	"server-pvc.yaml":                            server_pvc,
	"server-serviceaccount.yaml":                 server_serviceaccount,
	"server-service.yaml":                        server_service,
}

const alertmanager_configmap = `
{{- if and .Values.alertmanager.enabled (empty .Values.alertmanager.configMapOverrideName) -}}
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.alertmanager.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.alertmanager.fullname" . }}
data:
{{- $root := . -}}
{{- range $key, $value := .Values.alertmanagerFiles }}
  {{ $key }}: |
{{ toYaml $value | default "{}" | indent 4 }}
{{- end -}}
{{- end -}}`

const alertmanager_deployment = `
{{- if .Values.alertmanager.enabled -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.alertmanager.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.alertmanager.fullname" . }}
spec:
  replicas: {{ .Values.alertmanager.replicaCount }}
  {{- if .Values.server.strategy }}
  strategy:
{{ toYaml .Values.server.strategy | indent 4 }}
  {{- end }}
  selector:
    matchLabels:
      app: {{ template "prometheus.name" . }}
      component: "{{ .Values.alertmanager.name }}"
      release: {{ .Release.Name }}
  template:
    metadata:
    {{- if .Values.alertmanager.podAnnotations }}
      annotations:
{{ toYaml .Values.alertmanager.podAnnotations | indent 8 }}
    {{- end }}
      labels:
        app: {{ template "prometheus.name" . }}
        component: "{{ .Values.alertmanager.name }}"
        release: {{ .Release.Name }}
    spec:
{{- if .Values.alertmanager.affinity }}
      affinity:
{{ toYaml .Values.alertmanager.affinity | indent 8 }}
{{- end }}
{{- if .Values.alertmanager.schedulerName }}
      schedulerName: "{{ .Values.alertmanager.schedulerName }}"
{{- end }}
      serviceAccountName: {{ template "prometheus.serviceAccountName.alertmanager" . }}
{{- if .Values.alertmanager.priorityClassName }}
      priorityClassName: "{{ .Values.alertmanager.priorityClassName }}"
{{- end }}
      containers:
        - name: {{ template "prometheus.name" . }}-{{ .Values.alertmanager.name }}
          image: "{{ .Values.alertmanager.image.repository }}:{{ .Values.alertmanager.image.tag }}"
          imagePullPolicy: "{{ .Values.alertmanager.image.pullPolicy }}"
          env:
            {{- range $key, $value := .Values.alertmanager.extraEnv }}
            - name: {{ $key }}
              value: {{ $value }}
            {{- end }}
          args:
            - --config.file=/etc/config/alertmanager.yml
            - --storage.path={{ .Values.alertmanager.persistentVolume.mountPath }}
          {{- range $key, $value := .Values.alertmanager.extraArgs }}
            - --{{ $key }}={{ $value }}
          {{- end }}
          {{- if .Values.alertmanager.baseURL }}
            - --web.external-url={{ .Values.alertmanager.baseURL }}
          {{- end }}

          ports:
            - containerPort: 9093
          readinessProbe:
            httpGet:
              path: {{ .Values.alertmanager.prefixURL }}/#/status
              port: 9093
            initialDelaySeconds: 30
            timeoutSeconds: 30
          resources:
{{ toYaml .Values.alertmanager.resources | indent 12 }}
          volumeMounts:
            - name: config-volume
              mountPath: /etc/config
            - name: storage-volume
              mountPath: "{{ .Values.alertmanager.persistentVolume.mountPath }}"
              subPath: "{{ .Values.alertmanager.persistentVolume.subPath }}"

        - name: {{ template "prometheus.name" . }}-{{ .Values.alertmanager.name }}-{{ .Values.configmapReload.name }}
          image: "{{ .Values.configmapReload.image.repository }}:{{ .Values.configmapReload.image.tag }}"
          imagePullPolicy: "{{ .Values.configmapReload.image.pullPolicy }}"
          args:
            - --volume-dir=/etc/config
            - --webhook-url=http://localhost:9093{{ .Values.alertmanager.prefixURL }}/-/reload
          resources:
{{ toYaml .Values.configmapReload.resources | indent 12 }}
          volumeMounts:
            - name: config-volume
              mountPath: /etc/config
              readOnly: true
    {{- if .Values.alertmanager.nodeSelector }}
      nodeSelector:
{{ toYaml .Values.alertmanager.nodeSelector | indent 8 }}
    {{- end }}
    {{- if .Values.alertmanager.securityContext }}
      securityContext:
{{ toYaml .Values.alertmanager.securityContext | indent 8 }}
    {{- end }}
    {{- if .Values.alertmanager.tolerations }}
      tolerations:
{{ toYaml .Values.alertmanager.tolerations | indent 8 }}
    {{- end }}
    {{- if .Values.alertmanager.affinity }}
      affinity:
{{ toYaml .Values.alertmanager.affinity | indent 8 }}
    {{- end }}
      volumes:
        - name: config-volume
          configMap:
            name: {{ if .Values.alertmanager.configMapOverrideName }}{{ .Release.Name }}-{{ .Values.alertmanager.configMapOverrideName }}{{- else }}{{ template "prometheus.alertmanager.fullname" . }}{{- end }}
        - name: storage-volume
        {{- if .Values.alertmanager.persistentVolume.enabled }}
          persistentVolumeClaim:
            claimName: {{ if .Values.alertmanager.persistentVolume.existingClaim }}{{ .Values.alertmanager.persistentVolume.existingClaim }}{{- else }}{{ template "prometheus.alertmanager.fullname" . }}{{- end }}
        {{- else }}
          emptyDir: {}
        {{- end -}}
{{- end }}`

const alertmanager_ingress = `
{{- if and .Values.alertmanager.enabled .Values.alertmanager.ingress.enabled -}}
{{- $releaseName := .Release.Name -}}
{{- $serviceName := include "prometheus.alertmanager.fullname" . }}
{{- $servicePort := .Values.alertmanager.service.servicePort -}}
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
{{- if .Values.alertmanager.ingress.annotations }}
  annotations:
{{ toYaml .Values.alertmanager.ingress.annotations | indent 4 }}
{{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.alertmanager.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
{{- range $key, $value := .Values.alertmanager.ingress.extraLabels }}
    {{ $key }}: {{ $value }}
{{- end }}
  name: {{ template "prometheus.alertmanager.fullname" . }}
spec:
  rules:
  {{- range .Values.alertmanager.ingress.hosts }}
    {{- $url := splitList "/" . }}
    - host: {{ first $url }}
      http:
        paths:
          - path: /{{ rest $url | join "/" }}
            backend:
              serviceName: {{ $serviceName }}
              servicePort: {{ $servicePort }}
  {{- end -}}
{{- if .Values.alertmanager.ingress.tls }}
  tls:
{{ toYaml .Values.alertmanager.ingress.tls | indent 4 }}
  {{- end -}}
{{- end -}}`

const alertmanager_networkpolicy = `
{{- if .Values.networkPolicy.enabled }}
apiVersion: {{ template "prometheus.networkPolicy.apiVersion" . }}
kind: NetworkPolicy
metadata:
  name: {{ template "prometheus.alertmanager.fullname" . }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.alertmanager.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
spec:
  podSelector:
    matchLabels:
      app: {{ template "prometheus.name" . }}
      component: "{{ .Values.alertmanager.name }}"
      release: {{ .Release.Name }}
  ingress:
    - from:
      - podSelector:
          matchLabels:
            release: {{ .Release.Name }}
            component: "{{ .Values.server.name }}"
    - ports:
      - port: 9093
{{- end }}`

const alertmanager_pvc = `
{{- if and .Values.alertmanager.enabled .Values.alertmanager.persistentVolume.enabled -}}
{{- if not .Values.alertmanager.persistentVolume.existingClaim -}}
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  {{- if .Values.alertmanager.persistentVolume.annotations }}
  annotations:
{{ toYaml .Values.alertmanager.persistentVolume.annotations | indent 4 }}
  {{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.alertmanager.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.alertmanager.fullname" . }}
spec:
  accessModes:
{{ toYaml .Values.alertmanager.persistentVolume.accessModes | indent 4 }}
{{- if .Values.alertmanager.persistentVolume.storageClass }}
{{- if (eq "-" .Values.alertmanager.persistentVolume.storageClass) }}
  storageClassName: ""
{{- else }}
  storageClassName: "{{ .Values.alertmanager.persistentVolume.storageClass }}"
{{- end }}
{{- end }}
  resources:
    requests:
      storage: "{{ .Values.alertmanager.persistentVolume.size }}"
{{- end -}}
{{- end -}}`

const alertmanager_serviceaccount = `
{{- if .Values.serviceAccounts.alertmanager.create }}
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.alertmanager.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.serviceAccountName.alertmanager" . }}
{{- end }}`

const alertmanager_service = `
{{- if .Values.alertmanager.enabled -}}
apiVersion: v1
kind: Service
metadata:
{{- if .Values.alertmanager.service.annotations }}
  annotations:
{{ toYaml .Values.alertmanager.service.annotations | indent 4 }}
{{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.alertmanager.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
{{- if .Values.alertmanager.service.labels }}
{{ toYaml .Values.alertmanager.service.labels | indent 4 }}
{{- end }}
  name: {{ template "prometheus.alertmanager.fullname" . }}
spec:
{{- if .Values.alertmanager.service.clusterIP }}
  clusterIP: {{ .Values.alertmanager.service.clusterIP }}
{{- end }}
{{- if .Values.alertmanager.service.externalIPs }}
  externalIPs:
{{ toYaml .Values.alertmanager.service.externalIPs | indent 4 }}
{{- end }}
{{- if .Values.alertmanager.service.loadBalancerIP }}
  loadBalancerIP: {{ .Values.alertmanager.service.loadBalancerIP }}
{{- end }}
{{- if .Values.alertmanager.service.loadBalancerSourceRanges }}
  loadBalancerSourceRanges:
  {{- range $cidr := .Values.alertmanager.service.loadBalancerSourceRanges }}
    - {{ $cidr }}
  {{- end }}
{{- end }}
  ports:
    - name: http
      port: {{ .Values.alertmanager.service.servicePort }}
      protocol: TCP
      targetPort: 9093
    {{- if .Values.alertmanager.service.nodePort }}
      nodePort: {{ .Values.alertmanager.service.nodePort }}
    {{- end }}
{{- if .Values.alertmanager.service.enableMeshPeer }}
    - name: meshpeer
      port: 6783
      protocol: TCP
      targetPort: 6783
{{- end }}
  selector:
    app: {{ template "prometheus.name" . }}
    component: "{{ .Values.alertmanager.name }}"
    release: {{ .Release.Name }}
  type: "{{ .Values.alertmanager.service.type }}"
{{- end }}`

const prometheus_helpers = `
{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "prometheus.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "prometheus.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create a fully qualified alertmanager name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}

{{- define "prometheus.alertmanager.fullname" -}}
{{- if .Values.alertmanager.fullnameOverride -}}
{{- .Values.alertmanager.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- printf "%s-%s" .Release.Name .Values.alertmanager.name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s-%s" .Release.Name $name .Values.alertmanager.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create a fully qualified kube-state-metrics name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "prometheus.kubeStateMetrics.fullname" -}}
{{- if .Values.kubeStateMetrics.fullnameOverride -}}
{{- .Values.kubeStateMetrics.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- printf "%s-%s" .Release.Name .Values.kubeStateMetrics.name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s-%s" .Release.Name $name .Values.kubeStateMetrics.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create a fully qualified node-exporter name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "prometheus.nodeExporter.fullname" -}}
{{- if .Values.nodeExporter.fullnameOverride -}}
{{- .Values.nodeExporter.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- printf "%s-%s" .Release.Name .Values.nodeExporter.name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s-%s" .Release.Name $name .Values.nodeExporter.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create a fully qualified Prometheus server name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "prometheus.server.fullname" -}}
{{- if .Values.server.fullnameOverride -}}
{{- .Values.server.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- printf "%s-%s" .Release.Name .Values.server.name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s-%s" .Release.Name $name .Values.server.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create a fully qualified pushgateway name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "prometheus.pushgateway.fullname" -}}
{{- if .Values.pushgateway.fullnameOverride -}}
{{- .Values.pushgateway.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- printf "%s-%s" .Release.Name .Values.pushgateway.name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s-%s" .Release.Name $name .Values.pushgateway.name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Return the appropriate apiVersion for networkpolicy.
*/}}
{{- define "prometheus.networkPolicy.apiVersion" -}}
{{- if semverCompare ">=1.4-0, <1.7-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "extensions/v1beta1" -}}
{{- else if semverCompare "^1.7-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "networking.k8s.io/v1" -}}
{{- end -}}
{{- end -}}

{{/*
Create the name of the service account to use for the alertmanager component
*/}}
{{- define "prometheus.serviceAccountName.alertmanager" -}}
{{- if .Values.serviceAccounts.alertmanager.create -}}
    {{ default (include "prometheus.alertmanager.fullname" .) .Values.serviceAccounts.alertmanager.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccounts.alertmanager.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the service account to use for the kubeStateMetrics component
*/}}
{{- define "prometheus.serviceAccountName.kubeStateMetrics" -}}
{{- if .Values.serviceAccounts.kubeStateMetrics.create -}}
    {{ default (include "prometheus.kubeStateMetrics.fullname" .) .Values.serviceAccounts.kubeStateMetrics.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccounts.kubeStateMetrics.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the service account to use for the nodeExporter component
*/}}
{{- define "prometheus.serviceAccountName.nodeExporter" -}}
{{- if .Values.serviceAccounts.nodeExporter.create -}}
    {{ default (include "prometheus.nodeExporter.fullname" .) .Values.serviceAccounts.nodeExporter.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccounts.nodeExporter.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the service account to use for the pushgateway component
*/}}
{{- define "prometheus.serviceAccountName.pushgateway" -}}
{{- if .Values.serviceAccounts.pushgateway.create -}}
    {{ default (include "prometheus.pushgateway.fullname" .) .Values.serviceAccounts.pushgateway.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccounts.pushgateway.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the service account to use for the server component
*/}}
{{- define "prometheus.serviceAccountName.server" -}}
{{- if .Values.serviceAccounts.server.create -}}
    {{ default (include "prometheus.server.fullname" .) .Values.serviceAccounts.server.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccounts.server.name }}
{{- end -}}
{{- end -}}`

const kube_state_metrics_clusterrolebinding = `
{{- if .Values.rbac.create }}
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.kubeStateMetrics.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.kubeStateMetrics.fullname" . }}
subjects:
  - kind: ServiceAccount
    name: {{ template "prometheus.serviceAccountName.kubeStateMetrics" . }}
    namespace: {{ .Release.Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ template "prometheus.kubeStateMetrics.fullname" . }}
{{- end }}`

const kube_state_metrics_clusterrole = `
{{- if .Values.rbac.create }}
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.kubeStateMetrics.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.kubeStateMetrics.fullname" . }}
rules:
  - apiGroups:
      - ""
    resources:
      - namespaces
      - nodes
      - persistentvolumeclaims
      - pods
      - services
      - resourcequotas
      - replicationcontrollers
      - limitranges
      - persistentvolumeclaims
      - persistentvolumes
      - endpoints
      - secrets
      - configmaps
    verbs:
      - list
      - watch
  - apiGroups:
      - extensions
    resources:
      - daemonsets
      - deployments
      - replicasets
    verbs:
      - list
      - watch
  - apiGroups:
      - apps
    resources:
      - statefulsets
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - batch
    resources:
      - cronjobs
      - jobs
    verbs:
      - list
      - watch
  - apiGroups:
      - autoscaling
    resources:
      - horizontalpodautoscalers
    verbs:
      - list
      - watch
{{- end }}`

const kube_state_metrics_deployment = `
{{- if .Values.kubeStateMetrics.enabled -}}
apiVersion: apps/v1
kind: Deployment
metadata:
{{- if .Values.kubeStateMetrics.deploymentAnnotations }}
  annotations:
{{ toYaml .Values.kubeStateMetrics.deploymentAnnotations | indent 4 }}
{{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.kubeStateMetrics.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.kubeStateMetrics.fullname" . }}
spec:
  replicas: {{ .Values.kubeStateMetrics.replicaCount }}
  selector:
    matchLabels:
      app: {{ template "prometheus.name" . }}
      component: "{{ .Values.kubeStateMetrics.name }}"
      release: {{ .Release.Name }}
{{- if .Values.kubeStateMetrics.pod.labels }}
{{ toYaml .Values.kubeStateMetrics.pod.labels | indent 6 }}
{{- end }}
  template:
    metadata:
    {{- if .Values.kubeStateMetrics.podAnnotations }}
      annotations:
{{ toYaml .Values.kubeStateMetrics.podAnnotations | indent 8 }}
    {{- end }}
      labels:
        app: {{ template "prometheus.name" . }}
        component: "{{ .Values.kubeStateMetrics.name }}"
        release: {{ .Release.Name }}
{{- if .Values.kubeStateMetrics.pod.labels }}
{{ toYaml .Values.kubeStateMetrics.pod.labels | indent 8 }}
{{- end }}
    spec:
      serviceAccountName: {{ template "prometheus.serviceAccountName.kubeStateMetrics" . }}
{{- if .Values.kubeStateMetrics.priorityClassName }}
      priorityClassName: "{{ .Values.kubeStateMetrics.priorityClassName }}"
{{- end }}
      containers:
        - name: {{ template "prometheus.name" . }}-{{ .Values.kubeStateMetrics.name }}
          image: "{{ .Values.kubeStateMetrics.image.repository }}:{{ .Values.kubeStateMetrics.image.tag }}"
          imagePullPolicy: "{{ .Values.kubeStateMetrics.image.pullPolicy }}"
        {{- if .Values.kubeStateMetrics.args }}
          args:
          {{- range $key, $value := .Values.kubeStateMetrics.args }}
            - --{{ $key }}={{ $value }}
          {{- end }}
        {{- end }}
          ports:
            - name: metrics
              containerPort: 8080
          resources:
{{ toYaml .Values.kubeStateMetrics.resources | indent 12 }}
    {{- if .Values.kubeStateMetrics.nodeSelector }}
      nodeSelector:
{{ toYaml .Values.kubeStateMetrics.nodeSelector | indent 8 }}
    {{- end }}
    {{- if .Values.kubeStateMetrics.securityContext }}
      securityContext:
{{ toYaml .Values.kubeStateMetrics.securityContext | indent 8 }}
    {{- end }}
    {{- if .Values.kubeStateMetrics.tolerations }}
      tolerations:
{{ toYaml .Values.kubeStateMetrics.tolerations | indent 8 }}
    {{- end }}
    {{- if .Values.kubeStateMetrics.affinity }}
      affinity:
{{ toYaml .Values.kubeStateMetrics.affinity | indent 8 }}
    {{- end }}
{{- end }}`

const kube_state_metrics_networkpolicy = `
{{- if .Values.networkPolicy.enabled }}
apiVersion: {{ template "prometheus.networkPolicy.apiVersion" . }}
kind: NetworkPolicy
metadata:
  name: {{ template "prometheus.kubeStateMetrics.fullname" . }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.kubeStateMetrics.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
spec:
  podSelector:
    matchLabels:
      app: {{ template "prometheus.name" . }}
      component: "{{ .Values.kubeStateMetrics.name }}"
      release: {{ .Release.Name }}
  ingress:
  - from:
    - podSelector:
        matchLabels:
          release: {{ .Release.Name }}
          component: "{{ .Values.server.name }}"
  - ports:
    - port: 8080
{{- end }}`

const kube_state_metrics_serviceaccount = `
{{- if .Values.serviceAccounts.kubeStateMetrics.create }}
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.kubeStateMetrics.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.serviceAccountName.kubeStateMetrics" . }}
{{- end }}`

const kube_state_metrics_svc = `
{{- if .Values.kubeStateMetrics.enabled -}}
apiVersion: v1
kind: Service
metadata:
{{- if .Values.kubeStateMetrics.service.annotations }}
  annotations:
{{ toYaml .Values.kubeStateMetrics.service.annotations | indent 4 }}
{{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.kubeStateMetrics.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
{{- if .Values.kubeStateMetrics.service.labels }}
{{ toYaml .Values.kubeStateMetrics.service.labels | indent 4 }}
{{- end }}
  name: {{ template "prometheus.kubeStateMetrics.fullname" . }}
spec:
{{- if .Values.kubeStateMetrics.service.clusterIP }}
  clusterIP: {{ .Values.kubeStateMetrics.service.clusterIP }}
{{- end }}
{{- if .Values.kubeStateMetrics.service.externalIPs }}
  externalIPs:
{{ toYaml .Values.kubeStateMetrics.service.externalIPs | indent 4 }}
{{- end }}
{{- if .Values.kubeStateMetrics.service.loadBalancerIP }}
  loadBalancerIP: {{ .Values.kubeStateMetrics.service.loadBalancerIP }}
{{- end }}
{{- if .Values.kubeStateMetrics.service.loadBalancerSourceRanges }}
  loadBalancerSourceRanges:
  {{- range $cidr := .Values.kubeStateMetrics.service.loadBalancerSourceRanges }}
    - {{ $cidr }}
  {{- end }}
{{- end }}
  ports:
    - name: http
      port: {{ .Values.kubeStateMetrics.service.servicePort }}
      protocol: TCP
      targetPort: 8080
  selector:
    app: {{ template "prometheus.name" . }}
    component: "{{ .Values.kubeStateMetrics.name }}"
    release: {{ .Release.Name }}
  type: "{{ .Values.kubeStateMetrics.service.type }}"
{{- end }}`

const node_exporter_daemonset = `
{{- if .Values.nodeExporter.enabled -}}
apiVersion: apps/v1
kind: DaemonSet
metadata:
{{- if .Values.nodeExporter.deploymentAnnotations }}
  annotations:
{{ toYaml .Values.nodeExporter.deploymentAnnotations | indent 4 }}
{{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.nodeExporter.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.nodeExporter.fullname" . }}
spec:
  {{- if .Values.nodeExporter.updateStrategy }}
  updateStrategy:
{{ toYaml .Values.nodeExporter.updateStrategy | indent 4 }}
  {{- end }}
  selector:
    matchLabels:
      app: {{ template "prometheus.name" . }}
      component: "{{ .Values.nodeExporter.name }}"
      release: {{ .Release.Name }}
{{- if .Values.nodeExporter.pod.labels }}
{{ toYaml .Values.nodeExporter.pod.labels | indent 6 }}
{{- end }}
  template:
    metadata:
    {{- if .Values.nodeExporter.podAnnotations }}
      annotations:
{{ toYaml .Values.nodeExporter.podAnnotations | indent 8 }}
    {{- end }}
      labels:
        app: {{ template "prometheus.name" . }}
        component: "{{ .Values.nodeExporter.name }}"
        release: {{ .Release.Name }}
{{- if .Values.nodeExporter.pod.labels }}
{{ toYaml .Values.nodeExporter.pod.labels | indent 8 }}
{{- end }}
    spec:
      serviceAccountName: {{ template "prometheus.serviceAccountName.nodeExporter" . }}
{{- if .Values.nodeExporter.priorityClassName }}
      priorityClassName: "{{ .Values.nodeExporter.priorityClassName }}"
{{- end }}
      containers:
        - name: {{ template "prometheus.name" . }}-{{ .Values.nodeExporter.name }}
          image: "{{ .Values.nodeExporter.image.repository }}:{{ .Values.nodeExporter.image.tag }}"
          imagePullPolicy: "{{ .Values.nodeExporter.image.pullPolicy }}"
          args:
            - --path.procfs=/host/proc
            - --path.sysfs=/host/sys
          {{- range $key, $value := .Values.nodeExporter.extraArgs }}
          {{- if $value }}
            - --{{ $key }}={{ $value }}
          {{- else }}
            - --{{ $key }}
          {{- end }}
          {{- end }}
          ports:
            - name: metrics
              containerPort: 9100
              hostPort: {{ .Values.nodeExporter.service.hostPort }}
          resources:
{{ toYaml .Values.nodeExporter.resources | indent 12 }}
          volumeMounts:
            - name: proc
              mountPath: /host/proc
              readOnly:  true
            - name: sys
              mountPath: /host/sys
              readOnly: true
          {{- range .Values.nodeExporter.extraHostPathMounts }}
            - name: {{ .name }}
              mountPath: {{ .mountPath }}
              readOnly: {{ .readOnly }}
          {{- end }}
          {{- range .Values.nodeExporter.extraConfigmapMounts }}
            - name: {{ .name }}
              mountPath: {{ .mountPath }}
              readOnly: {{ .readOnly }}
          {{- end }}
    {{- if .Values.nodeExporter.hostNetwork }}
      hostNetwork: true
    {{- end }}
    {{- if .Values.nodeExporter.hostPID }}
      hostPID: true
    {{- end }}
    {{- if .Values.nodeExporter.tolerations }}
      tolerations:
{{ toYaml .Values.nodeExporter.tolerations | indent 8 }}
    {{- end }}
    {{- if .Values.nodeExporter.nodeSelector }}
      nodeSelector:
{{ toYaml .Values.nodeExporter.nodeSelector | indent 8 }}
    {{- end }}
    {{- if .Values.nodeExporter.securityContext }}
      securityContext:
{{ toYaml .Values.nodeExporter.securityContext | indent 8 }}
    {{- end }}
      volumes:
        - name: proc
          hostPath:
            path: /proc
        - name: sys
          hostPath:
            path: /sys
      {{- range .Values.nodeExporter.extraHostPathMounts }}
        - name: {{ .name }}
          hostPath:
            path: {{ .hostPath }}
      {{- end }}
      {{- range .Values.nodeExporter.extraConfigmapMounts }}
        - name: {{ .name }}
          configMap:
            name: {{ .configMap }}
      {{- end }}

{{- end -}}`

const node_exporter_serviceaccount = `
{{- if .Values.serviceAccounts.nodeExporter.create }}
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.nodeExporter.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.serviceAccountName.nodeExporter" . }}
{{- end }}`

const node_exporter_service = `
{{- if .Values.nodeExporter.enabled -}}
apiVersion: v1
kind: Service
metadata:
{{- if .Values.nodeExporter.service.annotations }}
  annotations:
{{ toYaml .Values.nodeExporter.service.annotations | indent 4 }}
{{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.nodeExporter.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
{{- if .Values.nodeExporter.service.labels }}
{{ toYaml .Values.nodeExporter.service.labels | indent 4 }}
{{- end }}
  name: {{ template "prometheus.nodeExporter.fullname" . }}
spec:
{{- if .Values.nodeExporter.service.clusterIP }}
  clusterIP: {{ .Values.nodeExporter.service.clusterIP }}
{{- end }}
{{- if .Values.nodeExporter.service.externalIPs }}
  externalIPs:
{{ toYaml .Values.nodeExporter.service.externalIPs | indent 4 }}
{{- end }}
{{- if .Values.nodeExporter.service.loadBalancerIP }}
  loadBalancerIP: {{ .Values.nodeExporter.service.loadBalancerIP }}
{{- end }}
{{- if .Values.nodeExporter.service.loadBalancerSourceRanges }}
  loadBalancerSourceRanges:
  {{- range $cidr := .Values.nodeExporter.service.loadBalancerSourceRanges }}
    - {{ $cidr }}
  {{- end }}
{{- end }}
  ports:
    - name: metrics
      port: {{ .Values.nodeExporter.service.servicePort }}
      protocol: TCP
      targetPort: 9100
  selector:
    app: {{ template "prometheus.name" . }}
    component: "{{ .Values.nodeExporter.name }}"
    release: {{ .Release.Name }}
  type: "{{ .Values.nodeExporter.service.type }}"
{{- end -}}`

const pushgateway_deployment = `
{{- if .Values.pushgateway.enabled -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.pushgateway.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.pushgateway.fullname" . }}
spec:
  replicas: {{ .Values.pushgateway.replicaCount }}
  selector:
    matchLabels:
      app: {{ template "prometheus.name" . }}
      component: "{{ .Values.pushgateway.name }}"
      release: {{ .Release.Name }}
  template:
    metadata:
    {{- if .Values.pushgateway.podAnnotations }}
      annotations:
{{ toYaml .Values.pushgateway.podAnnotations | indent 8 }}
    {{- end }}
      labels:
        app: {{ template "prometheus.name" . }}
        component: "{{ .Values.pushgateway.name }}"
        release: {{ .Release.Name }}
    spec:
      serviceAccountName: {{ template "prometheus.serviceAccountName.pushgateway" . }}
{{- if .Values.pushgateway.priorityClassName }}
      priorityClassName: "{{ .Values.pushgateway.priorityClassName }}"
{{- end }}
      containers:
        - name: {{ template "prometheus.name" . }}-{{ .Values.pushgateway.name }}
          image: "{{ .Values.pushgateway.image.repository }}:{{ .Values.pushgateway.image.tag }}"
          imagePullPolicy: "{{ .Values.pushgateway.image.pullPolicy }}"
          args:
          {{- range $key, $value := .Values.pushgateway.extraArgs }}
            - --{{ $key }}={{ $value }}
          {{- end }}
          ports:
            - containerPort: 9091
          readinessProbe:
            httpGet:
            {{- if (index .Values "pushgateway" "extraArgs" "web.route-prefix") }}
              path: /{{ index .Values "pushgateway" "extraArgs" "web.route-prefix" }}/#/status
            {{- else }}
              path: /#/status
            {{- end }}
              port: 9091
            initialDelaySeconds: 10
            timeoutSeconds: 10
          resources:
{{ toYaml .Values.pushgateway.resources | indent 12 }}
    {{- if .Values.pushgateway.nodeSelector }}
      nodeSelector:
{{ toYaml .Values.pushgateway.nodeSelector | indent 8 }}
    {{- end }}
    {{- if .Values.pushgateway.securityContext }}
      securityContext:
{{ toYaml .Values.pushgateway.securityContext | indent 8 }}
    {{- end }}
    {{- if .Values.pushgateway.tolerations }}
      tolerations:
{{ toYaml .Values.pushgateway.tolerations | indent 8 }}
    {{- end }}
    {{- if .Values.pushgateway.affinity }}
      affinity:
{{ toYaml .Values.pushgateway.affinity | indent 8 }}
    {{- end }}
{{- end }}`

const pushgateway_ingress = `
{{- if and .Values.pushgateway.enabled .Values.pushgateway.ingress.enabled -}}
{{- $releaseName := .Release.Name -}}
{{- $serviceName := include "prometheus.pushgateway.fullname" . }}
{{- $servicePort := .Values.pushgateway.service.servicePort -}}
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
{{- if .Values.pushgateway.ingress.annotations }}
  annotations:
{{ toYaml .Values.pushgateway.ingress.annotations | indent 4}}
{{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.pushgateway.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.pushgateway.fullname" . }}
spec:
  rules:
  {{- range .Values.pushgateway.ingress.hosts }}
    {{- $url := splitList "/" . }}
    - host: {{ first $url }}
      http:
        paths:
          - path: /{{ rest $url | join "/" }}
            backend:
              serviceName: {{ $serviceName }}
              servicePort: {{ $servicePort }}
  {{- end -}}
{{- if .Values.pushgateway.ingress.tls }}
  tls:
{{ toYaml .Values.pushgateway.ingress.tls | indent 4 }}
  {{- end -}}
{{- end -}}`

const pushgateway_serviceaccount = `
{{- if .Values.serviceAccounts.pushgateway.create }}
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.pushgateway.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.serviceAccountName.pushgateway" . }}
{{- end }}`

const pushgateway_service = `
{{- if .Values.pushgateway.enabled -}}
apiVersion: v1
kind: Service
metadata:
{{- if .Values.pushgateway.service.annotations }}
  annotations:
{{ toYaml .Values.pushgateway.service.annotations | indent 4}}
{{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.pushgateway.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
{{- if .Values.pushgateway.service.labels }}
{{ toYaml .Values.pushgateway.service.labels | indent 4}}
{{- end }}
  name: {{ template "prometheus.pushgateway.fullname" . }}
spec:
{{- if .Values.pushgateway.service.clusterIP }}
  clusterIP: {{ .Values.pushgateway.service.clusterIP }}
{{- end }}
{{- if .Values.pushgateway.service.externalIPs }}
  externalIPs:
{{ toYaml .Values.pushgateway.service.externalIPs | indent 4 }}
{{- end }}
{{- if .Values.pushgateway.service.loadBalancerIP }}
  loadBalancerIP: {{ .Values.pushgateway.service.loadBalancerIP }}
{{- end }}
{{- if .Values.pushgateway.service.loadBalancerSourceRanges }}
  loadBalancerSourceRanges:
  {{- range $cidr := .Values.pushgateway.service.loadBalancerSourceRanges }}
    - {{ $cidr }}
  {{- end }}
{{- end }}
  ports:
    - name: http
      port: {{ .Values.pushgateway.service.servicePort }}
      protocol: TCP
      targetPort: 9091
  selector:
    app: {{ template "prometheus.name" . }}
    component: "{{ .Values.pushgateway.name }}"
    release: {{ .Release.Name }}
  type: "{{ .Values.pushgateway.service.type }}"
{{- end }}`

const server_clusterrolebinding = `
{{- if .Values.rbac.create }}
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.server.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.server.fullname" . }}
subjects:
  - kind: ServiceAccount
    name: {{ template "prometheus.serviceAccountName.server" . }}
    namespace: {{ .Release.Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{ template "prometheus.server.fullname" . }}
{{- end }}`

const server_clusterrole = `
{{- if .Values.rbac.create }}
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.server.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.server.fullname" . }}
rules:
  - apiGroups:
      - ""
    resources:
      - nodes
      - nodes/proxy
      - services
      - endpoints
      - pods
      - ingresses
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - ""
    resources:
      - configmaps
    verbs:
      - get
  - apiGroups:
      - "extensions"
    resources:
      - ingresses/status
      - ingresses
    verbs:
      - get
      - list
      - watch
  - nonResourceURLs:
      - "/metrics"
    verbs:
      - get
{{- end }}`

const server_configmap = `
{{- if (empty .Values.server.configMapOverrideName) -}}
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.server.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.server.fullname" . }}
data:
{{- $root := . -}}
{{- range $key, $value := .Values.serverFiles }}
  {{ $key }}: |
{{- if eq $key "prometheus.yml" }}
    global:
{{ $root.Values.server.global | toYaml | indent 6 }}
{{- end }}
{{ toYaml $value | default "{}" | indent 4 }}
{{- if eq $key "prometheus.yml" -}}
{{- if $root.Values.alertmanager.enabled }}
    alerting:
      alertmanagers:
      - kubernetes_sd_configs:
          - role: pod
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
        bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
        {{- if $root.Values.alertmanager.prefixURL }}
        path_prefix: {{ $root.Values.alertmanager.prefixURL }}
        {{- end }}
        relabel_configs:
        - source_labels: [__meta_kubernetes_namespace]
          regex: {{ $root.Release.Namespace }}
          action: keep
        - source_labels: [__meta_kubernetes_pod_label_app]
          regex: {{ template "prometheus.name" $root }}
          action: keep
        - source_labels: [__meta_kubernetes_pod_label_component]
          regex: alertmanager
          action: keep
        - source_labels: [__meta_kubernetes_pod_container_port_number]
          regex:
          action: drop
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}`

const server_deployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
{{- if .Values.server.deploymentAnnotations }}
  annotations:
{{ toYaml .Values.server.deploymentAnnotations | indent 4 }}
{{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.server.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.server.fullname" . }}
spec:
  replicas: {{ .Values.server.replicaCount }}
  {{- if .Values.server.strategy }}
  strategy:
{{ toYaml .Values.server.strategy | indent 4 }}
  {{- end }}
  selector:
    matchLabels:
      app: {{ template "prometheus.name" . }}
      component: "{{ .Values.server.name }}"
      release: {{ .Release.Name }}
  template:
    metadata:
    {{- if .Values.server.podAnnotations }}
      annotations:
{{ toYaml .Values.server.podAnnotations | indent 8 }}
    {{- end }}
      labels:
        app: {{ template "prometheus.name" . }}
        component: "{{ .Values.server.name }}"
        release: {{ .Release.Name }}
    spec:
{{- if .Values.server.affinity }}
      affinity:
{{ toYaml .Values.server.affinity | indent 8 }}
{{- end }}
{{- if .Values.server.priorityClassName }}
      priorityClassName: "{{ .Values.server.priorityClassName }}"
{{- end }}
{{- if .Values.server.schedulerName }}
      schedulerName: "{{ .Values.server.schedulerName }}"
{{- end }}
      serviceAccountName: {{ template "prometheus.serviceAccountName.server" . }}
      {{- if .Values.initChownData.enabled }}
      initContainers:
      - name: "{{ .Values.initChownData.name }}"
        image: "{{ .Values.initChownData.image.repository }}:{{ .Values.initChownData.image.tag }}"
        imagePullPolicy: "{{ .Values.initChownData.image.pullPolicy }}"
        resources:
{{ toYaml .Values.initChownData.resources | indent 12 }}
        # 65534 is the nobody user that prometheus uses.
        command: ["chown", "-R", "65534:65534", "{{ .Values.server.persistentVolume.mountPath }}"]
        volumeMounts:
        - name: storage-volume
          mountPath: {{ .Values.server.persistentVolume.mountPath }}
          subPath: "{{ .Values.server.persistentVolume.subPath }}"
      {{- end }}
      containers:
        - name: {{ template "prometheus.name" . }}-{{ .Values.server.name }}-{{ .Values.configmapReload.name }}
          image: "{{ .Values.configmapReload.image.repository }}:{{ .Values.configmapReload.image.tag }}"
          imagePullPolicy: "{{ .Values.configmapReload.image.pullPolicy }}"
          args:
            - --volume-dir=/etc/config
            - --webhook-url=http://127.0.0.1:9090{{ .Values.server.prefixURL }}/-/reload
          {{- range $key, $value := .Values.configmapReload.extraArgs }}
            - --{{ $key }}={{ $value }}
          {{- end }}
          resources:
{{ toYaml .Values.configmapReload.resources | indent 12 }}
          volumeMounts:
            - name: config-volume
              mountPath: /etc/config
              readOnly: true
          {{- range .Values.configmapReload.extraConfigmapMounts }}
            - name: {{ $.Values.configmapReload.name }}-{{ .name }}
              mountPath: {{ .mountPath }}
              readOnly: {{ .readOnly }}
          {{- end }}

        - name: {{ template "prometheus.name" . }}-{{ .Values.server.name }}
          image: "{{ .Values.server.image.repository }}:{{ .Values.server.image.tag }}"
          imagePullPolicy: "{{ .Values.server.image.pullPolicy }}"
          args:
          {{- if .Values.server.retention }}
            - --storage.tsdb.retention={{ .Values.server.retention }}
          {{- end }}
            - --config.file=/etc/config/prometheus.yml
            - --storage.tsdb.path={{ .Values.server.persistentVolume.mountPath }}
            - --web.console.libraries=/etc/prometheus/console_libraries
            - --web.console.templates=/etc/prometheus/consoles
            - --web.enable-lifecycle
          {{- range $key, $value := .Values.server.extraArgs }}
            - --{{ $key }}={{ $value }}
          {{- end }}
          {{- if .Values.server.baseURL }}
            - --web.external-url={{ .Values.server.baseURL }}
          {{- end }}
          {{- if .Values.server.enableAdminApi }}
            - --web.enable-admin-api
          {{- end }}
          ports:
            - containerPort: 9090
          readinessProbe:
            httpGet:
              path: {{ .Values.server.prefixURL }}/-/ready
              port: 9090
            initialDelaySeconds: 30
            timeoutSeconds: 30
          livenessProbe:
            httpGet:
              path: {{ .Values.server.prefixURL }}/-/healthy
              port: 9090
            initialDelaySeconds: 30
            timeoutSeconds: 30
          resources:
{{ toYaml .Values.server.resources | indent 12 }}
          volumeMounts:
            - name: config-volume
              mountPath: /etc/config
            - name: storage-volume
              mountPath: {{ .Values.server.persistentVolume.mountPath }}
              subPath: "{{ .Values.server.persistentVolume.subPath }}"
          {{- range .Values.server.extraHostPathMounts }}
            - name: {{ .name }}
              mountPath: {{ .mountPath }}
              readOnly: {{ .readOnly }}
          {{- end }}
          {{- range .Values.server.extraConfigmapMounts }}
            - name: {{ $.Values.server.name }}-{{ .name }}
              mountPath: {{ .mountPath }}
              readOnly: {{ .readOnly }}
          {{- end }}
          {{- range .Values.server.extraSecretMounts }}
            - name: {{ .name }}
              mountPath: {{ .mountPath }}
              readOnly: {{ .readOnly }}
          {{- end }}
    {{- if .Values.server.nodeSelector }}
      nodeSelector:
{{ toYaml .Values.server.nodeSelector | indent 8 }}
    {{- end }}
    {{- if .Values.server.securityContext }}
      securityContext:
{{ toYaml .Values.server.securityContext | indent 8 }}
    {{- end }}
    {{- if .Values.server.tolerations }}
      tolerations:
{{ toYaml .Values.server.tolerations | indent 8 }}
    {{- end }}
    {{- if .Values.server.affinity }}
      affinity:
{{ toYaml .Values.server.affinity | indent 8 }}
    {{- end }}
      terminationGracePeriodSeconds: {{ .Values.server.terminationGracePeriodSeconds }}
      volumes:
        - name: config-volume
          configMap:
            name: {{ if .Values.server.configMapOverrideName }}{{ .Release.Name }}-{{ .Values.server.configMapOverrideName }}{{- else }}{{ template "prometheus.server.fullname" . }}{{- end }}
        - name: storage-volume
        {{- if .Values.server.persistentVolume.enabled }}
          persistentVolumeClaim:
            claimName: {{ if .Values.server.persistentVolume.existingClaim }}{{ .Values.server.persistentVolume.existingClaim }}{{- else }}{{ template "prometheus.server.fullname" . }}{{- end }}
        {{- else }}
          emptyDir: {}
        {{- end -}}
      {{- range .Values.server.extraHostPathMounts }}
        - name: {{ .name }}
          hostPath:
            path: {{ .hostPath }}
      {{- end }}
      {{- range .Values.configmapReload.extraConfigmapMounts }}
        - name: {{ $.Values.configmapReload.name }}-{{ .name }}
          configMap:
            name: {{ .configMap }}
      {{- end }}
      {{- range .Values.server.extraConfigmapMounts }}
        - name: {{ $.Values.server.name }}-{{ .name }}
          configMap:
            name: {{ .configMap }}
      {{- end }}
      {{- range .Values.server.extraSecretMounts }}
        - name: {{ .name }}
          secret:
            secretName: {{ .secretName }}
      {{- end }}
      {{- range .Values.configmapReload.extraConfigmapMounts }}
        - name: {{ .name }}
          configMap:
            name: {{ .configMap }}
      {{- end }}`

const server_ingress = `
{{- if .Values.server.ingress.enabled -}}
{{- $releaseName := .Release.Name -}}
{{- $serviceName := include "prometheus.server.fullname" . }}
{{- $servicePort := .Values.server.service.servicePort -}}
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
{{- if .Values.server.ingress.annotations }}
  annotations:
{{ toYaml .Values.server.ingress.annotations | indent 4 }}
{{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.server.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
{{- range $key, $value := .Values.server.ingress.extraLabels }}
    {{ $key }}: {{ $value }}
{{- end }}
  name: {{ template "prometheus.server.fullname" . }}
spec:
  rules:
  {{- range .Values.server.ingress.hosts }}
    {{- $url := splitList "/" . }}
    - host: {{ first $url }}
      http:
        paths:
          - path: /{{ rest $url | join "/" }}
            backend:
              serviceName: {{ $serviceName }}
              servicePort: {{ $servicePort }}
  {{- end -}}
{{- if .Values.server.ingress.tls }}
  tls:
{{ toYaml .Values.server.ingress.tls | indent 4 }}
  {{- end -}}
{{- end -}}`

const server_networkpolicy = `
{{- if .Values.networkPolicy.enabled }}
apiVersion: {{ template "prometheus.networkPolicy.apiVersion" . }}
kind: NetworkPolicy
metadata:
  name: {{ template "prometheus.server.fullname" . }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.server.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
spec:
  podSelector:
    matchLabels:
      app: {{ template "prometheus.name" . }}
      component: "{{ .Values.server.name }}"
      release: {{ .Release.Name }}
  ingress:
    - ports:
      - port: 9090
{{- end }}`

const server_pvc = `
{{- if .Values.server.persistentVolume.enabled -}}
{{- if not .Values.server.persistentVolume.existingClaim -}}
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  {{- if .Values.server.persistentVolume.annotations }}
  annotations:
{{ toYaml .Values.server.persistentVolume.annotations | indent 4 }}
  {{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.server.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.server.fullname" . }}
spec:
  accessModes:
{{ toYaml .Values.server.persistentVolume.accessModes | indent 4 }}
{{- if .Values.server.persistentVolume.storageClass }}
{{- if (eq "-" .Values.server.persistentVolume.storageClass) }}
  storageClassName: ""
{{- else }}
  storageClassName: "{{ .Values.server.persistentVolume.storageClass }}"
{{- end }}
{{- end }}
  resources:
    requests:
      storage: "{{ .Values.server.persistentVolume.size }}"
{{- end -}}
{{- end -}}`

const server_serviceaccount = `
{{- if .Values.serviceAccounts.server.create }}
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.server.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "prometheus.serviceAccountName.server" . }}
{{- end }}`

const server_service = `
apiVersion: v1
kind: Service
metadata:
{{- if .Values.server.service.annotations }}
  annotations:
{{ toYaml .Values.server.service.annotations | indent 4 }}
{{- end }}
  labels:
    app: {{ template "prometheus.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    component: "{{ .Values.server.name }}"
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
{{- if .Values.server.service.labels }}
{{ toYaml .Values.server.service.labels | indent 4 }}
{{- end }}
  name: {{ template "prometheus.server.fullname" . }}
spec:
{{- if .Values.server.service.clusterIP }}
  clusterIP: {{ .Values.server.service.clusterIP }}
{{- end }}
{{- if .Values.server.service.externalIPs }}
  externalIPs:
{{ toYaml .Values.server.service.externalIPs | indent 4 }}
{{- end }}
{{- if .Values.server.service.loadBalancerIP }}
  loadBalancerIP: {{ .Values.server.service.loadBalancerIP }}
{{- end }}
{{- if .Values.server.service.loadBalancerSourceRanges }}
  loadBalancerSourceRanges:
  {{- range $cidr := .Values.server.service.loadBalancerSourceRanges }}
    - {{ $cidr }}
  {{- end }}
{{- end }}
  ports:
    - name: http
      port: {{ .Values.server.service.servicePort }}
      protocol: TCP
      targetPort: 9090
    {{- if .Values.server.service.nodePort }}
      nodePort: {{ .Values.server.service.nodePort }}
    {{- end }}
  selector:
    app: {{ template "prometheus.name" . }}
    component: "{{ .Values.server.name }}"
    release: {{ .Release.Name }}
  type: "{{ .Values.server.service.type }}"`

const GAPGrafanaWebhook = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Release.Namespace }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    run: dingtalk
  name: webhook-dingtalk
  namespace: monitoring
spec:
  replicas: 1
  template:
    metadata:
      labels:
        run: dingtalk
    spec:
      containers:
      - name: dingtalk
        image: timonwong/prometheus-webhook-dingtalk:v0.3.0
        imagePullPolicy: IfNotPresent
        # 设置钉钉群聊自定义机器人后，使用实际 access_token 替换下面 xxxxxx部分
        args:
          - --ding.profile=webhook1=https://oapi.dingtalk.com/robot/send?access_token=xxxxxx
        ports:
        - containerPort: 8060
          protocol: TCP

---
apiVersion: v1
kind: Service
metadata:
  labels:
    run: dingtalk
  name: webhook-dingtalk
  namespace: monitoring
spec:
  ports:
  - port: 8060
    protocol: TCP
    targetPort: 8060
  selector:
    run: dingtalk
  sessionAffinity: None`

const GAPGrafanaDashboards = `
dashboards:
  default:
    Kubernetes_Nodes:
      json: |
        {
          "__inputs": [
            {
              "name": "MYDS_Prometheus",
              "label": "monitor-prometheus-server",
              "description": "",
              "type": "datasource",
              "pluginId": "prometheus",
              "pluginName": "Prometheus"
            }
          ],
          "__requires": [
            {
              "type": "grafana",
              "id": "grafana",
              "name": "Grafana",
              "version": "5.1.2"
            },
            {
              "type": "panel",
              "id": "graph",
              "name": "Graph",
              "version": "5.0.0"
            },
            {
              "type": "datasource",
              "id": "prometheus",
              "name": "Prometheus",
              "version": "5.0.0"
            },
            {
              "type": "panel",
              "id": "singlestat",
              "name": "Singlestat",
              "version": "5.0.0"
            }
          ],
          "annotations": {
            "list": [
              {
                "builtIn": 1,
                "datasource": "-- Grafana --",
                "enable": true,
                "hide": true,
                "iconColor": "rgba(0, 211, 255, 1)",
                "name": "Annotations & Alerts",
                "type": "dashboard"
              }
            ]
          },
          "description": "Kubernetes Nodes (prometheus)",
          "editable": true,
          "gnetId": 5324,
          "graphTooltip": 0,
          "id": null,
          "iteration": 1527950706714,
          "links": [],
          "panels": [
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "error": false,
              "fill": 1,
              "grid": {
                "threshold1Color": "rgba(216, 200, 27, 0.27)",
                "threshold2Color": "rgba(234, 112, 112, 0.22)"
              },
              "gridPos": {
                "h": 7,
                "w": 12,
                "x": 0,
                "y": 0
              },
              "id": 3,
              "isNew": false,
              "legend": {
                "alignAsTable": false,
                "avg": false,
                "current": false,
                "hideEmpty": false,
                "hideZero": false,
                "max": false,
                "min": false,
                "rightSide": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 2,
              "links": [],
              "nullPointMode": "connected",
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "100 - (avg by (cpu) (irate(node_cpu{mode=\"idle\", instance=\"$server\"}[5m])) * 100)",
                  "hide": false,
                  "intervalFactor": 10,
                  "legendFormat": "{{cpu}}",
                  "refId": "A",
                  "step": 50
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeShift": null,
              "title": "Idle CPU",
              "tooltip": {
                "msResolution": false,
                "shared": true,
                "sort": 0,
                "value_type": "cumulative"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "percent",
                  "label": "cpu usage",
                  "logBase": 1,
                  "max": 100,
                  "min": 0,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "error": false,
              "fill": 1,
              "grid": {
                "threshold1Color": "rgba(216, 200, 27, 0.27)",
                "threshold2Color": "rgba(234, 112, 112, 0.22)"
              },
              "gridPos": {
                "h": 7,
                "w": 12,
                "x": 12,
                "y": 0
              },
              "id": 9,
              "isNew": false,
              "legend": {
                "alignAsTable": false,
                "avg": false,
                "current": false,
                "hideEmpty": false,
                "hideZero": false,
                "max": false,
                "min": false,
                "rightSide": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 2,
              "links": [],
              "nullPointMode": "connected",
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "node_load1{instance=\"$server\"}",
                  "intervalFactor": 4,
                  "legendFormat": "load 1m",
                  "refId": "A",
                  "step": 20,
                  "target": ""
                },
                {
                  "expr": "node_load5{instance=\"$server\"}",
                  "intervalFactor": 4,
                  "legendFormat": "load 5m",
                  "refId": "B",
                  "step": 20,
                  "target": ""
                },
                {
                  "expr": "node_load15{instance=\"$server\"}",
                  "intervalFactor": 4,
                  "legendFormat": "load 15m",
                  "refId": "C",
                  "step": 20,
                  "target": ""
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeShift": null,
              "title": "System Load",
              "tooltip": {
                "msResolution": false,
                "shared": true,
                "sort": 0,
                "value_type": "cumulative"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "percentunit",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "error": false,
              "fill": 1,
              "grid": {
                "threshold1Color": "rgba(216, 200, 27, 0.27)",
                "threshold2Color": "rgba(234, 112, 112, 0.22)"
              },
              "gridPos": {
                "h": 7,
                "w": 18,
                "x": 0,
                "y": 7
              },
              "id": 4,
              "isNew": false,
              "legend": {
                "alignAsTable": false,
                "avg": false,
                "current": false,
                "hideEmpty": false,
                "hideZero": false,
                "max": false,
                "min": false,
                "rightSide": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 2,
              "links": [],
              "nullPointMode": "connected",
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [
                {
                  "alias": "node_memory_SwapFree{instance=\"172.17.0.1:9100\",job=\"prometheus\"}",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": true,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "node_memory_MemTotal{instance=\"$server\"} - node_memory_MemFree{instance=\"$server\"} - node_memory_Buffers{instance=\"$server\"} - node_memory_Cached{instance=\"$server\"}",
                  "hide": false,
                  "interval": "",
                  "intervalFactor": 2,
                  "legendFormat": "memory used",
                  "metric": "",
                  "refId": "C",
                  "step": 10
                },
                {
                  "expr": "node_memory_Buffers{instance=\"$server\"}",
                  "interval": "",
                  "intervalFactor": 2,
                  "legendFormat": "memory buffers",
                  "metric": "",
                  "refId": "E",
                  "step": 10
                },
                {
                  "expr": "node_memory_Cached{instance=\"$server\"}",
                  "intervalFactor": 2,
                  "legendFormat": "memory cached",
                  "metric": "",
                  "refId": "F",
                  "step": 10
                },
                {
                  "expr": "node_memory_MemFree{instance=\"$server\"}",
                  "intervalFactor": 2,
                  "legendFormat": "memory free",
                  "metric": "",
                  "refId": "D",
                  "step": 10
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeShift": null,
              "title": "Memory Usage",
              "tooltip": {
                "msResolution": false,
                "shared": true,
                "sort": 0,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "bytes",
                  "logBase": 1,
                  "min": "0",
                  "show": true
                },
                {
                  "format": "short",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "cacheTimeout": null,
              "colorBackground": false,
              "colorValue": false,
              "colors": [
                "rgba(50, 172, 45, 0.97)",
                "rgba(237, 129, 40, 0.89)",
                "rgba(245, 54, 54, 0.9)"
              ],
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "format": "percent",
              "gauge": {
                "maxValue": 100,
                "minValue": 0,
                "show": true,
                "thresholdLabels": false,
                "thresholdMarkers": true
              },
              "gridPos": {
                "h": 7,
                "w": 6,
                "x": 18,
                "y": 7
              },
              "hideTimeOverride": false,
              "id": 5,
              "interval": null,
              "links": [],
              "mappingType": 1,
              "mappingTypes": [
                {
                  "name": "value to text",
                  "value": 1
                },
                {
                  "name": "range to text",
                  "value": 2
                }
              ],
              "maxDataPoints": 100,
              "nullPointMode": "connected",
              "nullText": null,
              "postfix": "",
              "postfixFontSize": "50%",
              "prefix": "",
              "prefixFontSize": "50%",
              "rangeMaps": [
                {
                  "from": "null",
                  "text": "N/A",
                  "to": "null"
                }
              ],
              "sparkline": {
                "fillColor": "rgba(31, 118, 189, 0.18)",
                "full": false,
                "lineColor": "rgb(31, 120, 193)",
                "show": false
              },
              "tableColumn": "",
              "targets": [
                {
                  "expr": "((node_memory_MemTotal{instance=\"$server\"} - node_memory_MemFree{instance=\"$server\"}  - node_memory_Buffers{instance=\"$server\"} - node_memory_Cached{instance=\"$server\"}) / node_memory_MemTotal{instance=\"$server\"}) * 100",
                  "intervalFactor": 2,
                  "refId": "A",
                  "step": 60,
                  "target": ""
                }
              ],
              "thresholds": "80, 90",
              "title": "Memory Usage",
              "transparent": false,
              "type": "singlestat",
              "valueFontSize": "80%",
              "valueMaps": [
                {
                  "op": "=",
                  "text": "N/A",
                  "value": "null"
                }
              ],
              "valueName": "avg"
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "error": false,
              "fill": 1,
              "grid": {
                "threshold1Color": "rgba(216, 200, 27, 0.27)",
                "threshold2Color": "rgba(234, 112, 112, 0.22)"
              },
              "gridPos": {
                "h": 7,
                "w": 18,
                "x": 0,
                "y": 14
              },
              "id": 6,
              "isNew": true,
              "legend": {
                "alignAsTable": false,
                "avg": false,
                "current": false,
                "hideEmpty": false,
                "hideZero": false,
                "max": false,
                "min": false,
                "rightSide": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 2,
              "links": [],
              "nullPointMode": "connected",
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [
                {
                  "alias": "read",
                  "yaxis": 1
                },
                {
                  "alias": "{instance=\"172.17.0.1:9100\"}",
                  "yaxis": 2
                },
                {
                  "alias": "io time",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "sum by (instance) (rate(node_disk_bytes_read{instance=\"$server\"}[2m]))",
                  "hide": false,
                  "intervalFactor": 4,
                  "legendFormat": "read",
                  "refId": "A",
                  "step": 20,
                  "target": ""
                },
                {
                  "expr": "sum by (instance) (rate(node_disk_bytes_written{instance=\"$server\"}[2m]))",
                  "intervalFactor": 4,
                  "legendFormat": "written",
                  "refId": "B",
                  "step": 20
                },
                {
                  "expr": "sum by (instance) (rate(node_disk_io_time_ms{instance=\"$server\"}[2m]))",
                  "intervalFactor": 4,
                  "legendFormat": "io time",
                  "refId": "C",
                  "step": 20
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeShift": null,
              "title": "Disk I/O",
              "tooltip": {
                "msResolution": false,
                "shared": true,
                "sort": 0,
                "value_type": "cumulative"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "bytes",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "ms",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "cacheTimeout": null,
              "colorBackground": false,
              "colorValue": false,
              "colors": [
                "rgba(50, 172, 45, 0.97)",
                "rgba(237, 129, 40, 0.89)",
                "rgba(245, 54, 54, 0.9)"
              ],
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "format": "percentunit",
              "gauge": {
                "maxValue": 1,
                "minValue": 0,
                "show": true,
                "thresholdLabels": false,
                "thresholdMarkers": true
              },
              "gridPos": {
                "h": 7,
                "w": 6,
                "x": 18,
                "y": 14
              },
              "hideTimeOverride": false,
              "id": 7,
              "interval": null,
              "links": [],
              "mappingType": 1,
              "mappingTypes": [
                {
                  "name": "value to text",
                  "value": 1
                },
                {
                  "name": "range to text",
                  "value": 2
                }
              ],
              "maxDataPoints": 100,
              "nullPointMode": "connected",
              "nullText": null,
              "postfix": "",
              "postfixFontSize": "50%",
              "prefix": "",
              "prefixFontSize": "50%",
              "rangeMaps": [
                {
                  "from": "null",
                  "text": "N/A",
                  "to": "null"
                }
              ],
              "sparkline": {
                "fillColor": "rgba(31, 118, 189, 0.18)",
                "full": false,
                "lineColor": "rgb(31, 120, 193)",
                "show": false
              },
              "tableColumn": "",
              "targets": [
                {
                  "expr": "(sum(node_filesystem_size{device!=\"rootfs\",instance=\"$server\"}) - sum(node_filesystem_free{device!=\"rootfs\",instance=\"$server\"})) / sum(node_filesystem_size{device!=\"rootfs\",instance=\"$server\"})",
                  "intervalFactor": 2,
                  "refId": "A",
                  "step": 60,
                  "target": ""
                }
              ],
              "thresholds": "0.75, 0.9",
              "title": "Disk Space Usage",
              "transparent": false,
              "type": "singlestat",
              "valueFontSize": "80%",
              "valueMaps": [
                {
                  "op": "=",
                  "text": "N/A",
                  "value": "null"
                }
              ],
              "valueName": "current"
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "error": false,
              "fill": 1,
              "grid": {
                "threshold1Color": "rgba(216, 200, 27, 0.27)",
                "threshold2Color": "rgba(234, 112, 112, 0.22)"
              },
              "gridPos": {
                "h": 7,
                "w": 12,
                "x": 0,
                "y": 21
              },
              "id": 8,
              "isNew": false,
              "legend": {
                "alignAsTable": false,
                "avg": false,
                "current": false,
                "hideEmpty": false,
                "hideZero": false,
                "max": false,
                "min": false,
                "rightSide": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 2,
              "links": [],
              "nullPointMode": "connected",
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [
                {
                  "alias": "transmitted",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "rate(node_network_receive_bytes{instance=\"$server\",device!~\"lo\"}[5m])",
                  "hide": false,
                  "intervalFactor": 2,
                  "legendFormat": "{{device}}",
                  "refId": "A",
                  "step": 10,
                  "target": ""
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeShift": null,
              "title": "Network Received",
              "tooltip": {
                "msResolution": false,
                "shared": true,
                "sort": 0,
                "value_type": "cumulative"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "bytes",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "bytes",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "error": false,
              "fill": 1,
              "grid": {
                "threshold1Color": "rgba(216, 200, 27, 0.27)",
                "threshold2Color": "rgba(234, 112, 112, 0.22)"
              },
              "gridPos": {
                "h": 7,
                "w": 12,
                "x": 12,
                "y": 21
              },
              "id": 10,
              "isNew": false,
              "legend": {
                "alignAsTable": false,
                "avg": false,
                "current": false,
                "hideEmpty": false,
                "hideZero": false,
                "max": false,
                "min": false,
                "rightSide": false,
                "show": true,
                "total": false,
                "values": false
              },
              "lines": true,
              "linewidth": 2,
              "links": [],
              "nullPointMode": "connected",
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [
                {
                  "alias": "transmitted",
                  "yaxis": 2
                }
              ],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "rate(node_network_transmit_bytes{instance=\"$server\",device!~\"lo\"}[5m])",
                  "hide": false,
                  "intervalFactor": 2,
                  "legendFormat": "{{device}}",
                  "refId": "B",
                  "step": 10,
                  "target": ""
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeShift": null,
              "title": "Network Transmitted",
              "tooltip": {
                "msResolution": false,
                "shared": true,
                "sort": 0,
                "value_type": "cumulative"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "bytes",
                  "logBase": 1,
                  "show": true
                },
                {
                  "format": "bytes",
                  "logBase": 1,
                  "show": true
                }
              ],
              "yaxis": {
                "align": false,
                "alignLevel": null
              }
            }
          ],
          "refresh": false,
          "schemaVersion": 16,
          "style": "dark",
          "tags": [],
          "templating": {
            "list": [
              {
                "allValue": null,
                "current": {},
                "datasource": "MYDS_Prometheus",
                "hide": 0,
                "includeAll": false,
                "label": null,
                "multi": false,
                "name": "server",
                "options": [],
                "query": "label_values(node_boot_time, instance)",
                "refresh": 1,
                "regex": "",
                "sort": 0,
                "tagValuesQuery": "",
                "tags": [],
                "tagsQuery": "",
                "type": "query",
                "useTags": false
              }
            ]
          },
          "time": {
            "from": "now-1h",
            "to": "now"
          },
          "timepicker": {
            "refresh_intervals": [
              "5s",
              "10s",
              "30s",
              "1m",
              "5m",
              "15m",
              "30m",
              "1h",
              "2h",
              "1d"
            ],
            "time_options": [
              "5m",
              "15m",
              "1h",
              "6h",
              "12h",
              "24h",
              "2d",
              "7d",
              "30d"
            ]
          },
          "timezone": "browser",
          "title": "Kubernetes Nodes (prometheus)",
          "uid": "HcEiL8kmz",
          "version": 1
        }
    Kubernetes_Pods:
      json: |
        {
          "__inputs": [
            {
              "name": "MYDS_Prometheus",
              "label": "prometheus",
              "description": "",
              "type": "datasource",
              "pluginId": "prometheus",
              "pluginName": "Prometheus"
            }
          ],
          "__requires": [
            {
              "type": "grafana",
              "id": "grafana",
              "name": "Grafana",
              "version": "5.0.4"
            },
            {
              "type": "panel",
              "id": "graph",
              "name": "Graph",
              "version": "5.0.0"
            },
            {
              "type": "datasource",
              "id": "prometheus",
              "name": "Prometheus",
              "version": "5.0.0"
            }
          ],
          "annotations": {
            "list": [
              {
                "builtIn": 1,
                "datasource": "-- Grafana --",
                "enable": true,
                "hide": true,
                "iconColor": "rgba(0, 211, 255, 1)",
                "name": "Annotations & Alerts",
                "type": "dashboard"
              }
            ]
          },
          "editable": true,
          "gnetId": null,
          "graphTooltip": 0,
          "id": null,
          "iteration": 1526871489129,
          "links": [],
          "panels": [
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": "MYDS_Prometheus",
              "fill": 1,
              "gridPos": {
                "h": 9,
                "w": 12,
                "x": 0,
                "y": 0
              },
              "id": 4,
              "legend": {
                "alignAsTable": true,
                "avg": true,
                "current": true,
                "max": true,
                "min": true,
                "show": true,
                "total": false,
                "values": true
              },
              "lines": true,
              "linewidth": 1,
              "links": [],
              "nullPointMode": "null",
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "avg(rate (container_cpu_usage_seconds_total{image!=\"\",container_name!=\"POD\",namespace=~'$namespace$',pod_name=~'$pod_name$',container_name=~'$container_name$'}[5m]))",
                  "format": "time_series",
                  "hide": false,
                  "instant": false,
                  "intervalFactor": 1,
                  "legendFormat": "CPU Usage",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeShift": null,
              "title": "CPU Usage",
              "tooltip": {
                "shared": true,
                "sort": 0,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "short",
                  "label": null,
                  "logBase": 1,
                  "max": null,
                  "min": null,
                  "show": true
                },
                {
                  "format": "short",
                  "label": null,
                  "logBase": 1,
                  "max": null,
                  "min": null,
                  "show": true
                }
              ]
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": "MYDS_Prometheus",
              "fill": 1,
              "gridPos": {
                "h": 9,
                "w": 12,
                "x": 12,
                "y": 0
              },
              "id": 6,
              "legend": {
                "alignAsTable": true,
                "avg": true,
                "current": true,
                "max": true,
                "min": true,
                "show": true,
                "total": false,
                "values": true
              },
              "lines": true,
              "linewidth": 1,
              "links": [],
              "nullPointMode": "null",
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "avg(container_memory_usage_bytes{image!=\"\",container_name!=\"POD\",namespace=~'$namespace$',pod_name=~'$pod_name$',container_name=~'$container_name$'})",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "MEM Usage",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeShift": null,
              "title": "MEM Usage",
              "tooltip": {
                "shared": true,
                "sort": 0,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "short",
                  "label": null,
                  "logBase": 1,
                  "max": null,
                  "min": null,
                  "show": true
                },
                {
                  "format": "short",
                  "label": null,
                  "logBase": 1,
                  "max": null,
                  "min": null,
                  "show": true
                }
              ]
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": "MYDS_Prometheus",
              "fill": 1,
              "gridPos": {
                "h": 9,
                "w": 12,
                "x": 0,
                "y": 9
              },
              "id": 8,
              "legend": {
                "alignAsTable": true,
                "avg": true,
                "current": true,
                "max": true,
                "min": true,
                "show": true,
                "total": false,
                "values": true
              },
              "lines": true,
              "linewidth": 1,
              "links": [],
              "nullPointMode": "null",
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "sum (rate (container_network_receive_bytes_total{image!=\"\",namespace=~'$namespace$',pod_name=~'$pod_name$'}[5m]))",
                  "format": "time_series",
                  "hide": false,
                  "intervalFactor": 1,
                  "legendFormat": "Network Input",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeShift": null,
              "title": "Network Input Bytes",
              "tooltip": {
                "shared": true,
                "sort": 0,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "short",
                  "label": null,
                  "logBase": 1,
                  "max": null,
                  "min": null,
                  "show": true
                },
                {
                  "format": "short",
                  "label": null,
                  "logBase": 1,
                  "max": null,
                  "min": null,
                  "show": true
                }
              ]
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": "MYDS_Prometheus",
              "fill": 1,
              "gridPos": {
                "h": 9,
                "w": 12,
                "x": 12,
                "y": 9
              },
              "id": 10,
              "legend": {
                "alignAsTable": true,
                "avg": true,
                "current": true,
                "max": true,
                "min": true,
                "show": true,
                "total": false,
                "values": true
              },
              "lines": true,
              "linewidth": 1,
              "links": [],
              "nullPointMode": "null",
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "sum (rate (container_network_transmit_bytes_total{image!=\"\",namespace=~'$namespace$',pod_name=~'$pod_name$'}[5m]))",
                  "format": "time_series",
                  "hide": false,
                  "intervalFactor": 1,
                  "legendFormat": "Network Output",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeShift": null,
              "title": "Network Output Bytes",
              "tooltip": {
                "shared": true,
                "sort": 0,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "short",
                  "label": null,
                  "logBase": 1,
                  "max": null,
                  "min": null,
                  "show": true
                },
                {
                  "format": "short",
                  "label": null,
                  "logBase": 1,
                  "max": null,
                  "min": null,
                  "show": true
                }
              ]
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": "MYDS_Prometheus",
              "fill": 1,
              "gridPos": {
                "h": 9,
                "w": 12,
                "x": 0,
                "y": 18
              },
              "id": 12,
              "legend": {
                "alignAsTable": true,
                "avg": true,
                "current": true,
                "max": true,
                "min": true,
                "show": true,
                "total": false,
                "values": true
              },
              "lines": true,
              "linewidth": 1,
              "links": [],
              "nullPointMode": "null",
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "sum (rate (container_network_receive_packets_total{image!=\"\",namespace=~'$namespace$',pod_name=~'$pod_name$'}[5m]))",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "Input Packets",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeShift": null,
              "title": "Network Input Packets",
              "tooltip": {
                "shared": true,
                "sort": 0,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "short",
                  "label": null,
                  "logBase": 1,
                  "max": null,
                  "min": null,
                  "show": true
                },
                {
                  "format": "short",
                  "label": null,
                  "logBase": 1,
                  "max": null,
                  "min": null,
                  "show": true
                }
              ]
            },
            {
              "aliasColors": {},
              "bars": false,
              "dashLength": 10,
              "dashes": false,
              "datasource": "MYDS_Prometheus",
              "fill": 1,
              "gridPos": {
                "h": 9,
                "w": 12,
                "x": 12,
                "y": 18
              },
              "id": 14,
              "legend": {
                "alignAsTable": true,
                "avg": true,
                "current": true,
                "max": true,
                "min": true,
                "show": true,
                "total": false,
                "values": true
              },
              "lines": true,
              "linewidth": 1,
              "links": [],
              "nullPointMode": "null",
              "percentage": false,
              "pointradius": 5,
              "points": false,
              "renderer": "flot",
              "seriesOverrides": [],
              "spaceLength": 10,
              "stack": false,
              "steppedLine": false,
              "targets": [
                {
                  "expr": "sum (rate (container_network_transmit_packets_total{image!=\"\",namespace=~'$namespace$',pod_name=~'$pod_name$'}[5m]))",
                  "format": "time_series",
                  "intervalFactor": 1,
                  "legendFormat": "Output Packets",
                  "refId": "A"
                }
              ],
              "thresholds": [],
              "timeFrom": null,
              "timeShift": null,
              "title": "Network Output Packets",
              "tooltip": {
                "shared": true,
                "sort": 0,
                "value_type": "individual"
              },
              "type": "graph",
              "xaxis": {
                "buckets": null,
                "mode": "time",
                "name": null,
                "show": true,
                "values": []
              },
              "yaxes": [
                {
                  "format": "short",
                  "label": null,
                  "logBase": 1,
                  "max": null,
                  "min": null,
                  "show": true
                },
                {
                  "format": "short",
                  "label": null,
                  "logBase": 1,
                  "max": null,
                  "min": null,
                  "show": true
                }
              ]
            }
          ],
          "schemaVersion": 16,
          "style": "dark",
          "tags": [],
          "templating": {
            "list": [
              {
                "allValue": null,
                "current": {},
                "datasource": "MYDS_Prometheus",
                "hide": 0,
                "includeAll": true,
                "label": null,
                "multi": false,
                "name": "namespace",
                "options": [],
                "query": "label_values(container_memory_usage_bytes, namespace)",
                "refresh": 1,
                "regex": "",
                "sort": 0,
                "tagValuesQuery": "",
                "tags": [],
                "tagsQuery": "",
                "type": "query",
                "useTags": false
              },
              {
                "allValue": null,
                "current": {},
                "datasource": "MYDS_Prometheus",
                "hide": 0,
                "includeAll": true,
                "label": null,
                "multi": false,
                "name": "app_name",
                "options": [],
                "query": "label_values(container_memory_usage_bytes{namespace =~\"$namespace.*\"}, pod_name)",
                "refresh": 1,
                "regex": "/(.*)-.*/",
                "sort": 0,
                "tagValuesQuery": "",
                "tags": [],
                "tagsQuery": "",
                "type": "query",
                "useTags": false
              },
              {
                "allValue": null,
                "current": {},
                "datasource": "MYDS_Prometheus",
                "hide": 0,
                "includeAll": true,
                "label": null,
                "multi": false,
                "name": "pod_name",
                "options": [],
                "query": "label_values(container_memory_usage_bytes{pod_name =~\"$app_name.*\"}, pod_name)",
                "refresh": 1,
                "regex": "",
                "sort": 0,
                "tagValuesQuery": "",
                "tags": [],
                "tagsQuery": "",
                "type": "query",
                "useTags": false
              },
              {
                "allValue": null,
                "current": {},
                "datasource": "MYDS_Prometheus",
                "hide": 0,
                "includeAll": true,
                "label": null,
                "multi": false,
                "name": "container_name",
                "options": [],
                "query": "label_values(container_memory_usage_bytes{pod_name =~\"$pod_name.*\",container_name!=\"POD\"}, container_name)",
                "refresh": 1,
                "regex": "",
                "sort": 0,
                "tagValuesQuery": "",
                "tags": [],
                "tagsQuery": "",
                "type": "query",
                "useTags": false
              }
            ]
          },
          "time": {
            "from": "now-1h",
            "to": "now"
          },
          "timepicker": {
            "refresh_intervals": [
              "5s",
              "10s",
              "30s",
              "1m",
              "5m",
              "15m",
              "30m",
              "1h",
              "2h",
              "1d"
            ],
            "time_options": [
              "5m",
              "15m",
              "1h",
              "6h",
              "12h",
              "24h",
              "2d",
              "7d",
              "30d"
            ]
          },
          "timezone": "",
          "title": "Kubernetes Apps",
          "uid": "0xvoGCkmz",
          "version": 1
        }
    Kubernetes_Cluster_Health:
      json: |
        {
          "__inputs": [
            {
              "name": "MYDS_Prometheus",
              "label": "monitor-prometheus-server",
              "description": "",
              "type": "datasource",
              "pluginId": "prometheus",
              "pluginName": "Prometheus"
            }
          ],
          "__requires": [
            {
              "type": "grafana",
              "id": "grafana",
              "name": "Grafana",
              "version": "5.1.2"
            },
            {
              "type": "datasource",
              "id": "prometheus",
              "name": "Prometheus",
              "version": "5.0.0"
            },
            {
              "type": "panel",
              "id": "singlestat",
              "name": "Singlestat",
              "version": "5.0.0"
            }
          ],
          "annotations": {
            "list": [
              {
                "builtIn": 1,
                "datasource": "-- Grafana --",
                "enable": true,
                "hide": true,
                "iconColor": "rgba(0, 211, 255, 1)",
                "name": "Annotations & Alerts",
                "type": "dashboard"
              }
            ]
          },
          "description": "Kubernetes Cluster Health (Prometheus)",
          "editable": true,
          "gnetId": 5312,
          "graphTooltip": 0,
          "id": null,
          "links": [],
          "panels": [
            {
              "cacheTimeout": null,
              "colorBackground": false,
              "colorValue": true,
              "colors": [
                "rgba(50, 172, 45, 0.97)",
                "rgba(237, 129, 40, 0.89)",
                "rgba(245, 54, 54, 0.9)"
              ],
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "format": "none",
              "gauge": {
                "maxValue": 100,
                "minValue": 0,
                "show": false,
                "thresholdLabels": false,
                "thresholdMarkers": true
              },
              "gridPos": {
                "h": 6,
                "w": 6,
                "x": 0,
                "y": 0
              },
              "hideTimeOverride": false,
              "id": 1,
              "interval": null,
              "links": [],
              "mappingType": 1,
              "mappingTypes": [
                {
                  "name": "value to text",
                  "value": 1
                },
                {
                  "name": "range to text",
                  "value": 2
                }
              ],
              "maxDataPoints": 100,
              "nullPointMode": "connected",
              "nullText": null,
              "postfix": "",
              "postfixFontSize": "50%",
              "prefix": "",
              "prefixFontSize": "50%",
              "rangeMaps": [
                {
                  "from": "null",
                  "text": "N/A",
                  "to": "null"
                }
              ],
              "sparkline": {
                "fillColor": "rgba(31, 118, 189, 0.18)",
                "full": false,
                "lineColor": "rgb(31, 120, 193)",
                "show": false
              },
              "tableColumn": "",
              "targets": [
                {
                  "expr": "sum(up{job=~\"apiserver|kube-scheduler|kube-controller-manager\"} == 0)",
                  "format": "time_series",
                  "intervalFactor": 2,
                  "legendFormat": "",
                  "refId": "A",
                  "step": 600
                }
              ],
              "thresholds": "1, 3",
              "title": "Control Plane Components Down",
              "transparent": false,
              "type": "singlestat",
              "valueFontSize": "80%",
              "valueMaps": [
                {
                  "op": "=",
                  "text": "Everything UP and healthy",
                  "value": "null"
                },
                {
                  "op": "=",
                  "text": "",
                  "value": ""
                }
              ],
              "valueName": "avg"
            },
            {
              "cacheTimeout": null,
              "colorBackground": false,
              "colorValue": true,
              "colors": [
                "rgba(50, 172, 45, 0.97)",
                "rgba(237, 129, 40, 0.89)",
                "rgba(245, 54, 54, 0.9)"
              ],
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "format": "none",
              "gauge": {
                "maxValue": 100,
                "minValue": 0,
                "show": false,
                "thresholdLabels": false,
                "thresholdMarkers": true
              },
              "gridPos": {
                "h": 3,
                "w": 3,
                "x": 6,
                "y": 0
              },
              "hideTimeOverride": false,
              "id": 6,
              "interval": null,
              "links": [],
              "mappingType": 1,
              "mappingTypes": [
                {
                  "name": "value to text",
                  "value": 1
                },
                {
                  "name": "range to text",
                  "value": 2
                }
              ],
              "maxDataPoints": 100,
              "nullPointMode": "connected",
              "nullText": null,
              "postfix": "",
              "postfixFontSize": "50%",
              "prefix": "",
              "prefixFontSize": "50%",
              "rangeMaps": [
                {
                  "from": "null",
                  "text": "N/A",
                  "to": "null"
                }
              ],
              "sparkline": {
                "fillColor": "rgba(31, 118, 189, 0.18)",
                "full": false,
                "lineColor": "rgb(31, 120, 193)",
                "show": false
              },
              "tableColumn": "",
              "targets": [
                {
                  "expr": "sum(kube_node_status_condition{condition=\"DiskPressure\",status=\"true\"})",
                  "format": "time_series",
                  "intervalFactor": 2,
                  "legendFormat": "",
                  "refId": "A",
                  "step": 600
                }
              ],
              "thresholds": "1, 3",
              "title": "Node Disk Pressure",
              "transparent": false,
              "type": "singlestat",
              "valueFontSize": "80%",
              "valueMaps": [
                {
                  "op": "=",
                  "text": "N/A",
                  "value": "null"
                }
              ],
              "valueName": "current"
            },
            {
              "cacheTimeout": null,
              "colorBackground": false,
              "colorValue": true,
              "colors": [
                "rgba(50, 172, 45, 0.97)",
                "rgba(237, 129, 40, 0.89)",
                "rgba(245, 54, 54, 0.9)"
              ],
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "format": "none",
              "gauge": {
                "maxValue": 100,
                "minValue": 0,
                "show": false,
                "thresholdLabels": false,
                "thresholdMarkers": true
              },
              "gridPos": {
                "h": 3,
                "w": 3,
                "x": 9,
                "y": 0
              },
              "hideTimeOverride": false,
              "id": 8,
              "interval": null,
              "links": [],
              "mappingType": 1,
              "mappingTypes": [
                {
                  "name": "value to text",
                  "value": 1
                },
                {
                  "name": "range to text",
                  "value": 2
                }
              ],
              "maxDataPoints": 100,
              "nullPointMode": "connected",
              "nullText": null,
              "postfix": "",
              "postfixFontSize": "50%",
              "prefix": "",
              "prefixFontSize": "50%",
              "rangeMaps": [
                {
                  "from": "null",
                  "text": "N/A",
                  "to": "null"
                }
              ],
              "sparkline": {
                "fillColor": "rgba(31, 118, 189, 0.18)",
                "full": false,
                "lineColor": "rgb(31, 120, 193)",
                "show": false
              },
              "tableColumn": "",
              "targets": [
                {
                  "expr": "sum(kube_node_spec_unschedulable)",
                  "format": "time_series",
                  "intervalFactor": 2,
                  "legendFormat": "",
                  "refId": "A",
                  "step": 600
                }
              ],
              "thresholds": "1, 3",
              "title": "Nodes Unschedulable",
              "transparent": false,
              "type": "singlestat",
              "valueFontSize": "80%",
              "valueMaps": [
                {
                  "op": "=",
                  "text": "N/A",
                  "value": "null"
                }
              ],
              "valueName": "current"
            },
            {
              "cacheTimeout": null,
              "colorBackground": false,
              "colorValue": true,
              "colors": [
                "rgba(50, 172, 45, 0.97)",
                "rgba(237, 129, 40, 0.89)",
                "rgba(245, 54, 54, 0.9)"
              ],
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "format": "none",
              "gauge": {
                "maxValue": 100,
                "minValue": 0,
                "show": false,
                "thresholdLabels": false,
                "thresholdMarkers": true
              },
              "gridPos": {
                "h": 3,
                "w": 3,
                "x": 12,
                "y": 0
              },
              "hideTimeOverride": false,
              "id": 7,
              "interval": null,
              "links": [],
              "mappingType": 1,
              "mappingTypes": [
                {
                  "name": "value to text",
                  "value": 1
                },
                {
                  "name": "range to text",
                  "value": 2
                }
              ],
              "maxDataPoints": 100,
              "nullPointMode": "connected",
              "nullText": null,
              "postfix": "",
              "postfixFontSize": "50%",
              "prefix": "",
              "prefixFontSize": "50%",
              "rangeMaps": [
                {
                  "from": "null",
                  "text": "N/A",
                  "to": "null"
                }
              ],
              "sparkline": {
                "fillColor": "rgba(31, 118, 189, 0.18)",
                "full": false,
                "lineColor": "rgb(31, 120, 193)",
                "show": false
              },
              "tableColumn": "",
              "targets": [
                {
                  "expr": "sum(kube_node_status_condition{condition=\"MemoryPressure\",status=\"true\"})",
                  "format": "time_series",
                  "intervalFactor": 2,
                  "legendFormat": "",
                  "refId": "A",
                  "step": 600
                }
              ],
              "thresholds": "1, 3",
              "title": "Node Memory Pressure",
              "transparent": false,
              "type": "singlestat",
              "valueFontSize": "80%",
              "valueMaps": [
                {
                  "op": "=",
                  "text": "N/A",
                  "value": "null"
                }
              ],
              "valueName": "current"
            },
            {
              "cacheTimeout": null,
              "colorBackground": false,
              "colorValue": true,
              "colors": [
                "rgba(50, 172, 45, 0.97)",
                "rgba(237, 129, 40, 0.89)",
                "rgba(245, 54, 54, 0.9)"
              ],
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "format": "none",
              "gauge": {
                "maxValue": 100,
                "minValue": 0,
                "show": false,
                "thresholdLabels": false,
                "thresholdMarkers": true
              },
              "gridPos": {
                "h": 3,
                "w": 3,
                "x": 15,
                "y": 0
              },
              "hideTimeOverride": false,
              "id": 4,
              "interval": null,
              "links": [],
              "mappingType": 1,
              "mappingTypes": [
                {
                  "name": "value to text",
                  "value": 1
                },
                {
                  "name": "range to text",
                  "value": 2
                }
              ],
              "maxDataPoints": 100,
              "nullPointMode": "connected",
              "nullText": null,
              "postfix": "",
              "postfixFontSize": "50%",
              "prefix": "",
              "prefixFontSize": "50%",
              "rangeMaps": [
                {
                  "from": "null",
                  "text": "N/A",
                  "to": "null"
                }
              ],
              "sparkline": {
                "fillColor": "rgba(31, 118, 189, 0.18)",
                "full": false,
                "lineColor": "rgb(31, 120, 193)",
                "show": false
              },
              "tableColumn": "",
              "targets": [
                {
                  "expr": "count(increase(kube_pod_container_status_restarts[1h]) > 5)",
                  "format": "time_series",
                  "intervalFactor": 2,
                  "legendFormat": "",
                  "refId": "A",
                  "step": 600
                }
              ],
              "thresholds": "1, 3",
              "title": "Crashlooping Pods",
              "transparent": false,
              "type": "singlestat",
              "valueFontSize": "80%",
              "valueMaps": [
                {
                  "op": "=",
                  "text": "0",
                  "value": "null"
                }
              ],
              "valueName": "current"
            },
            {
              "cacheTimeout": null,
              "colorBackground": false,
              "colorValue": true,
              "colors": [
                "rgba(50, 172, 45, 0.97)",
                "rgba(237, 129, 40, 0.89)",
                "rgba(245, 54, 54, 0.9)"
              ],
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "format": "none",
              "gauge": {
                "maxValue": 100,
                "minValue": 0,
                "show": false,
                "thresholdLabels": false,
                "thresholdMarkers": true
              },
              "gridPos": {
                "h": 3,
                "w": 3,
                "x": 6,
                "y": 3
              },
              "hideTimeOverride": false,
              "id": 5,
              "interval": null,
              "links": [],
              "mappingType": 1,
              "mappingTypes": [
                {
                  "name": "value to text",
                  "value": 1
                },
                {
                  "name": "range to text",
                  "value": 2
                }
              ],
              "maxDataPoints": 100,
              "nullPointMode": "connected",
              "nullText": null,
              "postfix": "",
              "postfixFontSize": "50%",
              "prefix": "",
              "prefixFontSize": "50%",
              "rangeMaps": [
                {
                  "from": "null",
                  "text": "N/A",
                  "to": "null"
                }
              ],
              "sparkline": {
                "fillColor": "rgba(31, 118, 189, 0.18)",
                "full": false,
                "lineColor": "rgb(31, 120, 193)",
                "show": false
              },
              "tableColumn": "",
              "targets": [
                {
                  "expr": "sum(kube_node_status_condition{condition=\"Ready\",status!=\"true\"})",
                  "format": "time_series",
                  "intervalFactor": 2,
                  "legendFormat": "",
                  "refId": "A",
                  "step": 600
                }
              ],
              "thresholds": "1, 3",
              "title": "Node Not Ready",
              "transparent": false,
              "type": "singlestat",
              "valueFontSize": "80%",
              "valueMaps": [
                {
                  "op": "=",
                  "text": "N/A",
                  "value": "null"
                }
              ],
              "valueName": "current"
            },
            {
              "cacheTimeout": null,
              "colorBackground": false,
              "colorValue": true,
              "colors": [
                "rgba(50, 172, 45, 0.97)",
                "rgba(237, 129, 40, 0.89)",
                "rgba(245, 54, 54, 0.9)"
              ],
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "format": "none",
              "gauge": {
                "maxValue": 100,
                "minValue": 0,
                "show": false,
                "thresholdLabels": false,
                "thresholdMarkers": true
              },
              "gridPos": {
                "h": 3,
                "w": 3,
                "x": 9,
                "y": 3
              },
              "hideTimeOverride": false,
              "id": 2,
              "interval": null,
              "links": [],
              "mappingType": 1,
              "mappingTypes": [
                {
                  "name": "value to text",
                  "value": 1
                },
                {
                  "name": "range to text",
                  "value": 2
                }
              ],
              "maxDataPoints": 100,
              "nullPointMode": "connected",
              "nullText": null,
              "postfix": "",
              "postfixFontSize": "50%",
              "prefix": "",
              "prefixFontSize": "50%",
              "rangeMaps": [
                {
                  "from": "null",
                  "text": "N/A",
                  "to": "null"
                }
              ],
              "sparkline": {
                "fillColor": "rgba(31, 118, 189, 0.18)",
                "full": false,
                "lineColor": "rgb(31, 120, 193)",
                "show": false
              },
              "tableColumn": "",
              "targets": [
                {
                  "expr": "sum(ALERTS{alertstate=\"firing\",alertname!=\"DeadMansSwitch\"})",
                  "format": "time_series",
                  "intervalFactor": 2,
                  "legendFormat": "",
                  "refId": "A",
                  "step": 600
                }
              ],
              "thresholds": "1, 3",
              "title": "Alerts Firing",
              "transparent": false,
              "type": "singlestat",
              "valueFontSize": "80%",
              "valueMaps": [
                {
                  "op": "=",
                  "text": "0",
                  "value": "null"
                }
              ],
              "valueName": "current"
            },
            {
              "cacheTimeout": null,
              "colorBackground": false,
              "colorValue": true,
              "colors": [
                "rgba(50, 172, 45, 0.97)",
                "rgba(237, 129, 40, 0.89)",
                "rgba(245, 54, 54, 0.9)"
              ],
              "datasource": "MYDS_Prometheus",
              "editable": true,
              "format": "none",
              "gauge": {
                "maxValue": 100,
                "minValue": 0,
                "show": false,
                "thresholdLabels": false,
                "thresholdMarkers": true
              },
              "gridPos": {
                "h": 3,
                "w": 3,
                "x": 12,
                "y": 3
              },
              "hideTimeOverride": false,
              "id": 3,
              "interval": null,
              "links": [],
              "mappingType": 1,
              "mappingTypes": [
                {
                  "name": "value to text",
                  "value": 1
                },
                {
                  "name": "range to text",
                  "value": 2
                }
              ],
              "maxDataPoints": 100,
              "nullPointMode": "connected",
              "nullText": null,
              "postfix": "",
              "postfixFontSize": "50%",
              "prefix": "",
              "prefixFontSize": "50%",
              "rangeMaps": [
                {
                  "from": "null",
                  "text": "N/A",
                  "to": "null"
                }
              ],
              "sparkline": {
                "fillColor": "rgba(31, 118, 189, 0.18)",
                "full": false,
                "lineColor": "rgb(31, 120, 193)",
                "show": false
              },
              "tableColumn": "",
              "targets": [
                {
                  "expr": "sum(ALERTS{alertstate=\"pending\",alertname!=\"DeadMansSwitch\"})",
                  "format": "time_series",
                  "intervalFactor": 2,
                  "legendFormat": "",
                  "refId": "A",
                  "step": 600
                }
              ],
              "thresholds": "3, 5",
              "title": "Alerts Pending",
              "transparent": false,
              "type": "singlestat",
              "valueFontSize": "80%",
              "valueMaps": [
                {
                  "op": "=",
                  "text": "0",
                  "value": "null"
                }
              ],
              "valueName": "current"
            }
          ],
          "refresh": "10s",
          "schemaVersion": 16,
          "style": "dark",
          "tags": [],
          "templating": {
            "list": []
          },
          "time": {
            "from": "now-12h",
            "to": "now"
          },
          "timepicker": {
            "refresh_intervals": [
              "5s",
              "10s",
              "30s",
              "1m",
              "5m",
              "15m",
              "30m",
              "1h",
              "2h",
              "1d"
            ],
            "time_options": [
              "5m",
              "15m",
              "1h",
              "6h",
              "12h",
              "24h",
              "2d",
              "7d",
              "30d"
            ]
          },
          "timezone": "browser",
          "title": "Kubernetes Cluster Health (Prometheus)",
          "uid": "NEYiLUzik",
          "version": 1
        }`

const GAPGrafanaSet = `
service:
  type: NodePort
  nodePort: 31002

adminUser: admin
adminPassword: admin

datasources:
  datasources.yaml:
    apiVersion: 1
    datasources:
    - name: MYDS_Prometheus
      type: prometheus
      #url: http:// + 集群里prometheus-server的服务名
      #可以用 kubectl get svc --all-namespaces |grep prometheus-server查看
      url: http://monitor-prometheus-server
      access: proxy
      isDefault: true

dashboardProviders:
  dashboardproviders.yaml:
    apiVersion: 1
    providers:
    - name: 'default'
      orgId: 1
      folder: ''
      type: file
      disableDeletion: false
      editable: true
      options:
        path: /var/lib/grafana/dashboards`

const GAPGrafanaChart = `
name: grafana
version: 1.16.0
appVersion: 5.2.4
kubeVersion: "^1.8.0-0"
description: The leading tool for querying and visualizing time series and metrics.
home: https://grafana.net
icon: https://raw.githubusercontent.com/grafana/grafana/master/public/img/logo_transparent_400x.png
sources:
  - https://github.com/grafana/grafana
maintainers:
  - name: zanhsieh
    email: zanhsieh@gmail.com
  - name: rtluckie
    email: rluckie@cisco.com
engine: gotpl
`

const GAPGrafanaValues = `
rbac:
  create: true
  pspEnabled: true
serviceAccount:
  create: true
  name:

replicas: 1

deploymentStrategy: RollingUpdate

livenessProbe:
  httpGet:
    path: /api/health
    port: 3000

readinessProbe:
  httpGet:
    path: /api/health
    port: 3000
  initialDelaySeconds: 60
  timeoutSeconds: 30
  failureThreshold: 10
  periodSeconds: 10

image:
  repository: caas4/grafana
  tag: 5.2.4
  pullPolicy: IfNotPresent

  ## Optionally specify an array of imagePullSecrets.
  ## Secrets must be manually created in the namespace.
  ## ref: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
  ##
  # pullSecrets:
  #   - myRegistrKeySecretName

securityContext:
  runAsUser: 472
  fsGroup: 472

downloadDashboardsImage:
  repository: caas4/curl
  tag: latest
  pullPolicy: IfNotPresent

## Pod Annotations
# podAnnotations: {}

## Deployment annotations
# annotations: {}

## Expose the grafana service to be accessed from outside the cluster (LoadBalancer service).
## or access it from within the cluster (ClusterIP service). Set the service type and the port to serve it.
## ref: http://kubernetes.io/docs/user-guide/services/
##
service:
  type: ClusterIP
  port: 80
  annotations: {}
  labels: {}

ingress:
  enabled: false
  annotations: {}
    # kubernetes.io/ingress.class: nginx
    # kubernetes.io/tls-acme: "true"
  labels: {}
  path: /
  hosts:
    - chart-example.local
  tls: []
  #  - secretName: chart-example-tls
  #    hosts:
  #      - chart-example.local

resources: {}
#  limits:
#    cpu: 100m
#    memory: 128Mi
#  requests:
#    cpu: 100m
#    memory: 128Mi

## Node labels for pod assignment
## ref: https://kubernetes.io/docs/user-guide/node-selection/
#
nodeSelector: {}

## Tolerations for pod assignment
## ref: https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/
##
tolerations: []

## Affinity for pod assignment
## ref: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity
##
affinity: {}

## Enable persistence using Persistent Volume Claims
## ref: http://kubernetes.io/docs/user-guide/persistent-volumes/
##
persistence:
  enabled: false
  # storageClassName: default
  # accessModes:
  #   - ReadWriteOnce
  # size: 10Gi
  # annotations: {}
  # subPath: ""
  # existingClaim:

adminUser: admin
# adminPassword: strongpassword

## Use an alternate scheduler, e.g. "stork".
## ref: https://kubernetes.io/docs/tasks/administer-cluster/configure-multiple-schedulers/
##
# schedulerName:

## Extra environment variables that will be pass onto deployment pods
env: {}

## The name of a secret in the same kubernetes namespace which contain values to be added to the environment
## This can be useful for auth tokens, etc
envFromSecret: ""

## Additional grafana server secret mounts
# Defines additional mounts with secrets. Secrets must be manually created in the namespace.
extraSecretMounts: []
  # - name: secret-files
  #   mountPath: /etc/secrets
  #   secretName: grafana-secret-files
  #   readOnly: true

## Pass the plugins you want installed as a list.
##
plugins: []
  # - digrich-bubblechart-panel
  # - grafana-clock-panel

## Configure grafana datasources
## ref: http://docs.grafana.org/administration/provisioning/#datasources
##
datasources: {}
#  datasources.yaml:
#    apiVersion: 1
#    datasources:
#    - name: Prometheus
#      type: prometheus
#      url: http://prometheus-prometheus-server
#      access: proxy
#      isDefault: true

## Configure grafana dashboard providers
## ref: http://docs.grafana.org/administration/provisioning/#dashboards
##

##
dashboardProviders: {}
#  dashboardproviders.yaml:
#    apiVersion: 1
#    providers:
#    - name: 'default'
#      orgId: 1
#      folder: ''
#      type: file
#      disableDeletion: false
#      editable: true
#      options:

## Configure grafana dashboard to import
## NOTE: To use dashboards you must also enable/configure dashboardProviders
## ref: https://grafana.com/dashboards
##
## dashboards per provider, use provider name as key.
##
dashboards: {}
#  default:
#    some-dashboard:
#      json: |
#        $RAW_JSON
#    prometheus-stats:
#      gnetId: 2
#      revision: 2
#      datasource: Prometheus
#    local-dashboard:
#      url: https://example.com/repository/test.json

## Reference to external ConfigMap per provider. Use provider name as key and ConfiMap name as value.
## A provider dashboards must be defined either by external ConfigMaps or in values.yaml, not in both.
## ConfigMap data example:
##
## data:
##   example-dashboard.json: |
##     RAW_JSON
##
dashboardsConfigMaps: {}
#  default: ""

## Grafana's primary configuration
## NOTE: values in map will be converted to ini format
## ref: http://docs.grafana.org/installation/configuration/
##
grafana.ini:
  paths:
    data: /var/lib/grafana/data
    logs: /var/log/grafana
    plugins: /var/lib/grafana/plugins
    provisioning: /etc/grafana/provisioning
  analytics:
    check_for_updates: true
  log:
    mode: console
  grafana_net:
    url: https://grafana.net
## LDAP Authentication can be enabled with the following values on grafana.ini
## NOTE: Grafana will fail to start if the value for ldap.toml is invalid
  # auth.ldap:
  #   enabled: true
  #   allow_sign_up: true
  #   config_file: /etc/grafana/ldap.toml

## Grafana's LDAP configuration
## Templated by the template in _helpers.tpl
## NOTE: To enable the grafana.ini must be configured with auth.ldap.enabled
## ref: http://docs.grafana.org/installation/configuration/#auth-ldap
## ref: http://docs.grafana.org/installation/ldap/#configuration
ldap:
  existingSecret: ""
  config: ""
  # config: |-
  #   verbose_logging = true

  #   [[servers]]
  #   host = "my-ldap-server"
  #   port = 636
  #   use_ssl = true
  #   start_tls = false
  #   ssl_skip_verify = false
  #   bind_dn = "uid=%s,ou=users,dc=myorg,dc=com"

## Grafana's SMTP configuration
## NOTE: To enable, grafana.ini must be configured with smtp.enabled
## ref: http://docs.grafana.org/installation/configuration/#smtp
smtp:
  existingSecret: ""

## Sidecars that collect the configmaps with specified label and stores the included files them into the respective folders
## Requires at least Grafana 5 to work and can't be used together with parameters dashboardProviders, datasources and dashboards
sidecar:
  image: caas4/k8s-sidecar:0.0.3
  imagePullPolicy: IfNotPresent
  resources:
#   limits:
#     cpu: 100m
#     memory: 100Mi
#   requests:
#     cpu: 50m
#     memory: 50Mi
  dashboards:
    enabled: false
    # label that the configmaps with dashboards are marked with
    label: grafana_dashboard
    # folder in the pod that should hold the collected dashboards
    folder: /tmp/dashboards
  datasources:
    enabled: false
    # label that the configmaps with datasources are marked with
    label: grafana_datasource`

var GAPGrafanaTemplates = map[string]string{
	"clusterrolebinding.yaml":           clusterrolebinding,
	"clusterrole.yaml":                  clusterrole,
	"configmap_dashboard_provider.yaml": configmap_dashboard_provider,
	"configmap.yaml":                    configmap,
	"dashboards_json_configmap.yaml":    dashboards_json_configmap,
	"deployment.yaml":                   deployment,
	"_helpers.tpl":                      grafana_helpers,
	"ingress.yaml":                      ingress,
	"podsecuritypolicy.yaml":            podsecuritypolicy,
	"pvc.yaml":                          pvc,
	"rolebinding.yaml":                  rolebinding,
	"role.yaml":                         role,
	"secret":                            secret,
	"serviceaccount.yaml":               serviceaccount,
	"service.yaml":                      service,
}

const clusterrolebinding = `
{{- if .Values.rbac.create }}
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: {{ template "grafana.fullname" . }}-clusterrolebinding
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ template "grafana.chart" . }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
{{- with .Values.annotations }}
  annotations:
{{ toYaml . | indent 4 }}
{{- end }}
subjects:
  - kind: ServiceAccount
    name: {{ template "grafana.serviceAccountName" . }}
    namespace: {{ .Release.Namespace }}
roleRef:
  kind: ClusterRole
  name: {{ template "grafana.fullname" . }}-clusterrole
  apiGroup: rbac.authorization.k8s.io
{{- end}}`
const clusterrole = `
{{- if .Values.rbac.create }}
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ template "grafana.chart" . }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
{{- with .Values.annotations }}
  annotations:
{{ toYaml . | indent 4 }}
{{- end }}
  name: {{ template "grafana.fullname" . }}-clusterrole
{{- if or .Values.sidecar.dashboards.enabled .Values.sidecar.datasources.enabled }}
rules:
- apiGroups: [""] # "" indicates the core API group
  resources: ["configmaps"]
  verbs: ["get", "watch", "list"]
{{- else }}
rules: []
{{- end}}
{{- end}}`
const configmap_dashboard_provider = `
{{- if .Values.sidecar.dashboards.enabled }}
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ template "grafana.chart" . }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
{{- with .Values.annotations }}
  annotations:
{{ toYaml . | indent 4 }}
{{- end }}
  name: {{ template "grafana.fullname" . }}-config-dashboards
data:
  provider.yaml: |-
    apiVersion: 1
    providers:
    - name: 'default'
      orgId: 1
      folder: ''
      type: file
      disableDeletion: false
      options:
        path: {{ .Values.sidecar.dashboards.folder }}
{{- end}}`
const configmap = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ template "grafana.fullname" . }}
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ template "grafana.chart" . }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
data:
{{- if .Values.plugins }}
  plugins: {{ join "," .Values.plugins }}
{{- end }}
  grafana.ini: |
{{- range $key, $value := index .Values "grafana.ini" }}
    [{{ $key }}]
    {{- range $elem, $elemVal := $value }}
    {{ $elem }} = {{ $elemVal }}
    {{- end }}
{{- end }}

{{- if .Values.datasources }}
  {{- range $key, $value := .Values.datasources }}
  {{ $key }}: |
{{ toYaml $value | indent 4 }}
  {{- end -}}
{{- end -}}

{{- if .Values.dashboardProviders }}
  {{- range $key, $value := .Values.dashboardProviders }}
  {{ $key }}: |
{{ toYaml $value | indent 4 }}
  {{- end -}}
{{- end -}}

{{- if .Values.dashboards  }}
  download_dashboards.sh: |
    #!/usr/bin/env sh
    set -euf
    {{- if .Values.dashboardProviders }}
      {{- range $key, $value := .Values.dashboardProviders }}
        {{- range $value.providers }}
    mkdir -p {{ .options.path }}
        {{- end }}
      {{- end }}
    {{- end }}

  {{- range $provider, $dashboards := .Values.dashboards }}
    {{- range $key, $value := $dashboards }}
      {{- if (or (hasKey $value "gnetId") (hasKey $value "url")) }}
    curl -sk \
    --connect-timeout 60 \
    --max-time 60 \
    -H "Accept: application/json" \
    -H "Content-Type: application/json;charset=UTF-8" \
    {{- if $value.url -}}{{ $value.url }}{{- else -}} https://grafana.com/api/dashboards/{{ $value.gnetId }}/revisions/{{- if $value.revision -}}{{ $value.revision }}{{- else -}}1{{- end -}}/download{{- end -}}{{ if $value.datasource }}| sed 's|\"datasource\":[^,]*|\"datasource\": \"{{ $value.datasource }}\"|g'{{ end }} \
    > /var/lib/grafana/dashboards/{{ $provider }}/{{ $key }}.json
      {{- end }}
    {{- end }}
  {{- end }}
{{- end }}`
const dashboards_json_configmap = `
{{- if .Values.dashboards }}
  {{- range $provider, $dashboards := .Values.dashboards }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ template "grafana.fullname" $ }}-dashboards-{{ $provider }}
  labels:
    app: {{ template "grafana.name" $ }}
    chart: {{ template "grafana.chart" $ }}
    release: {{ $.Release.Name }}
    heritage: {{ $.Release.Service }}
    dashboard-provider: {{ $provider }}
data:
  {{- range $key, $value := $dashboards }}
    {{- if hasKey $value "json" }}
  {{ $key }}.json: |
{{ $value.json | indent 4 }}
    {{- end }}
  {{- end }}
  {{- end }}
{{- end }}
`
const deployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ template "grafana.fullname" . }}
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ template "grafana.chart" . }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
{{- with .Values.annotations }}
  annotations:
{{ toYaml . | indent 4 }}
{{- end }}
spec:
  replicas: {{ .Values.replicas }}
  selector:
    matchLabels:
      app: {{ template "grafana.name" . }}
      release: {{ .Release.Name }}
  strategy:
    type: {{ .Values.deploymentStrategy }}
  {{- if ne .Values.deploymentStrategy "RollingUpdate" }}
    rollingUpdate: null
  {{- end }}
  template:
    metadata:
      labels:
        app: {{ template "grafana.name" . }}
        release: {{ .Release.Name }}
{{- with .Values.podAnnotations }}
      annotations:
{{ toYaml . | indent 8 }}
{{- end }}
    spec:
      serviceAccountName: {{ template "grafana.serviceAccountName" . }}
{{- if .Values.schedulerName }}
      schedulerName: "{{ .Values.schedulerName }}"
{{- end }}
{{- if .Values.securityContext }}
      securityContext:
{{ toYaml .Values.securityContext | indent 8 }}
{{- end }}
{{- if .Values.dashboards }}
      initContainers:
        - name: download-dashboards
          image: "{{ .Values.downloadDashboardsImage.repository }}:{{ .Values.downloadDashboardsImage.tag }}"
          imagePullPolicy: {{ .Values.downloadDashboardsImage.pullPolicy }}
          command: ["sh", "/etc/grafana/download_dashboards.sh"]
          volumeMounts:
            - name: config
              mountPath: "/etc/grafana/download_dashboards.sh"
              subPath: download_dashboards.sh
            - name: storage
              mountPath: "/var/lib/grafana"
              subPath: {{ .Values.persistence.subPath }}
          {{- range .Values.extraSecretMounts }}
            - name: {{ .name }}
              mountPath: {{ .mountPath }}
              readOnly: {{ .readOnly }}
          {{- end }}
{{- end }}
      {{- if .Values.image.pullSecrets }}
      imagePullSecrets:
      {{- range .Values.image.pullSecrets }}
        - name: {{ . }}
      {{- end}}
      {{- end }}
      containers:
{{- if .Values.sidecar.dashboards.enabled }}
        - name: {{ template "grafana.name" . }}-sc-dashboard
          image: "{{ .Values.sidecar.image }}"
          imagePullPolicy: {{ .Values.sidecar.imagePullPolicy }}
          env:
            - name: LABEL
              value: "{{ .Values.sidecar.dashboards.label }}"
            - name: FOLDER
              value: "{{ .Values.sidecar.dashboards.folder }}"
          resources:
{{ toYaml .Values.sidecar.resources | indent 12 }}
          volumeMounts:
            - name: sc-dashboard-volume
              mountPath: {{ .Values.sidecar.dashboards.folder | quote }}
{{- end}}
{{- if .Values.sidecar.datasources.enabled }}
        - name: {{ template "grafana.name" . }}-sc-datasources
          image: "{{ .Values.sidecar.image }}"
          imagePullPolicy: {{ .Values.sidecar.imagePullPolicy }}
          env:
            - name: LABEL
              value: "{{ .Values.sidecar.datasources.label }}"
            - name: FOLDER
              value: "/etc/grafana/provisioning/datasources"
          resources:
{{ toYaml .Values.sidecar.resources | indent 12 }}
          volumeMounts:
            - name: sc-datasources-volume
              mountPath: "/etc/grafana/provisioning/datasources"
{{- end}}
        - name: {{ .Chart.Name }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          volumeMounts:
            - name: config
              mountPath: "/etc/grafana/grafana.ini"
              subPath: grafana.ini
            - name: ldap
              mountPath: "/etc/grafana/ldap.toml"
              subPath: ldap.toml
{{- if .Values.dashboards }}
  {{- range $provider, $dashboards := .Values.dashboards }}
    {{- range $key, $value := $dashboards }}
      {{- if hasKey $value "json" }}
            - name: dashboards-{{ $provider }}
              mountPath: "/var/lib/grafana/dashboards/{{ $provider }}/{{ $key }}.json"
              subPath: "{{ $key }}.json"
      {{- end }}
    {{- end }}
  {{- end }}
{{- end -}}
{{- if .Values.dashboardsConfigMaps }}
  {{- range keys .Values.dashboardsConfigMaps }}
            - name: dashboards-{{ . }}
              mountPath: "/var/lib/grafana/dashboards/{{ . }}"
  {{- end }}
{{- end }}
{{- if .Values.datasources }}
            - name: config
              mountPath: "/etc/grafana/provisioning/datasources/datasources.yaml"
              subPath: datasources.yaml
{{- end }}
{{- if .Values.dashboardProviders }}
            - name: config
              mountPath: "/etc/grafana/provisioning/dashboards/dashboardproviders.yaml"
              subPath: dashboardproviders.yaml
{{- end }}
{{- if .Values.sidecar.dashboards.enabled }}
            - name: sc-dashboard-volume
              mountPath: {{ .Values.sidecar.dashboards.folder | quote }}
            - name: sc-dashboard-provider
              mountPath: "/etc/grafana/provisioning/dashboards/sc-dashboardproviders.yaml"
              subPath: provider.yaml
{{- end}}
{{- if .Values.sidecar.datasources.enabled }}
            - name: sc-datasources-volume
              mountPath: "/etc/grafana/provisioning/datasources"
{{- end}}
            - name: storage
              mountPath: "/var/lib/grafana"
              subPath: {{ .Values.persistence.subPath }}
          {{- range .Values.extraSecretMounts }}
            - name: {{ .name }}
              mountPath: {{ .mountPath }}
              readOnly: {{ .readOnly }}
          {{- end }}
          ports:
            - name: service
              containerPort: {{ .Values.service.port }}
              protocol: TCP
            - name: grafana
              containerPort: 3000
              protocol: TCP
          env:
            - name: GF_SECURITY_ADMIN_USER
              valueFrom:
                secretKeyRef:
                  name: {{ template "grafana.fullname" . }}
                  key: admin-user
            - name: GF_SECURITY_ADMIN_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{ template "grafana.fullname" . }}
                  key: admin-password
            {{- if .Values.plugins }}
            - name: GF_INSTALL_PLUGINS
              valueFrom:
                configMapKeyRef:
                  name: {{ template "grafana.fullname" . }}
                  key: plugins
            {{- end }}
            {{- if .Values.smtp.existingSecret }}
            - name: GF_SMTP_USER
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.smtp.existingSecret }}
                  key: user
            - name: GF_SMTP_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.smtp.existingSecret }}
                  key: password
            {{- end }}
{{- range $key, $value := .Values.env }}
            - name: "{{ $key }}"
              value: "{{ $value }}"
{{- end }}
          {{- if .Values.envFromSecret }}
          envFrom:
            - secretRef:
                name: {{ .Values.envFromSecret }}
          {{- end }}
          livenessProbe:
{{ toYaml .Values.livenessProbe | indent 12 }}
          readinessProbe:
{{ toYaml .Values.readinessProbe | indent 12 }}
          resources:
{{ toYaml .Values.resources | indent 12 }}
    {{- with .Values.nodeSelector }}
      nodeSelector:
{{ toYaml . | indent 8 }}
    {{- end }}
    {{- with .Values.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
    {{- end }}
    {{- with .Values.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
    {{- end }}
      volumes:
        - name: config
          configMap:
            name: {{ template "grafana.fullname" . }}
        {{- if .Values.dashboards }}
          {{- range keys .Values.dashboards }}
        - name: dashboards-{{ . }}
          configMap:
            name: {{ template "grafana.fullname" $ }}-dashboards-{{ . }}
          {{- end }}
        {{- end }}
        {{- if .Values.dashboardsConfigMaps }}
          {{- range $provider, $name := .Values.dashboardsConfigMaps }}
        - name: dashboards-{{ $provider }}
          configMap:
            name: {{ $name }}
          {{- end }}
        {{- end }}
        - name: ldap
          secret:
            {{- if .Values.ldap.existingSecret }}
            secretName: {{ .Values.ldap.existingSecret }}
            {{- else }}
            secretName: {{ template "grafana.fullname" . }}
            {{- end }}
            items:
              - key: ldap-toml
                path: ldap.toml
        - name: storage
      {{- if .Values.persistence.enabled }}
          persistentVolumeClaim:
            claimName: {{ .Values.persistence.existingClaim | default (include "grafana.fullname" .) }}
      {{- else }}
          emptyDir: {}
      {{- end -}}
      {{- if .Values.sidecar.dashboards.enabled }}
        - name: sc-dashboard-volume
          emptyDir: {}
        - name: sc-dashboard-provider
          configMap:
            name: {{ template "grafana.fullname" . }}-config-dashboards
      {{- end }}
      {{- if .Values.sidecar.datasources.enabled }}
        - name: sc-datasources-volume
          emptyDir: {}
      {{- end -}}
      {{- range .Values.extraSecretMounts }}
        - name: {{ .name }}
          secret:
            secretName: {{ .secretName }}
            defaultMode: {{ .defaultMode }}
      {{- end }}`
const grafana_helpers = `
{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "grafana.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "grafana.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "grafana.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the service account
*/}}
{{- define "grafana.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "grafana.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}`

const ingress = `
{{- if .Values.ingress.enabled -}}
{{- $fullName := include "grafana.fullname" . -}}
{{- $servicePort := .Values.service.port -}}
{{- $ingressPath := .Values.ingress.path -}}
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: {{ $fullName }}
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ template "grafana.chart" . }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
{{- if .Values.ingress.labels }}
{{ toYaml .Values.ingress.labels | indent 4 }}
{{- end }}
{{- with .Values.ingress.annotations }}
  annotations:
{{ toYaml . | indent 4 }}
{{- end }}
spec:
{{- if .Values.ingress.tls }}
  tls:
  {{- range .Values.ingress.tls }}
    - hosts:
      {{- range .hosts }}
        - {{ . | quote }}
      {{- end }}
      secretName: {{ .secretName }}
  {{- end }}
{{- end }}
  rules:
  {{- range .Values.ingress.hosts }}
    - host: {{ . }}
      http:
        paths:
          - path: {{ $ingressPath }}
            backend:
              serviceName: {{ $fullName }}
              servicePort: {{ $servicePort }}
  {{- end }}
{{- end }}`

const podsecuritypolicy = `
{{- if .Values.rbac.pspEnabled }}
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: {{ template "grafana.fullname" . }}
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  annotations:
    seccomp.security.alpha.kubernetes.io/allowedProfileNames: 'docker/default'
    apparmor.security.beta.kubernetes.io/allowedProfileNames: 'runtime/default'
    seccomp.security.alpha.kubernetes.io/defaultProfileName:  'docker/default'
    apparmor.security.beta.kubernetes.io/defaultProfileName:  'runtime/default'
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  hostNetwork: false
  hostIPC: false
  hostPID: false
  runAsUser:
    rule: 'RunAsAny'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
  readOnlyRootFilesystem: false
{{- end }}`

const pvc = `
{{- if and .Values.persistence.enabled (not .Values.persistence.existingClaim) }}
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ template "grafana.fullname" . }}
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ template "grafana.chart" . }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
  {{- with .Values.persistence.annotations  }}
  annotations:
{{ toYaml . | indent 4 }}
  {{- end }}
spec:
  accessModes:
    {{- range .Values.persistence.accessModes }}
    - {{ . | quote }}
    {{- end }}
  resources:
    requests:
      storage: {{ .Values.persistence.size | quote }}
  storageClassName: {{ .Values.persistence.storageClassName }}
{{- end -}}`

const rolebinding = `
{{- if .Values.rbac.create -}}
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: {{ template "grafana.fullname" . }}
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: {{ template "grafana.fullname" . }}
subjects:
- kind: ServiceAccount
  name: {{ template "grafana.serviceAccountName" . }}
{{- end -}}`

const role = `
{{- if .Values.rbac.create }}
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: {{ template "grafana.fullname" . }}
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
{{- if .Values.rbac.pspEnabled }}
rules:
- apiGroups:      ['policy']
  resources:      ['podsecuritypolicies']
  verbs:          ['use']
  resourceNames:  [{{ template "grafana.fullname" . }}]
{{- end }}
{{- end }}`

const secret = `
apiVersion: v1
kind: Secret
metadata:
  name: {{ template "grafana.fullname" . }}
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ template "grafana.chart" . }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
type: Opaque
data:
  admin-user: {{ .Values.adminUser | b64enc | quote }}
  {{- if .Values.adminPassword }}
  admin-password: {{ .Values.adminPassword | b64enc | quote }}
  {{- else }}
  admin-password: {{ randAlphaNum 40 | b64enc | quote }}
  {{- end }}
  {{- if not .Values.ldap.existingSecret }}
  ldap-toml: {{ .Values.ldap.config | b64enc | quote }}
  {{- end }}`

const serviceaccount = `
{{- if .Values.serviceAccount.create }}
apiVersion: v1
kind: ServiceAccount
metadata:
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ .Chart.Name }}-{{ .Chart.Version }}
    heritage: {{ .Release.Service }}
    release: {{ .Release.Name }}
  name: {{ template "grafana.serviceAccountName" . }}
{{- end }}`

const service = `
apiVersion: v1
kind: Service
metadata:
  name: {{ template "grafana.fullname" . }}
  labels:
    app: {{ template "grafana.name" . }}
    chart: {{ template "grafana.chart" . }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
{{- if .Values.service.labels }}
{{ toYaml .Values.service.labels | indent 4 }}
{{- end }}
{{- with .Values.service.annotations }}
  annotations:
{{ toYaml . | indent 4 }}
{{- end }}
spec:
{{- if (or (eq .Values.service.type "ClusterIP") (empty .Values.service.type)) }}
  type: ClusterIP
  {{- if .Values.service.clusterIP }}
  clusterIP: {{ .Values.service.clusterIP }}
  {{end}}
{{- else if eq .Values.service.type "LoadBalancer" }}
  type: {{ .Values.service.type }}
  {{- if .Values.service.loadBalancerIP }}
  loadBalancerIP: {{ .Values.service.loadBalancerIP }}
  {{- end }}
  {{- if .Values.service.loadBalancerSourceRanges }}
  loadBalancerSourceRanges:
{{ toYaml .Values.service.loadBalancerSourceRanges | indent 4 }}
  {{- end -}}
{{- else }}
  type: {{ .Values.service.type }}
{{- end }}
{{- if .Values.service.externalIPs }}
  externalIPs:
{{ toYaml .Values.service.externalIPs | indent 4 }}
{{- end }}
  ports:
    - name: service
      port: {{ .Values.service.port }}
      protocol: TCP
      targetPort: 3000
{{ if (and (eq .Values.service.type "NodePort") (not (empty .Values.service.nodePort))) }}
      nodePort: {{.Values.service.nodePort}}
{{ end }}
  selector:
    app: {{ template "grafana.name" . }}
    release: {{ .Release.Name }}`
