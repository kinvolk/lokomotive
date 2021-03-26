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
reclaimPolicy: "{{ .ReclaimPolicy }}"
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
