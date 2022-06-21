package middle_platform

const Template = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{.Namespace}}

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kube-resource
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: kube-resource
    namespace: {{.Namespace}}

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-resource
  namespace: {{.Namespace}}

---
apiVersion: v1
kind: Secret
metadata:
  name: middle-admin-pass
  namespace: {{.Namespace}}
type: Opaque
data:
  password: {{b64enc .SystemAdminPassword}}

---
apiVersion: v1
kind: Secret
metadata:
  name: postgres-pass
  namespace: {{.Namespace}}
type: Opaque
data:
  password.txt: {{b64enc .DB.Password}}

{{if .LDAP}}
---
apiVersion: v1
kind: Secret
metadata:
  name: ldap-secret
  namespace: {{.Namespace}}
type: Opaque
data:
  ldap_password: {{if .LDAP.CaaSDeploy}}{{b64enc .LDAP.CaaSDeploy.AdminPassword}}{{else}}{{b64enc .LDAP.AdminPassword}}{{end}}
{{end}}

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: etcd-deployment
  namespace: {{.Namespace}}
  labels:
    app: etcd
spec:
  replicas: 1
  selector:
    matchLabels:
      app: etcd
  template:
    metadata:
      labels:
        app: etcd
    spec:
      containers:
        - name: etcd
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/etcd:3.4.4
          imagePullPolicy: IfNotPresent
          env:
            - name: ALLOW_NONE_AUTHENTICATION
              value: "yes"
          ports:
            - containerPort: 2379
              name: etcd1
            - containerPort: 2380
              name: etcd2
          livenessProbe:
            failureThreshold: 3
            periodSeconds: 10
            successThreshold: 1
            tcpSocket:
              port: 2379
            timeoutSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: etcd
  namespace: {{.Namespace}}
  labels:
    app: etcd
spec:
  ports:
    - port: 2379
      targetPort: 2379
      protocol: TCP
      name: etcd1
    - port: 2380
      targetPort: 2380
      protocol: TCP
      name: etcd2
  selector:
    app: etcd

---
apiVersion: batch/v1
kind: Job
metadata:
  labels:
    app: middle-platform
    job: middle-platform-db-init
  name: middle-platform-db-init-job
  namespace: {{.Namespace}}
spec:
  backoffLimit: 6
  completions: 1
  parallelism: 1
  template:
    metadata:
      labels:
        app: middle-platform
        job: middle-platform-db-init
      name: middle-platform-db-init
    spec:
      serviceAccountName: kube-resource
      initContainers:
        - name: wait4pgo
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/k8s-wait-for:v1.3
          imagePullPolicy: IfNotPresent
          args:
            - service
            - postgres-operator
            - -n
            - {{.PGO.Namespace}}
        - name: create-cluster
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/curl:latest
          imagePullPolicy: IfNotPresent
          command:
            - sh
          args:
            - -c
            - |
              set -ex
              curl \
                -u {{.PGO.AdminUser}}:{{.PGO.AdminPassword}} \
                -H "Content-Type:application/json" \
                -X POST \
                -d '{"AutofailFlag":true, "ClientVersion":"4.3.0", "Namespace":"{{.PGO.Namespace}}", "ReplicaCount":{{.DB.Replicas}}, "Name":"{{.DB.Host}}", "Username":"{{.DB.User}}", "Password":"{{.DB.Password}}", "Series":1, "Database":"{{.DB.Name}}", "PasswordSuperuser":"{{.DB.Password}}"}' \
                http://postgres-operator.{{.PGO.Namespace}}.svc:8000/clusters
        - name: wait4cluster
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/k8s-wait-for:v1.3
          imagePullPolicy: IfNotPresent
          args:
            - service
            - {{.DB.Host}}-replica
            - -n
            - {{.PGO.Namespace}}
      containers:
        - name: create-db
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/pgo-backrest:centos7-4.3.0
          imagePullPolicy: IfNotPresent
          command:
            - sh
          args:
            - -c
            - |
              set -ex
              if psql -lqt | cut -d\| -f 1 | grep -qw {{.DB.Name}}; then
                echo db {{.DB.Name}} exsits
              else
                echo creating db {{.DB.Name}}
                psql -c "CREATE DATABASE {{.DB.Name}}"
              fi
          env:
            - name: PGHOST
              value: {{.DB.Host}}.{{.PGO.Namespace}}.svc
            - name: PGUSER
              value: postgres
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-pass
                  key: password.txt
      dnsPolicy: ClusterFirst
      restartPolicy: OnFailure
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: caddy-config
  namespace: {{.Namespace}}
  labels:
    app: apigateway
data:
  Caddyfile: |
    :2018 {
      root /home
      {{if .EnableLicense}}license{{end}}
      authenticate {
        path /
        auth-service-name go.micro.middle-platform.auth
        service-registry-addr etcd:2379
        except * /capi/iam.io/v1/login /swagger
      }
      authorize {
          path /
          auth-service-name go.micro.middle-platform.auth
          service-registry-addr etcd:2379
          except * /capi/iam.io/v1/login /capi/iam.io/v1/current
      }
      audit {
          path /
          except GET /api /apis /capi/iam.io /capi/tenant.io /capi/resources.io
          except * /swagger
      }
      swagger
      # iam
      proxy /capi/iam.io http://api.{{.Namespace}}.svc {
        transparent
      }
      # tenant
      proxy /capi/tenant.io http://api.{{.Namespace}}.svc {
        transparent
      }
      proxy /capi/resources.io http://api.{{.Namespace}}.svc {
        transparent
      }
      proxy /capi/audit.io http://api.{{.Namespace}}.svc {
        transparent
      }
      # k8s api
      proxy /api https://kubernetes.default {
        header_upstream Authorization "Bearer {$KUBERNETES_TOKEN}"
        insecure_skip_verify
        transparent
        websocket
      }
      log / stdout "{remote} {when} {method} {uri} {proto} {status} {size} {latency_ms}ms"
    }

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apigateway-deployment
  namespace: {{.Namespace}}
  labels:
    app: apigateway
spec:
  replicas: 1
  selector:
    matchLabels:
      app: apigateway
  template:
    metadata:
      labels:
        app: apigateway
    spec:
      containers:
        - command:
            - /bin/sh
            - -c
            - export KUBERNETES_TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
              && api-gateway --conf=/etc/caddy/Caddyfile --log=stderr
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/middle-platform-api-gateway:caas-v1.0
          imagePullPolicy: IfNotPresent
          name: apigateway
          ports:
            - containerPort: 2018
              name: http
              protocol: TCP
          volumeMounts:
            - mountPath: /etc/caddy
              name: apigateway-config-volume
      serviceAccountName: kube-resource
      volumes:
        - configMap:
            defaultMode: 420
            items:
              - key: Caddyfile
                path: Caddyfile
            name: caddy-config
          name: apigateway-config-volume

---
apiVersion: v1
kind: Service
metadata:
  name: apigateway
  namespace: {{.Namespace}}
  labels:
    app: apigateway
spec:
  selector:
    app: apigateway
  ports:
    - name: http
      port: 80
      protocol: TCP
      targetPort: 2018
  type: ClusterIP

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-deployment
  namespace: {{.Namespace}}
  labels:
    app: api
spec:
  replicas: 1
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      initContainers:
        - name: wait4etcd
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/k8s-wait-for:v1.3
          imagePullPolicy: IfNotPresent
          args:
            - pod
            - -lapp=etcd
        - name: wait4cluster
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/k8s-wait-for:v1.3
          imagePullPolicy: IfNotPresent
          args:
            - service
            - {{.DB.Host}}-replica
            - -n
            - {{.PGO.Namespace}}
      containers:
        - command:
            - api
            - --logtostderr=true
          env:
            - name: DB_HOST
              value: {{.DB.Host}}.{{.PGO.Namespace}}.svc
            - name: DB_PORT
              value: "5432"
            - name: DB_TYPE
              value: postgres
            - name: DB_USER
              value: postgres
            - name: DB_MAX_IDLE_CONN
              value: "9"
            - name: DB_MAX_OPEN_CONN
              value: "9"
            - name: DB_NAME
              value: {{.DB.Name}}
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-pass
                  key: password.txt
            - name: ES_SERVER_ADDR
              value: http://elasticsearch-logging.efk:9200
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/middle-platform-api:caas-v1.0
          imagePullPolicy: IfNotPresent
          name: api
          ports:
            - containerPort: 8080
              name: http
              protocol: TCP
      serviceAccountName: kube-resource

---
apiVersion: v1
kind: Service
metadata:
  name: api
  namespace: {{.Namespace}}
  labels:
    app: api
spec:
  selector:
    app: api
  ports:
    - port: 80
      targetPort: 8080
      name: http
      protocol: TCP
  type: ClusterIP

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-deployment
  namespace: {{.Namespace}}
  labels:
    app: auth
spec:
  replicas: 1
  selector:
    matchLabels:
      app: auth
  template:
    metadata:
      labels:
        app: auth
    spec:
      initContainers:
        - name: wait4etcd
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/k8s-wait-for:v1.3
          imagePullPolicy: IfNotPresent
          args:
            - pod
            - -lapp=etcd
        - name: wait4cluster
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/k8s-wait-for:v1.3
          imagePullPolicy: IfNotPresent
          args:
            - service
            - {{.DB.Host}}-replica
            - -n
            - {{.PGO.Namespace}}
      containers:
        - name: auth
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/middle-platform-auth:caas-v1.0
          imagePullPolicy: IfNotPresent
          args:
            - --registry_address=etcd:2379
          env:
            - name: MICRO_REGISTRY
              value: "etcd"
            - name: DB_HOST
              value: {{.DB.Host}}.{{.PGO.Namespace}}.svc
            - name: DB_PORT
              value: "5432"
            - name: DB_TYPE
              value: postgres
            - name: DB_MAX_IDLE_CONN
              value: "9"
            - name: DB_MAX_OPEN_CONN
              value: "9"
            - name: DB_USER
              value: postgres
            - name: CASBIN_ADAPTER_DRIVER
              value: postgres
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-pass
                  key: password.txt
{{if .LDAP}}
            - name: ENABLE_LDAP
              value: "true"
            - name: LDAP_HOST
              value: {{if .LDAP.CaaSDeploy}}openldap.{{.LDAP.CaaSDeploy.Namespace}}.svc:389{{else}}{{.LDAP.Addr}}{{end}}
            - name: LDAP_USER
              value: {{if .LDAP.CaaSDeploy}}cn=admin,dc={{split "." .LDAP.CaaSDeploy.DN | index 0}},dc={{split "." .LDAP.CaaSDeploy.DN | index 1}}{{else}}cn={{.LDAP.AdminUser}},dc={{split "." .LDAP.DN | index 0}},dc={{split "." .LDAP.DN | index 1}}{{end}}
            - name: LDAP_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: ldap-secret
                  key: ldap_password
            - name: LDAP_USER_BASE
              value: ou=users,{{if .LDAP.CaaSDeploy}}dc={{split "." .LDAP.CaaSDeploy.DN | index 0}},dc={{split "." .LDAP.CaaSDeploy.DN | index 1}}{{else}}dc={{split "." .LDAP.DN | index 0}},dc={{split "." .LDAP.DN | index 1}}{{end}}
            - name: LDAP_GROUP_BASE
              value: ou=groups,{{if .LDAP.CaaSDeploy}}dc={{split "." .LDAP.CaaSDeploy.DN | index 0}},dc={{split "." .LDAP.CaaSDeploy.DN | index 1}}{{else}}dc={{split "." .LDAP.DN | index 0}},dc={{split "." .LDAP.DN | index 1}}{{end}}
{{else}}
            - name: ENABLE_LDAP
              value: "false"
{{end}}
          ports:
            - containerPort: 6060
              name: pprof
      serviceAccountName: kube-resource

---
apiVersion: v1
kind: Service
metadata:
  name: auth
  namespace: {{.Namespace}}
  labels:
    app: auth
spec:
  selector:
    app: auth
  ports:
    - port: 6060
      targetPort: 6060
      name: pprof
      protocol: TCP
  type: ClusterIP

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-deployment
  namespace: {{.Namespace}}
  labels:
    app: user
spec:
  replicas: 1
  selector:
    matchLabels:
      app: user
  template:
    metadata:
      labels:
        app: user
    spec:
      initContainers:
        - name: wait4etcd
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/k8s-wait-for:v1.3
          imagePullPolicy: IfNotPresent
          args:
            - pod
            - -lapp=etcd
        - name: wait4cluster
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/k8s-wait-for:v1.3
          imagePullPolicy: IfNotPresent
          args:
            - service
            - {{.DB.Host}}-replica
            - -n
            - {{.PGO.Namespace}}
        - name: wait4auth
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/k8s-wait-for:v1.3
          imagePullPolicy: IfNotPresent
          args:
            - pod
            - -lapp=auth
      containers:
        - name: user
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/middle-platform-user:caas-v1.0
          imagePullPolicy: IfNotPresent
          args:
            - --registry_address=etcd:2379
          env:
            - name: MICRO_REGISTRY
              value: "etcd"
            - name: ROOT_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: middle-admin-pass
                  key: password
            - name: DB_HOST
              value: {{.DB.Host}}.{{.PGO.Namespace}}.svc
            - name: DB_PORT
              value: "5432"
            - name: DB_TYPE
              value: postgres
            - name: DB_MAX_IDLE_CONN
              value: "9"
            - name: DB_MAX_OPEN_CONN
              value: "9"
            - name: DB_USER
              value: postgres
            - name: DB_NAME
              value: {{.DB.Name}}
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-pass
                  key: password.txt
{{if .LDAP}}
            - name: ENABLE_LDAP
              value: "true"
            - name: LDAP_HOST
              value: {{if .LDAP.CaaSDeploy}}openldap.{{.LDAP.CaaSDeploy.Namespace}}.svc:389{{else}}{{.LDAP.Addr}}{{end}}
            - name: LDAP_USER
              value: {{if .LDAP.CaaSDeploy}}cn=admin,dc={{split "." .LDAP.CaaSDeploy.DN | index 0}},dc={{split "." .LDAP.CaaSDeploy.DN | index 1}}{{else}}cn={{.LDAP.AdminUser}},dc={{split "." .LDAP.DN | index 0}},dc={{split "." .LDAP.DN | index 1}}{{end}}
            - name: LDAP_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: ldap-secret
                  key: ldap_password
            - name: LDAP_USER_BASE
              value: ou=users,{{if .LDAP.CaaSDeploy}}dc={{split "." .LDAP.CaaSDeploy.DN | index 0}},dc={{split "." .LDAP.CaaSDeploy.DN | index 1}}{{else}}dc={{split "." .LDAP.DN | index 0}},dc={{split "." .LDAP.DN | index 1}}{{end}}
            - name: LDAP_GROUP_BASE
              value: ou=groups,{{if .LDAP.CaaSDeploy}}dc={{split "." .LDAP.CaaSDeploy.DN | index 0}},dc={{split "." .LDAP.CaaSDeploy.DN | index 1}}{{else}}dc={{split "." .LDAP.DN | index 0}},dc={{split "." .LDAP.DN | index 1}}{{end}}
{{else}}
            - name: ENABLE_LDAP
              value: "false"
{{end}}
          ports:
            - containerPort: 6060
              name: pprof
      serviceAccountName: kube-resource

---
apiVersion: v1
kind: Service
metadata:
  name: user
  namespace: {{.Namespace}}
  labels:
    app: user
spec:
  selector:
    app: user
  ports:
    - port: 6060
      targetPort: 6060
      name: pprof
      protocol: TCP
  type: ClusterIP
`
