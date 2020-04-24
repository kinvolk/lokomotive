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

type ClusterConfig struct {
	Cluster    *cluster    `hcl:"cluster,block"`
	Backend    *backend    `hcl:"backend,block"`
	Components []component `hcl:"component,block"`
	Variables  []variable  `hcl:"variable,block"`
}

type HCLConfig struct {
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

func LoadConfig(lokocfgPath, lokocfgVarsPath string) (*HCLConfig, hcl.Diagnostics) {
	lokocfgPaths, err := loadLokocfgPaths(lokocfgPath)
	if err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  err.Error(),
			},
		}
	}

	hclParser := hclparse.NewParser()

	var hclFiles []*hcl.File
	for _, f := range lokocfgPaths {
		hclFile, diags := hclParser.ParseHCLFile(f)
		if len(diags) > 0 {
			return nil, diags
		}
		hclFiles = append(hclFiles, hclFile)
	}

	configBody := hcl.MergeFiles(hclFiles)

	exists, err := util.PathExists(lokocfgVarsPath)
	if err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("could not stat %q: %v", lokocfgVarsPath, err),
			},
		}
	}
	var userVals map[string]cty.Value
	var diags hcl.Diagnostics
	if exists {
		userVals, diags = LoadValuesFile(lokocfgVarsPath)
		if len(diags) > 0 {
			return nil, diags
		}
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

	return &HCLConfig{
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
func (c *HCLConfig) LoadComponentConfigBody(componentName string) *hcl.Body {
	for _, component := range c.ClusterConfig.Components {
		if componentName == component.Name {
			return &component.Config
		}
	}
	return nil
}
