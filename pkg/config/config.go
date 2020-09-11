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

// ReadHCL reads all HCL files at path and one HCL values file at valuesPath, constructs a Config
// and returns a pointer to it.
func ReadHCL(path, valuesPath string) (*Config, hcl.Diagnostics) {
	names, err := lokocfgFilenames(path)
	if err != nil {
		return nil, diag(fmt.Errorf("getting lokocfg paths: %w", err))
	}

	files, err := readFiles(names)
	if err != nil {
		return nil, diag(err)
	}

	body, diags := parseHCLFiles(files)
	if diags.HasErrors() {
		return nil, diags
	}

	varsExist, err := util.PathExists(valuesPath)
	if err != nil {
		return nil, diag(fmt.Errorf("could not stat %q: %w", valuesPath, err))
	}

	var userVals map[string]cty.Value
	if varsExist {
		userVals, diags = readValuesFile(valuesPath)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	var rootConfig RootConfig

	diags = gohcl.DecodeBody(body, nil, &rootConfig)
	if diags.HasErrors() {
		return nil, diags
	}

	variables, diags := populateVars(rootConfig.Variables, userVals)
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
		RootConfig:  &rootConfig,
		EvalContext: &evalContext,
	}, nil
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
		diags = append(diags, d...)
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
		return nil, fmt.Errorf("checking if %q is a directory: %w", p, err)
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

// populateVars accepts a slice of variables and a map of user-provided values. The function
// populates each variable with the corresponding user-provided value or with a default. If a
// variable doesn't have a matching value and there is no default, an error is returned (as
// hcl.Diagnostics).
func populateVars(vars []variable, values map[string]cty.Value) (map[string]cty.Value, hcl.Diagnostics) {
	res := map[string]cty.Value{}

	for _, v := range vars {
		if value, ok := values[v.Name]; ok {
			res[v.Name] = value

			continue
		}

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

// diag returns an hcl.Diagnostics with a single hcl.Diagnostic based on the provided error.
func diag(err error) hcl.Diagnostics {
	return hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
		},
	}
}
