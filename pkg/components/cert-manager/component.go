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
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/helm"
)

const name = "cert-manager"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	Email     string `hcl:"email,attr"`
	Namespace string `hcl:"namespace,optional"`
	Webhooks  bool   `hcl:"webhooks,optional"`
}

func newComponent() *component {
	return &component{
		Namespace: "cert-manager",
		Webhooks:  true,
	}
}

const chartValuesTmpl = `
email: {{.Email}}
webhook:
  enabled: {{.Webhooks}}
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
	helmChart, err := helm.LoadChartFromAssets(fmt.Sprintf("/components/%s/manifests", name))
	if err != nil {
		return nil, errors.Wrap(err, "load chart from assets")
	}

	values, err := util.RenderTemplate(chartValuesTmpl, c)
	if err != nil {
		return nil, errors.Wrap(err, "render chart values template")
	}

	renderedFiles, err := helm.RenderChart(helmChart, name, c.Namespace, values)
	if err != nil {
		return nil, errors.Wrap(err, "render chart")
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Namespace: c.Namespace,
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
