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

var chartValuesTmpl = `
storageClass:
  enable: {{.StorageClass.Enable}}
  default: {{.StorageClass.Default}}
  reclaimPolicy: {{.StorageClass.ReclaimPolicy}}

enableToolbox: {{.EnableToolbox}}

cephCluster:
  {{if .Resources}}
  resources:
    mon: {{.Resources.MONRaw}}
    mgr: {{.Resources.MGRRaw}}
    osd: {{.Resources.OSDRaw}}
    mds: {{.Resources.MDSRaw}}
    prepareosd: {{.Resources.PrepareOSDRaw}}
    crashcollector: {{.Resources.CrashCollectorRaw}}
    mgrSidecar: {{.Resources.MGRSidecarRaw}}
  {{end}}
  mon:
    count: {{.MonitorCount}}
  {{if .MetadataDevice}}
  metadataDevice: {{.MetadataDevice}}
  {{end}}

{{if .NodeAffinity}}
nodeAffinity:
  requiredDuringSchedulingIgnoredDuringExecution:
    nodeSelectorTerms:
    - matchExpressions: {{.NodeAffinityRaw}}
{{end}}

{{if .Tolerations}}
tolerations: {{.TolerationsRaw}}
{{end}}
  `
