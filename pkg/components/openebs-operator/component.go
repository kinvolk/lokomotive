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
