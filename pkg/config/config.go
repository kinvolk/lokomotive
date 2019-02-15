package config

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

type variable struct {
	Name    string         `hcl:"name,label"`
	Default hcl.Attributes `hcl:",remain"`
}

type cluster struct {
	Name   string   `hcl:"name,label"`
	Config hcl.Body `hcl:",remain"`
}

type component struct {
	Name   string   `hcl:"name,label"`
	Config hcl.Body `hcl:",remain"`
}

type RootConfig struct {
	Cluster    *cluster    `hcl:"cluster,block"`
	Components []component `hcl:"component,block"`
	Variables  []variable  `hcl:"variable,block"`
}

type Config struct {
	RootConfig  *RootConfig
	EvalContext *hcl.EvalContext
}

type ClusterConfig struct {
	ControllerCount int `hcl:"controller_count,attr"`
	WorkerCount     int `hcl:"worker_count,attr"`
}

func LoadConfig(configDir string) (*Config, hcl.Diagnostics) {
	// TODO(schu): support both a target directory with
	// multiple configuration files and a single file

	globPattern := configDir + "./*.lokocfg"
	configFiles, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("bad filepath glob pattern %q: %v", globPattern, err),
			},
		}
	}

	hclParser := hclparse.NewParser()

	var hclFiles []*hcl.File
	for _, f := range configFiles {
		hclFile, diags := hclParser.ParseHCLFile(f)
		if len(diags) > 0 {
			return nil, diags
		}
		hclFiles = append(hclFiles, hclFile)
	}

	configBody := hcl.MergeFiles(hclFiles)

	userVals, diags := LoadValuesFile("lokocfg.vars")
	if len(diags) > 0 {
		return nil, diags
	}

	var rootConfig RootConfig
	diags = gohcl.DecodeBody(configBody, nil, &rootConfig)
	if len(diags) > 0 {
		return nil, diags
	}

	variables := map[string]cty.Value{}
	for _, v := range rootConfig.Variables {
		if userVal, ok := userVals[v.Name]; ok {
			variables[v.Name] = userVal
			continue
		}
		if len(v.Default) == 0 {
			continue
		}
		defaultVal, diags := v.Default["default"].Expr.Value(nil)
		if len(diags) > 0 {
			return nil, diags
		}
		variables[v.Name] = defaultVal
	}

	evalContext := hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var": cty.ObjectVal(variables),
		},
	}

	return &Config{
		RootConfig:  &rootConfig,
		EvalContext: &evalContext,
	}, nil
}

// LoadValuesFile reads the file at the given path and parses it as a
// "values file" (flat key.value HCL config) for later use in the
// `EvalContext`.
//
// Adapted from
// https://github.com/hashicorp/terraform/blob/d4ac68423c4998279f33404db46809d27a5c2362/configs/parser_values.go#L8-L23
func LoadValuesFile(path string) (map[string]cty.Value, hcl.Diagnostics) {
	hclParser := hclparse.NewParser()
	varsFile, diags := hclParser.ParseHCLFile(path)
	if diags != nil {
		return nil, diags
	}

	body := varsFile.Body
	if body == nil {
		return nil, diags
	}

	vars := make(map[string]cty.Value)
	attrs, attrsDiags := body.JustAttributes()
	diags = append(diags, attrsDiags...)
	if attrs == nil {
		return vars, diags
	}

	for name, attr := range attrs {
		val, valDiags := attr.Expr.Value(nil)
		diags = append(diags, valDiags...)
		vars[name] = val
	}

	return vars, diags
}

func (c *Config) LoadClusterConfig() (*ClusterConfig, hcl.Diagnostics) {
	var clusterConfig ClusterConfig
	diags := gohcl.DecodeBody(c.RootConfig.Cluster.Config, c.EvalContext, &clusterConfig)
	return &clusterConfig, diags
}

// LoadComponentConfigBody returns nil if no component with the given
// name is found in the configuration
func (c *Config) LoadComponentConfigBody(componentName string) *hcl.Body {
	for _, component := range c.RootConfig.Components {
		if componentName == component.Name {
			return &component.Config
		}
	}
	return nil
}
