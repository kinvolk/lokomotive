package rook

// These manifests were taken from https://github.com/rook/rook/tree/release-1.0/cluster/examples/kubernetes/ceph with
// some modifications (explained below on a case to case basis).

// Namespace was split from https://github.com/rook/rook/blob/release-1.0/cluster/examples/kubernetes/ceph/common.yaml
// for easier management in code.
const namespace = `
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Namespace }}
`

// CRD resource definitions was split from https://github.com/rook/rook/blob/release-1.0/cluster/examples/kubernetes/ceph/common.yaml
// for easier management in code.
const crds = `
# The CRD declarations
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
            cephVersion:
              properties:
                allowUnsupported:
                  type: boolean
                image:
                  type: string
                name:
                  pattern: ^(luminous|mimic|nautilus)$
                  type: string
            dashboard:
              properties:
                enabled:
                  type: boolean
                urlPrefix:
                  type: string
                port:
                  type: integer
            dataDirHostPath:
              pattern: ^/(\S+)
              type: string
            mon:
              properties:
                allowMultiplePerNode:
                  type: boolean
                count:
                  maximum: 9
                  minimum: 1
                  type: integer
                preferredCount:
                  maximum: 9
                  minimum: 0
                  type: integer
              required:
              - count
            network:
              properties:
                hostNetwork:
                  type: boolean
            storage:
              properties:
                nodes:
                  items: {}
                  type: array
                useAllDevices: {}
                useAllNodes:
                  type: boolean
          required:
          - mon
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
  additionalPrinterColumns:
    - name: MdsCount
      type: string
      description: Number of MDSs
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
`

// RBAC resource definitions was split from
// https://github.com/rook/rook/blob/release-1.0/cluster/examples/kubernetes/ceph/common.yaml for easier management in
// code.  This section contains additional Role and Rolebinding named "rook-psp-privileged" that allows the service
// accounts in the {{ .Namespace }} namespace to use the privileged PSP. Without this, Rook will not work on a cluster
// with PSP enabled.
const rbac = `
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
---
#################################################################################################################
# Beginning of cluster-specific resources. The example will assume the cluster will be created in the "rook-ceph"
# namespace. If you want to create the cluster in a different namespace, you will need to modify these roles
# and bindings accordingly.
#################################################################################################################
# Service account for the Ceph OSDs. Must exist and cannot be renamed.
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
kind: Role
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: rook-ceph-osd
  namespace: {{ .Namespace }}
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: [ "get", "list", "watch", "create", "update", "delete" ]
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
# Create a Role to allow the use of the privileged PSP. For this to
# work, a PSP with the name "privileged" must exist. Remove if running
# on a cluster without PSPs enabled.
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: Role
metadata:
  name: rook-psp-privileged
  namespace: {{ .Namespace }}
rules:
- apiGroups:
  - extensions
  resourceNames:
  - privileged
  resources:
  - podsecuritypolicies
  verbs:
  - use
---
# Create a RoleBinding to allow the service accounts in the {{ .Namespace }}
# namespace to use the privileged PSP.
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: RoleBinding
metadata:
  name: rook-psp-privileged
  namespace: {{ .Namespace }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: rook-psp-privileged
subjects:
- kind: ServiceAccount
  name: rook-ceph-system
  namespace: {{ .Namespace }}
- kind: ServiceAccount
  name: default
  namespace: {{ .Namespace }}
- kind: ServiceAccount
  name: rook-ceph-mgr
  namespace: {{ .Namespace }}
- kind: ServiceAccount
  name: rook-ceph-osd
  namespace: {{ .Namespace }}
`

// Deployment resource definition was taken from https://github.com/rook/rook/blob/release-1.0/cluster/examples/kubernetes/ceph/operator.yaml
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
        image: rook/ceph:v1.0.4
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
        # (Optional) Rook Agent mount security mode. Can by "Any" or "Restricted".
        # "Any" uses Ceph admin credentials by default/fallback.
        # For using "Restricted" you must have a Ceph secret in each namespace storage should be consumed from and
        # set "mountUser" to the Ceph user, "mountSecret" to the Kubernetes secret name.
        # to the namespace in which the "mountSecret" Kubernetes secret namespace.
        # - name: AGENT_MOUNT_SECURITY_MODE
        #   value: "Any"
        # Set the path where the Rook agent can find the flex volumes
        - name: FLEXVOLUME_DIR_PATH
          value: "/var/lib/kubelet/volumeplugins"
        # Set the path where kernel modules can be found
        # - name: LIB_MODULES_DIR_PATH
        #  value: "<PathToLibModules>"
        # Mount any extra directories into the agent container
        # - name: AGENT_MOUNTS
        #  value: "somemount=/host/path:/container/path,someothermount=/host/path2:/container/path2"
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
      volumes:
      - name: rook-config
        emptyDir: {}
      - name: default-config-dir
        emptyDir: {}
`
