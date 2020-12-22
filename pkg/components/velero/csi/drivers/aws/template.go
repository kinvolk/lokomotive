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

package aws

const chartValuesTmpl = `
configuration:
  features: "EnableCSI"
  provider: aws
  backupStorageLocation:
    provider: velero.io/aws
    {{- if .Configuration.BackupStorageLocation.Name}}
    name: {{ .Configuration.BackupStorageLocation.Name }}
    {{- end }}
    bucket: {{ .Configuration.BackupStorageLocation.Bucket }}
    config:
      region: {{ .Configuration.BackupStorageLocation.Region }}
  volumeSnapshotLocation:
    provider: velero.io/aws
    {{- if .Configuration.VolumeSnapshotLocation.Name }}
    name: {{ .Configuration.VolumeSnapshotLocation.Name }}
    {{- end }}
    config:
      region: {{ .Configuration.VolumeSnapshotLocation.Region }}
credentials:
  secretContents:
  {{- if .Configuration.Credentials }}
    cloud: |
{{ .CredentialsIndented }}
  {{- end }}
initContainers:
- image: velero/velero-plugin-for-aws:v1.1.0
  imagePullPolicy: IfNotPresent
  name: velero-plugin-for-aws
  resources: {}
  terminationMessagePath: /dev/termination-log
  terminationMessagePolicy: File
  volumeMounts:
  - mountPath: /target
    name: plugins
- image: velero/velero-plugin-for-csi:v0.1.2
  imagePullPolicy: IfNotPresent
  name: velero-plugin-for-csi
  resources: {}
  terminationMessagePath: /dev/termination-log
  terminationMessagePolicy: File
  volumeMounts:
  - mountPath: /target
    name: plugins
`
