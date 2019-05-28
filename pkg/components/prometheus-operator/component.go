package prometheus

import (
	packr "github.com/gobuffalo/packr/v2"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
)

const name = "prometheus-operator"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	Namespace            string   `hcl:"namespace,attr"`
	GrafanaAdminPassword string   `hcl:"grafana_admin_password,attr"`
	EtcdEndpoints        []string `hcl:"etcd_endpoints,optional"`

	PrometheusOperatorNodeSelector map[string]string `hcl:"prometheus_operator_node_selector,optional"`

	PrometheusMetricsRetention string            `hcl:"prometheus_metrics_retention,optional"`
	PrometheusExternalURL      string            `hcl:"prometheus_external_url,optional"`
	PrometheusNodeSelector     map[string]string `hcl:"prometheus_node_selector,optional"`

	AlertManagerRetention    string            `hcl:"alertmanager_retention,optional"`
	AlertManagerExternalURL  string            `hcl:"alertmanager_external_url,optional"`
	AlertManagerConfig       string            `hcl:"alertmanager_config,optional"`
	AlertManagerNodeSelector map[string]string `hcl:"alertmanager_node_selector,optional"`
}

func newComponent() *component {
	defaultAlertManagerConfig := `
  config:
    global:
      resolve_timeout: 5m
    route:
      group_by:
      - job
      group_wait: 30s
      group_interval: 5m
      repeat_interval: 12h
      receiver: 'null'
      routes:
      - match:
          alertname: Watchdog
        receiver: 'null'
    receivers:
    - name: 'null'
`

	return &component{
		PrometheusMetricsRetention: "10d",
		AlertManagerRetention:      "120h",
		AlertManagerConfig:         defaultAlertManagerConfig,
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{
			components.HCLDiagConfigBodyNil,
		}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	box := packr.New(name, "./manifests/")

	helmChart, err := util.LoadChartFromBox(box)
	if err != nil {
		return nil, errors.Wrap(err, "load chart from box")
	}

	releaseOptions := &chartutil.ReleaseOptions{
		Name:      name,
		Namespace: c.Namespace,
		IsInstall: true,
	}

	values, err := util.RenderTemplate(chartValuesTmpl, c)
	if err != nil {
		return nil, errors.Wrap(err, "render chart values template")
	}

	chartConfig := &chart.Config{Raw: values}

	renderedFiles, err := util.RenderChart(helmChart, chartConfig, releaseOptions)
	if err != nil {
		return nil, errors.Wrap(err, "render chart")
	}

	return renderedFiles, nil
}

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}
