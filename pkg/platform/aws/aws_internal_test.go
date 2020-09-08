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

package aws

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/kinvolk/lokomotive/pkg/config"
)

// Can't put this test in aws_test.go because workerPools is unexported.
func TestNewConfig(t *testing.T) {
	tests := []struct {
		desc       string
		rawConfig  string
		wantError  bool
		wantConfig *Config
	}{
		{
			desc: "Basic config",
			rawConfig: `cluster "aws" {
  asset_dir           = "/foo"
  cluster_name        = "mycluster"
  dns_zone            = "testzone"
  dns_zone_id         = "testzoneID"
  region              = "eu-west-1"
  ssh_pubkeys         = ["testkey"]
  worker_pool "foo" {
    count   = 1
    ssh_pubkeys = ["testkey"]
  }
}`,
			wantConfig: &Config{
				AssetDir:           "/foo",
				ClusterName:        "mycluster",
				DNSZone:            "testzone",
				DNSZoneID:          "testzoneID",
				EnableAggregation:  true,
				EnableTLSBootstrap: true,
				Region:             "eu-west-1",
				SSHPubKeys:         []string{"testkey"},
				WorkerPools: []workerPool{
					{
						Name:       "foo",
						Count:      1,
						SSHPubKeys: []string{"testkey"},
					},
				},
			},
		},
	}

	for _, test := range tests {
		p := hclparse.NewParser()

		f, d := p.ParseHCL([]byte(test.rawConfig), "x.lokocfg")

		if d.HasErrors() && !test.wantError {
			t.Fatalf("Parsing HCL should succeed, got: %v", d)
		}

		b := hcl.MergeFiles([]*hcl.File{f})

		var rc config.RootConfig

		if d := gohcl.DecodeBody(b, nil, &rc); d.HasErrors() && !test.wantError {
			t.Fatalf("Decoding into root config should succeed, got: %v", d)
		}

		c, d := NewConfig(&rc.Cluster.Config, nil)
		if d.HasErrors() && !test.wantError {
			t.Fatalf("Unexpected error: %v", d)
		}

		if diff := cmp.Diff(test.wantConfig, c); diff != "" {
			t.Errorf("%q: unexpected config (-want +got)\n%s", test.desc, diff)
		}
	}
}

// TODO: Migrate tests.

// // loadConfigFromString loads config from string.
// func loadConfigFromString(t *testing.T, c string) hcl.Diagnostics {
// 	p := hclparse.NewParser()

// 	f, d := p.ParseHCL([]byte(c), "x.lokocfg")
// 	if d.HasErrors() {
// 		t.Fatalf("parsing HCL should succeed, got: %v", d)
// 	}

// 	configBody := hcl.MergeFiles([]*hcl.File{f})

// 	var rootConfig lokoconfig.RootConfig

// 	if d := gohcl.DecodeBody(configBody, nil, &rootConfig); d.HasErrors() {
// 		t.Fatalf("decoding root config should succeed, got: %v", d)
// 	}

// 	cc := &config{}

// 	return cc.LoadConfig(&rootConfig.Cluster.Config, &hcl.EvalContext{})
// }

// func TestLoadConfig(t *testing.T) {
// 	c := `
// cluster "aws" {
//   asset_dir           = "/fooo"
//   cluster_name        = "mycluster"
//   dns_zone            = "testzone"
//   dns_zone_id         = "testzoneID"
//   ssh_pubkeys         = ["testkey"]
//   worker_pool "foo" {
//     count   = 1
//     ssh_pubkeys = ["testkey"]
//   }
// }
// `

// 	if d := loadConfigFromString(t, c); d.HasErrors() {
// 		t.Fatalf("valid config should not return error, got: %v", d)
// 	}
// }

// //nolint: funlen
// func TestWorkerPoolPort(t *testing.T) {
// 	type testCase struct {
// 		// Config to test.
// 		cfg config
// 		// Expected output after running test.
// 		hasError bool
// 	}

// 	cases := []testCase{
// 		{
// 			cfg: config{
// 				WorkerPools: []workerPool{
// 					{
// 						Name: "pool-1",
// 					},
// 				},
// 			},
// 			hasError: false,
// 		},
// 		{
// 			cfg: config{
// 				WorkerPools: []workerPool{
// 					{
// 						Name:        "pool-1",
// 						LBHTTPPort:  80,
// 						LBHTTPSPort: 443,
// 					},
// 				},
// 			},
// 			hasError: false,
// 		},
// 		{
// 			cfg: config{
// 				WorkerPools: []workerPool{
// 					{
// 						Name: "pool-1",
// 					},
// 					{
// 						Name: "pool-2",
// 					},
// 				},
// 			},
// 			hasError: true,
// 		},
// 		{
// 			cfg: config{
// 				WorkerPools: []workerPool{
// 					{
// 						Name: "pool-1",
// 					},
// 					{
// 						Name:        "pool-2",
// 						LBHTTPPort:  80,
// 						LBHTTPSPort: 443,
// 					},
// 				},
// 			},
// 			hasError: true,
// 		},
// 		{
// 			cfg: config{
// 				WorkerPools: []workerPool{
// 					{
// 						Name:        "pool-1",
// 						LBHTTPPort:  8080,
// 						LBHTTPSPort: 8443,
// 					},
// 					{
// 						Name:        "pool-2",
// 						LBHTTPPort:  8080,
// 						LBHTTPSPort: 8443,
// 					},
// 				},
// 			},
// 			hasError: true,
// 		},
// 		{
// 			cfg: config{
// 				WorkerPools: []workerPool{
// 					{
// 						Name: "pool-1",
// 					},
// 					{
// 						Name:        "pool-2",
// 						LBHTTPPort:  8080,
// 						LBHTTPSPort: 8443,
// 					},
// 				},
// 			},
// 			hasError: false,
// 		},
// 	}

// 	for tcIdx, tc := range cases {
// 		output := tc.cfg.checkLBPortsUnique()
// 		if output.HasErrors() != tc.hasError {
// 			t.Errorf("In test %v, expected %v, got %v", tcIdx+1, tc.hasError, output.HasErrors())
// 		}
// 	}
// }
