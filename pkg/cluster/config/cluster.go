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

const defaultCertsValidityPeriodHours = 8760

// ClusterConfig defines the cluster configuration
type ClusterConfig struct {
	EnableAggregation        bool   `hcl:"enable_aggregation,optional"`
	CertsValidityPeriodHours int    `hcl:"certs_validity_period_hours,optional"`
	ClusterDomainSuffix      string `hcl:"cluster_domain_suffix,optional"`
}

// DefaultClusterConfig returns an instance of ClusterConfig with
// default values
func DefaultClusterConfig() *ClusterConfig {
	return &ClusterConfig{
		EnableAggregation:        true,
		CertsValidityPeriodHours: defaultCertsValidityPeriodHours,
		ClusterDomainSuffix:      "cluster.local",
	}
}

// Validate validates the cluster configuration
func (c *ClusterConfig) Validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics
	if c.CertsValidityPeriodHours <= 0 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error in 'cluster' block",
			Detail:   fmt.Sprintf("`certs_validity_period_hours` should be more than zero, got: %d", c.CertsValidityPeriodHours),
		})
	}

	if c.ClusterDomainSuffix == "" {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error in 'cluster' block",
			Detail:   fmt.Sprintf("`cluster_domain_suffix` cannot be empty"),
		})
	}

	return diagnostics
}
