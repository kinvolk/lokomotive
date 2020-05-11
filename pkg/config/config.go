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

package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/mitchellh/go-homedir"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	"github.com/kinvolk/lokomotive/pkg/util"
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

type backend struct {
	Name   string   `hcl:"name,label"`
	Config hcl.Body `hcl:",remain"`
}

// ClusterConfig represents the HCL configuration provided by the
// user in lokocfg file. Each fields represents an HCL block type.
type ClusterConfig struct {
	Cluster    *cluster    `hcl:"cluster,block"`
	Backend    *backend    `hcl:"backend,block"`
	Components []component `hcl:"component,block"`
	Variables  []variable  `hcl:"variable,block"`
}

// HCL represents the HCL configuration provided by
// the user and evaluation context in case any variables or functions
// are used in the Lokocfg file.
type HCL struct {
	ClusterConfig *ClusterConfig
	EvalContext   *hcl.EvalContext
}

func loadLokocfgPaths(configPath string) ([]string, error) {
	isDir, err := util.PathIsDir(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config path %q: %w", configPath, err)
	}
	var lokocfgPaths []string
	if isDir {
		globPattern := filepath.Join(configPath, "*.lokocfg")
		configFiles, err := filepath.Glob(globPattern)
		if err != nil {
			return nil, fmt.Errorf("bad filepath glob pattern %q: %w", globPattern, err)
		}
		lokocfgPaths = append(lokocfgPaths, configFiles...)
	} else {
		lokocfgPaths = append(lokocfgPaths, configPath)
	}
	return lokocfgPaths, nil
}

func LoadConfig(lokocfgPath, lokocfgVarsPath string) (*HCL, hcl.Diagnostics) {
	lokocfgPaths, err := loadLokocfgPaths(lokocfgPath)
	if err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  err.Error(),
			},
		}
	}

	lokocfgBytes := make(map[string][]byte)

	for _, f := range lokocfgPaths {
		data, err := ioutil.ReadFile(filepath.Clean(f))
		if err != nil {
			return nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Can't read %q: %v", f, err),
				},
			}
		}
		lokocfgBytes[f] = data
	}

	useLokocfg, err := util.PathExists(lokocfgVarsPath)
	if err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("could not stat %q: %v", lokocfgVarsPath, err),
			},
		}
	}
	lokocfgVarsBytes := []byte{}
	if useLokocfg {
		data, err := ioutil.ReadFile(filepath.Clean(lokocfgVarsPath))
		if err != nil {
			return nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Can't read %q: %v", lokocfgVarsPath, err),
				},
			}
		}
		lokocfgVarsBytes = data
	}

	return parseConfig(lokocfgBytes, lokocfgVarsBytes, lokocfgVarsPath)
}

// TODO: if common to pass them around, maybe consider creating a struct with fields lokocfgVars and
// lokocfgVarsPath.
func parseConfig(lokocfg map[string][]byte, lokocfgVars []byte, lokocfgVarsPath string) (*HCL, hcl.Diagnostics) {
	var hclFiles []*hcl.File
	for f, data := range lokocfg {
		hclParser := hclparse.NewParser()
		hclFile, diags := hclParser.ParseHCL(data, f)
		if len(diags) > 0 {
			return nil, diags
		}
		hclFiles = append(hclFiles, hclFile)
	}

	configBody := hcl.MergeFiles(hclFiles)

	var userVals map[string]cty.Value
	var diags hcl.Diagnostics

	userVals, diags = LoadValuesFile(lokocfgVarsPath, lokocfgVars)
	if len(diags) > 0 {
		return nil, diags
	}

	var clusterConfig ClusterConfig
	diags = gohcl.DecodeBody(configBody, nil, &clusterConfig)
	if len(diags) > 0 {
		return nil, diags
	}

	variables := map[string]cty.Value{}
	for _, v := range clusterConfig.Variables {
		if userVal, ok := userVals[v.Name]; ok {
			variables[v.Name] = userVal
			continue
		}
		if len(v.Default) == 0 {
			continue
		}
		defaultValue, hasDefaultValue := v.Default["default"]
		if !hasDefaultValue {
			continue
		}
		defaultVal, diags := defaultValue.Expr.Value(nil)
		if len(diags) > 0 {
			return nil, diags
		}
		variables[v.Name] = defaultVal
	}

	evalContext := hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var": cty.ObjectVal(variables),
		},
		Functions: map[string]function.Function{
			"pathexpand": evalFuncPathExpand(),
			"file":       evalFuncFile(),
		},
	}

	return &HCL{
		ClusterConfig: &clusterConfig,
		EvalContext:   &evalContext,
	}, nil
}

func evalFuncPathExpand() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			}},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			expandedPath, err := homedir.Expand(args[0].AsString())
			return cty.StringVal(expandedPath), err
		},
	})
}

func evalFuncFile() function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			}},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			filePath := args[0].AsString()
			content, err := ioutil.ReadFile(filePath)
			return cty.StringVal(string(content)), err
		},
	})
}

// LoadComponentConfigBody returns nil if no component with the given
// name is found in the configuration
func (c *HCL) LoadComponentConfigBody(componentName string) *hcl.Body {
	for _, component := range c.ClusterConfig.Components {
		if componentName == component.Name {
			return &component.Config
		}
	}
	return nil
}
