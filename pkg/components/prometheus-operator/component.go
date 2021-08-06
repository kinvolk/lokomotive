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
	"net/url"

	helmcontrollerapi "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/types"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"github.com/kinvolk/lokomotive/pkg/version"
)

const (
	// Name represents Prometheus Operator component name as it should be referenced in function calls
	// and in configuration.
	Name = "prometheus-operator"
)

// Monitor holds information about which Kubernetes components should be monitored with the default Prometheus instance.
type Monitor struct {
	Etcd                  bool `hcl:"etcd,optional"`
	KubeControllerManager bool `hcl:"kube_controller_manager,optional"`
	KubeScheduler         bool `hcl:"kube_scheduler,optional"`
	KubeProxy             bool `hcl:"kube_proxy,optional"`
	Kubelet               bool `hcl:"kubelet,optional"`
}

// CoreDNS holds information about how CoreDNS should be scraped.
type CoreDNS struct {
	Selector map[string]string `hcl:"selector,optional"`
}

// Grafana object collects sub component grafana related information.
type Grafana struct {
	AdminPassword string            `hcl:"admin_password,optional"`
	SecretEnv     map[string]string `hcl:"secret_env,optional"`
	Ingress       *types.Ingress    `hcl:"ingress,block"`
}

// Operator object collects sub component Prometheus operator related information.
type Operator struct {
	AdmissionWebhookTolerations    []util.Toleration `hcl:"admission_webhook_tolerations,block"`
	AdmissionWebhookTolerationsRaw string
	NodeSelector                   map[string]string `hcl:"node_selector,optional"`
	Tolerations                    []util.Toleration `hcl:"tolerations,block"`
	TolerationsRaw                 string
}

// Prometheus object collects sub component Prometheus related information.
type Prometheus struct {
	MetricsRetention            string            `hcl:"metrics_retention,optional"`
	NodeSelector                map[string]string `hcl:"node_selector,optional"`
	StorageSize                 string            `hcl:"storage_size,optional"`
	WatchLabeledServiceMonitors bool              `hcl:"watch_labeled_service_monitors,optional"`
	WatchLabeledPrometheusRules bool              `hcl:"watch_labeled_prometheus_rules,optional"`
	Ingress                     *types.Ingress    `hcl:"ingress,block"`
	ExternalLabels              map[string]string `hcl:"external_labels,optional"`
	ExternalURL                 string            `hcl:"external_url,optional"`
	Tolerations                 []util.Toleration `hcl:"tolerations,block"`
	TolerationsRaw              string
}

// AlertManager object collects sub component AlertManager related information.
type AlertManager struct {
	Config         string            `hcl:"config,optional"`
	ExternalURL    string            `hcl:"external_url,optional"`
	NodeSelector   map[string]string `hcl:"node_selector,optional"`
	Retention      string            `hcl:"retention,optional"`
	StorageSize    string            `hcl:"storage_size,optional"`
	Tolerations    []util.Toleration `hcl:"tolerations,block"`
	TolerationsRaw string
}

type component struct {
	Grafana *Grafana `hcl:"grafana,block"`

	Namespace string `hcl:"namespace,optional"`

	Operator *Operator `hcl:"operator,block"`

	Prometheus *Prometheus `hcl:"prometheus,block"`

	AlertManager *AlertManager `hcl:"alertmanager,block"`

	StorageClass string `hcl:"storage_class,optional"`

	DisableWebhooks bool `hcl:"disable_webhooks,optional"`

	Monitor *Monitor `hcl:"monitor,block"`
	CoreDNS *CoreDNS `hcl:"coredns,block"`
}

// NewConfig returns new Prometheus Operator component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
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
		Prometheus: &Prometheus{
			MetricsRetention:            "10d",
			StorageSize:                 "50Gi",
			WatchLabeledServiceMonitors: true,
			WatchLabeledPrometheusRules: true,
		},
		AlertManager: &AlertManager{
			Retention:   "120h",
			Config:      defaultAlertManagerConfig,
			StorageSize: "50Gi",
		},
		Namespace: "monitoring",
		Monitor: &Monitor{
			Etcd:                  true,
			KubeControllerManager: true,
			KubeScheduler:         true,
			KubeProxy:             true,
			Kubelet:               true,
		},
		CoreDNS: &CoreDNS{
			Selector: map[string]string{
				"k8s-app": "coredns",
				"tier":    "control-plane",
			},
		},
		Grafana: &Grafana{
			// This is done in order to make sure that Grafana admin user password is generated if
			// user does not provide one.
			// If this block is not provided here and user also does not specify any grafana related
			// config then admin password is set to "prom-operator".
			// See: https://github.com/kinvolk/lokomotive/pull/507#issuecomment-636049574
			AdminPassword: "",
		},
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		// return empty struct instead of hcl.Diagnostics{components.HCLDiagConfigBodyNil}
		// since all the component values are optional
		return hcl.Diagnostics{}
	}

	if err := gohcl.DecodeBody(*configBody, evalContext, c); err != nil {
		return err
	}

	if c.Grafana != nil && c.Grafana.Ingress != nil {
		c.Grafana.Ingress.SetDefaults()
	}

	if c.Prometheus != nil && c.Prometheus.Ingress != nil {
		c.Prometheus.Ingress.SetDefaults()
	}

	// If user has provided both `prometheus.ingress.host` and `prometheus.external_url`, the
	// hostnames should be the same.
	if c.Prometheus != nil && c.Prometheus.ExternalURL != "" && c.Prometheus.Ingress != nil {
		exu, err := url.Parse(c.Prometheus.ExternalURL)
		if err != nil {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary: fmt.Sprintf("parsing 'prometheus.external_url' with value %q as URL: %v",
						c.Prometheus.ExternalURL, err),
				},
			}
		}

		if exu.Host != c.Prometheus.Ingress.Host {
			return hcl.Diagnostics{
				{
					Severity: hcl.DiagError,
					Summary:  "'prometheus.external_url' and 'prometheus.ingress.host' do not match",
				},
			}
		}
	}

	return nil
}

func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
	}

	c.Prometheus.TolerationsRaw, err = util.RenderTolerations(c.Prometheus.Tolerations)
	if err != nil {
		return nil, fmt.Errorf("rendering prometheus tolerations: %w", err)
	}

	if c.Operator != nil {
		c.Operator.TolerationsRaw, err = util.RenderTolerations(c.Operator.Tolerations)
		if err != nil {
			return nil, fmt.Errorf("rendering operator tolerations: %w", err)
		}

		c.Operator.AdmissionWebhookTolerationsRaw, err = util.RenderTolerations(c.Operator.AdmissionWebhookTolerations) //nolint:lll
		if err != nil {
			return nil, fmt.Errorf("rendering operator admission webhook tolerations: %w", err)
		}
	}

	c.AlertManager.TolerationsRaw, err = util.RenderTolerations(c.AlertManager.Tolerations)
	if err != nil {
		return nil, fmt.Errorf("rendering alertmanager tolerations: %w", err)
	}

	values, err := template.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering chart values template: %w", err)
	}

	renderedFiles, err := util.RenderChart(helmChart, Name, c.Namespace, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: Name,
		Namespace: k8sutil.Namespace{
			Name: c.Namespace,
		},
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

func (c *component) GenerateHelmRelease() (*helmcontrollerapi.HelmRelease, error) {
	valuesYaml, err := template.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering chart values template: %w", err)
	}

	values, err := k8syaml.YAMLToJSON([]byte(valuesYaml))
	if err != nil {
		return nil, fmt.Errorf("converting YAML to JSON: %w", err)
	}

	return &helmcontrollerapi.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      Name,
			Namespace: "flux-system",
		},
		Spec: helmcontrollerapi.HelmReleaseSpec{
			Chart: helmcontrollerapi.HelmChartTemplate{
				Spec: helmcontrollerapi.HelmChartTemplateSpec{
					Chart: components.ComponentsPath + Name,
					SourceRef: helmcontrollerapi.CrossNamespaceObjectReference{
						Kind: "GitRepository",
						Name: "lokomotive-" + version.Version,
					},
				},
			},
			ReleaseName: Name,
			Install: &helmcontrollerapi.Install{
				CRDs:            helmcontrollerapi.CreateReplace,
				CreateNamespace: true,
				Remediation: &helmcontrollerapi.InstallRemediation{
					Retries: -1,
				},
			},
			Upgrade: &helmcontrollerapi.Upgrade{
				CRDs: helmcontrollerapi.CreateReplace,
			},
			Interval:        components.FluxInstallInterval,
			Timeout:         &components.FluxInstallTimeout,
			TargetNamespace: c.Namespace,
			Values: &apiextensionsv1.JSON{
				Raw: values,
			},
		},
	}, nil
}
