package restarter

const Template = `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: auto-restarter
  namespace: {{.Namespace}}

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: auto-restarter-role
rules:
  - apiGroups:
      - "*"
    resources:
      - deployments
      - statefulsets
    verbs:
      - get
      - list
      - watch
      - update
      - patch
  - apiGroups:
      - "*"
    resources:
      - cronjobs
      - endpoints
      - leases
    verbs:
      - get
      - list
      - watch
      - create
      - delete
      - update

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: auto-restarter-global
subjects:
- kind: ServiceAccount
  name: auto-restarter
  namespace: {{.Namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: auto-restarter-role

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auto-restarter
  namespace: {{.Namespace}}
  labels:
    app: auto-restarter
spec:
  replicas: {{.Replicas}}
  selector:
    matchLabels:
      app: auto-restarter
  template:
    metadata:
      labels:
        app: auto-restarter
    spec:
      containers:
        - name: auto-restarter
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/auto-restarter:{{.VersionTag}}
          command:
            - restarter
            - --leader-elect={{if (gt .Replicas 1)}}true{{else}}false{{end}}
            - --kubectl-image={{with .ImageRegistry}}{{.}}/{{end}}kubesphere/kubectl:v1.19.0
          imagePullPolicy: IfNotPresent
          livenessProbe:
            failureThreshold: 3
            httpGet:
              path: /healthz
              port: 8080
              scheme: HTTP
            initialDelaySeconds: 30
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 15
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 200m
              memory: 256Mi
          ports:
            - containerPort: 8080
      serviceAccountName: auto-restarter

---
apiVersion: v1
kind: Service
metadata:
  name: auto-restarter
  namespace: {{.Namespace}}
spec:
  selector:
    app: auto-restarter
  ports:
    - name: http
      protocol: TCP
      port: 80
      targetPort: 8080
`
