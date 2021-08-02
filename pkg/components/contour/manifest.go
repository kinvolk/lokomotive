// Copyright 2021 The Lokomotive Authors
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

// Package contour has code related to deployment of contour component.
package contour

const chartValuesTmpl = `
{{- if .EnableMonitoring }}
monitoring:
  enable: {{ .EnableMonitoring }}
{{- end }}

envoy:
  serviceType: {{ .ServiceType }}
  {{- if .Envoy }}
  metricsScrapeInterval: {{ .Envoy.MetricsScrapeInterval }}
  {{- end }}

{{- if .NodeAffinity }}
nodeAffinity:
  requiredDuringSchedulingIgnoredDuringExecution:
    nodeSelectorTerms:
    - matchExpressions: {{ .NodeAffinityRaw }}
{{- end}}

{{- if .Tolerations }}
tolerations: {{ .TolerationsRaw }}
{{- end }}
`
