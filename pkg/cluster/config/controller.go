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

const defaultControllerCount = 1

// ControllerConfig defines the configuration for controller block
type ControllerConfig struct {
	Count      int      `hcl:"count,optional"`
	SSHPubKeys []string `hcl:"ssh_pubkeys"`
}

// DefaultControllerConfig returns instance of ControllerConfig with
// default values
func DefaultControllerConfig() *ControllerConfig {
	return &ControllerConfig{
		Count: defaultControllerCount,
	}
}

// Validate validates the controller configuration
func (c *ControllerConfig) Validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics
	if c.Count < defaultControllerCount {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error in 'controller' block",
			Detail:   fmt.Sprintf("expected 'count' greater than 0, got: %d", c.Count),
		})
	}

	if len(c.SSHPubKeys) == 0 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error in 'controller' block",
			Detail:   fmt.Sprintf("expected atleast one public ssh-key in 'ssh_pubkeys', got: 0"),
		})
	}

	return diagnostics
}
