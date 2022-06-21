package dns

const Corefile = `.:53 {
    caas /etc/coredns/config.yaml
    log
    errors
{{if .Upstream }}
    forward . {{.Upstream}}
{{end}}
}`

const CaasPluginConfig = `server-name: {{.ServerName}}
server-ip: {{.ServerIp}}
initial-data-provider-service-url: {{.DataProviderURL}}
message-queues:
  name: nat
  auth-mode: basic
  address: {{.MQServers}}
  port: {{.MQPort}}
  subject: {{.MQSubject}}
  user-name: {{.MQUsername}}
  password: {{.MQPassword}}
mode: {{.Mode}}
ask-frequency: {{.AskFrequency}}`

const StaticPodTemplate = `apiVersion: v1
kind: Pod
metadata:
  labels:
    caas: dns
  name: caas-dns
  namespace: kube-system
spec:
  containers:
  - args:
    - -conf
    - /etc/coredns/Corefile
    image: {{.ImageRegistry}}/coredns:v1.8.3
    imagePullPolicy: IfNotPresent
    name: coredns
    ports:
    - containerPort: 53
      name: dns
      protocol: UDP
    - containerPort: 53
      name: dns-tcp
      protocol: TCP
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        add:
        - NET_BIND_SERVICE
        drop:
        - all
      periodSeconds: 10
      successThreshold: 1
      timeoutSeconds: 1
    resources:
      limits:
        memory: 2048Mi
      readOnlyRootFilesystem: true
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /etc/coredns
      name: dns-config
  dnsPolicy: Default
  priorityClassName: system-cluster-critical
  restartPolicy: Always
  hostNetwork: true
  volumes:
  - hostPath:
      path: /etc/coredns
      type: DirectoryOrCreate
    name: dns-config`

const ClusterCorednsConfig = `apiVersion: v1
data:
  Corefile: ".:53 {\n    errors\n    health {\n       lameduck 5s\n    }\n    ready\n
    \   kubernetes {{.ClusterName}} in-addr.arpa ip6.arpa {\n       pods insecure\n
    \      fallthrough in-addr.arpa ip6.arpa\n       ttl 30\n    }\n    prometheus
    :9153\n    forward . {{.CombinedAddress}} \n    cache 30\n    loop\n    reload\n    loadbalance\n}\n"
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system`