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

func loadLokocfgPaths(path, extension string) ([]string, error) {
	var paths []string

	isDir, err := util.PathIsDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file path %q: %w", path, err)
	}

	if isDir {
		globPattern := filepath.Join(path, fmt.Sprintf("*.%s", extension))

		hclFiles, err := filepath.Glob(globPattern)
		if err != nil {
			return nil, fmt.Errorf("bad filepath glob pattern %q: %w", globPattern, err)
		}

		paths = append(paths, hclFiles...)
	} else {
		paths = append(paths, path)
	}

	return paths, nil
}

// LoadHCLFiles loads all the hcl files present in the path provided into a
// map of file name and content in byte array
func LoadHCLFiles(path, extension string) (map[string][]byte, hcl.Diagnostics) {
	files := make(map[string][]byte)
	lokocfgPaths, err := loadLokocfgPaths(path, extension)
	if err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  err.Error(),
			},
		}
	}

	if len(lokocfgPaths) == 0 {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "could not find any `.lokocfg` files in the provided path",
			},
		}
	}

	for _, path := range lokocfgPaths {
		data, err := loadHCLFile(path)
		if err != nil {
			return nil, hcl.Diagnostics{
				&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  err.Error(),
				},
			}
		}

		files[path] = data
	}

	return files, nil
}

func loadHCLFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ParseHCLFiles parses the HCL files into an instance of Config.
//nolint:funlen
func ParseHCLFiles(lokocfgFiles, variablesFile map[string][]byte) (*HCLConfig, hcl.Diagnostics) {
	hclFiles := []*hcl.File{}
	varsFiles := []*hcl.File{}

	hclParser := hclparse.NewParser()

	for path, content := range lokocfgFiles {
		file, diags := hclParser.ParseHCL(content, path)
		if diags.HasErrors() {
			return nil, diags
		}

		hclFiles = append(hclFiles, file)
	}

	for path, content := range variablesFile {
		file, diags := hclParser.ParseHCL(content, path)
		if diags.HasErrors() {
			return nil, diags
		}

		varsFiles = append(varsFiles, file)
	}

	configBody := hcl.MergeFiles(hclFiles)

	var userVals map[string]cty.Value
	var diags hcl.Diagnostics

	if len(variablesFile) > 0 {
		userVals, diags = LoadValues(varsFiles)
		if len(diags) > 0 {
			return nil, diags
		}
	}

	var cfg ClusterConfig

	diags = gohcl.DecodeBody(configBody, nil, &cfg)
	if len(diags) > 0 {
		return nil, diags
	}

	variables := map[string]cty.Value{}

	for _, v := range cfg.Variables {
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
		ClusterConfig: &cfg,
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
