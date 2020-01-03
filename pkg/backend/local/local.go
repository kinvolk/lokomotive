package local

import (
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"

	"github.com/kinvolk/lokoctl/pkg/backend"
	"github.com/kinvolk/lokoctl/pkg/components/util"
)

type local struct {
	Path string `hcl:"path,optional"`
}

// init registers local as a backend.
func init() {
	backend.Register("local", NewLocalBackend())
}

// LoadConfig loads the configuration for the local backend.
func (l *local) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, l)
}

func NewLocalBackend() *local {
	return &local{}
}

// Render renders the Go template with local backend configuration.
func (l *local) Render() (string, error) {
	return util.RenderTemplate(backendConfigTmpl, l)
}

// Validate validates the local backend configuration.
func (l *local) Validate() error {
	return nil
}
