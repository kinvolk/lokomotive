package rook

import (
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
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
		Namespace: c.Namespace,
	}
}
