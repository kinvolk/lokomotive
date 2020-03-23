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

package util

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"

	"github.com/kinvolk/lokomotive/pkg/config"
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

	var clusterConfig config.ClusterConfig

	diagnostics := gohcl.DecodeBody(configBody, nil, &clusterConfig)
	if diagnostics.HasErrors() {
		return nil, diags
	}

	c := &config.Config{
		ClusterConfig: &clusterConfig,
	}

	return c.LoadComponentConfigBody(name), nil
}
