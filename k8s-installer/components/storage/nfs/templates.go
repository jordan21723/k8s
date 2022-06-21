package nfs

const NamespaceTemplate = `apiVersion: v1
kind: Namespace
metadata:
  name: nfs-client-provisioner`

const StorageClassTemplate = `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{.StorageClassName}}
{{if .IsDefaultSc}}
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
{{end}}
provisioner: fuseim.pri/ifs # or choose another name, must match deployment's env PROVISIONER_NAME'
parameters:
  archiveOnDelete: "{{.ReclaimPolicy}}" # When set to "false" your PVs will not be archive
{{if gt (len .MountOptions) 0}}
mountOptions:
{{range $i, $v := .MountOptions}}
  - {{$v}}
{{end}}
{{end}}`

const Template = `kind: Deployment
apiVersion: apps/v1
metadata:
  name: nfs-client-provisioner
  namespace: nfs-client-provisioner
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nfs-client-provisioner
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: nfs-client-provisioner
    spec:
      serviceAccountName: nfs-client-provisioner
      containers:
        - name: nfs-client-provisioner
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/nfs-client-provisioner:v3.1.0-k8s1.11
          volumeMounts:
            - name: nfs-client-root
              mountPath: /persistentvolumes
          env:
            - name: PROVISIONER_NAME
              value: fuseim.pri/ifs
            - name: NFS_SERVER
              value: {{.NFSServerAddress}}
            - name: NFS_PATH
              value: {{.NFSPath}}
      volumes:
        - name: nfs-client-root
          nfs:
            server: {{.NFSServerAddress}}
            path: {{.NFSPath}}`

const RBAC = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: nfs-client-provisioner
  # replace with namespace where provisioner is deployed
  namespace: nfs-client-provisioner
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: nfs-client-provisioner-runner
rules:
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["create", "update", "patch"]
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: run-nfs-client-provisioner
subjects:
  - kind: ServiceAccount
    name: nfs-client-provisioner
    # replace with namespace where provisioner is deployed
    namespace: nfs-client-provisioner
roleRef:
  kind: ClusterRole
  name: nfs-client-provisioner-runner
  apiGroup: rbac.authorization.k8s.io
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: leader-locking-nfs-client-provisioner
  # replace with namespace where provisioner is deployed
  namespace: nfs-client-provisioner
rules:
  - apiGroups: [""]
    resources: ["endpoints"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: leader-locking-nfs-client-provisioner
  # replace with namespace where provisioner is deployed
  namespace: nfs-client-provisioner
subjects:
  - kind: ServiceAccount
    name: nfs-client-provisioner
    # replace with namespace where provisioner is deployed
    namespace: nfs-client-provisioner
roleRef:
  kind: Role
  name: leader-locking-nfs-client-provisioner
  apiGroup: rbac.authorization.k8s.io`
