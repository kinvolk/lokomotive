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
	"net"

	"github.com/hashicorp/hcl/v2"
)

const defaultNetworkMTU = 1480

// NetworkConfig defines the network configuraition
type NetworkConfig struct {
	NetworkMTU      int    `hcl:"network_mtu,optional"`
	PodCIDR         string `hcl:"pod_cidr,optional"`
	ServiceCIDR     string `hcl:"service_cidr,optional"`
	EnableReporting bool   `hcl:"enable_reporting,optional"`
}

// DefaultNetworkConfig returns an instance of NetworkConfig with
// default values
func DefaultNetworkConfig() *NetworkConfig {
	return &NetworkConfig{
		NetworkMTU:      defaultNetworkMTU,
		PodCIDR:         "10.2.0.0/16",
		ServiceCIDR:     "10.3.0.0/16",
		EnableReporting: false,
	}
}

// Validate validates the network configuration
func (n *NetworkConfig) Validate() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if n.NetworkMTU <= 0 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error in 'network' block",
			Detail:   fmt.Sprintf("expected 'network_mtu' to be greater than zero, got: %d", n.NetworkMTU),
		})
	}

	if err := validCIDR(n.PodCIDR); err != nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error in 'network' block",
			Detail:   fmt.Sprintf("invalid 'pod_cidr': %s", n.PodCIDR),
		})
	}

	if err := validCIDR(n.ServiceCIDR); err != nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Validation error in 'network' block",
			Detail:   fmt.Sprintf("invalid 'service_cidr': %s", n.ServiceCIDR),
		})
	}

	return diagnostics
}

func validCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)

	return err
}
