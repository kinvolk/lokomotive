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

package metricsserver

import (
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
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
	p := filepath.Join(assets.ComponentsSource, name)
	helmChart, err := util.LoadChartFromAssets(p)
	if err != nil {
		return nil, errors.Wrap(err, "load chart from assets")
	}

	values, err := template.Render(chartValuesTmpl, c)
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
		Name:      name,
		Namespace: c.Namespace,
	}
}
