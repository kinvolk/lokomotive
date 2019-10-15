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

// init registers local as a backend
func init() {
	backend.Register("local", NewLocalBackend())
}

//Loadconfig loads configuration for local backend
func (l *local) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, l)
}

func NewLocalBackend() *local {
	return &local{}
}

// Render renders the go template with local backend configuration
func (l *local) Render() (string, error) {

	return util.RenderTemplate(backendConfigTmpl, l)
}
