package util

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"

	"github.com/kinvolk/lokoctl/pkg/config"
)

// GetComponentBody parses a string containing a component configuration in
// HCL and returns its body.
// Currently only the body of the first component is returned.
func GetComponentBody(configHCL string, name string) (*hcl.Body, hcl.Diagnostics) {
	hclParser := hclparse.NewParser()

	file, diags := hclParser.ParseHCL([]byte(configHCL), "x.lokocfg")
	if diags.HasErrors() {
		return nil, diags
	}

	configBody := hcl.MergeFiles([]*hcl.File{file})

	var rootConfig config.RootConfig

	diagnostics := gohcl.DecodeBody(configBody, nil, &rootConfig)
	if diagnostics.HasErrors() {
		return nil, diags
	}

	c := &config.Config{
		RootConfig: &rootConfig,
	}

	return c.LoadComponentConfigBody(name), nil
}
