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

	api "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const (
	// Name represents OpenEBS Operator component name as it should be referenced in function calls
	// and in configuration.
	Name = "openebs-operator"
)

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

webhook:
  # OpenEBS by default ships the failurePolicy as 'Fail', however this causes sometimes to unexpectedly
  # hit a known bug in OpenEBS https://github.com/openebs/openebs/issues/3046.
  # One of the workarounds suggested is to change the failurePolicy to "Ignore".
  failurePolicy: 'Ignore'
{{- if and .NDMSelectorLabel .NDMSelectorValue }}
  nodeSelector:
    "{{ .NDMSelectorLabel }}": "{{ .NDMSelectorValue }}"
{{- end }}
`

// NewConfig returns new OpenEBS Operator component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
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
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
	}

	values, err := template.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("render chart values template: %w", err)
	}

	renderedFiles, err := util.RenderChart(helmChart, Name, c.Metadata().Namespace.Name, values)
	if err != nil {
		return nil, fmt.Errorf("render chart: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: Name,
		Namespace: k8sutil.Namespace{
			Name: "openebs",
		},
	}
}

func (c *component) GenerateHelmRelease() (*api.HelmRelease, error) {
	return nil, components.NotImplementedErr
}
