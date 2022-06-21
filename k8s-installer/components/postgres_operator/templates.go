package postgres_operator

const Template = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{.Namespace}}

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pg-operator
  namespace: {{.Namespace}}

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: pg-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: pg-operator
    namespace: {{.Namespace}}

---
apiVersion: batch/v1
kind: Job
metadata:
  name: pgo-deploy
  namespace: {{.Namespace}}
spec:
  backoffLimit: 0
  template:
    metadata:
      name: pgo-deploy
    spec:
      serviceAccountName: pg-operator
      restartPolicy: Never
      containers:
        - name: pgo-deploy
          command: ["/pgo-deploy.sh"]
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/pgo-deployer:centos7-4.3.0
          imagePullPolicy: IfNotPresent
          env:
            - name: ANSIBLE_CONFIG
              value: "/ansible/ansible.cfg"
            - name: ARCHIVE_MODE
              value: "true"
            - name: ARCHIVE_TIMEOUT
              value: "60"
            - name: BACKREST
              value: "true"
            - name: BADGER
              value: "false"
            - name: CRUNCHY_DEBUG
              value: "false"
            - name: CREATE_RBAC
              value: "true"
            - name: CCP_IMAGE_PREFIX
              value: "{{with .ImageRegistry}}{{.}}/{{end}}caas4"
            - name: CCP_IMAGE_TAG
              value: "centos7-12.2-4.3.0"
            - name: DB_PASSWORD_LENGTH
              value: "24"
            - name: DB_PORT
              value: "5432"
            - name: DB_REPLICAS
              value: "0"
            - name: DB_USER
              value: "testuser"
            - name: DEFAULT_INSTANCE_MEMORY
              value: "128Mi"
            - name: DEFAULT_PGBACKREST_MEMORY
              value: ""
            - name: DEFAULT_PGBOUNCER_MEMORY
              value: ""
            - name: DEPLOY_ACTION
              value: "install"
            - name: DISABLE_AUTO_FAILOVER
              value: "false"
            - name: EXPORTERPORT
              value: "9187"
            - name: HOME
              value: "/tmp"
            - name: METRICS
              value: "false"
            - name: NAMESPACE
              value: {{.WatchNamespaces}}
            - name: NAMESPACE_MODE
              value: "dynamic"
            - name: PGBADGERPORT
              value: "10000"
            - name: PGO_ADMIN_PASSWORD
              value: "{{.AdminPassword}}"
            - name: PGO_ADMIN_PERMS
              value: "*"
            - name: PGO_ADMIN_ROLE_NAME
              value: "pgoadmin"
            - name: PGO_ADMIN_USERNAME
              value: "{{.AdminUser}}"
            - name: PGO_CLIENT_VERSION
              value: "v4.3.0"
            - name: PGO_IMAGE_PREFIX
              value: "{{with .ImageRegistry}}{{.}}/{{end}}caas4"
            - name: PGO_IMAGE_TAG
              value: "centos7-4.3.0"
            - name: PGO_INSTALLATION_NAME
              value: "devtest"
            - name: PGO_OPERATOR_NAMESPACE
              value: "{{.Namespace}}"
            - name: SCHEDULER_TIMEOUT
              value: "3600"
            - name: BACKREST_STORAGE
              value: "primarysite"
            - name: BACKUP_STORAGE
              value: "primarysite"
            - name: PRIMARY_STORAGE
              value: "primarysite"
            - name: REPLICA_STORAGE
              value: "primarysite"
            - name: WAL_STORAGE
              value: ""
            - name: STORAGE6_NAME
              value: "primarysite"
            - name: STORAGE6_ACCESS_MODE
              value: "ReadWriteOnce"
            - name: STORAGE6_SIZE
              value: "{{.StorageSize}}Gi"
            - name: STORAGE6_TYPE
              value: "dynamic"
            - name: STORAGE6_CLASS
              value: "{{.StorageClass}}"
            - name: STORAGE6_SUPPLEMENTAL_GROUPS
              value: "65534"
            - name: PGO_DISABLE_TLS
              value: "true"
            - name: PGO_APISERVER_PORT
              value: "8000"
            - name: PGO_TLS_NO_VERIFY
              value: "true"
            - name: PGO_CLIENT_CONTAINER_INSTALL
              value: "true"
            - name: PGO_APISERVER_URL
              value: "http://postgres-operator"
`
