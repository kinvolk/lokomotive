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
