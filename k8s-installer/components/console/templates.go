package console

const Template = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{.Namespace}}

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: console-config
  namespace: {{.Namespace}}
data:
  config.yaml: |
    # CaaS configuration

    server:
      http:
        hostname: localhost
        port: 8000
        static:
          development:
            /public: server/public
            /assets: src/assets
            /build: build
            /dist: dist
          production:
            /public: server/public
            /assets: dist/assets
            /dist: dist
      # redis config for multi replicas
      # redis:
      #   port: 6379
      #   host: 127.0.0.1
      redisTimeout: 5000
      # Mec CaaS console session params, not login session
      sessionKey: 'mec-caas:sess'
      sessionTimeout: 7200000 # unit: millisecond

      # backend service gateway server
      gatewayServer:
        url: http://apigateway.{{.MiddlePlatformNamespace}}.svc
        wsUrl: ws://apigateway.{{.MiddlePlatformNamespace}}.svc

      # prometheus server endpoint
      prometheusServer:
        url: http://monitor-prometheus-server.{{.PrometheusNamespace}}.svc

      # docker image search default url
      dockerHubUrl: https://hub.docker.com

    client:
      title: {{.ConsoleTitle}}

      version:
        CaaS: 4.1

      # branch tag
      tag: {{.VendorTag}}

      # current support lanaguages
      supportLangs:
        - label: '简体中文'
          value: 'zh'
        - label: 'English'
          value: 'en'
      defaultLang: 'zh'

      # disable harbor, hidden some pages related with harbor
      disableHarbor: true

      # platform management navigations
      globalNavs:
        - cate: platform
          title: Platform
          items:
            - {
                name: workspaces,
                title: Workspaces,
                icon: enterprise,
                # authAction: 'manage',
              }
            - {
                name: projects,
                title: NAV_PROJECTS,
                icon: project,
                # authAction: 'manage',
              }
            - { name: accounts, title: NAV_ACCOUNTS, icon: human }
            - { name: roles, title: Platform Roles, icon: role }
            - { name: logs, title: Platform Logs, icon: log }
        - cate: infrastructure
          title: Infrastructure
          items:
            - {
                name: infrastructure,
                title: Infrastructure,
                icon: cluster,
                authKey: 'nodes|storageclasses',
              }

      # infrastructure navigations
      infrastructureNavs:
        - cate: infrastructure
          title: Infrastructure
          items:
            - { name: nodes, title: Nodes, icon: laptop }
            - { name: storageclasses, title: Storage Classes, icon: database }

      # monitoring center navigations
      monitoringNavs:
        - cate: monitoring
          title: Monitoring
          items:
            - {
                name: monitor-cluster,
                title: Cluster Status,
                icon: linechart,
                authKey: monitoring,
              }
            - {
                name: monitor-resource,
                title: Application Resources,
                icon: linechart,
                authKey: monitoring,
              }
        - cate: alerting
          title: Alerting
          items:
            - {
                name: alert-message,
                title: Alerting Message,
                icon: loudspeaker,
                authKey: alerting,
              }
            - {
                name: alert-policy,
                title: Alerting Policy,
                icon: hammer,
                authKey: alerting,
              }

      # platform settings navigations
      platformSettingsNavs:
        - cate: 'platformsettings'
          title: Platform Settings
          items:
            - {
                name: mail-server,
                title: Mail Server,
                icon: mail,
                authKey: alerting,
              }
            - {
                name: log-collection,
                title: Log Collection,
                icon: file,
                authKey: logging,
              }

      # workspace page navigations
      workspaceNavs:
        - cate: 'workspace'
          items:
            # - { name: overview, title: Overview, icon: dashboard, skipAuth: true }
            - { name: projects, title: NAV_PROJECTS, icon: project }
            - { name: images, title: Images, icon: strategy-group }
            - { name: devops, title: DevOps Projects, icon: strategy-group }
            - {
                name: apps,
                title: App Templates,
                icon: appcenter,
              }
            - name: management
              title: Workspace Settings
              icon: cogwheel
              children:
                - { name: base-info, title: Basic Info, skipAuth: true }
                - { name: repos, title: App Repos }
                - { name: roles, title: Workspace Roles }
                - { name: members, title: Workspace Members }
                - { name: imagemembers, title: Imagespace Members }

      # apps manage page navigations
      manageAppNavs:
        - cate: 'monitoring'
          items:
            - { name: store, title: App Store, icon: appcenter }
            - { name: categories, title: App Categories, icon: tag }
            - { name: reviews, title: App Review, icon: safe-notice }

      # project page navigations
      projectNavs:
        - cate: 'project'
          items:
            - { name: overview, title: Overview, icon: dashboard, skipAuth: true }
            - name: app-workloads
              title: Application Workloads
              icon: appcenter
              children:
                - { name: services, title: Services }
                - name: workloads
                  title: Workloads
                  tabs:
                    - { name: deployments, title: Deployments }
                    - { name: statefulsets, title: StatefulSets }
                    - { name: daemonsets, title: DaemonSets }
                - name: jobs
                  title: Jobs
                  tabs:
                    - { name: jobs, title: Jobs }
                    - { name: cronjobs, title: CronJobs }
                - { name: pods, title: Pods }
                - { name: ingress, title: Ingress }
            - { name: volumes, title: Volumes, icon: storage }
            - name: config
              title: Configuration Center
              icon: hammer
              children:
                - { name: secrets, title: Secrets }
                - { name: configmaps, title: ConfigMaps }
            - { name: logs, title: logs, icon: log }
            - { name: certificates, title: certificates, icon: file }
            - {
                name: grayrelease,
                title: Grayscale Release,
                icon: bird,
                authKey: 'applications',
                ksModule: 'servicemesh',
              }
            - { name: s2ibuilders, title: Image Builder, icon: vnas }
            - name: monitoring
              title: Monitoring & Alerting
              icon: monitor
              children:
                - {
                    name: alert-message,
                    title: Alerting Message,
                    authKey: 'alerting',
                  }
                - {
                    name: alert-policy,
                    title: Alerting Policy,
                    authKey: 'alerting',
                  }
            - name: settings
              title: Project Settings
              icon: cogwheel
              children:
                - { name: base-info, title: Basic Info, skipAuth: true }
                - { name: roles, title: Project Roles }
                - { name: members, title: Project Members }
            - { name: test, title: Test, icon: vnas }
      # devops page navigations
      devopsNavs:
        - cate: ''
          items:
            - { name: pipelines, title: Pipelines, icon: application }
            - name: management
              title: DEVOPS_PROJECT_MANAGEMENT
              icon: cogwheel
              open: true
              children:
                - { name: base-info, title: Basic Info, skipAuth: true }
                - { name: credentials, title: Credentials }
                - { name: roles, title: DEVOPS_PROJECT_ROLES }
                - { name: members, title: DEVOPS_PROJECT_MEMBERS }

      # system workspace rules control
      systemWorkspace: system
      systemWorkspaceRules:
        devops: []
        members: [view, create, edit, delete]
        projects: [view, edit]
        roles: [view]
        workspaces: [view, edit]
      systemWorkspaceProjectRules:
        alerting: [view, create, edit, delete]
        members: [view, create, edit, delete]
        roles: [view, create, edit, delete]

      # preset infos
      presetUsers: ['admin', 'sonarqube']
      presetClusterRoles: [cluster-admin, cluster-regular, workspaces-manager]
      presetWorkspaceRoles: [admin, regular, viewer]
      presetDevOpsRoles: [owner, maintainer, developer, reporter]
      presetRoles: [admin, operator, viewer]

      # system annotations that need to be hidden for edit
      preservedAnnotations: ['.*caas.io/', 'openpitrix_runtime']

      # namespaces that need to be disable collection file log
      disabledLoggingSidecarNamespace: ['caas-logging-system']

      # loadbalancer annotations, default support qingcloud lb
      loadBalancerDefaultAnnotations:
        service.beta.kubernetes.io/qingcloud-load-balancer-eip-ids: ''
        service.beta.kubernetes.io/qingcloud-load-balancer-type: '0'

      # control error notify on page
      enableErrorNotify: true

      # enable image search when add image for a container
      enableImageSearch: true

      # development
      # disable authority check when developing
      disableAuthority: false

      # show kubeconfig
      enableKubeConfig: true

      # docs url for resources
      resourceDocs:
        applications: /application/app-template/
        composingapps: /application/composing-svc/
        deployments: /workload/deployments/
        statefulsets: /workload/statefulsets/
        daemonsets: /workload/daemonsets/
        jobs: /workload/jobs/
        cronjobs: /workload/cronjobs/
        volumes: /storage/volume/
        storageclasses: /infrastructure/storageclass/
        services: /ingress-service/services/
        grayrelease: /ingress-service/grayscale/
        ingresses: /ingress-service/ingress/
        secrets: /configuration/secrets/
        imageregistry: /configuration/image-registry/
        configmaps: /configuration/configmaps/
        internet: /project-setting/project-gateway/
        nodes: /infrastructure/nodes/
        cicds: /devops/pipeline/
        cridentials: /devops/credential/
        project_base_info: /project-setting/project-quota/
        project_roles: /project-setting/project-roles/
        project_members: /project-setting/project-members/
        s2i_template: /workload/s2i-template/
        helm_specification: /developer/helm-specification/
        helm_developer_guide: /developer/helm-developer-guide/

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: console-deployment
  namespace: {{.Namespace}}
spec:
  replicas: {{.Replicas}}
  selector:
    matchLabels:
      app: console
  template:
    metadata:
      labels:
        app: console
    spec:
      containers:
        - name: console
          image: {{with .ImageRegistry}}{{.}}/{{end}}caas4/console:{{.VendorTag}}
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8000
          volumeMounts:
            - name: console-config-volume
              mountPath: /root/caas4UI/server/config.yaml
              subPath: config.yaml
{{if .ConsoleResourcePath}}
            - name: logo
              mountPath: /root/caas4UI/dist/assets/image
{{end}}
      volumes:
        - name: console-config-volume
          configMap:
            name: console-config
            items:
              - key: config.yaml
                path: config.yaml
{{if .ConsoleResourcePath}}
        - name: logo
          hostPath:
            path: {{.ConsoleResourcePath}}
            type: Directory
{{end}}
---
apiVersion: v1
kind: Service
metadata:
  name: console
  namespace: {{.Namespace}}
spec:
  selector:
    app: console
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8000
      nodePort: 30001
  type: NodePort
`
