package ingressnginx

import (
	packr "github.com/gobuffalo/packr/v2"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
)

const name = "ingress-nginx"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	Namespace             string `hcl:"namespace,optional"`
	ServiceType           string `hcl:"service_type,optional"`
	InstallMode           string `hcl:"install_mode,optional"`
	ExternalTrafficPolicy string `hcl:"external_traffic_policy,optional"`
}

func newComponent() *component {
	return &component{
		Namespace:             "",
		ServiceType:           "ClusterIP",
		InstallMode:           "deployment",
		ExternalTrafficPolicy: "cluster",
	}
}

const chartValuesTmpl = `
namespace: {{.Namespace}}
serviceType: {{.ServiceType}}
installMode: {{.InstallMode}}
externalTrafficPolicy: {{.ExternalTrafficPolicy}}
`

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	if err := gohcl.DecodeBody(*configBody, evalContext, c); err != nil {
		return err
	}
	if c.InstallMode != "deployment" && c.InstallMode != "daemonset" {
		err := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "install_mode must be either 'deployment' or 'daemonset'",
			Detail:   "Make sure to set install_mode to either 'deployment' or 'daemonset' in lowercase",
		}
		return hcl.Diagnostics{err}
	}
	if c.ExternalTrafficPolicy != "cluster" && c.ExternalTrafficPolicy != "local" {
		err := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "external_traffic_policy must be either 'cluster' or 'local'",
			Detail:   "Make sure to set external_traffic_policy to either 'cluster' or 'local' in lowercase",
		}
		return hcl.Diagnostics{err}
	}
	return nil
}

func (c *component) RenderManifests() (map[string]string, error) {
	box := packr.New(name, "./manifests/")

	helmChart, err := util.LoadChartFromBox(box)
	if err != nil {
		return nil, errors.Wrap(err, "load chart from box")
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

func (c *component) Install(kubeconfig string) error {
	return util.Install(c, kubeconfig)
}
