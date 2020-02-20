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

package openebsoperator

import (
	"bytes"
	"html/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/components"
)

const name = "openebs-operator"

func init() {
	components.Register(name, newComponent())
}

type component struct {
	NDMSelectorLabel string `hcl:"ndm_selector_label,optional"`
	NDMSelectorValue string `hcl:"ndm_selector_value,optional"`
}

func newComponent() *component {
	return &component{}
}

func (c *component) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		// return empty struct instead of hcl.Diagnostics{components.HCLDiagConfigBodyNil}
		// since all the component values are optional
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, c)
}

func (c *component) RenderManifests() (map[string]string, error) {
	tmpl, err := template.New("installer").Parse(operatorInstallerTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "parse template failed")
	}
	var installerBuf bytes.Buffer
	if err := tmpl.Execute(&installerBuf, c); err != nil {
		return nil, errors.Wrap(err, "execute template failed")
	}
	return map[string]string{
		"openebs-operator.yml": installerBuf.String(),
	}, nil
}

func (c *component) Metadata() components.Metadata {
	return components.Metadata{
		Namespace: "openebs",
		Helm:      &components.HelmMetadata{},
	}
}
