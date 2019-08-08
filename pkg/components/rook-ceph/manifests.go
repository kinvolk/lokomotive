package rookceph

// CephCluster resource definition was taken from https://github.com/rook/rook/blob/release-1.0/cluster/examples/kubernetes/ceph/cluster.yaml
const cephCluster = `
apiVersion: ceph.rook.io/v1
kind: CephCluster
metadata:
  name: rook-ceph
  namespace: {{ .Namespace }}
spec:
  cephVersion:
    image: ceph/ceph:v14.2.1-20190430
    allowUnsupported: true
  dataDirHostPath: /var/lib/rook
  mon:
    count: {{ .MonitorCount }}
    allowMultiplePerNode: false
  dashboard:
    enabled: true
  network:
    hostNetwork: false
  # RBD is required for block device. At the moment we only need object storage so this can be skipped.
  rbdMirroring:
    workers: 0
  placement:
    all:
      {{- if .NodeSelectors }}
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
  storage:
    useAllNodes: true
    useAllDevices: true
    config:
      storeType: bluestore
      osdsPerDevice: "1" # this value can be overridden at the node or device level
    # directories:
    # - path: /var/lib/rook
    #   # /dev/md/node-local-storage/rook
`
