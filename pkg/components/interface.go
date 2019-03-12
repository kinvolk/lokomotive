package components

import (
	"github.com/hashicorp/hcl2/hcl"
)

type Component interface {
	LoadConfig(*hcl.Body, *hcl.EvalContext) hcl.Diagnostics
	RenderManifests() (map[string]string, error)
	Install(kubeconfigPath string) error
}
