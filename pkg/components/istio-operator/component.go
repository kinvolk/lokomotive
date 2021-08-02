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

package istiooperator

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	internaltemplate "github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const (
	// Name represents Istio Operator component name as it should be referenced in function calls
	// and in configuration.
	Name = "experimental-istio-operator"

	namespace = "istio-operator"
)

type component struct {
	Profile          string `hcl:"profile,optional"`
	EnableMonitoring bool   `hcl:"enable_monitoring,optional"`
}

// NewConfig returns new Istio Operator component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
	return &component{
		Profile:          "minimal",
		EnableMonitoring: false,
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diagnostics := hcl.Diagnostics{}

	if configBody == nil {
		return hcl.Diagnostics{}
	}

	d := gohcl.DecodeBody(*configBody, evalContext, c)
	if d.HasErrors() {
		return append(diagnostics, d...)
	}

	return diagnostics
}

func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := components.Chart("istio-operator")
	if err != nil {
		return nil, fmt.Errorf("loading chart from assets: %w", err)
	}

	values, err := internaltemplate.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering values template failed: %w", err)
	}

	// Generate YAML for the istio deployment.
	renderedFiles, err := util.RenderChart(helmChart, Name, namespace, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart failed: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: Name,
		Namespace: k8sutil.Namespace{
			Name: namespace,
			Labels: map[string]string{
				"istio-operator-managed": "Reconcile",
				"istio-injection":        "disabled",
			},
		},
	}
}
