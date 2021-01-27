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

package restic

const chartValuesTmpl = `
configuration:
  provider: {{ .Configuration.BackupStorageLocation.Provider }}
  backupStorageLocation:
    {{- if .Configuration.BackupStorageLocation.Provider}}
    provider: {{ .Configuration.BackupStorageLocation.Provider }}
    {{- end }}
    {{- if .Configuration.BackupStorageLocation.Name}}
    name: {{ .Configuration.BackupStorageLocation.Name }}
    {{- end }}
    bucket: {{ .Configuration.BackupStorageLocation.Bucket }}
    config:
      region: {{ .Configuration.BackupStorageLocation.Region }}
  {{- if not .Configuration.RequireVolumeAnnotation }}
  defaultVolumesToRestic: true
  {{- end }}
deployRestic: true
snapshotsEnabled: false
restic:
  privileged: true
  {{- if .Configuration.Tolerations }}
  tolerations: {{ .Configuration.TolerationsRaw }}
  {{- end }}
credentials:
  secretContents:
  {{- if .Configuration.Credentials }}
    cloud: |
{{ .CredentialsIndented }}
  {{- end }}
initContainers:
{{- if eq .Configuration.BackupStorageLocation.Provider "aws" }}
- image: velero/velero-plugin-for-aws:v1.1.0
  imagePullPolicy: IfNotPresent
  name: velero-plugin-for-aws
  resources: {}
  terminationMessagePath: /dev/termination-log
  terminationMessagePolicy: File
  volumeMounts:
  - mountPath: /target
    name: plugins
{{- end }}
{{- if eq .Configuration.BackupStorageLocation.Provider "gcp" }}
- image: velero/velero-plugin-for-gcp:v1.1.0
  imagePullPolicy: IfNotPresent
  name: velero-plugin-for-gcp
  resources: {}
  terminationMessagePath: /dev/termination-log
  terminationMessagePolicy: File
  volumeMounts:
  - mountPath: /target
    name: plugins
{{- end }}
{{- if eq .Configuration.BackupStorageLocation.Provider "azure" }}
- image: velero/velero-plugin-for-microsoft-azure:v1.1.0
  imagePullPolicy: IfNotPresent
  name: velero-plugin-for-azure
  resources: {}
  terminationMessagePath: /dev/termination-log
  terminationMessagePolicy: File
  volumeMounts:
  - mountPath: /target
    name: plugins
{{- end }}
`
