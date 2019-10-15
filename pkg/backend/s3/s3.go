package s3

import (
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"

	"github.com/kinvolk/lokoctl/pkg/backend"
	"github.com/kinvolk/lokoctl/pkg/components/util"
)

type s3 struct {
	Bucket string `hcl:"bucket"`
	Key    string `hcl:"key"`
	Region string `hcl:"region"`
}

// init registers s3 as a backend
func init() {
	backend.Register("s3", NewS3Backend())
}

//Loadconfig loads configuration for s3 backend
func (s *s3) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		return hcl.Diagnostics{}
	}
	return gohcl.DecodeBody(*configBody, evalContext, s)
}

func NewS3Backend() *s3 {
	return &s3{}
}

// Render renders the go template with s3 backend configuration
func (s *s3) Render() (string, error) {

	return util.RenderTemplate(backendConfigTmpl, s)
}
