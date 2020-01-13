package certmanager

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
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
	helmChart, err := util.LoadChartFromAssets(fmt.Sprintf("/components/%s/manifests", name))
	if err != nil {
		return nil, errors.Wrap(err, "load chart from assets")
	}

	values, err := util.RenderTemplate(chartValuesTmpl, c)
	if err != nil {
		return nil, errors.Wrap(err, "render chart values template")
	}

	renderedFiles, err := util.RenderChart(helmChart, name, c.Namespace, values)
	if err != nil {
		return nil, errors.Wrap(err, "render chart")
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Namespace: c.Namespace,
	}
}
