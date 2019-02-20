package certmanager

import (
	"time"

	packr "github.com/gobuffalo/packr/v2"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
	"github.com/kinvolk/lokoctl/pkg/k8sutil"
)

const name = "cert-manager"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	Email     string  `hcl:"email,attr"`
	Namespace *string `hcl:"namespace,attr"`
}

func newComponent() *component {
	defaultNamespace := ""
	return &component{
		Namespace: &defaultNamespace,
	}
}

const chartValuesTmpl = `
namespace: {{.Namespace}}
email: {{.Email}}
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
	box := packr.New(name, "./manifests/")

	helmChart, err := util.LoadChartFromBox(box)
	if err != nil {
		return nil, errors.Wrap(err, "load chart from box")
	}

	releaseOptions := &chartutil.ReleaseOptions{
		Name:      name,
		Namespace: *c.Namespace,
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

func (c *component) Install(kubeconfig string) error {
	renderedFiles, err := c.RenderManifests()
	if err != nil {
		return err
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{},
	)

	return k8sutil.CreateAssets(clientConfig, renderedFiles, 1*time.Minute)
}
