package openldap

const Template = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{.Namespace}}

---
apiVersion: v1
kind: Secret
metadata:
  name: openldap-secret
  namespace: {{.Namespace}}
type: Opaque
data:
  ldap_admin_password: {{b64enc .AdminPassword}}

---
apiVersion: v1
kind: Service
metadata:
  name: openldap
  namespace: {{.Namespace}}
spec:
  type: NodePort
  selector:
    app: openldap
  ports:
    - protocol: TCP
      name: ldapport
      port: 389
      targetPort: 389
    - protocol: TCP
      name: ldapport2
      port: 636
      targetPort: 636
    - protocol: TCP
      name: webport
      port: 443
      targetPort: 443

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: openldap-data-claim
  namespace: {{.Namespace}}
  labels:
    app: openldap
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: {{.StorageSize}}Gi
  storageClassName: {{.StorageClass}}

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: openldap-deployment
  namespace: {{.Namespace}}
  labels:
    app: openldap
spec:
  replicas: 1
  selector:
    matchLabels:
      app: openldap
  template:
    metadata:
      labels:
        app: openldap
    spec:
      containers:
        - name: openldap
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/openldap:1.3.0
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 389
            - containerPort: 636
          env:
            - name: LDAP_DOMAIN
              value: {{.DN}}
            - name: LDAP_ORGANISATION
              value: {{split "." .DN | index 0}}
            - name: LDAP_ADMIN_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: openldap-secret
                  key: ldap_admin_password
            - name: LDAP_REPLICATION
              value: "false"
            - name: LDAP_TLS
              value: "false"
            - name: LDAP_REMOVE_CONFIG_AFTER_SETUP
              value: "true"
          volumeMounts:
            - mountPath: /var/lib/ldap
              name: openldap-data-volume
              subPath: ldap-data
            - mountPath: /etc/ldap/slapd.d
              name: openldap-data-volume
              subPath: ldap-config
        - name: phpldapadmin
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/phpldapadmin:0.9.0
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 443
          env:
            - name: PHPLDAPADMIN_LDAP_HOSTS
              value: "localhost"
      volumes:
        - name: openldap-data-volume
          persistentVolumeClaim:
            claimName: openldap-data-claim
`
