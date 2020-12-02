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

package tinkerbell_test

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"

	lokoconfig "github.com/kinvolk/lokomotive/pkg/config"
	"github.com/kinvolk/lokomotive/pkg/platform/tinkerbell"
)

func baseConfig() *tinkerbell.Config {
	return &tinkerbell.Config{
		AssetDir:              "foo",
		Name:                  "foo",
		DNSZone:               "example",
		SSHPublicKeys:         []string{"foo"},
		ControllerIPAddresses: []string{"foo"},
		Sandbox: &tinkerbell.Sandbox{
			HostsCIDR:        "foo",
			FlatcarImagePath: "foo",
			PoolPath:         "foo",
		},
		WorkerPools: []tinkerbell.WorkerPool{
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
		mutatingF   func(*tinkerbell.Config)
		expectError bool
	}{
		"base config": {
			mutatingF: func(c *tinkerbell.Config) {},
		},
		"require asset dir": {
			mutatingF: func(c *tinkerbell.Config) {
				c.AssetDir = ""
			},
			expectError: true,
		},
		"require name": {
			mutatingF: func(c *tinkerbell.Config) {
				c.Name = ""
			},
			expectError: true,
		},
		"require DNS zone": {
			mutatingF: func(c *tinkerbell.Config) {
				c.DNSZone = ""
			},
			expectError: true,
		},
		"require SSH public key": {
			mutatingF: func(c *tinkerbell.Config) {
				c.SSHPublicKeys = []string{}
			},
			expectError: true,
		},
		"allow no worker pools": {
			mutatingF: func(c *tinkerbell.Config) {
				c.WorkerPools = nil
			},
		},
		"require worker pool name": {
			mutatingF: func(c *tinkerbell.Config) {
				c.WorkerPools[0].PoolName = ""
			},
			expectError: true,
		},
		"require worker pool IP address": {
			mutatingF: func(c *tinkerbell.Config) {
				c.WorkerPools[0].IPAddresses = []string{}
			},
			expectError: true,
		},
		"require sandbox hosts CIDR": {
			mutatingF: func(c *tinkerbell.Config) {
				c.Sandbox.HostsCIDR = ""
			},
			expectError: true,
		},
		"require sandbox Flatcar image path": {
			mutatingF: func(c *tinkerbell.Config) {
				c.Sandbox.FlatcarImagePath = ""
			},
			expectError: true,
		},
		"require sandbox pool path": {
			mutatingF: func(c *tinkerbell.Config) {
				c.Sandbox.PoolPath = ""
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

func loadConfigFromString(t *testing.T, c string) (*tinkerbell.Config, hcl.Diagnostics) {
	p := hclparse.NewParser()

	f, d := p.ParseHCL([]byte(c), "x.lokocfg")
	if d.HasErrors() {
		t.Fatalf("Parsing HCL should succeed, got: %v", d)
	}

	configBody := hcl.MergeFiles([]*hcl.File{f})

	var rootConfig lokoconfig.RootConfig

	if d := gohcl.DecodeBody(configBody, nil, &rootConfig); d.HasErrors() {
		t.Fatalf("Decoding root config should succeed, got: %v", d)
	}

	cc := &tinkerbell.Config{}

	return cc, cc.LoadConfig(&rootConfig.Cluster.Config, &hcl.EvalContext{})
}

func TestLoadConfigEmpty(t *testing.T) {
	c := `cluster "tinkerbell" {}`

	if _, d := loadConfigFromString(t, c); !d.HasErrors() {
		t.Fatalf("Empty config should not be valid")
	}
}

func TestLoadConfigTrimSSHKeys(t *testing.T) {
	c := `
cluster "tinkerbell" {
  asset_dir       = "foo"
  ssh_public_keys = [<<EOF
key


EOF
		,
	]

  # Tinkerbell hardware entry must exist with this IP address.
  controller_ip_addresses = [
    "10.17.3.4",
  ]

  name = "tink"

  dns_zone = "example.com"

  worker_pool "foo" {
    ip_addresses = [
      "10.17.3.5"
    ]

    ssh_public_keys = [<<EOF
key

EOF
			,
		]
  }
}
`

	cc, d := loadConfigFromString(t, c)
	if d.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %v", d)
	}

	if !reflect.DeepEqual(cc.SSHPublicKeys, []string{"key"}) {
		t.Errorf("Controllers SSH public keys should be trimmed from whitespace to ensure right Terraform rendering")
	}

	if !reflect.DeepEqual(cc.WorkerPools[0].SSHPublicKeys, []string{"key"}) {
		t.Errorf("Worker pools SSH public keys should be trimmed from whitespace to ensure right Terraform rendering")
	}
}

func TestMeta(t *testing.T) {
	assetDir := "foo"

	c := &tinkerbell.Config{
		AssetDir:              assetDir,
		ControllerIPAddresses: []string{"foo", "bar"},
		WorkerPools: []tinkerbell.WorkerPool{
			{
				IPAddresses: []string{"foo"},
			},
			{
				IPAddresses: []string{"foo", "bar", "baz"},
			},
		},
	}

	m := c.Meta()

	if m.AssetDir != assetDir {
		t.Errorf("Expected asset dir %q, got %q", assetDir, m.AssetDir)
	}

	expectedNodes := 6
	if m.ExpectedNodes != expectedNodes {
		t.Errorf("Expected %d nodes, got %d", expectedNodes, m.ExpectedNodes)
	}
}
