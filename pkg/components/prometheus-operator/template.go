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

package prometheus

const chartValuesTmpl = `
alertmanager:
{{.AlertManagerConfig}}
  alertmanagerSpec:
    retention: {{.AlertManagerRetention}}
    externalUrl: {{.AlertManagerExternalURL}}
    {{ if .AlertManagerNodeSelector }}
    nodeSelector:
      {{ range $key, $value := .AlertManagerNodeSelector }}
      {{ $key }}: {{ $value }}
      {{ end }}
    {{ end }}
grafana:
  adminPassword: {{.GrafanaAdminPassword}}
  testFramework:
    enabled: false
  rbac:
    pspUseAppArmor: false
kubeEtcd:
  endpoints: {{.EtcdEndpoints}}
prometheus-node-exporter:
  service: {}
{{ if .PrometheusOperatorNodeSelector }}
prometheusOperator:
  nodeSelector:
    {{ range $key, $value := .PrometheusOperatorNodeSelector }}
    {{ $key }}: {{ $value }}
    {{ end }}
{{ end }}
prometheus:
  prometheusSpec:
    externalUrl: {{.PrometheusExternalURL}}
    {{ if .PrometheusNodeSelector }}
    nodeSelector:
      {{ range $key, $value := .PrometheusNodeSelector }}
      {{ $key }}: {{ $value }}
      {{ end }}
    {{ end }}
    retention: {{.PrometheusMetricsRetention}}
`
