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
	"fmt"

	api "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	"github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const (
	// Name represents metrics-server component name as it should be referenced in function calls
	// and in configuration.
	Name = "metrics-server"
)

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

type component struct {
	Namespace string `hcl:"namespace,optional"`
}

// NewConfig returns new metrics-server component configuration with default values set.
//
//nolint:golint
func NewConfig() *component {
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
	helmChart, err := components.Chart(Name)
	if err != nil {
		return nil, fmt.Errorf("retrieving chart from assets: %w", err)
	}

	values, err := template.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering chart values template: %w", err)
	}

	renderedFiles, err := util.RenderChart(helmChart, Name, c.Namespace, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: Name,
		Namespace: k8sutil.Namespace{
			Name: c.Namespace,
		},
		Helm: components.HelmMetadata{
			// metrics-server provides Kubernetes API Resource, so when it is unavailable, it may
			// cause Kubernetes clients to fail creating the client objects, as the client discovery
			// will be returning the following error:
			//
			// failed to create Kubernetes client: unable to retrieve the complete list of server APIs:
			// metrics.k8s.io/v1beta1: the server is currently unable to handle the request
			//
			// Adding wait will ensure, that metrics API is available before proceeding with installation
			// of other components.
			Wait: true,
		},
	}
}

func (c *component) GenerateHelmRelease() (*api.HelmRelease, error) {
	return nil, components.NotImplementedErr
}
