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

package config_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/kinvolk/lokomotive/pkg/config"
)

//nolint:funlen
func TestReadHCL(t *testing.T) {
	tests := []struct {
		desc string
		conf []byte
		vars []byte
		// Number of HCL diagnostics we expect to get.
		wantDiags int
	}{
		{
			desc: "Cluster with vars",
			conf: []byte(`variable "cluster_name" {}

cluster "aws" {
  asset_dir        = pathexpand("./assets")
  cluster_name     = var.cluster_name
  controller_count = 1
  dns_zone         = "example.com"
  dns_zone_id      = "fake"
  ssh_pubkeys      = ["stuff"]
  
  worker_pool "my-wp-name" {
    count         = 3
    instance_type = "fake"
    ssh_pubkeys   = ["stuff"]
  }
}`),
			vars: []byte(`cluster_name = "test-cluster"`),
		},
		{
			desc: "Cluster without vars",
			conf: []byte(`cluster "aws" {
  asset_dir        = pathexpand("./assets")
  cluster_name     = var.cluster_name
  controller_count = 1
  dns_zone         = "example.com"
  dns_zone_id      = "fake"
  ssh_pubkeys      = ["stuff"]
  
  worker_pool "my-wp-name" {
    count         = 3
    instance_type = "fake"
    ssh_pubkeys   = ["stuff"]
  }
}`),
		},
		{
			desc: "Empty config",
			conf: []byte(``),
			// TODO: Is it OK that an empty config doesn't generate errors?
		},
		{
			desc: "Empty vars",
			conf: []byte(`variable "cluster_name" {}
cluster "aws" {}`),
			vars: []byte(``),
		},
		{
			desc:      "Malformed config",
			conf:      []byte(`oops`),
			wantDiags: 1,
		},
		{
			desc:      "Non-existent block",
			conf:      []byte(`foo "bar" {}`),
			wantDiags: 1,
		},
	}

	d, err := ioutil.TempDir("", "lokoctl-tests-")
	if err != nil {
		t.Fatalf("Could not create temporary directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(d); err != nil {
			t.Logf("Could not remove temporary directory %q: %v", d, err)
		}
	})

	for _, test := range tests {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			cf, err := ioutil.TempFile(d, "lokoctl-tests-")
			if err != nil {
				t.Fatalf("Could not create temporary file: %v", err)
			}

			err = ioutil.WriteFile(cf.Name(), test.conf, 0600)
			if err != nil {
				t.Fatalf("Could not write config to file: %v", err)
			}

			var varFile string

			if len(test.vars) != 0 {
				vf, err := ioutil.TempFile(d, "lokoctl-tests-")
				if err != nil {
					t.Fatalf("Could not create temporary file: %v", err)
				}

				err = ioutil.WriteFile(vf.Name(), test.vars, 0600)
				if err != nil {
					t.Fatalf("Could not write vars to file: %v", err)
				}

				varFile = vf.Name()
			}

			_, diags := config.ReadHCL(cf.Name(), varFile)

			if len(diags) != test.wantDiags {
				t.Fatalf("Unexpected diagnostics: got %v, want %v", len(diags), test.wantDiags)
			}
		})
	}
}

//nolint:funlen
func TestComponent(t *testing.T) {
	tests := []struct {
		desc          string
		conf          []byte
		name          string
		wantComponent bool
	}{
		{
			desc:          "Existing component",
			conf:          []byte(`component "foo" {}`),
			name:          "foo",
			wantComponent: true,
		},
		{
			desc: "Non-existent component",
			conf: []byte(`component "foo" {}`),
			name: "bar",
		},
		{
			desc: "No component blocks",
			conf: []byte(`cluster "foo" {}`),
			name: "foo",
		},
	}

	d, err := ioutil.TempDir("", "lokoctl-tests-")
	if err != nil {
		t.Fatalf("Could not create temporary directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(d); err != nil {
			t.Logf("Could not remove temporary directory %q: %v", d, err)
		}
	})

	for _, test := range tests {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			cf, err := ioutil.TempFile(d, "lokoctl-tests-")
			if err != nil {
				t.Fatalf("Could not create temporary file: %v", err)
			}

			err = ioutil.WriteFile(cf.Name(), test.conf, 0600)
			if err != nil {
				t.Fatalf("Could not write config to file: %v", err)
			}

			c, d := config.ReadHCL(cf.Name(), "")

			if len(d) != 0 {
				t.Fatalf("Got unexpected errors: %v", d.Errs())
			}

			if c.Component(test.name) == nil && test.wantComponent {
				t.Fatalf("Expected to get component %q but got none", test.name)
			}

			if c.Component(test.name) != nil && !test.wantComponent {
				t.Fatalf("Got unexpected component %q", test.name)
			}
		})
	}
}
