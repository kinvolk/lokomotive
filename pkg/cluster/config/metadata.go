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

	"github.com/hashicorp/hcl/v2"
	"github.com/mitchellh/go-homedir"
)

// Metadata defines the metadata configuration
type Metadata struct {
	AssetDir    string `hcl:"asset_dir"`
	ClusterName string `hcl:"cluster_name"`
}

// Validate validates the metadata configuration
func (m *Metadata) Validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics
	if m.AssetDir == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error in 'metadata' block",
			Detail:   fmt.Sprintf("`asset_dir` cannot be empty"),
		})
	}

	err := m.setAssetDirAfterExpand()
	if err != nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error in 'metadata' block",
			Detail:   err.Error(),
		})
	}

	if m.ClusterName == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error in 'metadata' block",
			Detail:   fmt.Sprintf("`cluster_name` cannot be empty"),
		})
	}

	return diagnostics
}

func (m *Metadata) setAssetDirAfterExpand() error {
	assetDir, err := homedir.Expand(m.AssetDir)
	if err != nil {
		return fmt.Errorf("`asset_dir` could not be expanded, got: %v", err)
	}

	m.AssetDir = assetDir

	return nil
}
