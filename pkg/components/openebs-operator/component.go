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

package openebsoperator

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

const name = "openebs-operator"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	NDMSelectorLabel string `hcl:"ndm_selector_label,optional"`
	NDMSelectorValue string `hcl:"ndm_selector_value,optional"`
}

const chartValuesTmpl = `
rbac:
  pspEnabled: true

{{- if and .NDMSelectorLabel .NDMSelectorValue }}
apiserver:
  nodeSelector:
    "{{ .NDMSelectorLabel }}": "{{ .NDMSelectorValue }}"
{{- end }}

{{- if and .NDMSelectorLabel .NDMSelectorValue }}
provisioner:
  nodeSelector:
    "{{ .NDMSelectorLabel }}": "{{ .NDMSelectorValue }}"
{{- end }}

{{- if and .NDMSelectorLabel .NDMSelectorValue }}
localprovisioner:
  nodeSelector:
    "{{ .NDMSelectorLabel }}": "{{ .NDMSelectorValue }}"
{{- end }}

{{- if and .NDMSelectorLabel .NDMSelectorValue }}
snapshotOperator:
  nodeSelector:
    "{{ .NDMSelectorLabel }}": "{{ .NDMSelectorValue }}"
{{- end }}

{{- if and .NDMSelectorLabel .NDMSelectorValue }}
ndm:
  nodeSelector:
    "{{ .NDMSelectorLabel }}": "{{ .NDMSelectorValue }}"
{{- end }}

{{- if and .NDMSelectorLabel .NDMSelectorValue }}
ndmOperator:
  nodeSelector:
    "{{ .NDMSelectorLabel }}": "{{ .NDMSelectorValue }}"
{{- end }}

{{- if and .NDMSelectorLabel .NDMSelectorValue }}
webhook:
  nodeSelector:
    "{{ .NDMSelectorLabel }}": "{{ .NDMSelectorValue }}"
{{- end }}
`

func newComponent() *component {
	return &component{}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		// return empty struct instead of hcl.Diagnostics{components.HCLDiagConfigBodyNil}
		// since all the component values are optional
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	p := filepath.Join(assets.ComponentsSource, name)
	helmChart, err := util.LoadChartFromAssets(p)
	if err != nil {
		return nil, fmt.Errorf("load chart from assets: %w", err)
	}

	values, err := template.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("render chart values template: %w", err)
	}

	renderedFiles, err := util.RenderChart(helmChart, name, c.Metadata().Namespace, values)
	if err != nil {
		return nil, fmt.Errorf("render chart: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name:      name,
		Namespace: "openebs",
	}
}
