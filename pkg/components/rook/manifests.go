package rook

// These manifests were taken from https://github.com/rook/rook/tree/v1.2.1/cluster/examples/kubernetes/ceph with
// some modifications (explained below on a case to case basis).

// Namespace was split from https://github.com/rook/rook/blob/v1.2.1/cluster/examples/kubernetes/ceph/common.yaml
// for easier management in code.
const namespace = `
# Namespace where the operator and other rook resources are created
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Namespace }}
`

// CRD resource definitions was split from https://github.com/rook/rook/blob/v1.2.1/cluster/examples/kubernetes/ceph/common.yaml
// for easier management in code.
const crds = `
# The CRD declarations
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephclusters.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephCluster
    listKind: CephClusterList
    plural: cephclusters
    singular: cephcluster
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            annotations: {}
            cephVersion:
              properties:
                allowUnsupported:
                  type: boolean
                image:
                  type: string
            dashboard:
              properties:
                enabled:
                  type: boolean
                urlPrefix:
                  type: string
                port:
                  type: integer
                  minimum: 0
                  maximum: 65535
                ssl:
                  type: boolean
            dataDirHostPath:
              pattern: ^/(\S+)
              type: string
            disruptionManagement:
              properties:
                machineDisruptionBudgetNamespace:
                  type: string
                managePodBudgets:
                  type: boolean
                osdMaintenanceTimeout:
                  type: integer
                manageMachineDisruptionBudgets:
                  type: boolean
            skipUpgradeChecks:
              type: boolean
            mon:
              properties:
                allowMultiplePerNode:
                  type: boolean
                count:
                  maximum: 9
                  minimum: 0
                  type: integer
                volumeClaimTemplate: {}
            mgr:
              properties:
                modules:
                  items:
                    properties:
                      name:
                        type: string
                      enabled:
                        type: boolean
            network:
              properties:
                hostNetwork:
                  type: boolean
                provider:
                  type: string
                selectors: {}
            storage:
              properties:
                disruptionManagement:
                  properties:
                    machineDisruptionBudgetNamespace:
                      type: string
                    managePodBudgets:
                      type: boolean
                    osdMaintenanceTimeout:
                      type: integer
                    manageMachineDisruptionBudgets:
                      type: boolean
                useAllNodes:
                  type: boolean
                nodes:
                  items:
                    properties:
                      name:
                        type: string
                      config:
                        properties:
                          metadataDevice:
                            type: string
                          storeType:
                            type: string
                            pattern: ^(filestore|bluestore)$
                          databaseSizeMB:
                            type: string
                          walSizeMB:
                            type: string
                          journalSizeMB:
                            type: string
                          osdsPerDevice:
                            type: string
                          encryptedDevice:
                            type: string
                            pattern: ^(true|false)$
                      useAllDevices:
                        type: boolean
                      deviceFilter:
                        type: string
                      devicePathFilter:
                        type: string
                      directories:
                        type: array
                        items:
                          properties:
                            path:
                              type: string
                      devices:
                        type: array
                        items:
                          properties:
                            name:
                              type: string
                            config: {}
                      resources: {}
                  type: array
                useAllDevices:
                  type: boolean
                deviceFilter:
                  type: string
                devicePathFilter:
                  type: string
                directories:
                  type: array
                  items:
                    properties:
                      path:
                        type: string
                config: {}
                storageClassDeviceSets: {}
            monitoring:
              properties:
                enabled:
                  type: boolean
                rulesNamespace:
                  type: string
            rbdMirroring:
              properties:
                workers:
                  type: integer
            removeOSDsIfOutAndSafeToRemove:
              type: boolean
            external:
              properties:
                enable:
                  type: boolean
            placement: {}
            resources: {}
  additionalPrinterColumns:
    - name: DataDirHostPath
      type: string
      description: Directory used on the K8s nodes
      JSONPath: .spec.dataDirHostPath
    - name: MonCount
      type: string
      description: Number of MONs
      JSONPath: .spec.mon.count
    - name: Age
      type: date
      JSONPath: .metadata.creationTimestamp
    - name: State
      type: string
      description: Current State
      JSONPath: .status.state
    - name: Health
      type: string
      description: Ceph Health
      JSONPath: .status.ceph.health

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephclients.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephClient
    listKind: CephClientList
    plural: cephclients
    singular: cephclient
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            caps:
              type: object

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephfilesystems.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephFilesystem
    listKind: CephFilesystemList
    plural: cephfilesystems
    singular: cephfilesystem
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            metadataServer:
              properties:
                activeCount:
                  minimum: 1
                  maximum: 10
                  type: integer
                activeStandby:
                  type: boolean
                annotations: {}
                placement: {}
                resources: {}
            metadataPool:
              properties:
                failureDomain:
                  type: string
                replicated:
                  properties:
                    size:
                      minimum: 1
                      maximum: 10
                      type: integer
                erasureCoded:
                  properties:
                    dataChunks:
                      type: integer
                    codingChunks:
                      type: integer
            dataPools:
              type: array
              items:
                properties:
                  failureDomain:
                    type: string
                  replicated:
                    properties:
                      size:
                        minimum: 1
                        maximum: 10
                        type: integer
                  erasureCoded:
                    properties:
                      dataChunks:
                        type: integer
                      codingChunks:
                        type: integer
            preservePoolsOnDelete:
              type: boolean
  additionalPrinterColumns:
    - name: ActiveMDS
      type: string
      description: Number of desired active MDS daemons
      JSONPath: .spec.metadataServer.activeCount
    - name: Age
      type: date
      JSONPath: .metadata.creationTimestamp

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephnfses.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephNFS
    listKind: CephNFSList
    plural: cephnfses
    singular: cephnfs
    shortNames:
    - nfs
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            rados:
              properties:
                pool:
                  type: string
                namespace:
                  type: string
            server:
              properties:
                active:
                  type: integer
                annotations: {}
                placement: {}
                resources: {}

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephobjectstores.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephObjectStore
    listKind: CephObjectStoreList
    plural: cephobjectstores
    singular: cephobjectstore
  scope: Namespaced
  version: v1
  validation:
    openAPIV3Schema:
      properties:
        spec:
          properties:
            gateway:
              properties:
                type:
                  type: string
                sslCertificateRef: {}
                port:
                  type: integer
                securePort: {}
                instances:
                  type: integer
                annotations: {}
                placement: {}
                resources: {}
            metadataPool:
              properties:
                failureDomain:
                  type: string
                replicated:
                  properties:
                    size:
                      type: integer
                erasureCoded:
                  properties:
                    dataChunks:
                      type: integer
                    codingChunks:
                      type: integer
            dataPool:
              properties:
                failureDomain:
                  type: string
                replicated:
                  properties:
                    size:
                      type: integer
                erasureCoded:
                  properties:
                    dataChunks:
                      type: integer
                    codingChunks:
                      type: integer
            preservePoolsOnDelete:
              type: boolean

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephobjectstoreusers.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephObjectStoreUser
    listKind: CephObjectStoreUserList
    plural: cephobjectstoreusers
    singular: cephobjectstoreuser
    shortNames:
    - rcou
    - objectuser
  scope: Namespaced
  version: v1

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: cephblockpools.ceph.rook.io
spec:
  group: ceph.rook.io
  names:
    kind: CephBlockPool
    listKind: CephBlockPoolList
    plural: cephblockpools
    singular: cephblockpool
  scope: Namespaced
  version: v1

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: volumes.rook.io
spec:
  group: rook.io
  names:
    kind: Volume
    listKind: VolumeList
    plural: volumes
    singular: volume
    shortNames:
    - rv
  scope: Namespaced
  version: v1alpha2

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: objectbuckets.objectbucket.io
spec:
  group: objectbucket.io
  versions:
    - name: v1alpha1
      served: true
      storage: true
  names:
    kind: ObjectBucket
    listKind: ObjectBucketList
    plural: objectbuckets
    singular: objectbucket
    shortNames:
      - ob
      - obs
  scope: Cluster
  subresources:
    status: {}

---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: objectbucketclaims.objectbucket.io
spec:
  versions:
    - name: v1alpha1
      served: true
      storage: true
  group: objectbucket.io
  names:
    kind: ObjectBucketClaim
    listKind: ObjectBucketClaimList
    plural: objectbucketclaims
    singular: objectbucketclaim
    shortNames:
      - obc
      - obcs
  scope: Namespaced
  subresources:
    status: {}
`

// RBAC resource definitions was split from
// https://github.com/rook/rook/blob/v1.2.1/cluster/examples/kubernetes/ceph/common.yaml for easier management in
// code.  This section contains additional Role and Rolebinding named "rook-psp-privileged" that allows the service
// accounts in the {{ .Namespace }} namespace to use the privileged PSP. Without this, Rook will not work on a cluster
// with PSP enabled.
const rbac = `
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-object-bucket
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rook-ceph-object-bucket
subjects:
  - kind: ServiceAccount
    name: rook-ceph-system
    namespace: {{ .Namespace }}

---
# The cluster role for managing all the cluster-specific resources in a namespace
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: rook-ceph-cluster-mgmt
  labels:
    operator: rook
    storage-backend: ceph
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.ceph.rook.io/aggregate-to-rook-ceph-cluster-mgmt: "true"
rules: []
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: rook-ceph-cluster-mgmt-rules
  labels:
    operator: rook
    storage-backend: ceph
    rbac.ceph.rook.io/aggregate-to-rook-ceph-cluster-mgmt: "true"
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  - pods
  - pods/log
  - services
  - configmaps
  verbs:
  - get
  - list
  - watch
  - patch
  - create
  - update
  - delete
- apiGroups:
  - apps
  resources:
  - deployments
  - daemonsets
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
---
# The role for the operator to manage resources in its own namespace
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: rook-ceph-system
  namespace: {{ .Namespace }}
  labels:
    operator: rook
    storage-backend: ceph
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - configmaps
  - services
  verbs:
  - get
  - list
  - watch
  - patch
  - create
  - update
  - delete
- apiGroups:
  - apps
  resources:
  - daemonsets
  - statefulsets
  - deployments
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
---
# The cluster role for managing the Rook CRDs
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: rook-ceph-global
  labels:
    operator: rook
    storage-backend: ceph
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.ceph.rook.io/aggregate-to-rook-ceph-global: "true"
rules: []
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: rook-ceph-global-rules
  labels:
    operator: rook
    storage-backend: ceph
    rbac.ceph.rook.io/aggregate-to-rook-ceph-global: "true"
rules:
- apiGroups:
  - ""
  resources:
  # Pod access is needed for fencing
  - pods
  # Node access is needed for determining nodes where mons should run
  - nodes
  - nodes/proxy
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - events
    # PVs and PVCs are managed by the Rook provisioner
  - persistentvolumes
  - persistentvolumeclaims
  - endpoints
  verbs:
  - get
  - list
  - watch
  - patch
  - create
  - update
  - delete
- apiGroups:
  - storage.k8s.io
  resources:
  - storageclasses
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
- apiGroups:
  - ceph.rook.io
  resources:
  - "*"
  verbs:
  - "*"
- apiGroups:
  - rook.io
  resources:
  - "*"
  verbs:
  - "*"
- apiGroups:
  - policy
  - apps
  resources:
  # This is for the clusterdisruption controller
  - poddisruptionbudgets
  # This is for both clusterdisruption and nodedrain controllers
  - deployments
  - replicasets
  verbs:
  - "*"
- apiGroups:
  - healthchecking.openshift.io
  resources:
  - machinedisruptionbudgets
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
- apiGroups:
  - machine.openshift.io
  resources:
  - machines
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
- apiGroups:
  - storage.k8s.io
  resources:
  - csidrivers
  verbs:
  - create
---
# Aspects of ceph-mgr that require cluster-wide access
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr-cluster
  labels:
    operator: rook
    storage-backend: ceph
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.ceph.rook.io/aggregate-to-rook-ceph-mgr-cluster: "true"
rules: []
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr-cluster-rules
  labels:
    operator: rook
    storage-backend: ceph
    rbac.ceph.rook.io/aggregate-to-rook-ceph-mgr-cluster: "true"
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  - nodes
  - nodes/proxy
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - list
  - get
  - watch
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-object-bucket
  labels:
    operator: rook
    storage-backend: ceph
    rbac.ceph.rook.io/aggregate-to-rook-ceph-mgr-cluster: "true"
rules:
- apiGroups:
  - ""
  verbs:
  - "*"
  resources:
  - secrets
  - configmaps
- apiGroups:
    - storage.k8s.io
  resources:
    - storageclasses
  verbs:
    - get
    - list
    - watch
- apiGroups:
  - "objectbucket.io"
  verbs:
  - "*"
  resources:
  - "*"

---
# The rook system service account used by the operator, agent, and discovery pods
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-ceph-system
  namespace: {{ .Namespace }}
  labels:
    operator: rook
    storage-backend: ceph
---
# Grant the operator, agent, and discovery agents access to resources in the namespace
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-system
  namespace: {{ .Namespace }}
  labels:
    operator: rook
    storage-backend: ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rook-ceph-system
subjects:
- kind: ServiceAccount
  name: rook-ceph-system
  namespace: {{ .Namespace }}
---
# Grant the rook system daemons cluster-wide access to manage the Rook CRDs, PVCs, and storage classes
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-global
  namespace: {{ .Namespace }}
  labels:
    operator: rook
    storage-backend: ceph
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rook-ceph-global
subjects:
- kind: ServiceAccount
  name: rook-ceph-system
  namespace: {{ .Namespace }}
#################################################################################################################
# Beginning of cluster-specific resources. The example will assume the cluster will be created in the "rook-ceph"
# namespace. If you want to create the cluster in a different namespace, you will need to modify these roles
# and bindings accordingly.
#################################################################################################################
# Service account for the Ceph OSDs. Must exist and cannot be renamed.

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-ceph-osd
  namespace: {{ .Namespace }}
---
# Service account for the Ceph Mgr. Must exist and cannot be renamed.
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-ceph-mgr
  namespace: {{ .Namespace }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-ceph-cmd-reporter
  namespace: {{ .Namespace }}

---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-osd
  namespace: {{ .Namespace }}
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: [ "get", "list", "watch", "create", "update", "delete" ]
- apiGroups: ["ceph.rook.io"]
  resources: ["cephclusters", "cephclusters/finalizers"]
  verbs: [ "get", "list", "create", "update", "delete" ]
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-osd
  namespace: {{ .Namespace }}
rules:
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - get
  - list
---
# Aspects of ceph-mgr that require access to the system namespace
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr-system
  namespace: {{ .Namespace }}
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.ceph.rook.io/aggregate-to-rook-ceph-mgr-system: "true"
rules: []
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr-system-rules
  namespace: {{ .Namespace }}
  labels:
      rbac.ceph.rook.io/aggregate-to-rook-ceph-mgr-system: "true"
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
---
# Aspects of ceph-mgr that operate within the cluster's namespace
kind: Role
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr
  namespace: {{ .Namespace }}
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - batch
  resources:
  - jobs
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete
- apiGroups:
  - ceph.rook.io
  resources:
  - "*"
  verbs:
  - "*"

---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-cmd-reporter
  namespace: {{ .Namespace }}
rules:
- apiGroups:
  - ""
  resources:
  - pods
  - configmaps
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - delete

---
# Allow the operator to create resources in this cluster's namespace
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-cluster-mgmt
  namespace: {{ .Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rook-ceph-cluster-mgmt
subjects:
- kind: ServiceAccount
  name: rook-ceph-system
  namespace: {{ .Namespace }}
---
# Allow the osd pods in this namespace to work with configmaps
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-osd
  namespace: {{ .Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rook-ceph-osd
subjects:
- kind: ServiceAccount
  name: rook-ceph-osd
  namespace: {{ .Namespace }}
---
# Allow the ceph mgr to access the cluster-specific resources necessary for the mgr modules
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr
  namespace: {{ .Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rook-ceph-mgr
subjects:
- kind: ServiceAccount
  name: rook-ceph-mgr
  namespace: {{ .Namespace }}
---
# Allow the ceph mgr to access the rook system resources necessary for the mgr modules
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr-system
  namespace: {{ .Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rook-ceph-mgr-system
subjects:
- kind: ServiceAccount
  name: rook-ceph-mgr
  namespace: {{ .Namespace }}
---
# Allow the ceph mgr to access cluster-wide resources necessary for the mgr modules
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-mgr-cluster
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rook-ceph-mgr-cluster
subjects:
- kind: ServiceAccount
  name: rook-ceph-mgr
  namespace: {{ .Namespace }}

---
# Allow the ceph osd to access cluster-wide resources necessary for determining their topology location
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-osd
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: rook-ceph-osd
subjects:
- kind: ServiceAccount
  name: rook-ceph-osd
  namespace: {{ .Namespace }}


---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-cmd-reporter
  namespace: {{ .Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rook-ceph-cmd-reporter
subjects:
- kind: ServiceAccount
  name: rook-ceph-cmd-reporter
  namespace: {{ .Namespace }}
#################################################################################################################
# Beginning of pod security policy resources. The example will assume the cluster will be created in the
# "rook-ceph" namespace. If you want to create the cluster in a different namespace, you will need to modify
# the roles and bindings accordingly.
#################################################################################################################

---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: rook-privileged
spec:
  privileged: true
  allowedCapabilities:
    # required by CSI
    - SYS_ADMIN
  # fsGroup - the flexVolume agent has fsGroup capabilities and could potentially be any group
  fsGroup:
    rule: RunAsAny
  # runAsUser, supplementalGroups - Rook needs to run some pods as root
  # Ceph pods could be run as the Ceph user, but that user isn't always known ahead of time
  runAsUser:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  # seLinux - seLinux context is unknown ahead of time; set if this is well-known
  seLinux:
    rule: RunAsAny
  volumes:
    # recommended minimum set
    - configMap
    - downwardAPI
    - emptyDir
    - persistentVolumeClaim
    - secret
    - projected
    # required for Rook
    - hostPath
    - flexVolume
  # allowedHostPaths can be set to Rook's known host volume mount points when they are fully-known
  # directory-based OSDs make this hard to nail down
  # allowedHostPaths:
  #   - pathPrefix: "/run/udev"  # for OSD prep
  #     readOnly: false
  #   - pathPrefix: "/dev"  # for OSD prep
  #     readOnly: false
  #   - pathPrefix: "/var/lib/rook"  # or whatever the dataDirHostPath value is set to
  #     readOnly: false
  # Ceph requires host IPC for setting up encrypted devices
  hostIPC: true
  # Ceph OSDs need to share the same PID namespace
  hostPID: true
  # hostNetwork can be set to 'false' if host networking isn't used
  hostNetwork: true
  hostPorts:
    # Ceph messenger protocol v1
    - min: 6789
      max: 6790 # <- support old default port
    # Ceph messenger protocol v2
    - min: 3300
      max: 3300
    # Ceph RADOS ports for OSDs, MDSes
    - min: 6800
      max: 7300
    # # Ceph dashboard port HTTP (not recommended)
    # - min: 7000
    #   max: 7000
    # Ceph dashboard port HTTPS
    - min: 8443
      max: 8443
    # Ceph mgr Prometheus Metrics
    - min: 9283
      max: 9283

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: 'psp:rook'
rules:
  - apiGroups:
      - policy
    resources:
      - podsecuritypolicies
    resourceNames:
      - rook-privileged
    verbs:
      - use
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rook-ceph-system-psp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: 'psp:rook'
subjects:
  - kind: ServiceAccount
    name: rook-ceph-system
    namespace: {{ .Namespace }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rook-ceph-default-psp
  namespace: {{ .Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:rook
subjects:
- kind: ServiceAccount
  name: default
  namespace: {{ .Namespace }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rook-ceph-osd-psp
  namespace: {{ .Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:rook
subjects:
- kind: ServiceAccount
  name: rook-ceph-osd
  namespace: {{ .Namespace }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rook-ceph-mgr-psp
  namespace: {{ .Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:rook
subjects:
- kind: ServiceAccount
  name: rook-ceph-mgr
  namespace: {{ .Namespace }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: rook-ceph-cmd-reporter-psp
  namespace: {{ .Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: psp:rook
subjects:
- kind: ServiceAccount
  name: rook-ceph-cmd-reporter
  namespace: {{ .Namespace }}

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-csi-cephfs-plugin-sa
  namespace: {{ .Namespace }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-csi-cephfs-provisioner-sa
  namespace: {{ .Namespace }}

---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: {{ .Namespace }}
  name: cephfs-external-provisioner-cfg
rules:
  - apiGroups: [""]
    resources: ["endpoints"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list", "create", "delete"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]

---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-provisioner-role-cfg
  namespace: {{ .Namespace }}
subjects:
  - kind: ServiceAccount
    name: rook-csi-cephfs-provisioner-sa
    namespace: {{ .Namespace }}
roleRef:
  kind: Role
  name: cephfs-external-provisioner-cfg
  apiGroup: rbac.authorization.k8s.io

---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-nodeplugin
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.ceph.rook.io/aggregate-to-cephfs-csi-nodeplugin: "true"
rules: []
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-nodeplugin-rules
  labels:
    rbac.ceph.rook.io/aggregate-to-cephfs-csi-nodeplugin: "true"
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "update"]
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list"]
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-external-provisioner-runner
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.ceph.rook.io/aggregate-to-cephfs-external-provisioner-runner: "true"
rules: []
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-external-provisioner-runner-rules
  labels:
    rbac.ceph.rook.io/aggregate-to-cephfs-external-provisioner-runner: "true"
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete", "update"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list", "watch", "create", "update", "patch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rook-csi-cephfs-plugin-sa-psp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: 'psp:rook'
subjects:
  - kind: ServiceAccount
    name: rook-csi-cephfs-plugin-sa
    namespace: {{ .Namespace }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rook-csi-cephfs-provisioner-sa-psp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: 'psp:rook'
subjects:
  - kind: ServiceAccount
    name: rook-csi-cephfs-provisioner-sa
    namespace: {{ .Namespace }}
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-nodeplugin
subjects:
  - kind: ServiceAccount
    name: rook-csi-cephfs-plugin-sa
    namespace: {{ .Namespace }}
roleRef:
  kind: ClusterRole
  name: cephfs-csi-nodeplugin
  apiGroup: rbac.authorization.k8s.io

---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: cephfs-csi-provisioner-role
subjects:
  - kind: ServiceAccount
    name: rook-csi-cephfs-provisioner-sa
    namespace: {{ .Namespace }}
roleRef:
  kind: ClusterRole
  name: cephfs-external-provisioner-runner
  apiGroup: rbac.authorization.k8s.io

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-csi-rbd-plugin-sa
  namespace: {{ .Namespace }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: rook-csi-rbd-provisioner-sa
  namespace: {{ .Namespace }}

---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: {{ .Namespace }}
  name: rbd-external-provisioner-cfg
rules:
  - apiGroups: [""]
    resources: ["endpoints"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: ["coordination.k8s.io"]
    resources: ["leases"]
    verbs: ["get", "watch", "list", "delete", "update", "create"]

---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-csi-provisioner-role-cfg
  namespace: {{ .Namespace }}
subjects:
  - kind: ServiceAccount
    name: rook-csi-rbd-provisioner-sa
    namespace: {{ .Namespace }}
roleRef:
  kind: Role
  name: rbd-external-provisioner-cfg
  apiGroup: rbac.authorization.k8s.io

---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-csi-nodeplugin
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.ceph.rook.io/aggregate-to-rbd-csi-nodeplugin: "true"
rules: []
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-csi-nodeplugin-rules
  labels:
    rbac.ceph.rook.io/aggregate-to-rbd-csi-nodeplugin: "true"
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "update"]
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list"]
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-external-provisioner-runner
aggregationRule:
  clusterRoleSelectors:
  - matchLabels:
      rbac.ceph.rook.io/aggregate-to-rbd-external-provisioner-runner: "true"
rules: []
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-external-provisioner-runner-rules
  labels:
    rbac.ceph.rook.io/aggregate-to-rbd-external-provisioner-runner: "true"
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get", "list"]
  - apiGroups: [""]
    resources: ["persistentvolumes"]
    verbs: ["get", "list", "watch", "create", "delete", "update"]
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["volumeattachments"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["storage.k8s.io"]
    resources: ["storageclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["events"]
    verbs: ["list", "watch", "create", "update", "patch"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshots"]
    verbs: ["get", "list", "watch", "update"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshotcontents"]
    verbs: ["create", "get", "list", "watch", "update", "delete"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshotclasses"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apiextensions.k8s.io"]
    resources: ["customresourcedefinitions"]
    verbs: ["create", "list", "watch", "delete", "get", "update"]
  - apiGroups: ["snapshot.storage.k8s.io"]
    resources: ["volumesnapshots/status"]
    verbs: ["update"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rook-csi-rbd-plugin-sa-psp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: 'psp:rook'
subjects:
  - kind: ServiceAccount
    name: rook-csi-rbd-plugin-sa
    namespace: {{ .Namespace }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: rook-csi-rbd-provisioner-sa-psp
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: 'psp:rook'
subjects:
  - kind: ServiceAccount
    name: rook-csi-rbd-provisioner-sa
    namespace: {{ .Namespace }}
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-csi-nodeplugin
subjects:
  - kind: ServiceAccount
    name: rook-csi-rbd-plugin-sa
    namespace: {{ .Namespace }}
roleRef:
  kind: ClusterRole
  name: rbd-csi-nodeplugin
  apiGroup: rbac.authorization.k8s.io
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: rbd-csi-provisioner-role
subjects:
  - kind: ServiceAccount
    name: rook-csi-rbd-provisioner-sa
    namespace: {{ .Namespace }}
roleRef:
  kind: ClusterRole
  name: rbd-external-provisioner-runner
  apiGroup: rbac.authorization.k8s.io
`

// Deployment resource definition was taken from https://github.com/rook/rook/blob/v1.2.1/cluster/examples/kubernetes/ceph/operator.yaml
// The only change was made to uncomment the configuration of `FLEXVOLUME_DIR_PATH` which is required for Rook to work.
const deploymentOperator = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rook-ceph-operator
  namespace: {{ .Namespace }}
  labels:
    operator: rook
    storage-backend: ceph
spec:
  selector:
    matchLabels:
      app: rook-ceph-operator
  replicas: 1
  template:
    metadata:
      labels:
        app: rook-ceph-operator
    spec:
      {{- if .NodeSelectors }}
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                {{- range $item := .NodeSelectors }}
                - key: {{ $item.Key }}
                  operator: {{ $item.Operator }}
                  {{- if $item.Values }}
                  values:
                    {{- range $val := $item.Values }}
                    - {{ $val }}
                    {{- end }}
                  {{- end }}
                {{- end }}
      {{- end}}
      {{- if .TolerationsRaw }}
      tolerations: {{ .TolerationsRaw }}
      {{- end }}
      serviceAccountName: rook-ceph-system
      containers:
      - name: rook-ceph-operator
        image: rook/ceph:v1.2.1
        args: ["ceph", "operator"]
        volumeMounts:
        - mountPath: /var/lib/rook
          name: rook-config
        - mountPath: /etc/ceph
          name: default-config-dir
        env:
        # If the operator should only watch for cluster CRDs in the same namespace, set this to "true".
        # If this is not set to true, the operator will watch for cluster CRDs in all namespaces.
        - name: ROOK_CURRENT_NAMESPACE_ONLY
          value: "true"
        # To disable RBAC, uncomment the following:
        # - name: RBAC_ENABLED
        #  value: "false"
        {{- if .AgentTolerationEffect }}
        # Rook Agent toleration. Will tolerate all taints with all keys.
        # Choose between NoSchedule, PreferNoSchedule and NoExecute:
        - name: AGENT_TOLERATION
          value: {{ .AgentTolerationEffect }}
        {{- end }}
        {{- if .AgentTolerationKey }}
        # (Optional) Rook Agent toleration key. Set this to the key of the taint you want to tolerate
        - name: AGENT_TOLERATION_KEY
          value: {{ .AgentTolerationKey }}
        {{- end }}
        # (Optional) Rook Agent tolerations list. Put here list of taints you want to tolerate in YAML format.
        # - name: AGENT_TOLERATIONS
        #   value: |
        #     - effect: NoSchedule
        #       key: node-role.kubernetes.io/controlplane
        #       operator: Exists
        #     - effect: NoExecute
        #       key: node-role.kubernetes.io/etcd
        #       operator: Exists
        # (Optional) Rook Agent priority class name to set on the pod(s)
        # - name: AGENT_PRIORITY_CLASS_NAME
        #   value: "<PriorityClassName>"
        # (Optional) Rook Agent NodeAffinity.
        # - name: AGENT_NODE_AFFINITY
        #   value: "role=storage-node; storage=rook,ceph"
        # (Optional) Rook Agent mount security mode. Can by 'Any' or 'Restricted'.
        # 'Any' uses Ceph admin credentials by default/fallback.
        # For using 'Restricted' you must have a Ceph secret in each namespace storage should be consumed from and
        # set 'mountUser' to the Ceph user, 'mountSecret' to the Kubernetes secret name.
        # to the namespace in which the 'mountSecret' Kubernetes secret namespace.
        # - name: AGENT_MOUNT_SECURITY_MODE
        #   value: "Any"
        # Set the path where the Rook agent can find the flex volumes
        - name: FLEXVOLUME_DIR_PATH
          value: "/var/lib/kubelet/volumeplugins"
        # Set the path where kernel modules can be found
        # - name: LIB_MODULES_DIR_PATH
        #   value: "<PathToLibModules>"
        # Mount any extra directories into the agent container
        # - name: AGENT_MOUNTS
        #   value: "somemount=/host/path:/container/path,someothermount=/host/path2:/container/path2"
        {{- if .DiscoverTolerationEffect }}
        # Rook Discover toleration. Will tolerate all taints with all keys.
        # Choose between NoSchedule, PreferNoSchedule and NoExecute:
        - name: DISCOVER_TOLERATION
          value: {{ .DiscoverTolerationEffect }}
        {{- end }}
        {{- if .DiscoverTolerationKey }}
        # (Optional) Rook Discover toleration key. Set this to the key of the taint you want to tolerate
        - name: DISCOVER_TOLERATION_KEY
          value: {{ .DiscoverTolerationKey }}
        {{- end }}
        # (Optional) Rook Discover tolerations list. Put here list of taints you want to tolerate in YAML format.
        # - name: DISCOVER_TOLERATIONS
        #   value: |
        #     - effect: NoSchedule
        #       key: node-role.kubernetes.io/controlplane
        #       operator: Exists
        #     - effect: NoExecute
        #       key: node-role.kubernetes.io/etcd
        #       operator: Exists
        # (Optional) Rook Discover priority class name to set on the pod(s)
        # - name: DISCOVER_PRIORITY_CLASS_NAME
        #   value: "<PriorityClassName>"
        # (Optional) Discover Agent NodeAffinity.
        # - name: DISCOVER_AGENT_NODE_AFFINITY
        #   value: "role=storage-node; storage=rook, ceph"
        # Allow rook to create multiple file systems. Note: This is considered
        # an experimental feature in Ceph as described at
        # http://docs.ceph.com/docs/master/cephfs/experimental-features/#multiple-filesystems-within-a-ceph-cluster
        # which might cause mons to crash as seen in https://github.com/rook/rook/issues/1027
        - name: ROOK_ALLOW_MULTIPLE_FILESYSTEMS
          value: "false"

        # The logging level for the operator: INFO | DEBUG
        - name: ROOK_LOG_LEVEL
          value: "INFO"

        # The interval to check the health of the ceph cluster and update the status in the custom resource.
        - name: ROOK_CEPH_STATUS_CHECK_INTERVAL
          value: "60s"

        # The interval to check if every mon is in the quorum.
        - name: ROOK_MON_HEALTHCHECK_INTERVAL
          value: "45s"

        # The duration to wait before trying to failover or remove/replace the
        # current mon with a new mon (useful for compensating flapping network).
        - name: ROOK_MON_OUT_TIMEOUT
          value: "600s"

        # The duration between discovering devices in the rook-discover daemonset.
        - name: ROOK_DISCOVER_DEVICES_INTERVAL
          value: "60m"

        # Whether to start pods as privileged that mount a host path, which includes the Ceph mon and osd pods.
        # This is necessary to workaround the anyuid issues when running on OpenShift.
        # For more details see https://github.com/rook/rook/issues/1314#issuecomment-355799641
        - name: ROOK_HOSTPATH_REQUIRES_PRIVILEGED
          value: "false"

        # In some situations SELinux relabelling breaks (times out) on large filesystems, and doesn't work with cephfs ReadWriteMany volumes (last relabel wins).
        # Disable it here if you have similar issues.
        # For more details see https://github.com/rook/rook/issues/2417
        - name: ROOK_ENABLE_SELINUX_RELABELING
          value: "true"

        # In large volumes it will take some time to chown all the files. Disable it here if you have performance issues.
        # For more details see https://github.com/rook/rook/issues/2254
        - name: ROOK_ENABLE_FSGROUP
          value: "true"

        # Disable automatic orchestration when new devices are discovered
        - name: ROOK_DISABLE_DEVICE_HOTPLUG
          value: "false"

        # Provide customised regex as the values using comma. For eg. regex for rbd based volume, value will be like "(?i)rbd[0-9]+".
        # In case of more than one regex, use comma to seperate between them.
        # Default regex will be "(?i)dm-[0-9]+,(?i)rbd[0-9]+,(?i)nbd[0-9]+"
        # Add regex expression after putting a comma to blacklist a disk
        # If value is empty, the default regex will be used.
        - name: DISCOVER_DAEMON_UDEV_BLACKLIST
          value: "(?i)dm-[0-9]+,(?i)rbd[0-9]+,(?i)nbd[0-9]+"

        # Whether to enable the flex driver. By default it is enabled and is fully supported, but will be deprecated in some future release
        # in favor of the CSI driver.
        - name: ROOK_ENABLE_FLEX_DRIVER
          value: "false"

        # Whether to start the discovery daemon to watch for raw storage devices on nodes in the cluster.
        # This daemon does not need to run if you are only going to create your OSDs based on StorageClassDeviceSets with PVCs.
        - name: ROOK_ENABLE_DISCOVERY_DAEMON
          value: "true"

        # Enable the default version of the CSI CephFS driver. To start another version of the CSI driver, see image properties below.
        - name: ROOK_CSI_ENABLE_CEPHFS
          value: "true"

        # Enable the default version of the CSI RBD driver. To start another version of the CSI driver, see image properties below.
        - name: ROOK_CSI_ENABLE_RBD
          value: "true"
        - name: ROOK_CSI_ENABLE_GRPC_METRICS
          value: "true"
        # Enable deployment of snapshotter container in ceph-csi provisioner.
        - name: CSI_ENABLE_SNAPSHOTTER
          value: "true"
        # Enable Ceph Kernel clients on kernel < 4.17 which support quotas for Cephfs
        # If you disable the kernel client, your application may be disrupted during upgrade.
        # See the upgrade guide: https://rook.io/docs/rook/v1.2/ceph-upgrade.html
        - name: CSI_FORCE_CEPHFS_KERNEL_CLIENT
          value: "true"
        # CSI CephFS plugin daemonset update strategy, supported values are OnDelete and RollingUpdate.
        # Default value is RollingUpdate.
        #- name: CSI_CEPHFS_PLUGIN_UPDATE_STRATEGY
        #  value: "OnDelete"
        # CSI Rbd plugin daemonset update strategy, supported values are OnDelete and RollingUpdate.
        # Default value is RollingUpdate.
        #- name: CSI_RBD_PLUGIN_UPDATE_STRATEGY
        #  value: "OnDelete"
        # The default version of CSI supported by Rook will be started. To change the version
        # of the CSI driver to something other than what is officially supported, change
        # these images to the desired release of the CSI driver.
        #- name: ROOK_CSI_CEPH_IMAGE
        #  value: "quay.io/cephcsi/cephcsi:v1.2.2"
        #- name: ROOK_CSI_REGISTRAR_IMAGE
        #  value: "quay.io/k8scsi/csi-node-driver-registrar:v1.1.0"
        #- name: ROOK_CSI_PROVISIONER_IMAGE
        #  value: "quay.io/k8scsi/csi-provisioner:v1.4.0"
        #- name: ROOK_CSI_SNAPSHOTTER_IMAGE
        #  value: "quay.io/k8scsi/csi-snapshotter:v1.2.2"
        #- name: ROOK_CSI_ATTACHER_IMAGE
        #  value: "quay.io/k8scsi/csi-attacher:v1.2.0"
        # kubelet directory path, if kubelet configured to use other than /var/lib/kubelet path.
        #- name: ROOK_CSI_KUBELET_DIR_PATH
        #  value: "/var/lib/kubelet"
        # (Optional) Ceph Provisioner NodeAffinity.
        # - name: CSI_PROVISIONER_NODE_AFFINITY
        #   value: "role=storage-node; storage=rook, ceph"
        # (Optional) CEPH CSI provisioner tolerations list. Put here list of taints you want to tolerate in YAML format.
        #  CSI provisioner would be best to start on the same nodes as other ceph daemons.
        # - name: CSI_PROVISIONER_TOLERATIONS
        #   value: |
        #     - effect: NoSchedule
        #       key: node-role.kubernetes.io/controlplane
        #       operator: Exists
        #     - effect: NoExecute
        #       key: node-role.kubernetes.io/etcd
        #       operator: Exists
        # (Optional) Ceph CSI plugin NodeAffinity.
        # - name: CSI_PLUGIN_NODE_AFFINITY
        #   value: "role=storage-node; storage=rook, ceph"
        # (Optional) CEPH CSI plugin tolerations list. Put here list of taints you want to tolerate in YAML format.
        # CSI plugins need to be started on all the nodes where the clients need to mount the storage.
        # - name: CSI_PLUGIN_TOLERATIONS
        #   value: |
        #     - effect: NoSchedule
        #       key: node-role.kubernetes.io/controlplane
        #       operator: Exists
        #     - effect: NoExecute
        #       key: node-role.kubernetes.io/etcd
        #       operator: Exists
        # Configure CSI cephfs grpc and liveness metrics port
        #- name: CSI_CEPHFS_GRPC_METRICS_PORT
        #  value: "9091"
        #- name: CSI_CEPHFS_LIVENESS_METRICS_PORT
        #  value: "9081"
        # Configure CSI rbd grpc and liveness metrics port
        #- name: CSI_RBD_GRPC_METRICS_PORT
        #  value: "9090"
        #- name: CSI_RBD_LIVENESS_METRICS_PORT
        #  value: "9080"

        # Time to wait until the node controller will move Rook pods to other
        # nodes after detecting an unreachable node.
        # Pods affected by this setting are:
        # mgr, rbd, mds, rgw, nfs, PVC based mons and osds, and ceph toolbox
        # The value used in this variable replaces the default value of 300 secs
        # added automatically by k8s as Toleration for
        # <node.kubernetes.io/unreachable>
        # The total amount of time to reschedule Rook pods in healthy nodes
        # before detecting a <not ready node> condition will be the sum of:
        #  --> node-monitor-grace-period: 40 seconds (k8s kube-controller-manager flag)
        #  --> ROOK_UNREACHABLE_NODE_TOLERATION_SECONDS: 5 seconds
        - name: ROOK_UNREACHABLE_NODE_TOLERATION_SECONDS
          value: "5"

        # The name of the node to pass with the downward API
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        # The pod name to pass with the downward API
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        # The pod namespace to pass with the downward API
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
      # Uncomment it to run rook operator on the host network
      #hostNetwork: true
      volumes:
      - name: rook-config
        emptyDir: {}
      - name: default-config-dir
        emptyDir: {}
`
