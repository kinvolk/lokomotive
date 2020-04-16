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
)

// FlatcarConfig defines the flatcar configuration
type FlatcarConfig struct {
	Channel string `hcl:"channel,optional"`
	Version string `hcl:"version,optional"`
}

// DefaultFlatcarConfig returns an instance of FlatcarConfig
// with default values
func DefaultFlatcarConfig() *FlatcarConfig {
	return &FlatcarConfig{
		Version: "current",
		Channel: "stable",
	}
}

func getSupportedChannels() []string {
	return []string{"stable", "alpha", "beta", "edge"}
}

// Validate validates the flatcar configuration
func (f *FlatcarConfig) Validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics
	// TODO: Validate release version provided by the user
	// preferably by generating http query  against flatcar releases url
	if !isPresent(f.Channel) {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error in 'flatcar' block",
			Detail:   fmt.Sprintf("unsupported channel '%s'", f.Channel),
		})
	}

	if f.Version == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error in 'flatcar' block",
			Detail:   fmt.Sprintf("'version' cannot be empty '%s'", f.Version),
		})
	}

	return diagnostics
}

func isPresent(c string) bool {
	for _, channel := range getSupportedChannels() {
		if c == channel {
			return true
		}
	}

	return false
}
