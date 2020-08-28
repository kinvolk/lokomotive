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
	Platform string   `hcl:"platform,label"`
	Config   hcl.Body `hcl:",remain"`
}

type component struct {
	Name   string   `hcl:"name,label"`
	Config hcl.Body `hcl:",remain"`
}

type backend struct {
	Name   string   `hcl:"name,label"`
	Config hcl.Body `hcl:",remain"`
}

type RootConfig struct {
	Cluster    *cluster    `hcl:"cluster,block"`
	Backend    *backend    `hcl:"backend,block"`
	Components []component `hcl:"component,block"`
	Variables  []variable  `hcl:"variable,block"`
}

type Config struct {
	RootConfig  *RootConfig
	EvalContext *hcl.EvalContext
}

// Component finds a component named name and returns its configuration as a pointer to an
// hcl.Body. If the component isn't found, nil is returned.
func (c *Config) Component(name string) *hcl.Body {
	for _, component := range c.RootConfig.Components {
		if name == component.Name {
			return &component.Config
		}
	}

	return nil
}

// ReadHCL reads all HCL files at path and one HCL variables file at varPath, constructs a Config
// and returns a pointer to it.
func ReadHCL(path, varPath string) (*Config, hcl.Diagnostics) {
	rootConfig, diags := readConfig(path)
	if diags.HasErrors() {
		return nil, diags
	}

	variables, diags := rootConfig.variablesWithValues(varPath)
	if diags.HasErrors() {
		return nil, diags
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

	return &Config{
		RootConfig:  rootConfig,
		EvalContext: &evalContext,
	}, nil
}

// variablesWithValues converts configured variables into values map and optionally merges them with
// values from values file.
func (c *RootConfig) variablesWithValues(valuesPath string) (map[string]cty.Value, hcl.Diagnostics) {
	variablesMap, diags := c.variablesMap()
	if diags.HasErrors() {
		return nil, diags
	}

	valuesExist, err := util.PathExists(valuesPath)
	if err != nil {
		return nil, diag(fmt.Sprintf("could not stat %q: %v", valuesPath, err))
	}

	var values map[string]cty.Value

	if valuesExist {
		values, diags = readValuesFile(valuesPath)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	return mergeValuesMap(variablesMap, values), nil
}

// readConfig reads lokocfg configuration files and converts it into
// initialized RootConfig struct.
func readConfig(configPath string) (*RootConfig, hcl.Diagnostics) {
	// Read configuration files.
	names, err := lokocfgFilenames(configPath)
	if err != nil {
		return nil, diag(fmt.Sprintf("getting lokocfg paths: %v", err))
	}

	files, err := readFiles(names)
	if err != nil {
		return nil, diag(err.Error())
	}

	body, diags := parseHCLFiles(files)
	if diags.HasErrors() {
		return nil, diags
	}

	var rootConfig RootConfig

	return &rootConfig, gohcl.DecodeBody(body, nil, &rootConfig)
}

// readFiles reads all files based on the provided names and returns a map where the key is the
// name of the file and the value is a slice of bytes holding the file's contents.
func readFiles(names []string) (map[string][]byte, error) {
	files := map[string][]byte{}

	for _, fn := range names {
		data, err := ioutil.ReadFile(fn) //nolint:gosec
		if err != nil {
			return nil, fmt.Errorf("reading file %q: %w", fn, err)
		}

		files[fn] = data
	}

	return files, nil
}

// parseHCLFiles reads the provided HCL files, merges them into a single hcl.Body and returns it.
func parseHCLFiles(files map[string][]byte) (hcl.Body, hcl.Diagnostics) {
	parser := hclparse.NewParser()

	hclFiles := []*hcl.File{}

	var diags hcl.Diagnostics

	for name, data := range files {
		hclFile, d := parser.ParseHCL(data, name)
		if len(d) > 0 {
			diags = append(diags, d...)
		}

		hclFiles = append(hclFiles, hclFile)
	}

	if diags.HasErrors() {
		return nil, diags
	}

	return hcl.MergeFiles(hclFiles), nil
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

// lokocfgFilenames returns a slice containing the names of all .lokocfg files found under path p.
// If p is not a directory, p is returned by itself in the slice.
func lokocfgFilenames(p string) ([]string, error) {
	isDir, err := util.PathIsDir(p)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config path %q: %w", p, err)
	}

	if !isDir {
		return []string{p}, nil
	}

	globPattern := filepath.Join(p, "*.lokocfg")

	names, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, fmt.Errorf("bad filepath glob pattern %q: %w", globPattern, err)
	}

	return names, nil
}

// variablesMap converts Variables field into values map.
func (r *RootConfig) variablesMap() (map[string]cty.Value, hcl.Diagnostics) {
	res := map[string]cty.Value{}

	for _, v := range r.Variables {
		if len(v.Default) == 0 {
			continue
		}

		defaultValue, hasDefaultValue := v.Default["default"]
		if !hasDefaultValue {
			continue
		}

		dv, diags := defaultValue.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, diags
		}

		res[v.Name] = dv
	}

	return res, nil
}

// mergeValuesMap merges zero or more values map into one, where the next one
// takes precedence over previous ones.
func mergeValuesMap(values ...map[string]cty.Value) map[string]cty.Value {
	res := map[string]cty.Value{}

	for _, vm := range values {
		for k, v := range vm {
			res[k] = v
		}
	}

	return res
}

// diag returns an hcl.Diagnostics with a single hcl.Diagnostic based on the provided error string.
func diag(err string) hcl.Diagnostics {
	return hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err,
		},
	}
}
