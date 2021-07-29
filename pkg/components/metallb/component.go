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

package metallb

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
	// Name represents MetalLB component name as it should be referenced in function calls
	// and in configuration.
	Name = "metallb"
)

type component struct {
	AddressPools            map[string][]string `hcl:"address_pools"`
	ControllerNodeSelectors map[string]string   `hcl:"controller_node_selectors,optional"`
	SpeakerNodeSelectors    map[string]string   `hcl:"speaker_node_selectors,optional"`
	ControllerTolerations   []util.Toleration   `hcl:"controller_toleration,block"`
	SpeakerTolerations      []util.Toleration   `hcl:"speaker_toleration,block"`
	ServiceMonitor          bool                `hcl:"service_monitor,optional"`

	ControllerTolerationsJSON string
	SpeakerTolerationsJSON    string
}

// NewConfig returns new MetalLB component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
	return &component{}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
	}

	// Here are `nodeSelectors` and `tolerations` that are set by upstream. To make sure that we
	// don't miss them out we set them manually here. We cannot make these changes in the template
	// because we have parameterized these fields.
	if c.SpeakerNodeSelectors == nil {
		c.SpeakerNodeSelectors = map[string]string{}
	}
	// MetalLB only supports Linux, so force this selector, even if it's already specified by the
	// user.
	c.SpeakerNodeSelectors["beta.kubernetes.io/os"] = "linux"

	if c.ControllerNodeSelectors == nil {
		c.ControllerNodeSelectors = map[string]string{}
	}
	c.ControllerNodeSelectors["beta.kubernetes.io/os"] = "linux"
	c.ControllerNodeSelectors["node.kubernetes.io/master"] = ""

	c.ControllerTolerations = append(c.ControllerTolerations, util.Toleration{
		Effect: "NoSchedule",
		Key:    "node-role.kubernetes.io/master",
	})

	t, err := util.RenderTolerations(c.SpeakerTolerations)
	if err != nil {
		return nil, fmt.Errorf("marshaling speaker tolerations: %w", err)
	}
	c.SpeakerTolerationsJSON = t

	t, err = util.RenderTolerations(c.ControllerTolerations)
	if err != nil {
		return nil, fmt.Errorf("rendering controller tolerations: %w", err)
	}
	c.ControllerTolerationsJSON = t

	values, err := template.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering values template failed: %w", err)
	}

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
			Name: "metallb-system",
		},
	}
}
