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

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	utilpkg "github.com/kinvolk/lokomotive/pkg/util"
)

const name = "prometheus-operator"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	GrafanaAdminPassword string   `hcl:"grafana_admin_password,attr"`
	Namespace            string   `hcl:"namespace,optional"`
	EtcdEndpoints        []string `hcl:"etcd_endpoints,optional"`

	PrometheusOperatorNodeSelector map[string]string `hcl:"prometheus_operator_node_selector,optional"`

	PrometheusMetricsRetention  string            `hcl:"prometheus_metrics_retention,optional"`
	PrometheusExternalURL       string            `hcl:"prometheus_external_url,optional"`
	PrometheusNodeSelector      map[string]string `hcl:"prometheus_node_selector,optional"`
	WatchLabeledServiceMonitors bool              `hcl:"watch_labeled_service_monitors,optional"`
	WatchLabeledPrometheusRules bool              `hcl:"watch_labeled_prometheus_rules,optional"`

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
		PrometheusMetricsRetention:  "10d",
		AlertManagerRetention:       "120h",
		AlertManagerConfig:          defaultAlertManagerConfig,
		Namespace:                   "monitoring",
		WatchLabeledServiceMonitors: true,
		WatchLabeledPrometheusRules: true,
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
	helmChart, err := util.LoadChartFromAssets(fmt.Sprintf("/components/%s/manifests", name))
	if err != nil {
		return nil, errors.Wrap(err, "load chart from assets")
	}

	values, err := utilpkg.RenderTemplate(chartValuesTmpl, c)
	if err != nil {
		return nil, errors.Wrap(err, "render chart values template")
	}

	renderedFiles, err := util.RenderChart(helmChart, name, c.Namespace, values)
	if err != nil {
		return nil, errors.Wrap(err, "render chart")
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name:      name,
		Namespace: c.Namespace,
		Helm: components.HelmMetadata{
			// Prometheus-operator registers admission webhooks, so we should wait for the webhook to
			// become ready before proceeding with installing other components, as it may fail.
			// If webhooks are registered with 'failurePolicy: Fail', then kube-apiserver will reject
			// creating objects requiring the webhook until the webhook itself becomes ready. So if the
			// next component after prometheus-operator creates e.g. a Prometheus CR and the webhook is not ready
			// yet, it will fail. 'Wait' serializes the process, so Helm will only return without error, when
			// all deployments included in the component, including the webhook, become ready.
			Wait: true,
		},
	}
}
