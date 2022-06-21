package velero

const Chart = `
apiVersion: v2
appVersion: 1.6.0
description: A Helm chart for velero
name: velero
version: 2.19.1
home: https://github.com/vmware-tanzu/velero
icon: https://cdn-images-1.medium.com/max/1600/1*-9mb3AKnKdcL_QD3CMnthQ.png
sources:
- https://github.com/vmware-tanzu/velero
maintainers:
  - name: ashish-amarnath
    email: ashisham@vmware.com
  - name: carlisia
    email: carlisiac@vmware.com
  - name: jenting
    email: jenting.hsiao@suse.com
  - name: nrb
    email: brubakern@vmware.com
`

const Values = `
##
## Configuration settings that directly affect the Velero deployment YAML.
##

# Details of the container image to use in the Velero deployment & daemonset (if
# enabling restic). Required.
image:
  repository: velero/velero
  tag: v1.6.0
  # Digest value example: sha256:d238835e151cec91c6a811fe3a89a66d3231d9f64d09e5f3c49552672d271f38. If used, it will
  # take precedence over the image.tag.
  # digest:
  pullPolicy: IfNotPresent
  # One or more secrets to be used when pulling images
  imagePullSecrets: []
  # - registrySecretName

# Annotations to add to the Velero deployment's. Optional.
#
# If you are using reloader use the following annotation with your VELERO_SECRET_NAME
annotations: {}
# secret.reloader.stakater.com/reload: "<VELERO_SECRET_NAME>"

# Labels to add to the Velero deployment's. Optional.
labels: {}

# Annotations to add to the Velero deployment's pod template. Optional.
#
# If using kube2iam or kiam, use the following annotation with your AWS_ACCOUNT_ID
# and VELERO_ROLE_NAME filled in:
podAnnotations: {}
  #  iam.amazonaws.com/role: "arn:aws:iam::<AWS_ACCOUNT_ID>:role/<VELERO_ROLE_NAME>"

# Additional pod labels for Velero deployment's template. Optional
# ref: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
podLabels: {}

# Resource requests/limits to specify for the Velero deployment.
# https://velero.io/docs/v1.6/customize-installation/#customize-resource-requests-and-limits
resources:
  requests:
    cpu: 500m
    memory: 128Mi
  limits:
    cpu: 1000m
    memory: 512Mi

# Configure the dnsPolicy of the Velero deployment
# See: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy
dnsPolicy: ClusterFirst

# Init containers to add to the Velero deployment's pod spec. At least one plugin provider image is required.
initContainers: []
  # - name: velero-plugin-for-aws
  #   image: velero/velero-plugin-for-aws:v1.2.0
  #   imagePullPolicy: IfNotPresent
  #   volumeMounts:
  #     - mountPath: /target
  #       name: plugins

# SecurityContext to use for the Velero deployment. Optional.
# Set fsGroup for 'AWS IAM Roles for Service Accounts'
# see more informations at: https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
securityContext: {}
  # fsGroup: 1337

# Pod priority class name to use for the Velero deployment. Optional.
priorityClassName: ""

# Tolerations to use for the Velero deployment. Optional.
tolerations: []

# Affinity to use for the Velero deployment. Optional.
affinity: {}

# Node selector to use for the Velero deployment. Optional.
nodeSelector: {}

# Extra volumes for the Velero deployment. Optional.
extraVolumes: []

# Extra volumeMounts for the Velero deployment. Optional.
extraVolumeMounts: []

# Settings for Velero's prometheus metrics. Enabled by default.
metrics:
  enabled: true
  scrapeInterval: 30s
  scrapeTimeout: 10s

  # service metdata if metrics are enabled
  service:
    annotations: {}
    labels: {}

  # Pod annotations for Prometheus
  podAnnotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "8085"
    prometheus.io/path: "/metrics"

  serviceMonitor:
    enabled: false
    additionalLabels: {}
    # ServiceMonitor namespace. Default to Velero namespace.
    # namespace:

##
## End of deployment-related settings.
##


##
## Parameters for the 'default' BackupStorageLocation and VolumeSnapshotLocation,
## and additional server settings.
##
configuration:
  # Cloud provider being used (e.g. aws, azure, gcp).
  provider:

  # Parameters for the 'default' BackupStorageLocation. See
  # https://velero.io/docs/v1.6/api-types/backupstoragelocation/
  backupStorageLocation:
    # name is the name of the backup storage location where backups should be stored. If a name is not provided,
    # a backup storage location will be created with the name "default". Optional.
    name:
    # provider is the name for the backup storage location provider. If omitted
    # 'configuration.provider' will be used instead.
    provider:
    # bucket is the name of the bucket to store backups in. Required.
    bucket:
    # caCert defines a base64 encoded CA bundle to use when verifying TLS connections to the provider.
    caCert:
    # prefix is the directory under which all Velero data should be stored within the bucket. Optional.
    prefix:
    # Additional provider-specific configuration. See link above
    # for details of required/optional fields for your provider.
    config: {}
    #  region:
    #  s3ForcePathStyle:
    #  s3Url:
    #  kmsKeyId:
    #  resourceGroup:
    #  The ID of the subscription containing the storage account, if different from the cluster’s subscription. (Azure only)
    #  subscriptionId:
    #  storageAccount:
    #  publicUrl:
    #  Name of the GCP service account to use for this backup storage location. Specify the
    #  service account here if you want to use workload identity instead of providing the key file.(GCP only)
    #  serviceAccount:

  # Parameters for the 'default' VolumeSnapshotLocation. See
  # https://velero.io/docs/v1.6/api-types/volumesnapshotlocation/
  volumeSnapshotLocation:
    # name is the name of the volume snapshot location where snapshots are being taken. Required.
    name:
    # provider is the name for the volume snapshot provider. If omitted
    # 'configuration.provider' will be used instead.
    provider:
    # Additional provider-specific configuration. See link above
    # for details of required/optional fields for your provider.
    config: {}
  #    region:
  #    apiTimeout:
  #    resourceGroup:
  #    The ID of the subscription where volume snapshots should be stored, if different from the cluster’s subscription. If specified, also requires "configuration.volumeSnapshotLocation.config.resourceGroup" to be set. (Azure only)
  #    subscriptionId:
  #    incremental:
  #    snapshotLocation:
  #    project:

  # These are server-level settings passed as CLI flags to the "velero server" command. Velero
  # uses default values if they're not passed in, so they only need to be explicitly specified
  # here if using a non-default value. The "velero server" default values are shown in the
  # comments below.
  # --------------------
  # "velero server" default: 1m
  backupSyncPeriod:
  # "velero server" default: 1h
  resticTimeout:
  # "velero server" default: namespaces,persistentvolumes,persistentvolumeclaims,secrets,configmaps,serviceaccounts,limitranges,pods
  restoreResourcePriorities:
  # "velero server" default: false
  restoreOnlyMode:
  # "velero server" default: 20.0
  clientQPS:
  # "velero server" default: 30
  clientBurst:
  # "velero server" default: empty
  disableControllers:
  #

  # additional key/value pairs to be used as environment variables such as "AWS_CLUSTER_NAME: 'yourcluster.domain.tld'"
  extraEnvVars: {}

  # Comma separated list of velero feature flags. default: empty
  features:

  # Set log-level for Velero pod. Default: info. Other options: debug, warning, error, fatal, panic.
  logLevel:

  # Set log-format for Velero pod. Default: text. Other option: json.
  logFormat:

  # Set true for backup all pod volumes without having to apply annotation on the pod when used restic Default: false. Other option: false.
  defaultVolumesToRestic:

##
## End of backup/snapshot location settings.
##


##
## Settings for additional Velero resources.
##

rbac:
  # Whether to create the Velero role and role binding to give all permissions to the namespace to Velero.
  create: true
  # Whether to create the cluster role binding to give administrator permissions to Velero
  clusterAdministrator: true

# Information about the Kubernetes service account Velero uses.
serviceAccount:
  server:
    create: true
    name:
    annotations:
    labels:

# Info about the secret to be used by the Velero deployment, which
# should contain credentials for the cloud provider IAM account you've
# set up for Velero.
credentials:
  # Whether a secret should be used as the source of IAM account
  # credentials. Set to false if, for example, using kube2iam or
  # kiam to provide IAM credentials for the Velero pod.
  useSecret: true
  # Name of the secret to create if "useSecret" is true and "existingSecret" is empty
  name:
  # Name of a pre-existing secret (if any) in the Velero namespace
  # that should be used to get IAM account credentials. Optional.
  existingSecret:
  # Data to be stored in the Velero secret, if "useSecret" is true and "existingSecret" is empty.
  # As of the current Velero release, Velero only uses one secret key/value at a time.
  # The key must be named "cloud", and the value corresponds to the entire content of your IAM credentials file.
  # Note that the format will be different for different providers, please check their documentation.
  # Here is a list of documentation for plugins maintained by the Velero team:
  # [AWS] https://github.com/vmware-tanzu/velero-plugin-for-aws/blob/main/README.md
  # [GCP] https://github.com/vmware-tanzu/velero-plugin-for-gcp/blob/main/README.md
  # [Azure] https://github.com/vmware-tanzu/velero-plugin-for-microsoft-azure/blob/main/README.md
  secretContents: {}
  #  cloud: |
  #    [default]
  #    aws_access_key_id=<REDACTED>
  #    aws_secret_access_key=<REDACTED>
  # additional key/value pairs to be used as environment variables such as "DIGITALOCEAN_TOKEN: <your-key>". Values will be stored in the secret.
  extraEnvVars: {}
  # Name of a pre-existing secret (if any) in the Velero namespace
  # that will be used to load environment variables into velero and restic.
  # Secret should be in format - https://kubernetes.io/docs/concepts/configuration/secret/#use-case-as-container-environment-variables
  extraSecretRef: ""

# Whether to create backupstoragelocation crd, if false => do not create a default backup location
backupsEnabled: true
# Whether to create volumesnapshotlocation crd, if false => disable snapshot feature
snapshotsEnabled: true

# Whether to deploy the restic daemonset.
deployRestic: false

restic:
  podVolumePath: /var/lib/kubelet/pods
  privileged: false
  # Pod priority class name to use for the Restic daemonset. Optional.
  priorityClassName: ""
  # Resource requests/limits to specify for the Restic daemonset deployment. Optional.
  # https://velero.io/docs/v1.6/customize-installation/#customize-resource-requests-and-limits
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 1000m
      memory: 1024Mi

  # Tolerations to use for the Restic daemonset. Optional.
  tolerations: []

  # Annotations to set for the Restic daemonset. Optional.
  annotations: {}

  # labels to set for the Restic daemonset. Optional.
  labels: {}

  # Extra volumes for the Restic daemonset. Optional.
  extraVolumes: []

  # Extra volumeMounts for the Restic daemonset. Optional.
  extraVolumeMounts: []

  # Configure the dnsPolicy of the Restic daemonset
  # See: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy
  dnsPolicy: ClusterFirst

  # SecurityContext to use for the Velero deployment. Optional.
  # Set fsGroup for 'AWS IAM Roles for Service Accounts'
  # see more informations at: https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
  securityContext: {}
    # fsGroup: 1337

  # Node selector to use for the Restic daemonset. Optional.
  nodeSelector: {}

# Backup schedules to create.
# Eg:
# schedules:
#   mybackup:
#     labels:
#       myenv: foo
#     annotations:
#       myenv: foo
#     schedule: "0 0 * * *"
#     template:
#       ttl: "240h"
#       includedNamespaces:
#       - foo
schedules: {}

# Velero ConfigMaps.
# Eg:
# configMaps:
#   restic-restore-action-config:
#     labels:
#       velero.io/plugin-config: ""
#       velero.io/restic: RestoreItemAction
#     data:
#       image: velero/velero-restic-restore-helper:v1.6.0
configMaps: {}

##
## End of additional Velero resource settings.
##
`

var Crds = map[string]string{
	"backupstoragelocations.yaml": backupstoragelocations,
	"deletebackuprequests.yaml": deletebackuprequests,
	"podvolumebackups.yaml": podvolumebackups,
	"resticrepositories.yaml": resticrepositories,
	"schedules.yaml": schedules,
	"volumesnapshotlocations.yaml": volumesnapshotlocations,
	"backups.yaml": backups,
	"downloadrequests.yaml": downloadrequests,
	"podvolumerestores.yaml": podvolumerestores,
	"restores.yaml": restores,
	"serverstatusrequests.yaml": serverstatusrequests,
}

var Templates = map[string]string{
	"backupstoragelocation.yaml" : backupstoragelocation,
	"configmaps.yaml" : configmaps,
	"role.yaml" : role,
	"serviceaccount-server.yaml" : serviceaccountserver,
	"upgrade-crds.yaml" : upgradecrds,
	"cleanup-crds.yaml" : cleanupcrds,
	"deployment.yaml" : deployment,
	"restic-daemonset.yaml" : resticdaemonset,
	"schedule.yaml" : schedule,
	"servicemonitor.yaml" : servicemonitor,
	"volumesnapshotlocation.yaml" : volumesnapshotlocation,
	"clusterrolebinding.yaml" : clusterrolebinding,
	"_helpers.tpl" : helpers,
	"rolebinding.yaml" : rolebinding,
	"secret.yaml" : secret,
	"service.yaml" : service,
}

// Crds
const (
	backupstoragelocations = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    app.kubernetes.io/name: velero
    component: velero
  annotations:
    controller-gen.kubebuilder.io/version: v0.3.0
  creationTimestamp: null
  name: backupstoragelocations.velero.io
spec:
  additionalPrinterColumns:
  - JSONPath: .status.phase
    description: Backup Storage Location status such as Available/Unavailable
    name: Phase
    type: string
  - JSONPath: .status.lastValidationTime
    description: LastValidationTime is the last time the backup store location was
      validated
    name: Last Validated
    type: date
  - JSONPath: .metadata.creationTimestamp
    name: Age
    type: date
  - JSONPath: .spec.default
    description: Default backup storage location
    name: Default
    type: boolean
  group: velero.io
  names:
    kind: BackupStorageLocation
    listKind: BackupStorageLocationList
    plural: backupstoragelocations
    shortNames:
    - bsl
    singular: backupstoragelocation
  preserveUnknownFields: false
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: BackupStorageLocation is a location where Velero stores backup
        objects
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: BackupStorageLocationSpec defines the desired state of a Velero
            BackupStorageLocation
          properties:
            accessMode:
              description: AccessMode defines the permissions for the backup storage
                location.
              enum:
              - ReadOnly
              - ReadWrite
              type: string
            backupSyncPeriod:
              description: BackupSyncPeriod defines how frequently to sync backup
                API objects from object storage. A value of 0 disables sync.
              nullable: true
              type: string
            config:
              additionalProperties:
                type: string
              description: Config is for provider-specific configuration fields.
              type: object
            credential:
              description: Credential contains the credential information intended
                to be used with this location
              properties:
                key:
                  description: The key of the secret to select from.  Must be a valid
                    secret key.
                  type: string
                name:
                  description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                    TODO: Add other useful fields. apiVersion, kind, uid?'
                  type: string
                optional:
                  description: Specify whether the Secret or its key must be defined
                  type: boolean
              required:
              - key
              type: object
            default:
              description: Default indicates this location is the default backup storage
                location.
              type: boolean
            objectStorage:
              description: ObjectStorageLocation specifies the settings necessary
                to connect to a provider's object storage.
              properties:
                bucket:
                  description: Bucket is the bucket to use for object storage.
                  type: string
                caCert:
                  description: CACert defines a CA bundle to use when verifying TLS
                    connections to the provider.
                  format: byte
                  type: string
                prefix:
                  description: Prefix is the path inside a bucket to use for Velero
                    storage. Optional.
                  type: string
              required:
              - bucket
              type: object
            provider:
              description: Provider is the provider of the backup storage.
              type: string
            validationFrequency:
              description: ValidationFrequency defines how frequently to validate
                the corresponding object storage. A value of 0 disables validation.
              nullable: true
              type: string
          required:
          - objectStorage
          - provider
          type: object
        status:
          description: BackupStorageLocationStatus defines the observed state of BackupStorageLocation
          properties:
            accessMode:
              description: "AccessMode is an unused field. \n Deprecated: there is
                now an AccessMode field on the Spec and this field will be removed
                entirely as of v2.0."
              enum:
              - ReadOnly
              - ReadWrite
              type: string
            lastSyncedRevision:
              description: "LastSyncedRevision is the value of the "metadata/revision"
                file in the backup storage location the last time the BSL's contents
                were synced into the cluster. \n Deprecated: this field is no longer
                updated or used for detecting changes to the location's contents and
                will be removed entirely in v2.0."
              type: string
            lastSyncedTime:
              description: LastSyncedTime is the last time the contents of the location
                were synced into the cluster.
              format: date-time
              nullable: true
              type: string
            lastValidationTime:
              description: LastValidationTime is the last time the backup store location
                was validated the cluster.
              format: date-time
              nullable: true
              type: string
            phase:
              description: Phase is the current state of the BackupStorageLocation.
              enum:
              - Available
              - Unavailable
              type: string
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
	deletebackuprequests = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    app.kubernetes.io/name: velero
    component: velero
  annotations:
    controller-gen.kubebuilder.io/version: v0.3.0
  creationTimestamp: null
  name: deletebackuprequests.velero.io
spec:
  group: velero.io
  names:
    kind: DeleteBackupRequest
    listKind: DeleteBackupRequestList
    plural: deletebackuprequests
    singular: deletebackuprequest
  preserveUnknownFields: false
  scope: Namespaced
  validation:
    openAPIV3Schema:
      description: DeleteBackupRequest is a request to delete one or more backups.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: DeleteBackupRequestSpec is the specification for which backups
            to delete.
          properties:
            backupName:
              type: string
          required:
          - backupName
          type: object
        status:
          description: DeleteBackupRequestStatus is the current status of a DeleteBackupRequest.
          properties:
            errors:
              description: Errors contains any errors that were encountered during
                the deletion process.
              items:
                type: string
              nullable: true
              type: array
            phase:
              description: Phase is the current state of the DeleteBackupRequest.
              enum:
              - New
              - InProgress
              - Processed
              type: string
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
	podvolumebackups = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    app.kubernetes.io/name: velero
    component: velero
  annotations:
    controller-gen.kubebuilder.io/version: v0.3.0
  creationTimestamp: null
  name: podvolumebackups.velero.io
spec:
  group: velero.io
  names:
    kind: PodVolumeBackup
    listKind: PodVolumeBackupList
    plural: podvolumebackups
    singular: podvolumebackup
  preserveUnknownFields: false
  scope: Namespaced
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: PodVolumeBackupSpec is the specification for a PodVolumeBackup.
          properties:
            backupStorageLocation:
              description: BackupStorageLocation is the name of the backup storage
                location where the restic repository is stored.
              type: string
            node:
              description: Node is the name of the node that the Pod is running on.
              type: string
            pod:
              description: Pod is a reference to the pod containing the volume to
                be backed up.
              properties:
                apiVersion:
                  description: API version of the referent.
                  type: string
                fieldPath:
                  description: 'If referring to a piece of an object instead of an
                    entire object, this string should contain a valid JSON/Go field
                    access statement, such as desiredState.manifest.containers[2].
                    For example, if the object reference is to a container within
                    a pod, this would take on a value like: "spec.containers{name}"
                    (where "name" refers to the name of the container that triggered
                    the event) or if no container name is specified "spec.containers[2]"
                    (container with index 2 in this pod). This syntax is chosen only
                    to have some well-defined way of referencing a part of an object.
                    TODO: this design is not final and this field is subject to change
                    in the future.'
                  type: string
                kind:
                  description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
                  type: string
                name:
                  description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names'
                  type: string
                namespace:
                  description: 'Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/'
                  type: string
                resourceVersion:
                  description: 'Specific resourceVersion to which this reference is
                    made, if any. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency'
                  type: string
                uid:
                  description: 'UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids'
                  type: string
              type: object
            repoIdentifier:
              description: RepoIdentifier is the restic repository identifier.
              type: string
            tags:
              additionalProperties:
                type: string
              description: Tags are a map of key-value pairs that should be applied
                to the volume backup as tags.
              type: object
            volume:
              description: Volume is the name of the volume within the Pod to be backed
                up.
              type: string
          required:
          - backupStorageLocation
          - node
          - pod
          - repoIdentifier
          - volume
          type: object
        status:
          description: PodVolumeBackupStatus is the current status of a PodVolumeBackup.
          properties:
            completionTimestamp:
              description: CompletionTimestamp records the time a backup was completed.
                Completion time is recorded even on failed backups. Completion time
                is recorded before uploading the backup object. The server's time
                is used for CompletionTimestamps
              format: date-time
              nullable: true
              type: string
            message:
              description: Message is a message about the pod volume backup's status.
              type: string
            path:
              description: Path is the full path within the controller pod being backed
                up.
              type: string
            phase:
              description: Phase is the current state of the PodVolumeBackup.
              enum:
              - New
              - InProgress
              - Completed
              - Failed
              type: string
            progress:
              description: Progress holds the total number of bytes of the volume
                and the current number of backed up bytes. This can be used to display
                progress information about the backup operation.
              properties:
                bytesDone:
                  format: int64
                  type: integer
                totalBytes:
                  format: int64
                  type: integer
              type: object
            snapshotID:
              description: SnapshotID is the identifier for the snapshot of the pod
                volume.
              type: string
            startTimestamp:
              description: StartTimestamp records the time a backup was started. Separate
                from CreationTimestamp, since that value changes on restores. The
                server's time is used for StartTimestamps
              format: date-time
              nullable: true
              type: string
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
	resticrepositories = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    app.kubernetes.io/name: velero
    component: velero
  annotations:
    controller-gen.kubebuilder.io/version: v0.3.0
  creationTimestamp: null
  name: resticrepositories.velero.io
spec:
  group: velero.io
  names:
    kind: ResticRepository
    listKind: ResticRepositoryList
    plural: resticrepositories
    singular: resticrepository
  preserveUnknownFields: false
  scope: Namespaced
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: ResticRepositorySpec is the specification for a ResticRepository.
          properties:
            backupStorageLocation:
              description: BackupStorageLocation is the name of the BackupStorageLocation
                that should contain this repository.
              type: string
            maintenanceFrequency:
              description: MaintenanceFrequency is how often maintenance should be
                run.
              type: string
            resticIdentifier:
              description: ResticIdentifier is the full restic-compatible string for
                identifying this repository.
              type: string
            volumeNamespace:
              description: VolumeNamespace is the namespace this restic repository
                contains pod volume backups for.
              type: string
          required:
          - backupStorageLocation
          - maintenanceFrequency
          - resticIdentifier
          - volumeNamespace
          type: object
        status:
          description: ResticRepositoryStatus is the current status of a ResticRepository.
          properties:
            lastMaintenanceTime:
              description: LastMaintenanceTime is the last time maintenance was run.
              format: date-time
              nullable: true
              type: string
            message:
              description: Message is a message about the current status of the ResticRepository.
              type: string
            phase:
              description: Phase is the current state of the ResticRepository.
              enum:
              - New
              - Ready
              - NotReady
              type: string
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
	schedules = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    app.kubernetes.io/name: velero
    component: velero
  annotations:
    controller-gen.kubebuilder.io/version: v0.3.0
  creationTimestamp: null
  name: schedules.velero.io
spec:
  group: velero.io
  names:
    kind: Schedule
    listKind: ScheduleList
    plural: schedules
    singular: schedule
  preserveUnknownFields: false
  scope: Namespaced
  validation:
    openAPIV3Schema:
      description: Schedule is a Velero resource that represents a pre-scheduled or
        periodic Backup that should be run.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: ScheduleSpec defines the specification for a Velero schedule
          properties:
            schedule:
              description: Schedule is a Cron expression defining when to run the
                Backup.
              type: string
            template:
              description: Template is the definition of the Backup to be run on the
                provided schedule
              properties:
                defaultVolumesToRestic:
                  description: DefaultVolumesToRestic specifies whether restic should
                    be used to take a backup of all pod volumes by default.
                  type: boolean
                excludedNamespaces:
                  description: ExcludedNamespaces contains a list of namespaces that
                    are not included in the backup.
                  items:
                    type: string
                  nullable: true
                  type: array
                excludedResources:
                  description: ExcludedResources is a slice of resource names that
                    are not included in the backup.
                  items:
                    type: string
                  nullable: true
                  type: array
                hooks:
                  description: Hooks represent custom behaviors that should be executed
                    at different phases of the backup.
                  properties:
                    resources:
                      description: Resources are hooks that should be executed when
                        backing up individual instances of a resource.
                      items:
                        description: BackupResourceHookSpec defines one or more BackupResourceHooks
                          that should be executed based on the rules defined for namespaces,
                          resources, and label selector.
                        properties:
                          excludedNamespaces:
                            description: ExcludedNamespaces specifies the namespaces
                              to which this hook spec does not apply.
                            items:
                              type: string
                            nullable: true
                            type: array
                          excludedResources:
                            description: ExcludedResources specifies the resources
                              to which this hook spec does not apply.
                            items:
                              type: string
                            nullable: true
                            type: array
                          includedNamespaces:
                            description: IncludedNamespaces specifies the namespaces
                              to which this hook spec applies. If empty, it applies
                              to all namespaces.
                            items:
                              type: string
                            nullable: true
                            type: array
                          includedResources:
                            description: IncludedResources specifies the resources
                              to which this hook spec applies. If empty, it applies
                              to all resources.
                            items:
                              type: string
                            nullable: true
                            type: array
                          labelSelector:
                            description: LabelSelector, if specified, filters the
                              resources to which this hook spec applies.
                            nullable: true
                            properties:
                              matchExpressions:
                                description: matchExpressions is a list of label selector
                                  requirements. The requirements are ANDed.
                                items:
                                  description: A label selector requirement is a selector
                                    that contains values, a key, and an operator that
                                    relates the key and values.
                                  properties:
                                    key:
                                      description: key is the label key that the selector
                                        applies to.
                                      type: string
                                    operator:
                                      description: operator represents a key's relationship
                                        to a set of values. Valid operators are In,
                                        NotIn, Exists and DoesNotExist.
                                      type: string
                                    values:
                                      description: values is an array of string values.
                                        If the operator is In or NotIn, the values
                                        array must be non-empty. If the operator is
                                        Exists or DoesNotExist, the values array must
                                        be empty. This array is replaced during a
                                        strategic merge patch.
                                      items:
                                        type: string
                                      type: array
                                  required:
                                  - key
                                  - operator
                                  type: object
                                type: array
                              matchLabels:
                                additionalProperties:
                                  type: string
                                description: matchLabels is a map of {key,value} pairs.
                                  A single {key,value} in the matchLabels map is equivalent
                                  to an element of matchExpressions, whose key field
                                  is "key", the operator is "In", and the values array
                                  contains only "value". The requirements are ANDed.
                                type: object
                            type: object
                          name:
                            description: Name is the name of this hook.
                            type: string
                          post:
                            description: PostHooks is a list of BackupResourceHooks
                              to execute after storing the item in the backup. These
                              are executed after all "additional items" from item
                              actions are processed.
                            items:
                              description: BackupResourceHook defines a hook for a
                                resource.
                              properties:
                                exec:
                                  description: Exec defines an exec hook.
                                  properties:
                                    command:
                                      description: Command is the command and arguments
                                        to execute.
                                      items:
                                        type: string
                                      minItems: 1
                                      type: array
                                    container:
                                      description: Container is the container in the
                                        pod where the command should be executed.
                                        If not specified, the pod's first container
                                        is used.
                                      type: string
                                    onError:
                                      description: OnError specifies how Velero should
                                        behave if it encounters an error executing
                                        this hook.
                                      enum:
                                      - Continue
                                      - Fail
                                      type: string
                                    timeout:
                                      description: Timeout defines the maximum amount
                                        of time Velero should wait for the hook to
                                        complete before considering the execution
                                        a failure.
                                      type: string
                                  required:
                                  - command
                                  type: object
                              required:
                              - exec
                              type: object
                            type: array
                          pre:
                            description: PreHooks is a list of BackupResourceHooks
                              to execute prior to storing the item in the backup.
                              These are executed before any "additional items" from
                              item actions are processed.
                            items:
                              description: BackupResourceHook defines a hook for a
                                resource.
                              properties:
                                exec:
                                  description: Exec defines an exec hook.
                                  properties:
                                    command:
                                      description: Command is the command and arguments
                                        to execute.
                                      items:
                                        type: string
                                      minItems: 1
                                      type: array
                                    container:
                                      description: Container is the container in the
                                        pod where the command should be executed.
                                        If not specified, the pod's first container
                                        is used.
                                      type: string
                                    onError:
                                      description: OnError specifies how Velero should
                                        behave if it encounters an error executing
                                        this hook.
                                      enum:
                                      - Continue
                                      - Fail
                                      type: string
                                    timeout:
                                      description: Timeout defines the maximum amount
                                        of time Velero should wait for the hook to
                                        complete before considering the execution
                                        a failure.
                                      type: string
                                  required:
                                  - command
                                  type: object
                              required:
                              - exec
                              type: object
                            type: array
                        required:
                        - name
                        type: object
                      nullable: true
                      type: array
                  type: object
                includeClusterResources:
                  description: IncludeClusterResources specifies whether cluster-scoped
                    resources should be included for consideration in the backup.
                  nullable: true
                  type: boolean
                includedNamespaces:
                  description: IncludedNamespaces is a slice of namespace names to
                    include objects from. If empty, all namespaces are included.
                  items:
                    type: string
                  nullable: true
                  type: array
                includedResources:
                  description: IncludedResources is a slice of resource names to include
                    in the backup. If empty, all resources are included.
                  items:
                    type: string
                  nullable: true
                  type: array
                labelSelector:
                  description: LabelSelector is a metav1.LabelSelector to filter with
                    when adding individual objects to the backup. If empty or nil,
                    all objects are included. Optional.
                  nullable: true
                  properties:
                    matchExpressions:
                      description: matchExpressions is a list of label selector requirements.
                        The requirements are ANDed.
                      items:
                        description: A label selector requirement is a selector that
                          contains values, a key, and an operator that relates the
                          key and values.
                        properties:
                          key:
                            description: key is the label key that the selector applies
                              to.
                            type: string
                          operator:
                            description: operator represents a key's relationship
                              to a set of values. Valid operators are In, NotIn, Exists
                              and DoesNotExist.
                            type: string
                          values:
                            description: values is an array of string values. If the
                              operator is In or NotIn, the values array must be non-empty.
                              If the operator is Exists or DoesNotExist, the values
                              array must be empty. This array is replaced during a
                              strategic merge patch.
                            items:
                              type: string
                            type: array
                        required:
                        - key
                        - operator
                        type: object
                      type: array
                    matchLabels:
                      additionalProperties:
                        type: string
                      description: matchLabels is a map of {key,value} pairs. A single
                        {key,value} in the matchLabels map is equivalent to an element
                        of matchExpressions, whose key field is "key", the operator
                        is "In", and the values array contains only "value". The requirements
                        are ANDed.
                      type: object
                  type: object
                orderedResources:
                  additionalProperties:
                    type: string
                  description: OrderedResources specifies the backup order of resources
                    of specific Kind. The map key is the Kind name and value is a
                    list of resource names separated by commas. Each resource name
                    has format "namespace/resourcename".  For cluster resources, simply
                    use "resourcename".
                  nullable: true
                  type: object
                snapshotVolumes:
                  description: SnapshotVolumes specifies whether to take cloud snapshots
                    of any PV's referenced in the set of objects included in the Backup.
                  nullable: true
                  type: boolean
                storageLocation:
                  description: StorageLocation is a string containing the name of
                    a BackupStorageLocation where the backup should be stored.
                  type: string
                ttl:
                  description: TTL is a time.Duration-parseable string describing
                    how long the Backup should be retained for.
                  type: string
                volumeSnapshotLocations:
                  description: VolumeSnapshotLocations is a list containing names
                    of VolumeSnapshotLocations associated with this backup.
                  items:
                    type: string
                  type: array
              type: object
            useOwnerReferencesInBackup:
              description: UseOwnerReferencesBackup specifies whether to use OwnerReferences
                on backups created by this Schedule.
              nullable: true
              type: boolean
          required:
          - schedule
          - template
          type: object
        status:
          description: ScheduleStatus captures the current state of a Velero schedule
          properties:
            lastBackup:
              description: LastBackup is the last time a Backup was run for this Schedule
                schedule
              format: date-time
              nullable: true
              type: string
            phase:
              description: Phase is the current phase of the Schedule
              enum:
              - New
              - Enabled
              - FailedValidation
              type: string
            validationErrors:
              description: ValidationErrors is a slice of all validation errors (if
                applicable)
              items:
                type: string
              type: array
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
	volumesnapshotlocations = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    app.kubernetes.io/name: velero
    component: velero
  annotations:
    controller-gen.kubebuilder.io/version: v0.3.0
  creationTimestamp: null
  name: volumesnapshotlocations.velero.io
spec:
  group: velero.io
  names:
    kind: VolumeSnapshotLocation
    listKind: VolumeSnapshotLocationList
    plural: volumesnapshotlocations
    singular: volumesnapshotlocation
  preserveUnknownFields: false
  scope: Namespaced
  validation:
    openAPIV3Schema:
      description: VolumeSnapshotLocation is a location where Velero stores volume
        snapshots.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: VolumeSnapshotLocationSpec defines the specification for a
            Velero VolumeSnapshotLocation.
          properties:
            config:
              additionalProperties:
                type: string
              description: Config is for provider-specific configuration fields.
              type: object
            provider:
              description: Provider is the provider of the volume storage.
              type: string
          required:
          - provider
          type: object
        status:
          description: VolumeSnapshotLocationStatus describes the current status of
            a Velero VolumeSnapshotLocation.
          properties:
            phase:
              description: VolumeSnapshotLocationPhase is the lifecycle phase of a
                Velero VolumeSnapshotLocation.
              enum:
              - Available
              - Unavailable
              type: string
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
	backups = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    app.kubernetes.io/name: velero
    component: velero
  annotations:
    controller-gen.kubebuilder.io/version: v0.3.0
  creationTimestamp: null
  name: backups.velero.io
spec:
  group: velero.io
  names:
    kind: Backup
    listKind: BackupList
    plural: backups
    singular: backup
  preserveUnknownFields: false
  scope: Namespaced
  validation:
    openAPIV3Schema:
      description: Backup is a Velero resource that represents the capture of Kubernetes
        cluster state at a point in time (API objects and associated volume state).
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: BackupSpec defines the specification for a Velero backup.
          properties:
            defaultVolumesToRestic:
              description: DefaultVolumesToRestic specifies whether restic should
                be used to take a backup of all pod volumes by default.
              type: boolean
            excludedNamespaces:
              description: ExcludedNamespaces contains a list of namespaces that are
                not included in the backup.
              items:
                type: string
              nullable: true
              type: array
            excludedResources:
              description: ExcludedResources is a slice of resource names that are
                not included in the backup.
              items:
                type: string
              nullable: true
              type: array
            hooks:
              description: Hooks represent custom behaviors that should be executed
                at different phases of the backup.
              properties:
                resources:
                  description: Resources are hooks that should be executed when backing
                    up individual instances of a resource.
                  items:
                    description: BackupResourceHookSpec defines one or more BackupResourceHooks
                      that should be executed based on the rules defined for namespaces,
                      resources, and label selector.
                    properties:
                      excludedNamespaces:
                        description: ExcludedNamespaces specifies the namespaces to
                          which this hook spec does not apply.
                        items:
                          type: string
                        nullable: true
                        type: array
                      excludedResources:
                        description: ExcludedResources specifies the resources to
                          which this hook spec does not apply.
                        items:
                          type: string
                        nullable: true
                        type: array
                      includedNamespaces:
                        description: IncludedNamespaces specifies the namespaces to
                          which this hook spec applies. If empty, it applies to all
                          namespaces.
                        items:
                          type: string
                        nullable: true
                        type: array
                      includedResources:
                        description: IncludedResources specifies the resources to
                          which this hook spec applies. If empty, it applies to all
                          resources.
                        items:
                          type: string
                        nullable: true
                        type: array
                      labelSelector:
                        description: LabelSelector, if specified, filters the resources
                          to which this hook spec applies.
                        nullable: true
                        properties:
                          matchExpressions:
                            description: matchExpressions is a list of label selector
                              requirements. The requirements are ANDed.
                            items:
                              description: A label selector requirement is a selector
                                that contains values, a key, and an operator that
                                relates the key and values.
                              properties:
                                key:
                                  description: key is the label key that the selector
                                    applies to.
                                  type: string
                                operator:
                                  description: operator represents a key's relationship
                                    to a set of values. Valid operators are In, NotIn,
                                    Exists and DoesNotExist.
                                  type: string
                                values:
                                  description: values is an array of string values.
                                    If the operator is In or NotIn, the values array
                                    must be non-empty. If the operator is Exists or
                                    DoesNotExist, the values array must be empty.
                                    This array is replaced during a strategic merge
                                    patch.
                                  items:
                                    type: string
                                  type: array
                              required:
                              - key
                              - operator
                              type: object
                            type: array
                          matchLabels:
                            additionalProperties:
                              type: string
                            description: matchLabels is a map of {key,value} pairs.
                              A single {key,value} in the matchLabels map is equivalent
                              to an element of matchExpressions, whose key field is
                              "key", the operator is "In", and the values array contains
                              only "value". The requirements are ANDed.
                            type: object
                        type: object
                      name:
                        description: Name is the name of this hook.
                        type: string
                      post:
                        description: PostHooks is a list of BackupResourceHooks to
                          execute after storing the item in the backup. These are
                          executed after all "additional items" from item actions
                          are processed.
                        items:
                          description: BackupResourceHook defines a hook for a resource.
                          properties:
                            exec:
                              description: Exec defines an exec hook.
                              properties:
                                command:
                                  description: Command is the command and arguments
                                    to execute.
                                  items:
                                    type: string
                                  minItems: 1
                                  type: array
                                container:
                                  description: Container is the container in the pod
                                    where the command should be executed. If not specified,
                                    the pod's first container is used.
                                  type: string
                                onError:
                                  description: OnError specifies how Velero should
                                    behave if it encounters an error executing this
                                    hook.
                                  enum:
                                  - Continue
                                  - Fail
                                  type: string
                                timeout:
                                  description: Timeout defines the maximum amount
                                    of time Velero should wait for the hook to complete
                                    before considering the execution a failure.
                                  type: string
                              required:
                              - command
                              type: object
                          required:
                          - exec
                          type: object
                        type: array
                      pre:
                        description: PreHooks is a list of BackupResourceHooks to
                          execute prior to storing the item in the backup. These are
                          executed before any "additional items" from item actions
                          are processed.
                        items:
                          description: BackupResourceHook defines a hook for a resource.
                          properties:
                            exec:
                              description: Exec defines an exec hook.
                              properties:
                                command:
                                  description: Command is the command and arguments
                                    to execute.
                                  items:
                                    type: string
                                  minItems: 1
                                  type: array
                                container:
                                  description: Container is the container in the pod
                                    where the command should be executed. If not specified,
                                    the pod's first container is used.
                                  type: string
                                onError:
                                  description: OnError specifies how Velero should
                                    behave if it encounters an error executing this
                                    hook.
                                  enum:
                                  - Continue
                                  - Fail
                                  type: string
                                timeout:
                                  description: Timeout defines the maximum amount
                                    of time Velero should wait for the hook to complete
                                    before considering the execution a failure.
                                  type: string
                              required:
                              - command
                              type: object
                          required:
                          - exec
                          type: object
                        type: array
                    required:
                    - name
                    type: object
                  nullable: true
                  type: array
              type: object
            includeClusterResources:
              description: IncludeClusterResources specifies whether cluster-scoped
                resources should be included for consideration in the backup.
              nullable: true
              type: boolean
            includedNamespaces:
              description: IncludedNamespaces is a slice of namespace names to include
                objects from. If empty, all namespaces are included.
              items:
                type: string
              nullable: true
              type: array
            includedResources:
              description: IncludedResources is a slice of resource names to include
                in the backup. If empty, all resources are included.
              items:
                type: string
              nullable: true
              type: array
            labelSelector:
              description: LabelSelector is a metav1.LabelSelector to filter with
                when adding individual objects to the backup. If empty or nil, all
                objects are included. Optional.
              nullable: true
              properties:
                matchExpressions:
                  description: matchExpressions is a list of label selector requirements.
                    The requirements are ANDed.
                  items:
                    description: A label selector requirement is a selector that contains
                      values, a key, and an operator that relates the key and values.
                    properties:
                      key:
                        description: key is the label key that the selector applies
                          to.
                        type: string
                      operator:
                        description: operator represents a key's relationship to a
                          set of values. Valid operators are In, NotIn, Exists and
                          DoesNotExist.
                        type: string
                      values:
                        description: values is an array of string values. If the operator
                          is In or NotIn, the values array must be non-empty. If the
                          operator is Exists or DoesNotExist, the values array must
                          be empty. This array is replaced during a strategic merge
                          patch.
                        items:
                          type: string
                        type: array
                    required:
                    - key
                    - operator
                    type: object
                  type: array
                matchLabels:
                  additionalProperties:
                    type: string
                  description: matchLabels is a map of {key,value} pairs. A single
                    {key,value} in the matchLabels map is equivalent to an element
                    of matchExpressions, whose key field is "key", the operator is
                    "In", and the values array contains only "value". The requirements
                    are ANDed.
                  type: object
              type: object
            orderedResources:
              additionalProperties:
                type: string
              description: OrderedResources specifies the backup order of resources
                of specific Kind. The map key is the Kind name and value is a list
                of resource names separated by commas. Each resource name has format
                "namespace/resourcename".  For cluster resources, simply use "resourcename".
              nullable: true
              type: object
            snapshotVolumes:
              description: SnapshotVolumes specifies whether to take cloud snapshots
                of any PV's referenced in the set of objects included in the Backup.
              nullable: true
              type: boolean
            storageLocation:
              description: StorageLocation is a string containing the name of a BackupStorageLocation
                where the backup should be stored.
              type: string
            ttl:
              description: TTL is a time.Duration-parseable string describing how
                long the Backup should be retained for.
              type: string
            volumeSnapshotLocations:
              description: VolumeSnapshotLocations is a list containing names of VolumeSnapshotLocations
                associated with this backup.
              items:
                type: string
              type: array
          type: object
        status:
          description: BackupStatus captures the current status of a Velero backup.
          properties:
            completionTimestamp:
              description: CompletionTimestamp records the time a backup was completed.
                Completion time is recorded even on failed backups. Completion time
                is recorded before uploading the backup object. The server's time
                is used for CompletionTimestamps
              format: date-time
              nullable: true
              type: string
            errors:
              description: Errors is a count of all error messages that were generated
                during execution of the backup.  The actual errors are in the backup's
                log file in object storage.
              type: integer
            expiration:
              description: Expiration is when this Backup is eligible for garbage-collection.
              format: date-time
              nullable: true
              type: string
            formatVersion:
              description: FormatVersion is the backup format version, including major,
                minor, and patch version.
              type: string
            phase:
              description: Phase is the current state of the Backup.
              enum:
              - New
              - FailedValidation
              - InProgress
              - Completed
              - PartiallyFailed
              - Failed
              - Deleting
              type: string
            progress:
              description: Progress contains information about the backup's execution
                progress. Note that this information is best-effort only -- if Velero
                fails to update it during a backup for any reason, it may be inaccurate/stale.
              nullable: true
              properties:
                itemsBackedUp:
                  description: ItemsBackedUp is the number of items that have actually
                    been written to the backup tarball so far.
                  type: integer
                totalItems:
                  description: TotalItems is the total number of items to be backed
                    up. This number may change throughout the execution of the backup
                    due to plugins that return additional related items to back up,
                    the velero.io/exclude-from-backup label, and various other filters
                    that happen as items are processed.
                  type: integer
              type: object
            startTimestamp:
              description: StartTimestamp records the time a backup was started. Separate
                from CreationTimestamp, since that value changes on restores. The
                server's time is used for StartTimestamps
              format: date-time
              nullable: true
              type: string
            validationErrors:
              description: ValidationErrors is a slice of all validation errors (if
                applicable).
              items:
                type: string
              nullable: true
              type: array
            version:
              description: 'Version is the backup format major version. Deprecated:
                Please see FormatVersion'
              type: integer
            volumeSnapshotsAttempted:
              description: VolumeSnapshotsAttempted is the total number of attempted
                volume snapshots for this backup.
              type: integer
            volumeSnapshotsCompleted:
              description: VolumeSnapshotsCompleted is the total number of successfully
                completed volume snapshots for this backup.
              type: integer
            warnings:
              description: Warnings is a count of all warning messages that were generated
                during execution of the backup. The actual warnings are in the backup's
                log file in object storage.
              type: integer
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
	downloadrequests = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    app.kubernetes.io/name: velero
    component: velero
  annotations:
    controller-gen.kubebuilder.io/version: v0.3.0
  creationTimestamp: null
  name: downloadrequests.velero.io
spec:
  group: velero.io
  names:
    kind: DownloadRequest
    listKind: DownloadRequestList
    plural: downloadrequests
    singular: downloadrequest
  preserveUnknownFields: false
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: DownloadRequest is a request to download an artifact from backup
        object storage, such as a backup log file.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: DownloadRequestSpec is the specification for a download request.
          properties:
            target:
              description: Target is what to download (e.g. logs for a backup).
              properties:
                kind:
                  description: Kind is the type of file to download.
                  enum:
                  - BackupLog
                  - BackupContents
                  - BackupVolumeSnapshots
                  - BackupResourceList
                  - RestoreLog
                  - RestoreResults
                  type: string
                name:
                  description: Name is the name of the kubernetes resource with which
                    the file is associated.
                  type: string
              required:
              - kind
              - name
              type: object
          required:
          - target
          type: object
        status:
          description: DownloadRequestStatus is the current status of a DownloadRequest.
          properties:
            downloadURL:
              description: DownloadURL contains the pre-signed URL for the target
                file.
              type: string
            expiration:
              description: Expiration is when this DownloadRequest expires and can
                be deleted by the system.
              format: date-time
              nullable: true
              type: string
            phase:
              description: Phase is the current state of the DownloadRequest.
              enum:
              - New
              - Processed
              type: string
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
	podvolumerestores = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    app.kubernetes.io/name: velero
    component: velero
  annotations:
    controller-gen.kubebuilder.io/version: v0.3.0
  creationTimestamp: null
  name: podvolumerestores.velero.io
spec:
  group: velero.io
  names:
    kind: PodVolumeRestore
    listKind: PodVolumeRestoreList
    plural: podvolumerestores
    singular: podvolumerestore
  preserveUnknownFields: false
  scope: Namespaced
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: PodVolumeRestoreSpec is the specification for a PodVolumeRestore.
          properties:
            backupStorageLocation:
              description: BackupStorageLocation is the name of the backup storage
                location where the restic repository is stored.
              type: string
            pod:
              description: Pod is a reference to the pod containing the volume to
                be restored.
              properties:
                apiVersion:
                  description: API version of the referent.
                  type: string
                fieldPath:
                  description: 'If referring to a piece of an object instead of an
                    entire object, this string should contain a valid JSON/Go field
                    access statement, such as desiredState.manifest.containers[2].
                    For example, if the object reference is to a container within
                    a pod, this would take on a value like: "spec.containers{name}"
                    (where "name" refers to the name of the container that triggered
                    the event) or if no container name is specified "spec.containers[2]"
                    (container with index 2 in this pod). This syntax is chosen only
                    to have some well-defined way of referencing a part of an object.
                    TODO: this design is not final and this field is subject to change
                    in the future.'
                  type: string
                kind:
                  description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
                  type: string
                name:
                  description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names'
                  type: string
                namespace:
                  description: 'Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/'
                  type: string
                resourceVersion:
                  description: 'Specific resourceVersion to which this reference is
                    made, if any. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency'
                  type: string
                uid:
                  description: 'UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids'
                  type: string
              type: object
            repoIdentifier:
              description: RepoIdentifier is the restic repository identifier.
              type: string
            snapshotID:
              description: SnapshotID is the ID of the volume snapshot to be restored.
              type: string
            volume:
              description: Volume is the name of the volume within the Pod to be restored.
              type: string
          required:
          - backupStorageLocation
          - pod
          - repoIdentifier
          - snapshotID
          - volume
          type: object
        status:
          description: PodVolumeRestoreStatus is the current status of a PodVolumeRestore.
          properties:
            completionTimestamp:
              description: CompletionTimestamp records the time a restore was completed.
                Completion time is recorded even on failed restores. The server's
                time is used for CompletionTimestamps
              format: date-time
              nullable: true
              type: string
            message:
              description: Message is a message about the pod volume restore's status.
              type: string
            phase:
              description: Phase is the current state of the PodVolumeRestore.
              enum:
              - New
              - InProgress
              - Completed
              - Failed
              type: string
            progress:
              description: Progress holds the total number of bytes of the snapshot
                and the current number of restored bytes. This can be used to display
                progress information about the restore operation.
              properties:
                bytesDone:
                  format: int64
                  type: integer
                totalBytes:
                  format: int64
                  type: integer
              type: object
            startTimestamp:
              description: StartTimestamp records the time a restore was started.
                The server's time is used for StartTimestamps
              format: date-time
              nullable: true
              type: string
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
	restores = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    app.kubernetes.io/name: velero
    component: velero
  annotations:
    controller-gen.kubebuilder.io/version: v0.3.0
  creationTimestamp: null
  name: restores.velero.io
spec:
  group: velero.io
  names:
    kind: Restore
    listKind: RestoreList
    plural: restores
    singular: restore
  preserveUnknownFields: false
  scope: Namespaced
  validation:
    openAPIV3Schema:
      description: Restore is a Velero resource that represents the application of
        resources from a Velero backup to a target Kubernetes cluster.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: RestoreSpec defines the specification for a Velero restore.
          properties:
            backupName:
              description: BackupName is the unique name of the Velero backup to restore
                from.
              type: string
            excludedNamespaces:
              description: ExcludedNamespaces contains a list of namespaces that are
                not included in the restore.
              items:
                type: string
              nullable: true
              type: array
            excludedResources:
              description: ExcludedResources is a slice of resource names that are
                not included in the restore.
              items:
                type: string
              nullable: true
              type: array
            hooks:
              description: Hooks represent custom behaviors that should be executed
                during or post restore.
              properties:
                resources:
                  items:
                    description: RestoreResourceHookSpec defines one or more RestoreResrouceHooks
                      that should be executed based on the rules defined for namespaces,
                      resources, and label selector.
                    properties:
                      excludedNamespaces:
                        description: ExcludedNamespaces specifies the namespaces to
                          which this hook spec does not apply.
                        items:
                          type: string
                        nullable: true
                        type: array
                      excludedResources:
                        description: ExcludedResources specifies the resources to
                          which this hook spec does not apply.
                        items:
                          type: string
                        nullable: true
                        type: array
                      includedNamespaces:
                        description: IncludedNamespaces specifies the namespaces to
                          which this hook spec applies. If empty, it applies to all
                          namespaces.
                        items:
                          type: string
                        nullable: true
                        type: array
                      includedResources:
                        description: IncludedResources specifies the resources to
                          which this hook spec applies. If empty, it applies to all
                          resources.
                        items:
                          type: string
                        nullable: true
                        type: array
                      labelSelector:
                        description: LabelSelector, if specified, filters the resources
                          to which this hook spec applies.
                        nullable: true
                        properties:
                          matchExpressions:
                            description: matchExpressions is a list of label selector
                              requirements. The requirements are ANDed.
                            items:
                              description: A label selector requirement is a selector
                                that contains values, a key, and an operator that
                                relates the key and values.
                              properties:
                                key:
                                  description: key is the label key that the selector
                                    applies to.
                                  type: string
                                operator:
                                  description: operator represents a key's relationship
                                    to a set of values. Valid operators are In, NotIn,
                                    Exists and DoesNotExist.
                                  type: string
                                values:
                                  description: values is an array of string values.
                                    If the operator is In or NotIn, the values array
                                    must be non-empty. If the operator is Exists or
                                    DoesNotExist, the values array must be empty.
                                    This array is replaced during a strategic merge
                                    patch.
                                  items:
                                    type: string
                                  type: array
                              required:
                              - key
                              - operator
                              type: object
                            type: array
                          matchLabels:
                            additionalProperties:
                              type: string
                            description: matchLabels is a map of {key,value} pairs.
                              A single {key,value} in the matchLabels map is equivalent
                              to an element of matchExpressions, whose key field is
                              "key", the operator is "In", and the values array contains
                              only "value". The requirements are ANDed.
                            type: object
                        type: object
                      name:
                        description: Name is the name of this hook.
                        type: string
                      postHooks:
                        description: PostHooks is a list of RestoreResourceHooks to
                          execute during and after restoring a resource.
                        items:
                          description: RestoreResourceHook defines a restore hook
                            for a resource.
                          properties:
                            exec:
                              description: Exec defines an exec restore hook.
                              properties:
                                command:
                                  description: Command is the command and arguments
                                    to execute from within a container after a pod
                                    has been restored.
                                  items:
                                    type: string
                                  minItems: 1
                                  type: array
                                container:
                                  description: Container is the container in the pod
                                    where the command should be executed. If not specified,
                                    the pod's first container is used.
                                  type: string
                                execTimeout:
                                  description: ExecTimeout defines the maximum amount
                                    of time Velero should wait for the hook to complete
                                    before considering the execution a failure.
                                  type: string
                                onError:
                                  description: OnError specifies how Velero should
                                    behave if it encounters an error executing this
                                    hook.
                                  enum:
                                  - Continue
                                  - Fail
                                  type: string
                                waitTimeout:
                                  description: WaitTimeout defines the maximum amount
                                    of time Velero should wait for the container to
                                    be Ready before attempting to run the command.
                                  type: string
                              required:
                              - command
                              type: object
                            init:
                              description: Init defines an init restore hook.
                              properties:
                                initContainers:
                                  description: InitContainers is list of init containers
                                    to be added to a pod during its restore.
                                  items:
                                    description: A single application container that
                                      you want to run within a pod.
                                    properties:
                                      args:
                                        description: 'Arguments to the entrypoint.
                                          The docker image''s CMD is used if this
                                          is not provided. Variable references $(VAR_NAME)
                                          are expanded using the container''s environment.
                                          If a variable cannot be resolved, the reference
                                          in the input string will be unchanged. The
                                          $(VAR_NAME) syntax can be escaped with a
                                          double $$, ie: $$(VAR_NAME). Escaped references
                                          will never be expanded, regardless of whether
                                          the variable exists or not. Cannot be updated.
                                          More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell'
                                        items:
                                          type: string
                                        type: array
                                      command:
                                        description: 'Entrypoint array. Not executed
                                          within a shell. The docker image''s ENTRYPOINT
                                          is used if this is not provided. Variable
                                          references $(VAR_NAME) are expanded using
                                          the container''s environment. If a variable
                                          cannot be resolved, the reference in the
                                          input string will be unchanged. The $(VAR_NAME)
                                          syntax can be escaped with a double $$,
                                          ie: $$(VAR_NAME). Escaped references will
                                          never be expanded, regardless of whether
                                          the variable exists or not. Cannot be updated.
                                          More info: https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell'
                                        items:
                                          type: string
                                        type: array
                                      env:
                                        description: List of environment variables
                                          to set in the container. Cannot be updated.
                                        items:
                                          description: EnvVar represents an environment
                                            variable present in a Container.
                                          properties:
                                            name:
                                              description: Name of the environment
                                                variable. Must be a C_IDENTIFIER.
                                              type: string
                                            value:
                                              description: 'Variable references $(VAR_NAME)
                                                are expanded using the previous defined
                                                environment variables in the container
                                                and any service environment variables.
                                                If a variable cannot be resolved,
                                                the reference in the input string
                                                will be unchanged. The $(VAR_NAME)
                                                syntax can be escaped with a double
                                                $$, ie: $$(VAR_NAME). Escaped references
                                                will never be expanded, regardless
                                                of whether the variable exists or
                                                not. Defaults to "".'
                                              type: string
                                            valueFrom:
                                              description: Source for the environment
                                                variable's value. Cannot be used if
                                                value is not empty.
                                              properties:
                                                configMapKeyRef:
                                                  description: Selects a key of a
                                                    ConfigMap.
                                                  properties:
                                                    key:
                                                      description: The key to select.
                                                      type: string
                                                    name:
                                                      description: 'Name of the referent.
                                                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                                        TODO: Add other useful fields.
                                                        apiVersion, kind, uid?'
                                                      type: string
                                                    optional:
                                                      description: Specify whether
                                                        the ConfigMap or its key must
                                                        be defined
                                                      type: boolean
                                                  required:
                                                  - key
                                                  type: object
                                                fieldRef:
                                                  description: 'Selects a field of
                                                    the pod: supports metadata.name,
                                                    metadata.namespace, "metadata.labels[<KEY>]",
                                                    "metadata.annotations[<KEY>]",
                                                    spec.nodeName, spec.serviceAccountName,status.hostIP, status.podIP, status.podIPs.'
                                                  properties:
                                                    apiVersion:
                                                      description: Version of the
                                                        schema the FieldPath is written
                                                        in terms of, defaults to "v1".
                                                      type: string
                                                    fieldPath:
                                                      description: Path of the field
                                                        to select in the specified
                                                        API version.
                                                      type: string
                                                  required:
                                                  - fieldPath
                                                  type: object
                                                resourceFieldRef:
                                                  description: 'Selects a resource
                                                    of the container: only resources
                                                    limits and requests (limits.cpu,
                                                    limits.memory, limits.ephemeral-storage,
                                                    requests.cpu, requests.memory
                                                    and requests.ephemeral-storage)
                                                    are currently supported.'
                                                  properties:
                                                    containerName:
                                                      description: 'Container name:
                                                        required for volumes, optional
                                                        for env vars'
                                                      type: string
                                                    divisor:
                                                      anyOf:
                                                      - type: integer
                                                      - type: string
                                                      description: Specifies the output
                                                        format of the exposed resources,
                                                        defaults to "1"
                                                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                                      x-kubernetes-int-or-string: true
                                                    resource:
                                                      description: 'Required: resource
                                                        to select'
                                                      type: string
                                                  required:
                                                  - resource
                                                  type: object
                                                secretKeyRef:
                                                  description: Selects a key of a
                                                    secret in the pod's namespace
                                                  properties:
                                                    key:
                                                      description: The key of the
                                                        secret to select from.  Must
                                                        be a valid secret key.
                                                      type: string
                                                    name:
                                                      description: 'Name of the referent.
                                                        More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                                        TODO: Add other useful fields.
                                                        apiVersion, kind, uid?'
                                                      type: string
                                                    optional:
                                                      description: Specify whether
                                                        the Secret or its key must
                                                        be defined
                                                      type: boolean
                                                  required:
                                                  - key
                                                  type: object
                                              type: object
                                          required:
                                          - name
                                          type: object
                                        type: array
                                      envFrom:
                                        description: List of sources to populate environment
                                          variables in the container. The keys defined
                                          within a source must be a C_IDENTIFIER.
                                          All invalid keys will be reported as an
                                          event when the container is starting. When
                                          a key exists in multiple sources, the value
                                          associated with the last source will take
                                          precedence. Values defined by an Env with
                                          a duplicate key will take precedence. Cannot
                                          be updated.
                                        items:
                                          description: EnvFromSource represents the
                                            source of a set of ConfigMaps
                                          properties:
                                            configMapRef:
                                              description: The ConfigMap to select
                                                from
                                              properties:
                                                name:
                                                  description: 'Name of the referent.
                                                    More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                                    TODO: Add other useful fields.
                                                    apiVersion, kind, uid?'
                                                  type: string
                                                optional:
                                                  description: Specify whether the
                                                    ConfigMap must be defined
                                                  type: boolean
                                              type: object
                                            prefix:
                                              description: An optional identifier
                                                to prepend to each key in the ConfigMap.
                                                Must be a C_IDENTIFIER.
                                              type: string
                                            secretRef:
                                              description: The Secret to select from
                                              properties:
                                                name:
                                                  description: 'Name of the referent.
                                                    More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                                    TODO: Add other useful fields.
                                                    apiVersion, kind, uid?'
                                                  type: string
                                                optional:
                                                  description: Specify whether the
                                                    Secret must be defined
                                                  type: boolean
                                              type: object
                                          type: object
                                        type: array
                                      image:
                                        description: 'Docker image name. More info:
                                          https://kubernetes.io/docs/concepts/containers/images
                                          This field is optional to allow higher level
                                          config management to default or override
                                          container images in workload controllers
                                          like Deployments and StatefulSets.'
                                        type: string
                                      imagePullPolicy:
                                        description: 'Image pull policy. One of Always,
                                          Never, IfNotPresent. Defaults to Always
                                          if :latest tag is specified, or IfNotPresent
                                          otherwise. Cannot be updated. More info:
                                          https://kubernetes.io/docs/concepts/containers/images#updating-images'
                                        type: string
                                      lifecycle:
                                        description: Actions that the management system
                                          should take in response to container lifecycle
                                          events. Cannot be updated.
                                        properties:
                                          postStart:
                                            description: 'PostStart is called immediately
                                              after a container is created. If the
                                              handler fails, the container is terminated
                                              and restarted according to its restart
                                              policy. Other management of the container
                                              blocks until the hook completes. More
                                              info: https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks'
                                            properties:
                                              exec:
                                                description: One and only one of the
                                                  following should be specified. Exec
                                                  specifies the action to take.
                                                properties:
                                                  command:
                                                    description: Command is the command
                                                      line to execute inside the container,
                                                      the working directory for the
                                                      command  is root ('/') in the
                                                      container's filesystem. The
                                                      command is simply exec'd, it
                                                      is not run inside a shell, so
                                                      traditional shell instructions
                                                      ('|', etc) won't work. To use
                                                      a shell, you need to explicitly
                                                      call out to that shell. Exit
                                                      status of 0 is treated as live/healthy
                                                      and non-zero is unhealthy.
                                                    items:
                                                      type: string
                                                    type: array
                                                type: object
                                              httpGet:
                                                description: HTTPGet specifies the
                                                  http request to perform.
                                                properties:
                                                  host:
                                                    description: Host name to connect
                                                      to, defaults to the pod IP.
                                                      You probably want to set "Host"
                                                      in httpHeaders instead.
                                                    type: string
                                                  httpHeaders:
                                                    description: Custom headers to
                                                      set in the request. HTTP allows
                                                      repeated headers.
                                                    items:
                                                      description: HTTPHeader describes
                                                        a custom header to be used
                                                        in HTTP probes
                                                      properties:
                                                        name:
                                                          description: The header
                                                            field name
                                                          type: string
                                                        value:
                                                          description: The header
                                                            field value
                                                          type: string
                                                      required:
                                                      - name
                                                      - value
                                                      type: object
                                                    type: array
                                                  path:
                                                    description: Path to access on
                                                      the HTTP server.
                                                    type: string
                                                  port:
                                                    anyOf:
                                                    - type: integer
                                                    - type: string
                                                    description: Name or number of
                                                      the port to access on the container.
                                                      Number must be in the range
                                                      1 to 65535. Name must be an
                                                      IANA_SVC_NAME.
                                                    x-kubernetes-int-or-string: true
                                                  scheme:
                                                    description: Scheme to use for
                                                      connecting to the host. Defaults
                                                      to HTTP.
                                                    type: string
                                                required:
                                                - port
                                                type: object
                                              tcpSocket:
                                                description: 'TCPSocket specifies
                                                  an action involving a TCP port.
                                                  TCP hooks not yet supported TODO:
                                                  implement a realistic TCP lifecycle
                                                  hook'
                                                properties:
                                                  host:
                                                    description: 'Optional: Host name
                                                      to connect to, defaults to the
                                                      pod IP.'
                                                    type: string
                                                  port:
                                                    anyOf:
                                                    - type: integer
                                                    - type: string
                                                    description: Number or name of
                                                      the port to access on the container.
                                                      Number must be in the range
                                                      1 to 65535. Name must be an
                                                      IANA_SVC_NAME.
                                                    x-kubernetes-int-or-string: true
                                                required:
                                                - port
                                                type: object
                                            type: object
                                          preStop:
                                            description: 'PreStop is called immediately
                                              before a container is terminated due
                                              to an API request or management event
                                              such as liveness/startup probe failure,
                                              preemption, resource contention, etc.
                                              The handler is not called if the container
                                              crashes or exits. The reason for termination
                                              is passed to the handler. The Pod''s
                                              termination grace period countdown begins
                                              before the PreStop hooked is executed.
                                              Regardless of the outcome of the handler,
                                              the container will eventually terminate
                                              within the Pod''s termination grace
                                              period. Other management of the container
                                              blocks until the hook completes or until
                                              the termination grace period is reached.
                                              More info: https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks'
                                            properties:
                                              exec:
                                                description: One and only one of the
                                                  following should be specified. Exec
                                                  specifies the action to take.
                                                properties:
                                                  command:
                                                    description: Command is the command
                                                      line to execute inside the container,
                                                      the working directory for the
                                                      command  is root ('/') in the
                                                      container's filesystem. The
                                                      command is simply exec'd, it
                                                      is not run inside a shell, so
                                                      traditional shell instructions
                                                      ('|', etc) won't work. To use
                                                      a shell, you need to explicitly
                                                      call out to that shell. Exit
                                                      status of 0 is treated as live/healthy
                                                      and non-zero is unhealthy.
                                                    items:
                                                      type: string
                                                    type: array
                                                type: object
                                              httpGet:
                                                description: HTTPGet specifies the
                                                  http request to perform.
                                                properties:
                                                  host:
                                                    description: Host name to connect
                                                      to, defaults to the pod IP.
                                                      You probably want to set "Host"
                                                      in httpHeaders instead.
                                                    type: string
                                                  httpHeaders:
                                                    description: Custom headers to
                                                      set in the request. HTTP allows
                                                      repeated headers.
                                                    items:
                                                      description: HTTPHeader describes
                                                        a custom header to be used
                                                        in HTTP probes
                                                      properties:
                                                        name:
                                                          description: The header
                                                            field name
                                                          type: string
                                                        value:
                                                          description: The header
                                                            field value
                                                          type: string
                                                      required:
                                                      - name
                                                      - value
                                                      type: object
                                                    type: array
                                                  path:
                                                    description: Path to access on
                                                      the HTTP server.
                                                    type: string
                                                  port:
                                                    anyOf:
                                                    - type: integer
                                                    - type: string
                                                    description: Name or number of
                                                      the port to access on the container.
                                                      Number must be in the range
                                                      1 to 65535. Name must be an
                                                      IANA_SVC_NAME.
                                                    x-kubernetes-int-or-string: true
                                                  scheme:
                                                    description: Scheme to use for
                                                      connecting to the host. Defaults
                                                      to HTTP.
                                                    type: string
                                                required:
                                                - port
                                                type: object
                                              tcpSocket:
                                                description: 'TCPSocket specifies
                                                  an action involving a TCP port.
                                                  TCP hooks not yet supported TODO:
                                                  implement a realistic TCP lifecycle
                                                  hook'
                                                properties:
                                                  host:
                                                    description: 'Optional: Host name
                                                      to connect to, defaults to the
                                                      pod IP.'
                                                    type: string
                                                  port:
                                                    anyOf:
                                                    - type: integer
                                                    - type: string
                                                    description: Number or name of
                                                      the port to access on the container.
                                                      Number must be in the range
                                                      1 to 65535. Name must be an
                                                      IANA_SVC_NAME.
                                                    x-kubernetes-int-or-string: true
                                                required:
                                                - port
                                                type: object
                                            type: object
                                        type: object
                                      livenessProbe:
                                        description: 'Periodic probe of container
                                          liveness. Container will be restarted if
                                          the probe fails. Cannot be updated. More
                                          info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                        properties:
                                          exec:
                                            description: One and only one of the following
                                              should be specified. Exec specifies
                                              the action to take.
                                            properties:
                                              command:
                                                description: Command is the command
                                                  line to execute inside the container,
                                                  the working directory for the command  is
                                                  root ('/') in the container's filesystem.
                                                  The command is simply exec'd, it
                                                  is not run inside a shell, so traditional
                                                  shell instructions ('|', etc) won't
                                                  work. To use a shell, you need to
                                                  explicitly call out to that shell.
                                                  Exit status of 0 is treated as live/healthy
                                                  and non-zero is unhealthy.
                                                items:
                                                  type: string
                                                type: array
                                            type: object
                                          failureThreshold:
                                            description: Minimum consecutive failures
                                              for the probe to be considered failed
                                              after having succeeded. Defaults to
                                              3. Minimum value is 1.
                                            format: int32
                                            type: integer
                                          httpGet:
                                            description: HTTPGet specifies the http
                                              request to perform.
                                            properties:
                                              host:
                                                description: Host name to connect
                                                  to, defaults to the pod IP. You
                                                  probably want to set "Host" in httpHeaders
                                                  instead.
                                                type: string
                                              httpHeaders:
                                                description: Custom headers to set
                                                  in the request. HTTP allows repeated
                                                  headers.
                                                items:
                                                  description: HTTPHeader describes
                                                    a custom header to be used in
                                                    HTTP probes
                                                  properties:
                                                    name:
                                                      description: The header field
                                                        name
                                                      type: string
                                                    value:
                                                      description: The header field
                                                        value
                                                      type: string
                                                  required:
                                                  - name
                                                  - value
                                                  type: object
                                                type: array
                                              path:
                                                description: Path to access on the
                                                  HTTP server.
                                                type: string
                                              port:
                                                anyOf:
                                                - type: integer
                                                - type: string
                                                description: Name or number of the
                                                  port to access on the container.
                                                  Number must be in the range 1 to
                                                  65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                              scheme:
                                                description: Scheme to use for connecting
                                                  to the host. Defaults to HTTP.
                                                type: string
                                            required:
                                            - port
                                            type: object
                                          initialDelaySeconds:
                                            description: 'Number of seconds after
                                              the container has started before liveness
                                              probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                            format: int32
                                            type: integer
                                          periodSeconds:
                                            description: How often (in seconds) to
                                              perform the probe. Default to 10 seconds.
                                              Minimum value is 1.
                                            format: int32
                                            type: integer
                                          successThreshold:
                                            description: Minimum consecutive successes
                                              for the probe to be considered successful
                                              after having failed. Defaults to 1.
                                              Must be 1 for liveness and startup.
                                              Minimum value is 1.
                                            format: int32
                                            type: integer
                                          tcpSocket:
                                            description: 'TCPSocket specifies an action
                                              involving a TCP port. TCP hooks not
                                              yet supported TODO: implement a realistic
                                              TCP lifecycle hook'
                                            properties:
                                              host:
                                                description: 'Optional: Host name
                                                  to connect to, defaults to the pod
                                                  IP.'
                                                type: string
                                              port:
                                                anyOf:
                                                - type: integer
                                                - type: string
                                                description: Number or name of the
                                                  port to access on the container.
                                                  Number must be in the range 1 to
                                                  65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                            required:
                                            - port
                                            type: object
                                          timeoutSeconds:
                                            description: 'Number of seconds after
                                              which the probe times out. Defaults
                                              to 1 second. Minimum value is 1. More
                                              info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                            format: int32
                                            type: integer
                                        type: object
                                      name:
                                        description: Name of the container specified
                                          as a DNS_LABEL. Each container in a pod
                                          must have a unique name (DNS_LABEL). Cannot
                                          be updated.
                                        type: string
                                      ports:
                                        description: List of ports to expose from
                                          the container. Exposing a port here gives
                                          the system additional information about
                                          the network connections a container uses,
                                          but is primarily informational. Not specifying
                                          a port here DOES NOT prevent that port from
                                          being exposed. Any port which is listening
                                          on the default "0.0.0.0" address inside
                                          a container will be accessible from the
                                          network. Cannot be updated.
                                        items:
                                          description: ContainerPort represents a
                                            network port in a single container.
                                          properties:
                                            containerPort:
                                              description: Number of port to expose
                                                on the pod's IP address. This must
                                                be a valid port number, 0 < x < 65536.
                                              format: int32
                                              type: integer
                                            hostIP:
                                              description: What host IP to bind the
                                                external port to.
                                              type: string
                                            hostPort:
                                              description: Number of port to expose
                                                on the host. If specified, this must
                                                be a valid port number, 0 < x < 65536.
                                                If HostNetwork is specified, this
                                                must match ContainerPort. Most containers
                                                do not need this.
                                              format: int32
                                              type: integer
                                            name:
                                              description: If specified, this must
                                                be an IANA_SVC_NAME and unique within
                                                the pod. Each named port in a pod
                                                must have a unique name. Name for
                                                the port that can be referred to by
                                                services.
                                              type: string
                                            protocol:
                                              description: Protocol for port. Must
                                                be UDP, TCP, or SCTP. Defaults to
                                                "TCP".
                                              type: string
                                          required:
                                          - containerPort
                                          - protocol
                                          type: object
                                        type: array
                                        x-kubernetes-list-map-keys:
                                        - containerPort
                                        - protocol
                                        x-kubernetes-list-type: map
                                      readinessProbe:
                                        description: 'Periodic probe of container
                                          service readiness. Container will be removed
                                          from service endpoints if the probe fails.
                                          Cannot be updated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                        properties:
                                          exec:
                                            description: One and only one of the following
                                              should be specified. Exec specifies
                                              the action to take.
                                            properties:
                                              command:
                                                description: Command is the command
                                                  line to execute inside the container,
                                                  the working directory for the command  is
                                                  root ('/') in the container's filesystem.
                                                  The command is simply exec'd, it
                                                  is not run inside a shell, so traditional
                                                  shell instructions ('|', etc) won't
                                                  work. To use a shell, you need to
                                                  explicitly call out to that shell.
                                                  Exit status of 0 is treated as live/healthy
                                                  and non-zero is unhealthy.
                                                items:
                                                  type: string
                                                type: array
                                            type: object
                                          failureThreshold:
                                            description: Minimum consecutive failures
                                              for the probe to be considered failed
                                              after having succeeded. Defaults to
                                              3. Minimum value is 1.
                                            format: int32
                                            type: integer
                                          httpGet:
                                            description: HTTPGet specifies the http
                                              request to perform.
                                            properties:
                                              host:
                                                description: Host name to connect
                                                  to, defaults to the pod IP. You
                                                  probably want to set "Host" in httpHeaders
                                                  instead.
                                                type: string
                                              httpHeaders:
                                                description: Custom headers to set
                                                  in the request. HTTP allows repeated
                                                  headers.
                                                items:
                                                  description: HTTPHeader describes
                                                    a custom header to be used in
                                                    HTTP probes
                                                  properties:
                                                    name:
                                                      description: The header field
                                                        name
                                                      type: string
                                                    value:
                                                      description: The header field
                                                        value
                                                      type: string
                                                  required:
                                                  - name
                                                  - value
                                                  type: object
                                                type: array
                                              path:
                                                description: Path to access on the
                                                  HTTP server.
                                                type: string
                                              port:
                                                anyOf:
                                                - type: integer
                                                - type: string
                                                description: Name or number of the
                                                  port to access on the container.
                                                  Number must be in the range 1 to
                                                  65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                              scheme:
                                                description: Scheme to use for connecting
                                                  to the host. Defaults to HTTP.
                                                type: string
                                            required:
                                            - port
                                            type: object
                                          initialDelaySeconds:
                                            description: 'Number of seconds after
                                              the container has started before liveness
                                              probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                            format: int32
                                            type: integer
                                          periodSeconds:
                                            description: How often (in seconds) to
                                              perform the probe. Default to 10 seconds.
                                              Minimum value is 1.
                                            format: int32
                                            type: integer
                                          successThreshold:
                                            description: Minimum consecutive successes
                                              for the probe to be considered successful
                                              after having failed. Defaults to 1.
                                              Must be 1 for liveness and startup.
                                              Minimum value is 1.
                                            format: int32
                                            type: integer
                                          tcpSocket:
                                            description: 'TCPSocket specifies an action
                                              involving a TCP port. TCP hooks not
                                              yet supported TODO: implement a realistic
                                              TCP lifecycle hook'
                                            properties:
                                              host:
                                                description: 'Optional: Host name
                                                  to connect to, defaults to the pod
                                                  IP.'
                                                type: string
                                              port:
                                                anyOf:
                                                - type: integer
                                                - type: string
                                                description: Number or name of the
                                                  port to access on the container.
                                                  Number must be in the range 1 to
                                                  65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                            required:
                                            - port
                                            type: object
                                          timeoutSeconds:
                                            description: 'Number of seconds after
                                              which the probe times out. Defaults
                                              to 1 second. Minimum value is 1. More
                                              info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                            format: int32
                                            type: integer
                                        type: object
                                      resources:
                                        description: 'Compute Resources required by
                                          this container. Cannot be updated. More
                                          info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                                        properties:
                                          limits:
                                            additionalProperties:
                                              anyOf:
                                              - type: integer
                                              - type: string
                                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                              x-kubernetes-int-or-string: true
                                            description: 'Limits describes the maximum
                                              amount of compute resources allowed.
                                              More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                                            type: object
                                          requests:
                                            additionalProperties:
                                              anyOf:
                                              - type: integer
                                              - type: string
                                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                              x-kubernetes-int-or-string: true
                                            description: 'Requests describes the minimum
                                              amount of compute resources required.
                                              If Requests is omitted for a container,
                                              it defaults to Limits if that is explicitly
                                              specified, otherwise to an implementation-defined
                                              value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                                            type: object
                                        type: object
                                      securityContext:
                                        description: 'Security options the pod should
                                          run with. More info: https://kubernetes.io/docs/concepts/policy/security-context/
                                          More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context/'
                                        properties:
                                          allowPrivilegeEscalation:
                                            description: 'AllowPrivilegeEscalation
                                              controls whether a process can gain
                                              more privileges than its parent process.
                                              This bool directly controls if the no_new_privs
                                              flag will be set on the container process.
                                              AllowPrivilegeEscalation is true always
                                              when the container is: 1) run as Privileged
                                              2) has CAP_SYS_ADMIN'
                                            type: boolean
                                          capabilities:
                                            description: The capabilities to add/drop
                                              when running containers. Defaults to
                                              the default set of capabilities granted
                                              by the container runtime.
                                            properties:
                                              add:
                                                description: Added capabilities
                                                items:
                                                  description: Capability represent
                                                    POSIX capabilities type
                                                  type: string
                                                type: array
                                              drop:
                                                description: Removed capabilities
                                                items:
                                                  description: Capability represent
                                                    POSIX capabilities type
                                                  type: string
                                                type: array
                                            type: object
                                          privileged:
                                            description: Run container in privileged
                                              mode. Processes in privileged containers
                                              are essentially equivalent to root on
                                              the host. Defaults to false.
                                            type: boolean
                                          procMount:
                                            description: procMount denotes the type
                                              of proc mount to use for the containers.
                                              The default is DefaultProcMount which
                                              uses the container runtime defaults
                                              for readonly paths and masked paths.
                                              This requires the ProcMountType feature
                                              flag to be enabled.
                                            type: string
                                          readOnlyRootFilesystem:
                                            description: Whether this container has
                                              a read-only root filesystem. Default
                                              is false.
                                            type: boolean
                                          runAsGroup:
                                            description: The GID to run the entrypoint
                                              of the container process. Uses runtime
                                              default if unset. May also be set in
                                              PodSecurityContext.  If set in both
                                              SecurityContext and PodSecurityContext,
                                              the value specified in SecurityContext
                                              takes precedence.
                                            format: int64
                                            type: integer
                                          runAsNonRoot:
                                            description: Indicates that the container
                                              must run as a non-root user. If true,
                                              the Kubelet will validate the image
                                              at runtime to ensure that it does not
                                              run as UID 0 (root) and fail to start
                                              the container if it does. If unset or
                                              false, no such validation will be performed.
                                              May also be set in PodSecurityContext.  If
                                              set in both SecurityContext and PodSecurityContext,
                                              the value specified in SecurityContext
                                              takes precedence.
                                            type: boolean
                                          runAsUser:
                                            description: The UID to run the entrypoint
                                              of the container process. Defaults to
                                              user specified in image metadata if
                                              unspecified. May also be set in PodSecurityContext.  If
                                              set in both SecurityContext and PodSecurityContext,
                                              the value specified in SecurityContext
                                              takes precedence.
                                            format: int64
                                            type: integer
                                          seLinuxOptions:
                                            description: The SELinux context to be
                                              applied to the container. If unspecified,
                                              the container runtime will allocate
                                              a random SELinux context for each container.  May
                                              also be set in PodSecurityContext.  If
                                              set in both SecurityContext and PodSecurityContext,
                                              the value specified in SecurityContext
                                              takes precedence.
                                            properties:
                                              level:
                                                description: Level is SELinux level
                                                  label that applies to the container.
                                                type: string
                                              role:
                                                description: Role is a SELinux role
                                                  label that applies to the container.
                                                type: string
                                              type:
                                                description: Type is a SELinux type
                                                  label that applies to the container.
                                                type: string
                                              user:
                                                description: User is a SELinux user
                                                  label that applies to the container.
                                                type: string
                                            type: object
                                          seccompProfile:
                                            description: The seccomp options to use
                                              by this container. If seccomp options
                                              are provided at both the pod & container
                                              level, the container options override
                                              the pod options.
                                            properties:
                                              localhostProfile:
                                                description: localhostProfile indicates
                                                  a profile defined in a file on the
                                                  node should be used. The profile
                                                  must be preconfigured on the node
                                                  to work. Must be a descending path,
                                                  relative to the kubelet's configured
                                                  seccomp profile location. Must only
                                                  be set if type is "Localhost".
                                                type: string
                                              type:
                                                description: "type indicates which
                                                  kind of seccomp profile will be
                                                  applied. Valid options are: \n Localhost
                                                  - a profile defined in a file on
                                                  the node should be used. RuntimeDefault
                                                  - the container runtime default
                                                  profile should be used. Unconfined
                                                  - no profile should be applied."
                                                type: string
                                            required:
                                            - type
                                            type: object
                                          windowsOptions:
                                            description: The Windows specific settings
                                              applied to all containers. If unspecified,
                                              the options from the PodSecurityContext
                                              will be used. If set in both SecurityContext
                                              and PodSecurityContext, the value specified
                                              in SecurityContext takes precedence.
                                            properties:
                                              gmsaCredentialSpec:
                                                description: GMSACredentialSpec is
                                                  where the GMSA admission webhook
                                                  (https://github.com/kubernetes-sigs/windows-gmsa)
                                                  inlines the contents of the GMSA
                                                  credential spec named by the GMSACredentialSpecName
                                                  field.
                                                type: string
                                              gmsaCredentialSpecName:
                                                description: GMSACredentialSpecName
                                                  is the name of the GMSA credential
                                                  spec to use.
                                                type: string
                                              runAsUserName:
                                                description: The UserName in Windows
                                                  to run the entrypoint of the container
                                                  process. Defaults to the user specified
                                                  in image metadata if unspecified.
                                                  May also be set in PodSecurityContext.
                                                  If set in both SecurityContext and
                                                  PodSecurityContext, the value specified
                                                  in SecurityContext takes precedence.
                                                type: string
                                            type: object
                                        type: object
                                      startupProbe:
                                        description: 'StartupProbe indicates that
                                          the Pod has successfully initialized. If
                                          specified, no other probes are executed
                                          until this completes successfully. If this
                                          probe fails, the Pod will be restarted,
                                          just as if the livenessProbe failed. This
                                          can be used to provide different probe parameters
                                          at the beginning of a Pod''s lifecycle,
                                          when it might take a long time to load data
                                          or warm a cache, than during steady-state
                                          operation. This cannot be updated. This
                                          is a beta feature enabled by the StartupProbe
                                          feature flag. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                        properties:
                                          exec:
                                            description: One and only one of the following
                                              should be specified. Exec specifies
                                              the action to take.
                                            properties:
                                              command:
                                                description: Command is the command
                                                  line to execute inside the container,
                                                  the working directory for the command  is
                                                  root ('/') in the container's filesystem.
                                                  The command is simply exec'd, it
                                                  is not run inside a shell, so traditional
                                                  shell instructions ('|', etc) won't
                                                  work. To use a shell, you need to
                                                  explicitly call out to that shell.
                                                  Exit status of 0 is treated as live/healthy
                                                  and non-zero is unhealthy.
                                                items:
                                                  type: string
                                                type: array
                                            type: object
                                          failureThreshold:
                                            description: Minimum consecutive failures
                                              for the probe to be considered failed
                                              after having succeeded. Defaults to
                                              3. Minimum value is 1.
                                            format: int32
                                            type: integer
                                          httpGet:
                                            description: HTTPGet specifies the http
                                              request to perform.
                                            properties:
                                              host:
                                                description: Host name to connect
                                                  to, defaults to the pod IP. You
                                                  probably want to set "Host" in httpHeaders
                                                  instead.
                                                type: string
                                              httpHeaders:
                                                description: Custom headers to set
                                                  in the request. HTTP allows repeated
                                                  headers.
                                                items:
                                                  description: HTTPHeader describes
                                                    a custom header to be used in
                                                    HTTP probes
                                                  properties:
                                                    name:
                                                      description: The header field
                                                        name
                                                      type: string
                                                    value:
                                                      description: The header field
                                                        value
                                                      type: string
                                                  required:
                                                  - name
                                                  - value
                                                  type: object
                                                type: array
                                              path:
                                                description: Path to access on the
                                                  HTTP server.
                                                type: string
                                              port:
                                                anyOf:
                                                - type: integer
                                                - type: string
                                                description: Name or number of the
                                                  port to access on the container.
                                                  Number must be in the range 1 to
                                                  65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                              scheme:
                                                description: Scheme to use for connecting
                                                  to the host. Defaults to HTTP.
                                                type: string
                                            required:
                                            - port
                                            type: object
                                          initialDelaySeconds:
                                            description: 'Number of seconds after
                                              the container has started before liveness
                                              probes are initiated. More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                            format: int32
                                            type: integer
                                          periodSeconds:
                                            description: How often (in seconds) to
                                              perform the probe. Default to 10 seconds.
                                              Minimum value is 1.
                                            format: int32
                                            type: integer
                                          successThreshold:
                                            description: Minimum consecutive successes
                                              for the probe to be considered successful
                                              after having failed. Defaults to 1.
                                              Must be 1 for liveness and startup.
                                              Minimum value is 1.
                                            format: int32
                                            type: integer
                                          tcpSocket:
                                            description: 'TCPSocket specifies an action
                                              involving a TCP port. TCP hooks not
                                              yet supported TODO: implement a realistic
                                              TCP lifecycle hook'
                                            properties:
                                              host:
                                                description: 'Optional: Host name
                                                  to connect to, defaults to the pod
                                                  IP.'
                                                type: string
                                              port:
                                                anyOf:
                                                - type: integer
                                                - type: string
                                                description: Number or name of the
                                                  port to access on the container.
                                                  Number must be in the range 1 to
                                                  65535. Name must be an IANA_SVC_NAME.
                                                x-kubernetes-int-or-string: true
                                            required:
                                            - port
                                            type: object
                                          timeoutSeconds:
                                            description: 'Number of seconds after
                                              which the probe times out. Defaults
                                              to 1 second. Minimum value is 1. More
                                              info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes'
                                            format: int32
                                            type: integer
                                        type: object
                                      stdin:
                                        description: Whether this container should
                                          allocate a buffer for stdin in the container
                                          runtime. If this is not set, reads from
                                          stdin in the container will always result
                                          in EOF. Default is false.
                                        type: boolean
                                      stdinOnce:
                                        description: Whether the container runtime
                                          should close the stdin channel after it
                                          has been opened by a single attach. When
                                          stdin is true the stdin stream will remain
                                          open across multiple attach sessions. If
                                          stdinOnce is set to true, stdin is opened
                                          on container start, is empty until the first
                                          client attaches to stdin, and then remains
                                          open and accepts data until the client disconnects,
                                          at which time stdin is closed and remains
                                          closed until the container is restarted.
                                          If this flag is false, a container processes
                                          that reads from stdin will never receive
                                          an EOF. Default is false
                                        type: boolean
                                      terminationMessagePath:
                                        description: 'Optional: Path at which the
                                          file to which the container''s termination
                                          message will be written is mounted into
                                          the container''s filesystem. Message written
                                          is intended to be brief final status, such
                                          as an assertion failure message. Will be
                                          truncated by the node if greater than 4096
                                          bytes. The total message length across all
                                          containers will be limited to 12kb. Defaults
                                          to /dev/termination-log. Cannot be updated.'
                                        type: string
                                      terminationMessagePolicy:
                                        description: Indicate how the termination
                                          message should be populated. File will use
                                          the contents of terminationMessagePath to
                                          populate the container status message on
                                          both success and failure. FallbackToLogsOnError
                                          will use the last chunk of container log
                                          output if the termination message file is
                                          empty and the container exited with an error.
                                          The log output is limited to 2048 bytes
                                          or 80 lines, whichever is smaller. Defaults
                                          to File. Cannot be updated.
                                        type: string
                                      tty:
                                        description: Whether this container should
                                          allocate a TTY for itself, also requires
                                          'stdin' to be true. Default is false.
                                        type: boolean
                                      volumeDevices:
                                        description: volumeDevices is the list of
                                          block devices to be used by the container.
                                        items:
                                          description: volumeDevice describes a mapping
                                            of a raw block device within a container.
                                          properties:
                                            devicePath:
                                              description: devicePath is the path
                                                inside of the container that the device
                                                will be mapped to.
                                              type: string
                                            name:
                                              description: name must match the name
                                                of a persistentVolumeClaim in the
                                                pod
                                              type: string
                                          required:
                                          - devicePath
                                          - name
                                          type: object
                                        type: array
                                      volumeMounts:
                                        description: Pod volumes to mount into the
                                          container's filesystem. Cannot be updated.
                                        items:
                                          description: VolumeMount describes a mounting
                                            of a Volume within a container.
                                          properties:
                                            mountPath:
                                              description: Path within the container
                                                at which the volume should be mounted.  Must
                                                not contain ':'.
                                              type: string
                                            mountPropagation:
                                              description: mountPropagation determines
                                                how mounts are propagated from the
                                                host to container and the other way
                                                around. When not set, MountPropagationNone
                                                is used. This field is beta in 1.10.
                                              type: string
                                            name:
                                              description: This must match the Name
                                                of a Volume.
                                              type: string
                                            readOnly:
                                              description: Mounted read-only if true,
                                                read-write otherwise (false or unspecified).
                                                Defaults to false.
                                              type: boolean
                                            subPath:
                                              description: Path within the volume
                                                from which the container's volume
                                                should be mounted. Defaults to ""
                                                (volume's root).
                                              type: string
                                            subPathExpr:
                                              description: Expanded path within the
                                                volume from which the container's
                                                volume should be mounted. Behaves
                                                similarly to SubPath but environment
                                                variable references $(VAR_NAME) are
                                                expanded using the container's environment.
                                                Defaults to "" (volume's root). SubPathExpr
                                                and SubPath are mutually exclusive.
                                              type: string
                                          required:
                                          - mountPath
                                          - name
                                          type: object
                                        type: array
                                      workingDir:
                                        description: Container's working directory.
                                          If not specified, the container runtime's
                                          default will be used, which might be configured
                                          in the container image. Cannot be updated.
                                        type: string
                                    required:
                                    - name
                                    type: object
                                  type: array
                                timeout:
                                  description: Timeout defines the maximum amount
                                    of time Velero should wait for the initContainers
                                    to complete.
                                  type: string
                              type: object
                          type: object
                        type: array
                    required:
                    - name
                    type: object
                  type: array
              type: object
            includeClusterResources:
              description: IncludeClusterResources specifies whether cluster-scoped
                resources should be included for consideration in the restore. If
                null, defaults to true.
              nullable: true
              type: boolean
            includedNamespaces:
              description: IncludedNamespaces is a slice of namespace names to include
                objects from. If empty, all namespaces are included.
              items:
                type: string
              nullable: true
              type: array
            includedResources:
              description: IncludedResources is a slice of resource names to include
                in the restore. If empty, all resources in the backup are included.
              items:
                type: string
              nullable: true
              type: array
            labelSelector:
              description: LabelSelector is a metav1.LabelSelector to filter with
                when restoring individual objects from the backup. If empty or nil,
                all objects are included. Optional.
              nullable: true
              properties:
                matchExpressions:
                  description: matchExpressions is a list of label selector requirements.
                    The requirements are ANDed.
                  items:
                    description: A label selector requirement is a selector that contains
                      values, a key, and an operator that relates the key and values.
                    properties:
                      key:
                        description: key is the label key that the selector applies
                          to.
                        type: string
                      operator:
                        description: operator represents a key's relationship to a
                          set of values. Valid operators are In, NotIn, Exists and
                          DoesNotExist.
                        type: string
                      values:
                        description: values is an array of string values. If the operator
                          is In or NotIn, the values array must be non-empty. If the
                          operator is Exists or DoesNotExist, the values array must
                          be empty. This array is replaced during a strategic merge
                          patch.
                        items:
                          type: string
                        type: array
                    required:
                    - key
                    - operator
                    type: object
                  type: array
                matchLabels:
                  additionalProperties:
                    type: string
                  description: matchLabels is a map of {key,value} pairs. A single
                    {key,value} in the matchLabels map is equivalent to an element
                    of matchExpressions, whose key field is "key", the operator is
                    "In", and the values array contains only "value". The requirements
                    are ANDed.
                  type: object
              type: object
            namespaceMapping:
              additionalProperties:
                type: string
              description: NamespaceMapping is a map of source namespace names to
                target namespace names to restore into. Any source namespaces not
                included in the map will be restored into namespaces of the same name.
              type: object
            preserveNodePorts:
              description: PreserveNodePorts specifies whether to restore old nodePorts
                from backup.
              nullable: true
              type: boolean
            restorePVs:
              description: RestorePVs specifies whether to restore all included PVs
                from snapshot (via the cloudprovider).
              nullable: true
              type: boolean
            scheduleName:
              description: ScheduleName is the unique name of the Velero schedule
                to restore from. If specified, and BackupName is empty, Velero will
                restore from the most recent successful backup created from this schedule.
              type: string
          required:
          - backupName
          type: object
        status:
          description: RestoreStatus captures the current status of a Velero restore
          properties:
            completionTimestamp:
              description: CompletionTimestamp records the time the restore operation
                was completed. Completion time is recorded even on failed restore.
                The server's time is used for StartTimestamps
              format: date-time
              nullable: true
              type: string
            errors:
              description: Errors is a count of all error messages that were generated
                during execution of the restore. The actual errors are stored in object
                storage.
              type: integer
            failureReason:
              description: FailureReason is an error that caused the entire restore
                to fail.
              type: string
            phase:
              description: Phase is the current state of the Restore
              enum:
              - New
              - FailedValidation
              - InProgress
              - Completed
              - PartiallyFailed
              - Failed
              type: string
            progress:
              description: Progress contains information about the restore's execution
                progress. Note that this information is best-effort only -- if Velero
                fails to update it during a restore for any reason, it may be inaccurate/stale.
              nullable: true
              properties:
                itemsRestored:
                  description: ItemsRestored is the number of items that have actually
                    been restored so far
                  type: integer
                totalItems:
                  description: TotalItems is the total number of items to be restored.
                    This number may change throughout the execution of the restore
                    due to plugins that return additional related items to restore
                  type: integer
              type: object
            startTimestamp:
              description: StartTimestamp records the time the restore operation was
                started. The server's time is used for StartTimestamps
              format: date-time
              nullable: true
              type: string
            validationErrors:
              description: ValidationErrors is a slice of all validation errors (if
                applicable)
              items:
                type: string
              nullable: true
              type: array
            warnings:
              description: Warnings is a count of all warning messages that were generated
                during execution of the restore. The actual warnings are stored in
                object storage.
              type: integer
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
	serverstatusrequests = `
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    app.kubernetes.io/name: velero
    component: velero
  annotations:
    controller-gen.kubebuilder.io/version: v0.3.0
  creationTimestamp: null
  name: serverstatusrequests.velero.io
spec:
  group: velero.io
  names:
    kind: ServerStatusRequest
    listKind: ServerStatusRequestList
    plural: serverstatusrequests
    shortNames:
    - ssr
    singular: serverstatusrequest
  preserveUnknownFields: false
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: ServerStatusRequest is a request to access current status information
        about the Velero server.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: ServerStatusRequestSpec is the specification for a ServerStatusRequest.
          type: object
        status:
          description: ServerStatusRequestStatus is the current status of a ServerStatusRequest.
          properties:
            phase:
              description: Phase is the current lifecycle phase of the ServerStatusRequest.
              enum:
              - New
              - Processed
              type: string
            plugins:
              description: Plugins list information about the plugins running on the
                Velero server
              items:
                description: PluginInfo contains attributes of a Velero plugin
                properties:
                  kind:
                    type: string
                  name:
                    type: string
                required:
                - kind
                - name
                type: object
              nullable: true
              type: array
            processedTimestamp:
              description: ProcessedTimestamp is when the ServerStatusRequest was
                processed by the ServerStatusRequestController.
              format: date-time
              nullable: true
              type: string
            serverVersion:
              description: ServerVersion is the Velero server version.
              type: string
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`
)

// Templates
const (
	backupstoragelocation = `
{{- if .Values.backupsEnabled }}
apiVersion: velero.io/v1
kind: BackupStorageLocation
metadata:
  name: {{ include "velero.backupStorageLocation.name" . }}
  annotations:
    "helm.sh/hook": post-install,post-upgrade,post-rollback
    "helm.sh/hook-delete-policy": before-hook-creation
  labels:
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
spec:
  provider: {{ include "velero.backupStorageLocation.provider" . }}
{{- with .Values.configuration.backupStorageLocation }}
  objectStorage:
    bucket: {{ .bucket  }}
    {{- with .prefix }}
    prefix: {{ . }}
    {{- end }}
    {{- with .caCert }}
    caCert: {{ . }}
    {{- end }}
{{- with .config }}
  config:
{{- range $key, $value := . }}
{{- $key | nindent 4 }}: {{ $value | quote }}
{{- end }}
{{- end }}
{{- end }}
{{- end }}
`
	configmaps = `
{{- range $configMapName, $configMap := .Values.configMaps }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "velero.fullname" $ }}-{{ $configMapName }}
  labels:
    app.kubernetes.io/name: {{ include "velero.name" $ }}
    app.kubernetes.io/instance: {{ $.Release.Name }}
    app.kubernetes.io/managed-by: {{ $.Release.Service }}
    helm.sh/chart: {{ include "velero.chart" $ }}
  {{- with $configMap.labels }}
    {{- toYaml . | nindent 4 }}
  {{- end }}
data:
  {{- toYaml $configMap.data | nindent 2 }}
---
{{- end }}
`
	role = `
{{- if .Values.rbac.create }}
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: {{ include "velero.fullname" . }}-server
  labels:
    app.kubernetes.io/component: server
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
rules:
- apiGroups:
    - "*"
  resources:
    - "*"
  verbs:
    - "*"

{{- end }}
`
	serviceaccountserver = `
{{- if .Values.serviceAccount.server.create }}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "velero.serverServiceAccount" . }}
{{- if .Values.serviceAccount.server.annotations }}
  annotations:
{{ toYaml .Values.serviceAccount.server.annotations | nindent 4 }}
{{- end }}
  labels:
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
{{- with .Values.serviceAccount.server.labels }}
  {{- toYaml . | nindent 4 }}
{{- end }}
{{- end }}
`
	upgradecrds = `
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ template "velero.fullname" . }}-upgrade-crds
  namespace: {{ .Release.Namespace }}
  annotations:
    "helm.sh/hook": post-install,post-upgrade,post-rollback
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
  labels:
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
spec:
  backoffLimit: 3
  template:
    metadata:
      name: velero-upgrade-crds
    spec:
      serviceAccountName: {{ include "velero.serverServiceAccount" . }}
      initContainers:
        - name: velero
      {{- if .Values.image.digest }}
          image: "{{ .Values.image.repository }}@{{ .Values.image.digest }}"
      {{- else }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
      {{- end }}
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          command:
            - /bin/sh
            - -c
            - /velero install --crds-only --dry-run -o yaml > /tmp/crds.yaml
          volumeMounts:
            - mountPath: /tmp
              name: crds
      containers:
        - name: kubectl
          # image: docker.io/bitnami/kubectl:1.14.3
          image: caas4/kubectl:1.14.3
          imagePullPolicy: IfNotPresent
          command:
            - /bin/sh
            - -c
            - kubectl apply -f /tmp/crds.yaml
          volumeMounts:
            - mountPath: /tmp
              name: crds
      volumes:
        - name: crds
          emptyDir: {}
      restartPolicy: OnFailure
`
    cleanupcrds = `
{{- if .Values.cleanUpCRDs  }}
# This job is meant primarily for cleaning up on CI systems.
# Using this on production systems, especially those that have multiple releases of Velero, will be destructive.
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ template "velero.fullname" . }}-cleanup-crds
  namespace: {{ .Release.Namespace }}
  annotations:
    "helm.sh/hook": pre-delete
    "helm.sh/hook-delete-policy": hook-succeeded
  labels:
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
spec:
  backoffLimit: 3
  template:
    metadata:
      name: velero-cleanup-crds
    spec:
      serviceAccountName: {{ include "velero.serverServiceAccount" . }}
      containers:
        - name: kubectl
          # image: docker.io/bitnami/kubectl:1.14.3
          image: caas4/kubectl:1.14.3
          imagePullPolicy: IfNotPresent
          command:
            - /bin/sh
            - -c
            - >
              kubectl delete restore --all;
              kubectl delete backup --all;
              kubectl delete backupstoragelocation --all;
              kubectl delete volumesnapshotlocation --all;
              kubectl delete podvolumerestore --all;
              kubectl delete crd -l app.kubernetes.io/name=velero;
      restartPolicy: OnFailure
{{- end }}
`
	deployment = `
{{- if .Values.configuration.provider -}}
{{- $provider := .Values.configuration.provider -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "velero.fullname" . }}
  {{- with .Values.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  labels:
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
    component: velero
    {{- with .Values.labels }}
      {{- toYaml . | nindent 4 }}
    {{- end }}
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/instance: {{ .Release.Name }}
      app.kubernetes.io/name: {{ include "velero.name" . }}
  template:
    metadata:
      labels:
        name: velero
        app.kubernetes.io/name: {{ include "velero.name" . }}
        app.kubernetes.io/instance: {{ .Release.Name }}
        app.kubernetes.io/managed-by: {{ .Release.Service }}
        helm.sh/chart: {{ include "velero.chart" . }}
        {{- if .Values.podLabels }}
          {{- toYaml .Values.podLabels | nindent 8 }}
        {{- end }}
    {{- if or .Values.podAnnotations .Values.metrics.enabled }}
      annotations:
      {{- with .Values.podAnnotations }}
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.metrics.podAnnotations }}
        {{- toYaml . | nindent 8 }}
      {{- end }}
    {{- end }}
    spec:
    {{- if .Values.image.imagePullSecrets }}
      imagePullSecrets:
      {{- range .Values.image.imagePullSecrets }}
        - name: {{ . }}
      {{- end }}
    {{- end }}
      restartPolicy: Always
      serviceAccountName: {{ include "velero.serverServiceAccount" . }}
      {{- if .Values.priorityClassName }}
      priorityClassName: {{ include "velero.priorityClassName" . }}
      {{- end }}
      containers:
        - name: velero
      {{- if .Values.image.digest }}
          image: "{{ .Values.image.repository }}@{{ .Values.image.digest }}"
      {{- else }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
      {{- end }}
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          {{- if .Values.metrics.enabled }}
          ports:
            - name: monitoring
              containerPort: 8085
          {{- end }}
          command:
            - /velero
          args:
            - server
          {{- with .Values.configuration }}
            {{- with .backupSyncPeriod }}
            - --backup-sync-period={{ . }}
            {{- end }}
            {{- with .resticTimeout }}
            - --restic-timeout={{ . }}
            {{- end }}
            {{- if .restoreOnlyMode }}
            - --restore-only
            {{- end }}
            {{- with .restoreResourcePriorities }}
            - --restore-resource-priorities={{ . }}
            {{- end }}
            {{- with .features }}
            - --features={{ . }}
            {{- end }}
            {{- with .logLevel }}
            - --log-level={{ . }}
            {{- end }}
            {{- with .logFormat }}
            - --log-format={{ . }}
            {{- end }}
            {{- if .defaultVolumesToRestic }}
            - --default-volumes-to-restic
            {{- end }}
            {{- with .clientQPS }}
            - --client-qps={{ . }}
            {{- end }}
            {{- with .clientBurst }}
            - --client-burst={{ . }}
            {{- end }}
            {{- with .disableControllers }}
            - --disable-controllers={{ . }}
            {{- end }}
          {{- end }}
          {{- with .Values.resources }}
          resources:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          volumeMounts:
            - name: plugins
              mountPath: /plugins
        {{- if .Values.credentials.useSecret }}
            - name: cloud-credentials
              mountPath: /credentials
            - name: scratch
              mountPath: /scratch
        {{- end }}
            {{- if .Values.extraVolumeMounts }}
            {{- toYaml .Values.extraVolumeMounts | nindent 12 }}
            {{- end }}
        {{- if .Values.credentials.extraSecretRef }}
          envFrom:
          - secretRef:
              name: {{ .Values.credentials.extraSecretRef }}
        {{- end }}
          env:
            - name: VELERO_SCRATCH_DIR
              value: /scratch
            - name: VELERO_NAMESPACE
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: metadata.namespace
            - name: LD_LIBRARY_PATH
              value: /plugins
          {{- if .Values.credentials.useSecret }}
            {{- if eq $provider "aws" }}
            - name: AWS_SHARED_CREDENTIALS_FILE
              value: /credentials/cloud
            {{- else if eq $provider "gcp"}}
            - name: GOOGLE_APPLICATION_CREDENTIALS
              value: /credentials/cloud
            {{- else if eq $provider "azure" }}
            - name: AZURE_CREDENTIALS_FILE
              value: /credentials/cloud
            {{- else if eq $provider "alibabacloud" }}
            - name: ALIBABA_CLOUD_CREDENTIALS_FILE
              value: /credentials/cloud
            {{- end }}
          {{- end }}
          {{- with .Values.configuration.extraEnvVars }}
          {{- range $key, $value := . }}
            - name: {{ default "none" $key }}
              value: {{ default "none" $value }}
          {{- end }}
          {{- end }}
          {{- with .Values.credentials.extraEnvVars }}
          {{- range $key, $value := . }}
            - name: {{ default "none" $key }}
              valueFrom:
                secretKeyRef:
                  name: {{ include "velero.fullname" $ }}
                  key: {{ default "none" $key }}
          {{- end }}
          {{- end }}
      dnsPolicy: {{ .Values.dnsPolicy }}
{{- if .Values.initContainers }}
      initContainers:
        {{- toYaml .Values.initContainers | nindent 8 }}
{{- end }}
      volumes:
        {{- if .Values.credentials.useSecret }}
        - name: cloud-credentials
          secret:
            secretName: {{ include "velero.secretName" . }}
        {{- end }}
        - name: plugins
          emptyDir: {}
        - name: scratch
          emptyDir: {}
        {{- if .Values.extraVolumes }}
        {{- toYaml .Values.extraVolumes | nindent 8 }}
        {{- end }}
    {{- with .Values.securityContext }}
      securityContext:
        {{- toYaml . | nindent 8 }}
    {{- end }}
    {{- with .Values.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
    {{- end }}
    {{- with .Values.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
    {{- end }}
    {{- with .Values.affinity }}
      affinity:
        {{- toYaml . | nindent 8 }}
    {{- end }}
{{- end -}}
`
	resticdaemonset = `
{{- if .Values.deployRestic }}
{{- $provider := .Values.configuration.provider -}}
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: restic
  {{- with .Values.restic.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  labels:
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
    {{- with .Values.restic.labels }}
      {{- toYaml . | nindent 4 }}
    {{- end }}
spec:
  selector:
    matchLabels:
      name: restic
  template:
    metadata:
      labels:
        name: restic
        app.kubernetes.io/name: {{ include "velero.name" . }}
        app.kubernetes.io/instance: {{ .Release.Name }}
        app.kubernetes.io/managed-by: {{ .Release.Service }}
        helm.sh/chart: {{ include "velero.chart" . }}
      {{- if .Values.podLabels }}
        {{ toYaml .Values.podLabels | nindent 8 }}
      {{- end }}
      {{- with .Values.podAnnotations }}
      annotations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
    spec:
    {{- if .Values.image.imagePullSecrets }}
      imagePullSecrets:
      {{- range .Values.image.imagePullSecrets }}
        - name: {{ . }}
      {{- end }}
    {{- end }}
      serviceAccountName: {{ include "velero.serverServiceAccount" . }}
      securityContext:
        runAsUser: 0
      {{- if .Values.restic.priorityClassName }}
      priorityClassName: {{ include "velero.restic.priorityClassName" . }}
      {{- end }}
      volumes:
        {{- if .Values.credentials.useSecret }}
        - name: cloud-credentials
          secret:
            secretName: {{ include "velero.secretName" . }}
        {{- end }}
        - name: host-pods
          hostPath:
            path: {{ .Values.restic.podVolumePath }}
        - name: scratch
          emptyDir: {}
        {{- if .Values.restic.extraVolumes }}
        {{- toYaml .Values.restic.extraVolumes | nindent 8 }}
        {{- end }}
      dnsPolicy: {{ .Values.restic.dnsPolicy }}
      containers:
        - name: restic
        {{- if .Values.image.digest }}
          image: "{{ .Values.image.repository }}@{{ .Values.image.digest }}"
        {{- else }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        {{- end }}
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          command:
            - /velero
          args:
            - restic
            - server
          {{- with .Values.configuration }}
            {{- with .features }}
            - --features={{ . }}
            {{- end }}
            {{- with .logLevel }}
            - --log-level={{ . }}
            {{- end }}
            {{- with .logFormat }}
            - --log-format={{ . }}
            {{- end }}
          {{- end }}
          volumeMounts:
            {{- if .Values.credentials.useSecret }}
            - name: cloud-credentials
              mountPath: /credentials
            {{- end }}
            - name: host-pods
              mountPath: /host_pods
              mountPropagation: HostToContainer
            - name: scratch
              mountPath: /scratch
            {{- if .Values.restic.extraVolumeMounts }}
            {{- toYaml .Values.restic.extraVolumeMounts | nindent 12 }}
            {{- end }}
        {{- if .Values.credentials.extraSecretRef }}
          envFrom:
          - secretRef:
              name: {{ .Values.credentials.extraSecretRef }}
        {{- end }}
          env:
            - name: VELERO_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            - name: VELERO_SCRATCH_DIR
              value: /scratch
          {{- if .Values.credentials.useSecret }}
            {{- if eq $provider "aws" }}
            - name: AWS_SHARED_CREDENTIALS_FILE
              value: /credentials/cloud
            {{- else if eq $provider "gcp" }}
            - name: GOOGLE_APPLICATION_CREDENTIALS
              value: /credentials/cloud
            {{- else if eq $provider "azure" }}
            - name: AZURE_CREDENTIALS_FILE
              value: /credentials/cloud
            {{- else if eq $provider "alibabacloud" }}
            - name: ALIBABA_CLOUD_CREDENTIALS_FILE
              value: /credentials/cloud
            {{- end }}
          {{- end }}
          {{- with .Values.configuration.extraEnvVars }}
          {{- range $key, $value := . }}
            - name: {{ default "none" $key }}
              value: {{ default "none" $value }}
          {{- end }}
          {{- end }}
          {{- with .Values.credentials.extraEnvVars }}
          {{- range $key, $value := . }}
            - name: {{ default "none" $key }}
              valueFrom:
                secretKeyRef:
                  name: {{ include "velero.fullname" $ }}
                  key: {{ default "none" $key }}
          {{- end }}
          {{- end }}
          securityContext:
            privileged: {{ .Values.restic.privileged }}
            {{- with .Values.restic.securityContext }}
              {{- toYaml . | nindent 12 }}
            {{- end }}
          {{- with .Values.restic.resources }}
          resources:
            {{- toYaml . | nindent 12 }}
          {{- end }}
      {{- with .Values.restic.tolerations }}
      tolerations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.restic.nodeSelector }}
      nodeSelector:
        {{- toYaml . | nindent 8 }}
      {{- end }}
{{- end }}
`
	schedule = `
{{- range $scheduleName, $schedule := .Values.schedules }}
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: {{ include "velero.fullname" $ }}-{{ $scheduleName }}
  annotations:
  {{- if $schedule.annotations }}
    {{- toYaml $schedule.annotations | nindent 4 }}
  {{- end }}
    "helm.sh/hook": post-install,post-upgrade,post-rollback
    "helm.sh/hook-delete-policy": before-hook-creation
  labels:
    app.kubernetes.io/name: {{ include "velero.name" $ }}
    app.kubernetes.io/instance: {{ $.Release.Name }}
    app.kubernetes.io/managed-by: {{ $.Release.Service }}
    helm.sh/chart: {{ include "velero.chart" $ }}
    {{- if $schedule.labels }}
      {{- toYaml $schedule.labels | nindent 4 }}
    {{- end }}
spec:
  schedule: {{ $schedule.schedule | quote }}
{{- with $schedule.template }}
  template:
    {{- toYaml . | nindent 4 }}
{{- end }}
---
{{- end }}
`
	servicemonitor = `
{{- if and .Values.metrics.enabled .Values.metrics.serviceMonitor.enabled }}
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: {{ include "velero.fullname" . }}
  {{- if .Values.metrics.serviceMonitor.namespace }}
  namespace: {{ .Values.metrics.serviceMonitor.namespace }}
  {{- end }}
  labels:
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
  {{- with .Values.metrics.serviceMonitor.additionalLabels }}
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  namespaceSelector:
    matchNames:
      - {{ .Release.Namespace }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "velero.name" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
  endpoints:
  - port: monitoring
    interval: {{ .Values.metrics.scrapeInterval }}
    scrapeTimeout: {{ .Values.metrics.scrapeTimeout }}
{{- end }}
`
	volumesnapshotlocation = `
{{- if .Values.snapshotsEnabled }}
apiVersion: velero.io/v1
kind: VolumeSnapshotLocation
metadata:
  name: {{ include "velero.volumeSnapshotLocation.name" . }}
  annotations:
    "helm.sh/hook": post-install,post-upgrade,post-rollback
    "helm.sh/hook-delete-policy": before-hook-creation
  labels:
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
spec:
  provider: {{ include "velero.volumeSnapshotLocation.provider" . }}
{{- with .Values.configuration.volumeSnapshotLocation.config }}
  config:
{{- range $key, $value := . }}
{{- $key | nindent 4 }}: {{ $value | quote }}
{{- end }}
{{- end -}}
{{- end }}
`
	clusterrolebinding = `
{{- if and .Values.rbac.create .Values.rbac.clusterAdministrator }}
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{ include "velero.fullname" . }}-server
  labels:
    app.kubernetes.io/component: server
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
subjects:
  - kind: ServiceAccount
    namespace: {{ .Release.Namespace }}
    name: {{ include "velero.serverServiceAccount" . }}
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
{{- end }}
`
	helpers = `
{{- if and .Values.rbac.create .Values.rbac.clusterAdministrator }}
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{ include "velero.fullname" . }}-server
  labels:
    app.kubernetes.io/component: server
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
subjects:
  - kind: ServiceAccount
    namespace: {{ .Release.Namespace }}
    name: {{ include "velero.serverServiceAccount" . }}
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
{{- end }}
[root@master templates]# cat _helpers
cat: _helpers: No such file or directory
[root@master templates]# cat _helpers.tpl 
{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "velero.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "velero.fullname" -}}
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
{{- define "velero.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the service account to use for creating or deleting the velero server
*/}}
{{- define "velero.serverServiceAccount" -}}
{{- if .Values.serviceAccount.server.create -}}
    {{ default (printf "%s-%s" (include "velero.fullname" .) "server") .Values.serviceAccount.server.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.server.name }}
{{- end -}}
{{- end -}}

{{/*
Create the name for the credentials secret.
*/}}
{{- define "velero.secretName" -}}
{{- if .Values.credentials.existingSecret -}}
  {{- .Values.credentials.existingSecret -}}
{{- else -}}
  {{ default (include "velero.fullname" .) .Values.credentials.name }}
{{- end -}}
{{- end -}}

{{/*
Create the Velero priority class name.
*/}}
{{- define "velero.priorityClassName" -}}
{{- if .Values.priorityClassName -}}
  {{- .Values.priorityClassName -}}
{{- else -}}
  {{- include "velero.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
Create the Restic priority class name.
*/}}
{{- define "velero.restic.priorityClassName" -}}
{{- if .Values.restic.priorityClassName -}}
  {{- .Values.restic.priorityClassName -}}
{{- else -}}
  {{- include "velero.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
Create the backup storage location name
*/}}
{{- define "velero.backupStorageLocation.name" -}}
{{- with .Values.configuration.backupStorageLocation -}}
{{ default "default" .name }}
{{- end -}}
{{- end -}}

{{/*
Create the backup storage location provider
*/}}
{{- define "velero.backupStorageLocation.provider" -}}
{{- with .Values.configuration -}}
{{ default .provider .backupStorageLocation.provider }}
{{- end -}}
{{- end -}}

{{/*
Create the volume snapshot location name
*/}}
{{- define "velero.volumeSnapshotLocation.name" -}}
{{- with .Values.configuration.volumeSnapshotLocation -}}
{{ default "default" .name }}
{{- end -}}
{{- end -}}

{{/*
Create the volume snapshot location provider
*/}}
{{- define "velero.volumeSnapshotLocation.provider" -}}
{{- with .Values.configuration -}}
{{ default .provider .volumeSnapshotLocation.provider }}
{{- end -}}
{{- end -}}
`
	rolebinding = `
{{- if .Values.rbac.create }}
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ include "velero.fullname" . }}-server
  labels:
    app.kubernetes.io/component: server
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
subjects:
  - kind: ServiceAccount
    namespace: {{ .Release.Namespace }}
    name: {{ include "velero.serverServiceAccount" . }}
roleRef:
  kind: Role
  name: {{ include "velero.fullname" . }}-server
  apiGroup: rbac.authorization.k8s.io
{{- end }}
`
	secret = `
{{- if and .Values.credentials.useSecret (not .Values.credentials.existingSecret) -}}
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "velero.secretName" . }}
  labels:
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
type: Opaque
data:
{{- range $key, $value := .Values.credentials.secretContents }}
  {{ $key }}: {{ $value | b64enc | quote }}
{{- end }}
{{- range $key, $value := .Values.credentials.extraEnvVars }}
  {{ $key }}: {{ $value | b64enc | quote }}
{{- end }}
{{- end -}}
`
	service = `
{{- if .Values.metrics.enabled }}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "velero.fullname" . }}
  {{- with .Values.metrics.service.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  labels:
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
    app.kubernetes.io/managed-by: {{ .Release.Service }}
    helm.sh/chart: {{ include "velero.chart" . }}
    {{- with .Values.metrics.service.labels }}
      {{- toYaml . | nindent 4 }}
    {{- end }}
spec:
  type: ClusterIP
  ports:
    - name: monitoring
      port: 8085
      targetPort: monitoring
  selector:
    name: velero
    app.kubernetes.io/name: {{ include "velero.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
`
)