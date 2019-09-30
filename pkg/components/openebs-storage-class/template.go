package openebsstorageclass

const storageClassTmpl = `
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: {{ .Name }}
  annotations:
    openebs.io/cas-type: cstor
    storageclass.kubernetes.io/is-default-class: "{{ .Default }}"
    cas.openebs.io/config: |
      - name: StoragePoolClaim
        value: "cstor-pool-{{ .Name }}"
      - name: ReplicaCount
        value: "{{ .ReplicaCount }}"
provisioner: openebs.io/provisioner-iscsi
reclaimPolicy: Delete
`

const storagePoolTmpl = `
apiVersion: openebs.io/v1alpha1
kind: StoragePoolClaim
metadata:
  name: cstor-pool-{{ .Name }}
spec:
  name: cstor-pool-{{ .Name }}
  type: disk
  maxPools: 3
  poolSpec:
    poolType: striped
  {{- if .Disks }}
  blockDevices:
    blockDeviceList:
    {{range .Disks -}}
    - {{.}}
    {{end}}
  {{- end }}
`
