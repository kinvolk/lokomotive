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

package linkerd

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/linkerd/linkerd2/pkg/tls"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/kinvolk/lokomotive/internal"
	internaltemplate "github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

const (
	name           = "experimental-linkerd"
	certCommonName = "identity.linkerd.cluster.local"
)

//nolint:gochecknoinits
func init() {
	components.Register(name, newComponent())
}

type component struct {
	ControllerReplicas int  `hcl:"controller_replicas,optional"`
	EnableMonitoring   bool `hcl:"enable_monitoring,optional"`

	Cert cert
}

type cert struct {
	CA     string
	Cert   string
	Key    string
	Expiry string
}

func newComponent() *component {
	return &component{
		ControllerReplicas: 1,
		EnableMonitoring:   false,
	}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	diagnostics := hcl.Diagnostics{}

	if configBody == nil {
		return hcl.Diagnostics{}
	}

	d := gohcl.DecodeBody(*configBody, evalContext, c)
	if d.HasErrors() {
		return append(diagnostics, d...)
	}

	if c.ControllerReplicas < 1 {
		return hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "'controller_replicas' should be more than or equal to one",
			},
		}
	}

	return diagnostics
}

func (c *component) RenderManifests() (map[string]string, error) {
	// linkerd2 is the name of the upstream chart.
	helmChart, err := components.Chart("linkerd2")
	if err != nil {
		return nil, fmt.Errorf("loading chart from assets: %w", err)
	}

	if c.Cert, err = generateCertificates(); err != nil {
		return nil, err
	}

	values, err := internaltemplate.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering values template failed: %w", err)
	}

	values, err = mergeValuesFiles(values, valuesHA)
	if err != nil {
		return nil, fmt.Errorf("merging values failed: %w", err)
	}

	// Generate YAML for the Linkerd deployment.
	renderedFiles, err := util.RenderChart(helmChart, name, c.Metadata().Namespace.Name, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart failed: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name: name,
		Namespace: k8sutil.Namespace{
			Name: "linkerd",
			Annotations: map[string]string{
				"linkerd.io/inject": "disabled",
			},
			Labels: map[string]string{
				"linkerd.io/is-control-plane":          "true",
				"config.linkerd.io/admission-webhooks": "disabled",
				"linkerd.io/control-plane-ns":          "linkerd",
			},
		},
		Helm: components.HelmMetadata{
			Wait: true,
		},
	}
}

func mergeValuesFiles(dst, src string) (string, error) {
	d, err := chartutil.ReadValues([]byte(dst))
	if err != nil {
		return "", fmt.Errorf("could not read values from destination: %w", err)
	}

	s, err := chartutil.ReadValues([]byte(src))
	if err != nil {
		return "", fmt.Errorf("could not read values from source: %w", err)
	}

	return chartutil.Values(chartutil.CoalesceTables(d, s)).YAML()
}

func generateCertificates() (cert, error) {
	root, err := tls.GenerateRootCAWithDefaults(certCommonName)
	if err != nil {
		return cert{}, fmt.Errorf("generating certificates: %w", err)
	}

	return cert{
		Key:    internal.Indent(root.Cred.EncodePrivateKeyPEM(), 8),
		Cert:   internal.Indent(root.Cred.Crt.EncodeCertificatePEM(), 8),
		CA:     internal.Indent(root.Cred.Crt.EncodeCertificatePEM(), 4),
		Expiry: root.Cred.Crt.Certificate.NotAfter.String(),
	}, nil
}
