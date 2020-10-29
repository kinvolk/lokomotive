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

package certmanager

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const (
	// Name represents cert-manager component name as it should be referenced in function calls
	// and in configuration.
	Name = "cert-manager"
)

func init() {
	components.Register(Name, newComponent())
}

type component struct {
	Email          string `hcl:"email,attr"`
	Namespace      string `hcl:"namespace,optional"`
	Webhooks       bool   `hcl:"webhooks,optional"`
	ServiceMonitor bool   `hcl:"service_monitor,optional"`
}

func newComponent() *component {
	return &component{
		Namespace:      "cert-manager",
		Webhooks:       true,
		ServiceMonitor: false,
	}
}

const chartValuesTmpl = `
email: {{.Email}}
webhook:
  enabled: {{.Webhooks}}
{{ if .ServiceMonitor }}
prometheus:
  servicemonitor:
    enabled: true
    labels:
      release: prometheus-operator
{{ end }}
global:
  podSecurityPolicy:
    enabled: true
    useAppArmor: false
installCRDs: true
`

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{
			components.HCLDiagConfigBodyNil,
		}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
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
			Labels: map[string]string{
				"certmanager.k8s.io/disable-validation": "true",
			},
		},
		Helm: components.HelmMetadata{
			// Cert-manager registers admission webhooks, so we should wait for the webhook to
			// become ready before proceeding with installing other components, as it may fail.
			// If webhooks are registered with 'failurePolicy: Fail', then kube-apiserver will reject
			// creating objects requiring the webhook until the webhook itself becomes ready. So if the
			// next component after cert-manager creates e.g. an Ingress object and the webhook is not ready
			// yet, it will fail. 'Wait' serializes the process, so Helm will only return without error, when
			// all deployments included in the component, including the webhook, become ready.
			Wait: true,
		},
	}
}
