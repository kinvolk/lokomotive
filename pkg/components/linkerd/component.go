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
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/linkerd/linkerd2/pkg/tls"

	internaltemplate "github.com/kinvolk/lokomotive/internal/template"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

const (
	name = "linkerd"
)

// nolint:gochecknoinits
func init() {
	components.Register(name, newComponent())
}

type component struct {
	CA                 string
	Cert               string
	Key                string
	Expiry             string
	ControllerReplicas int  `hcl:"controller_replicas,optional"`
	EnableMonitoring   bool `hcl:"enable_monitoring,optional"`
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
		return hcl.Diagnostics{
			components.HCLDiagConfigBodyNil,
		}
	}

	d := gohcl.DecodeBody(*configBody, evalContext, c)
	if d.HasErrors() {
		diagnostics = append(diagnostics, d...)
		return diagnostics
	}

	return diagnostics
}

func (c *component) RenderManifests() (map[string]string, error) {
	helmChart, err := util.LoadChartFromAssets("/components/linkerd2")
	if err != nil {
		return nil, fmt.Errorf("load chart from assets: %w", err)
	}

	// Generate certs
	root, err := tls.GenerateRootCAWithDefaults("identity.linkerd.cluster.local")
	if err != nil {
		return nil, fmt.Errorf("could not generate cert: %w", err)
	}

	c.Key = indent(root.Cred.EncodePrivateKeyPEM(), 8)
	c.Cert = indent(root.Cred.Crt.EncodeCertificatePEM(), 8)
	c.CA = indent(root.Cred.Crt.EncodeCertificatePEM(), 4)
	c.Expiry = root.Cred.Crt.Certificate.NotAfter.String()

	values, err := internaltemplate.Render(chartValuesTmpl, c)
	if err != nil {
		return nil, fmt.Errorf("rendering values template failed: %w", err)
	}

	// Generate YAML for the istio deployment.
	renderedFiles, err := util.RenderChart(helmChart, name, c.Metadata().Namespace, values)
	if err != nil {
		return nil, fmt.Errorf("rendering chart failed: %w", err)
	}

	return renderedFiles, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Name:      name,
		Namespace: "linkerd",
		Helm: components.HelmMetadata{
			Wait: true,
		},
	}
}

func indent(data string, indent int) string {
	lines := strings.Split(data, "\n")

	var gap string
	for i := 0; i < indent; i++ {
		gap += " "
	}

	for ind := range lines {
		lines[ind] = gap + lines[ind]
	}

	return strings.Join(lines, "\n")
}
