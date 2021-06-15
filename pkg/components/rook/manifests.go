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

package rook

const chartValuesTmpl = `
resources:
  limits:
    memory: 512Mi

{{- if .NodeSelector }}
nodeSelector: {{ .NodeSelectorRaw }}
{{- end }}

{{- if .Tolerations }}
tolerations: {{ .TolerationsRaw }}
{{- end }}

# This is disabled by default from now. But baremetal setups should have this daemon running.
enableDiscoveryDaemon: true

csi:
  # This is set explicitly because the default port 9091 conflicts with Calico's metrics port.
  cephfsGrpcMetricsPort: 9092

  {{- if .Tolerations }}
  provisionerTolerations: {{ .TolerationsRaw }}
  {{- end }}
  {{- if .NodeSelector }}
  provisionerNodeAffinity: {{ .RookNodeAffinity }}
  {{- end }}

  {{- if .CSIPluginTolerations }}
  pluginTolerations: {{ .CSIPluginTolerationsRaw }}
  {{- end }}
  {{- if .CSIPluginNodeSelector }}
  pluginNodeAffinity: {{ .CSIPluginNodeAffinity }}
  {{- end }}

agent:
  flexVolumeDirPath: "/var/lib/kubelet/volumeplugins"
  {{- if and .AgentTolerationKey .AgentTolerationEffect }}
  tolerationKey: {{ .AgentTolerationKey }}
  toleration: {{ .AgentTolerationEffect }}
  {{- end }}

  {{- if .NodeSelector }}
  nodeAffinity: {{ .RookNodeAffinity }}
  {{- end }}

  {{- if .Tolerations }}
  tolerations: {{ .TolerationsRaw }}
  {{- end }}

discover:
  {{- if and .DiscoverTolerationKey .DiscoverTolerationEffect }}
  tolerationKey: {{ .DiscoverTolerationKey }}
  toleration: {{ .DiscoverTolerationEffect }}
  {{- end }}

  {{- if .NodeSelector }}
  nodeAffinity: {{ .RookNodeAffinity }}
  {{- end }}

  {{- if .Tolerations }}
  tolerations: {{ .TolerationsRaw }}
  {{- end }}

enableMonitoring: {{ .EnableMonitoring }}
`
