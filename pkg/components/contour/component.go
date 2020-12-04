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

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	internaltemplate "github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const (
	// Name represents Contour component name as it should be referenced in function calls
	// and in configuration.
	Name = "contour"

	serviceTypeNodePort     = "NodePort"
	serviceTypeLoadBalancer = "LoadBalancer"
)

// This annotation is added to Envoy service.
type component struct {
	EnableMonitoring bool                `hcl:"enable_monitoring,optional"`
	NodeAffinity     []util.NodeAffinity `hcl:"node_affinity,block"`
	NodeAffinityRaw  string
	ServiceType      string            `hcl:"service_type,optional"`
	Tolerations      []util.Toleration `hcl:"toleration,block"`
	TolerationsRaw   string
	Envoy            *envoy `hcl:"envoy,block"`
}

type envoy struct {
	MetricsScrapeInterval string `hcl:"metrics_scrape_interval,optional"`
}

// NewConfig returns new Contour component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
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
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
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
	renderedFiles, err := util.RenderChart(helmChart, Name, c.Metadata().Namespace.Name, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart failed: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: Name,
		Namespace: k8sutil.Namespace{
			Name: "projectcontour",
		},
	}
}
