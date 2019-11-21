package metricsserver

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

const name = "metrics-server"

// * --kubelet-preferred-address-types=InternalIP to be able to properly the kubelet.
//  I am not sure why this option is needed, but tried the alternatives
//  for this and didn't work
//  And this option does the trick for others
//  people too: https://github.com/kubernetes-incubator/metrics-server/issues/237#issuecomment-504427772
//
// * Use --kubelet-insecure-tls for the self-signed kubelets certificates
//   When we are able to remove the option above, we may be able to use
//   --kubelet-certificate-authority but, meanwhile, this is needed to
//   communicate with kubelets.
//   Something like: --kubelet-certificate-authority=/run/secrets/kubernetes.io/serviceaccount/ca.crt
//   But this doesn't work out of the box, it seems no permissions to open the ca.crt file.
//   We should investigate when we can change to not use the InternalIP
//   or use a cert that signs also the IP of the kubelet
const chartValuesTmpl = `
args:
- --kubelet-insecure-tls=true
- --kubelet-preferred-address-types=InternalIP
`

func init() {
	components.Register(name, newComponent())
}

type component struct {
	Namespace string `hcl:"namespace,optional"`
}

func newComponent() *component {
	return &component{
		Namespace: "kube-system",
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}

	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := util.LoadChartFromAssets(fmt.Sprintf("/components/%s", name))
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

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Namespace: c.Namespace,
	}
}
