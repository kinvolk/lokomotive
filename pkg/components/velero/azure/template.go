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

package azure

const chartValuesTmpl = `
configuration:
  provider: azure
  backupStorageLocation:
    name: azure
    provider: velero.io/azure
    bucket: {{ .BackupStorageLocation.Bucket }}
    config:
      resourceGroup: {{ .BackupStorageLocation.ResourceGroup }}
      storageAccount: {{ .BackupStorageLocation.StorageAccount }}
  volumeSnapshotLocation:
    {{- if .VolumeSnapshotLocation.Name }}
    name: {{ .VolumeSnapshotLocation.Name }}
    {{- end }}
    provider: velero.io/azure
    config:
      {{- if .VolumeSnapshotLocation.ResourceGroup }}
      resourceGroup: {{ .VolumeSnapshotLocation.ResourceGroup }}
      {{- end }}
      apitimeout: {{ .VolumeSnapshotLocation.APITimeout }}
credentials:
  secretContents:
    cloud: |
      AZURE_SUBSCRIPTION_ID: "{{ .SubscriptionID }}"
      AZURE_TENANT_ID: "{{ .TenantID }}"
      AZURE_CLIENT_ID: "{{ .ClientID }}"
      AZURE_CLIENT_SECRET: "{{ .ClientSecret }}"
      AZURE_RESOURCE_GROUP: "{{ .ResourceGroup }}"
initContainers:
- image: velero/velero-plugin-for-microsoft-azure:v1.0.0
  imagePullPolicy: IfNotPresent
  name: velero-plugin-for-azure
  resources: {}
  terminationMessagePath: /dev/termination-log
  terminationMessagePolicy: File
  volumeMounts:
  - mountPath: /target
    name: plugins
`
