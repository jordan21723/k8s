package minio

const MinioDeployment = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{.Namespace}}

---
apiVersion: apps/v1
kind: Deployment
metadata:
  namespace: {{.Namespace}}
  name: minio
  labels:
    component: minio
spec:
  strategy:
    type: Recreate
  selector:
    matchLabels:
      component: minio
  template:
    metadata:
      labels:
        component: minio
    spec:
      volumes:
      - name: storage
        hostPath:
          path: {{.BackupPath}}/storage
      - name: config
        hostPath:
          path: {{.BackupPath}}/config
      containers:
      - name: minio
        image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/minio:latest
        imagePullPolicy: IfNotPresent
        args:
        - server
        - /storage
        - --config-dir=/config
        env:
        - name: MINIO_ACCESS_KEY
          value: {{.UserName}}
        - name: MINIO_SECRET_KEY
          value: {{.PassWord}}
        ports:
        - containerPort: 9000
        volumeMounts:
        - name: storage
          mountPath: "/storage"
        - name: config
          mountPath: "/config"

---
apiVersion: v1
kind: Service
metadata:
  namespace: {{.Namespace}}
  name: minio
  labels:
    component: minio
spec:
  # ClusterIP is recommended for production environments.
  # Change to NodePort if needed per documentation,
  # but only if you run Minio in a test/trial environment, for example with Minikube.
  type: NodePort
  ports:
    - port: 9000
      targetPort: 9000
      protocol: TCP
      nodePort: 30069
  selector:
    component: minio

---
apiVersion: batch/v1
kind: Job
metadata:
  namespace: {{.Namespace}}
  name: minio-setup
  labels:
    component: minio
spec:
  template:
    metadata:
      name: minio-setup
    spec:
      restartPolicy: OnFailure
      volumes:
      - name: config
        emptyDir: {}
      containers:
      - name: mc
        image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/mc:latest
        image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/mc:latest
        imagePullPolicy: IfNotPresent
        command:
        - /bin/sh
        - -c
        - "mc --config-dir=/config config host add velero http://minio.minio.svc:9000 {{.UserName}} {{.PassWord}} && mc --config-dir=/config mb -p velero/velero"
        volumeMounts:
        - name: config
          mountPath: "/config"

`