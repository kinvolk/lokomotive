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

package contour

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	internaltemplate "github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

const (
	name                    = "contour"
	serviceTypeNodePort     = "NodePort"
	serviceTypeLoadBalancer = "LoadBalancer"
)

func init() {
	components.Register(name, newComponent())
}

// This annotation is added to Envoy service.
type component struct {
	EnableMonitoring bool                `hcl:"enable_monitoring,optional"`
	NodeAffinity     []util.NodeAffinity `hcl:"node_affinity,block"`
	NodeAffinityRaw  string
	ServiceType      string            `hcl:"service_type,optional"`
	Tolerations      []util.Toleration `hcl:"toleration,block"`
	TolerationsRaw   string
}

func newComponent() *component {
	return &component{
		ServiceType: serviceTypeLoadBalancer,
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diagnostics := hcl.Diagnostics{}

	if configBody == nil {
		return hcl.Diagnostics{
			components.HCLDiagConfigBodyNil,
		}
	}

	d := gohcl.DecodeBody(*configBody, evalContext, c)
	if d.HasErrors() {
		diagnostics = append(diagnostics, d...)
		return diagnostics
	}

	// Validate service type.
	if c.ServiceType != serviceTypeNodePort && c.ServiceType != serviceTypeLoadBalancer {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Unknown service type %q", c.ServiceType),
		})
	}

	return diagnostics
}

func (c *component) RenderManifests() (map[string]string, error) {
	p := filepath.Join(assets.ComponentsSource, name)
	helmChart, err := util.LoadChartFromAssets(p)
	if err != nil {
		return nil, fmt.Errorf("load chart from assets: %w", err)
	}

	c.TolerationsRaw, err = util.RenderTolerations(c.Tolerations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal operator tolerations: %w", err)
	}

	c.NodeAffinityRaw, err = util.RenderNodeAffinity(c.NodeAffinity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal node affinity: %w", err)
	}

	values, err := internaltemplate.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering values template failed: %w", err)
	}

	// Generate YAML for the Contour deployment.
	renderedFiles, err := util.RenderChart(helmChart, name, c.Metadata().Namespace, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart failed: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name:      name,
		Namespace: "projectcontour",
	}
}
