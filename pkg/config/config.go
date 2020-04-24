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

// loadLokocfgPaths loads the files that match the pattern dictated
// by the file extension. If a directory is passed, then all matching
// files are collected. If a single file is passed that that file is
// returned.
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
// map of file name and content in byte array.
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

// loadHCLfile loads the hcl file provided by the path and returns
// the contents of the file as a slice of bytes.
func loadHCLFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return data, nil
}

// getHCLFiles takes a map of file name and its contents and returns
// a slice of HCL File instance.
func getHCLFiles(files map[string][]byte) ([]*hcl.File, hcl.Diagnostics) {
	var diagnostics hcl.Diagnostics

	hclFiles := []*hcl.File{}

	hclParser := hclparse.NewParser()

	for path, content := range files {
		f, diags := hclParser.ParseHCL(content, path)
		diagnostics = append(diagnostics, diags...)
		hclFiles = append(hclFiles, f)
	}

	return hclFiles, diagnostics
}

// parseVariables parses the variables used in the HCL configuration.
func parseVariables(userVals map[string]cty.Value, cfg ClusterConfig) (map[string]cty.Value, hcl.Diagnostics) {
	var diagnostics hcl.Diagnostics

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
		diagnostics = append(diagnostics, diags...)

		variables[v.Name] = defaultVal
	}

	return variables, diagnostics
}

// ParseHCLFiles parses the HCL files into an instance of HCL struct.
func ParseHCLFiles(configFiles, variableFiles map[string][]byte) (*HCL, hcl.Diagnostics) {
	var diagnostics hcl.Diagnostics

	configHCLFiles, diags := getHCLFiles(configFiles)
	diagnostics = append(diagnostics, diags...)

	configBody := hcl.MergeFiles(configHCLFiles)

	var clusterConfig ClusterConfig

	diags = gohcl.DecodeBody(configBody, nil, &clusterConfig)
	diagnostics = append(diagnostics, diags...)

	varHCLFiles, diags := getHCLFiles(variableFiles)
	diagnostics = append(diagnostics, diags...)

	var userVals map[string]cty.Value

	userVals, diags = LoadValues(varHCLFiles)
	diagnostics = append(diagnostics, diags...)

	variables, diags := parseVariables(userVals, clusterConfig)
	diagnostics = append(diagnostics, diags...)

	if diagnostics.HasErrors() {
		return nil, diagnostics
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
