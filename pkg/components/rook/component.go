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

package rook

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

const name = "rook"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	Namespace                string            `hcl:"namespace,optional"`
	NodeSelector             util.NodeSelector `hcl:"node_selector,optional"`
	NodeSelectorRaw          string
	RookNodeAffinity         string
	Tolerations              []util.Toleration `hcl:"toleration,block"`
	TolerationsRaw           string
	AgentTolerationKey       string `hcl:"agent_toleration_key,optional"`
	AgentTolerationEffect    string `hcl:"agent_toleration_effect,optional"`
	DiscoverTolerationKey    string `hcl:"discover_toleration_key,optional"`
	DiscoverTolerationEffect string `hcl:"discover_toleration_effect,optional"`
	EnableMonitoring         bool   `hcl:"enable_monitoring,optional"`
}

func newComponent() *component {
	return &component{
		Namespace: "rook",
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
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

	c.TolerationsRaw, err = util.RenderTolerations(c.Tolerations)
	if err != nil {
		return nil, fmt.Errorf("rendering tolerations failed: %w", err)
	}

	c.NodeSelectorRaw, err = c.NodeSelector.Render()
	if err != nil {
		return nil, fmt.Errorf("rendering node selector failed: %w", err)
	}

	c.RookNodeAffinity = convertNodeSelector(c.NodeSelector)

	values, err := template.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering values template failed: %w", err)
	}

	// Generate YAML for the Rook operator deployment.
	renderedFiles, err := util.RenderChart(helmChart, name, c.Metadata().Namespace, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart failed: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name:      name,
		Namespace: c.Namespace,
	}
}

// convertNodeSelector converts the key value pair in the map to the format:
// key1=value1; key2=value2;
func convertNodeSelector(m map[string]string) string {
	var ret string

	for k, v := range m {
		ret += fmt.Sprintf("%s=%s; ", k, v)
	}

	return ret
}
