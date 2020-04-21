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
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/pkg/errors"
)

const name = "rook"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	Namespace                string              `hcl:"namespace,optional"`
	NodeSelectors            []util.NodeSelector `hcl:"node_selector,block"`
	Tolerations              []util.Toleration   `hcl:"toleration,block"`
	TolerationsRaw           string
	AgentTolerationKey       string `hcl:"agent_toleration_key,optional"`
	AgentTolerationEffect    string `hcl:"agent_toleration_effect,optional"`
	DiscoverTolerationKey    string `hcl:"discover_toleration_key,optional"`
	DiscoverTolerationEffect string `hcl:"discover_toleration_effect,optional"`
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
	// Generate YAML for namespace.
	namespaceStr, err := util.RenderTemplate(namespace, c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render template")
	}

	// Generate YAML for RBAC resources.
	rbacStr, err := util.RenderTemplate(rbac, c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render template")
	}

	// Generate YAML for the Rook operator deployment.
	c.TolerationsRaw, err = util.RenderTolerations(c.Tolerations)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal operator tolerations")
	}

	deploymentStr, err := util.RenderTemplate(deploymentOperator, c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render template")
	}

	return map[string]string{
		"namespace.yaml":           namespaceStr,
		"crds.yaml":                crds,
		"rbac.yaml":                rbacStr,
		"deployment_operator.yaml": deploymentStr,
	}, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name:      name,
		Namespace: c.Namespace,
	}
}
