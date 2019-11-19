package certmanager

import (
	"fmt"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

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

	releaseOptions := &chartutil.ReleaseOptions{
		Name:      name,
		Namespace: c.Namespace,
		IsInstall: true,
	}

	values, err := util.RenderTemplate(chartValuesTmpl, c)
	if err != nil {
		return nil, errors.Wrap(err, "render chart values template")
	}

	chartConfig := &chart.Config{Raw: values}

	renderedFiles, err := util.RenderChart(helmChart, chartConfig, releaseOptions)
	if err != nil {
		return nil, errors.Wrap(err, "render chart")
	}

	return renderedFiles, nil
}
