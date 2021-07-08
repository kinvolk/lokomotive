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

package vmware_test

import (
	"testing"

	"github.com/kinvolk/lokomotive/pkg/dns"
	"github.com/kinvolk/lokomotive/pkg/platform/vmware"
)

func baseConfig() *vmware.Config {
	return &vmware.Config{
		AssetDir: "foo",
		Name:     "foo",
		DNS: dns.Config{
			Provider: "manual",
			Zone:     "example.com",
		},
		Datacenter:            "dc",
		Datastore:             "datastore",
		ComputeCluster:        "cluster",
		Network:               "network",
		Template:              "mytemplate",
		SSHPublicKeys:         []string{"foo"},
		ControllerIPAddresses: []string{"foo"},
		HostsCIDR:             "10.10.0.0/8",
		WorkerPools: []vmware.WorkerPool{
			{
				PoolName:      "foo",
				IPAddresses:   []string{"foo"},
				SSHPublicKeys: []string{"foo"},
			},
		},
	}
}

//nolint:funlen
func TestConfigValidation(t *testing.T) {
	cases := map[string]struct {
		mutatingF   func(*vmware.Config)
		expectError bool
	}{
		"base config": {
			mutatingF: func(c *vmware.Config) {},
		},
		"require asset dir": {
			mutatingF: func(c *vmware.Config) {
				c.AssetDir = ""
			},
			expectError: true,
		},
		"require name": {
			mutatingF: func(c *vmware.Config) {
				c.Name = ""
			},
			expectError: true,
		},
		"require datacenter": {
			mutatingF: func(c *vmware.Config) {
				c.Datacenter = ""
			},
			expectError: true,
		},
		"require datastore": {
			mutatingF: func(c *vmware.Config) {
				c.Datastore = ""
			},
			expectError: true,
		},
		"require computer cluster": {
			mutatingF: func(c *vmware.Config) {
				c.ComputeCluster = ""
			},
			expectError: true,
		},
		"require network": {
			mutatingF: func(c *vmware.Config) {
				c.Network = ""
			},
			expectError: true,
		},
		"require template": {
			mutatingF: func(c *vmware.Config) {
				c.Template = ""
			},
			expectError: true,
		},
		"allow no folder": {
			mutatingF: func(c *vmware.Config) {
				c.Folder = ""
			},
		},
		"require host CIDR": {
			mutatingF: func(c *vmware.Config) {
				c.HostsCIDR = ""
			},
			expectError: true,
		},
		"require SSH public key": {
			mutatingF: func(c *vmware.Config) {
				c.SSHPublicKeys = []string{}
			},
			expectError: true,
		},
		"allow no worker pools": {
			mutatingF: func(c *vmware.Config) {
				c.WorkerPools = nil
			},
		},
		"require worker pool name": {
			mutatingF: func(c *vmware.Config) {
				c.WorkerPools[0].PoolName = ""
			},
			expectError: true,
		},
		"require worker pool IP address": {
			mutatingF: func(c *vmware.Config) {
				c.WorkerPools[0].IPAddresses = []string{}
			},
			expectError: true,
		},
	}

	for n, spec := range cases {
		spec := spec

		t.Run(n, func(t *testing.T) {
			c := baseConfig()
			spec.mutatingF(c)

			diags := c.Validate()

			if spec.expectError && !diags.HasErrors() {
				t.Fatalf("Expected validation error")
			}

			if !spec.expectError && diags.HasErrors() {
				t.Fatalf("Unexpected validation error: %v", diags)
			}
		})
	}
}
