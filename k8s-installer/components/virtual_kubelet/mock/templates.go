package mock

const mockConfig = `{
    "{{.Hostname}}": {
        "cpu": "{{.Mock.CpuLimit}}",
        "memory": "{{.Mock.MemoryLimit}}",
        "pods": "{{.Mock.PodLimit}}"
    }
}`

const clusterRoleBinding1 = `apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: {{.Hostname}}
  namespace: ""
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - apiGroup: rbac.authorization.k8s.io
    kind: User
    name: {{.Hostname}}`

const ClusterRoleBinding2 = `apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: system:kube-nodes
  namespace: ""
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:node-proxier
subjects:
  - apiGroup: rbac.authorization.k8s.io
    kind: Group
    name: system:nodes`

const systemdTemplate = `[Unit]
Description=k8s-install-client
After=network-online.target

[Service]
Environment="HOME=/root"
Type=simple
Restart=on-failure
RestartSec=5s
TimeoutStartSec=0
ExecStart=/usr/bin/virtual-kubelet --provider mock \
--nodename {{.Hostname}} \
--provider-config /etc/virtual_kubelet/config.json \
--kubeconfig /etc/kubernetes/kubelet.conf
ExecReload=/bin/kill -HUP 
KillMode=process

[Install]
WantedBy=multi-user.target`