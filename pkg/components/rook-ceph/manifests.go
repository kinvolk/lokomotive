// Copyright 2020 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rookceph

// CephCluster resource definition was taken from:
// https://github.com/rook/rook/blob/v1.4.6/cluster/examples/kubernetes/ceph/cluster.yaml
var template = map[string]string{
	"ceph-cluster.yaml": `
apiVersion: ceph.rook.io/v1
kind: CephCluster
metadata:
  name: rook-ceph
  namespace: {{ .Namespace }}
spec:
  cephVersion:
    image: ceph/ceph:v15.2.5-20200916
    allowUnsupported: false
  dataDirHostPath: /var/lib/rook
  skipUpgradeChecks: false
  continueUpgradeAfterChecksEvenIfNotHealthy: false
  mon:
    count: {{ .MonitorCount }}
    allowMultiplePerNode: false
  mgr:
    modules:
    - name: pg_autoscaler
      enabled: true
  dashboard:
    enabled: true
    ssl: true
  monitoring:
    enabled: true
    rulesNamespace: {{ .Namespace }}
  network:
  crashCollector:
    disable: false
  cleanupPolicy:
    confirmation: ""
    sanitizeDisks:
      method: quick
      dataSource: zero
      iteration: 1
    allowUninstallWithVolumes: false
  placement:
    all:
      {{- if .NodeAffinity }}
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            - matchExpressions:
              {{- range $item := .NodeAffinity }}
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
  annotations:
  resources:
  removeOSDsIfOutAndSafeToRemove: false
  storage: # cluster level storage configuration and selection
    useAllNodes: true
    useAllDevices: true
    config:
      {{- if .MetadataDevice }}
      metadataDevice: "{{ .MetadataDevice }}"
      {{- end }}
      storeType: bluestore
      osdsPerDevice: "1" # this value can be overridden at the node or device level
  disruptionManagement:
    managePodBudgets: false
    osdMaintenanceTimeout: 30
    manageMachineDisruptionBudgets: false
    machineDisruptionBudgetNamespace: openshift-machine-api

  # healthChecks
  # Valid values for daemons are 'mon', 'osd', 'status'
  healthCheck:
    daemonHealth:
      mon:
        disabled: false
        interval: 45s
      osd:
        disabled: false
        interval: 60s
      status:
        disabled: false
        interval: 60s
    # Change pod liveness probe, it works for all mon,mgr,osd daemons
    livenessProbe:
      mon:
        disabled: false
      mgr:
        disabled: false
      osd:
        disabled: false
`,

	"storage-class.yaml": `
{{- if .StorageClass.Enable }}
apiVersion: ceph.rook.io/v1
kind: CephBlockPool
metadata:
  name: replicapool
  namespace: {{ .Namespace }}
spec:
  failureDomain: host
  replicated:
    size: 3
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: rook-ceph-block
  annotations:
    {{- if .StorageClass.Default }}
    storageclass.kubernetes.io/is-default-class: "true"
    {{- end }}
allowVolumeExpansion: true
provisioner: {{ .Namespace }}.rbd.csi.ceph.com
parameters:
  clusterID: {{ .Namespace }}
  # Ceph pool into which the RBD image shall be created
  pool: replicapool

  # RBD image format. Defaults to "2".
  imageFormat: "2"

  # RBD image features. Available for imageFormat: "2". CSI RBD currently supports only 'layering' feature.
  imageFeatures: layering

  # The secrets contain Ceph admin credentials.
  csi.storage.k8s.io/provisioner-secret-name: rook-csi-rbd-provisioner
  csi.storage.k8s.io/provisioner-secret-namespace: {{ .Namespace }}
  csi.storage.k8s.io/node-stage-secret-name: rook-csi-rbd-node
  csi.storage.k8s.io/node-stage-secret-namespace: {{ .Namespace }}
  csi.storage.k8s.io/controller-expand-secret-name: rook-csi-rbd-provisioner
  csi.storage.k8s.io/controller-expand-secret-namespace: {{ .Namespace }}

  # Specify the filesystem type of the volume. If not specified, csi-provisioner
  # will set default as 'ext4'.
  csi.storage.k8s.io/fstype: xfs

# Delete the rbd volume when a PVC is deleted
reclaimPolicy: Delete
{{- end }}
`,

	"ceph-toolbox.yaml": `
{{- if .EnableToolbox }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rook-ceph-tools
  namespace: {{ .Namespace }}
  labels:
    app: rook-ceph-tools
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rook-ceph-tools
  template:
    metadata:
      labels:
        app: rook-ceph-tools
    spec:
      dnsPolicy: ClusterFirstWithHostNet
      containers:
      - name: rook-ceph-tools
        image: rook/ceph:v1.4.6
        command: ["/tini"]
        args: ["-g", "--", "/usr/local/bin/toolbox.sh"]
        imagePullPolicy: IfNotPresent
        env:
        - name: ROOK_CEPH_USERNAME
          valueFrom:
            secretKeyRef:
              name: rook-ceph-mon
              key: ceph-username
        - name: ROOK_CEPH_SECRET
          valueFrom:
            secretKeyRef:
              name: rook-ceph-mon
              key: ceph-secret
        volumeMounts:
        - mountPath: /etc/ceph
          name: ceph-config
        - name: mon-endpoint-volume
          mountPath: /etc/rook
      volumes:
      - name: mon-endpoint-volume
        configMap:
          name: rook-ceph-mon-endpoints
          items:
          - key: data
            path: mon-endpoints
      - name: ceph-config
        emptyDir: {}
{{- if .TolerationsRaw }}
      tolerations: {{ .TolerationsRaw }}
{{- end }}
{{- end }}
`,
}
