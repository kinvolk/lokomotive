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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

//nolint:funlen
func TestReadFiles(t *testing.T) {
	tests := []struct {
		desc  string
		files map[string][]byte
	}{
		{
			desc: "Single file",
			files: map[string][]byte{
				"foo.txt": []byte(`foo`),
			},
		},
		{
			desc: "Multiple files",
			files: map[string][]byte{
				"foo.txt": []byte(`foo`),
				"bar.txt": []byte(`bar`),
				"baz.txt": []byte(`baz`),
			},
		},
		{
			desc: "Empty file",
			files: map[string][]byte{
				"foo.txt": []byte(``),
			},
		},
		{
			desc:  "No files",
			files: map[string][]byte{},
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
		// test.files maps a short filename to its contents. We need full paths to verify the
		// result so we create a copy of the original map with full paths as the keys.
		wantFiles := map[string][]byte{}

		for fn, data := range test.files {
			p := filepath.Join(d, fn)
			wantFiles[p] = data

			if err := ioutil.WriteFile(p, data, 0600); err != nil {
				t.Fatalf("Could not write test file %q: %v", p, err)
			}
		}

		in := []string{}
		for fn := range wantFiles {
			in = append(in, fn)
		}

		out, err := readFiles(in)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !reflect.DeepEqual(out, wantFiles) {
			t.Fatalf("Unexpected files: got %v, want %v", out, wantFiles)
		}
	}
}

//nolint:funlen
func TestParseHCLFiles(t *testing.T) {
	tests := []struct {
		desc  string
		files map[string][]byte
		// Number of HCL diagnostics we expect to get.
		wantDiags int
	}{
		{
			desc: "Single file",
			files: map[string][]byte{
				"cluster.lokocfg": []byte(`cluster "aws" {
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
}

component "foo" {}
component "bar" {}`),
			},
		},
		{
			desc: "Multiple files",
			files: map[string][]byte{
				"cluster.lokocfg": []byte(`cluster "aws" {
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
				"components.lokocfg": []byte(`component "foo" {}
component "bar" {}`),
			},
		},
		{
			desc: "Diagnostics from multiple files",
			files: map[string][]byte{
				"this.lokocfg": []byte(`oops1`),
				"that.lokocfg": []byte(`oops2`),
			},
			wantDiags: 2,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			_, diags := parseHCLFiles(test.files)

			if len(diags) != test.wantDiags {
				t.Fatalf("Unexpected diagnostics: got %v, want %v", len(diags), test.wantDiags)
			}

			// TODO: Verify the resulting body is correct.
		})
	}
}

//nolint:funlen
func TestLokocfgFilenames(t *testing.T) {
	d, err := ioutil.TempDir("", "lokoctl-tests-")
	if err != nil {
		t.Fatalf("Could not create temporary directory: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(d); err != nil {
			t.Logf("Could not remove temporary directory %q: %v", d, err)
		}
	})

	tests := []struct {
		desc          string
		files         []string
		path          string
		wantFilenames []string
	}{
		{
			desc:          "Single file",
			files:         []string{"test.lokocfg"},
			path:          filepath.Join(d, "test.lokocfg"),
			wantFilenames: []string{filepath.Join(d, "test.lokocfg")},
		},
		{
			desc:  "Multiple files",
			files: []string{"test1.lokocfg", "test2.lokocfg"},
			path:  d,
			wantFilenames: []string{
				filepath.Join(d, "test1.lokocfg"),
				filepath.Join(d, "test2.lokocfg"),
			},
		},
		{
			desc:          "Unrelated files are ignored",
			files:         []string{"test.lokocfg", "test.txt"},
			path:          d,
			wantFilenames: []string{filepath.Join(d, "test.lokocfg")},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			for _, f := range test.files {
				p := filepath.Join(d, f)

				if _, err = os.Create(p); err != nil {
					t.Fatalf("Could not create file %q: %v", p, err)
				}

				t.Cleanup(func() {
					if err = os.Remove(p); err != nil {
						t.Fatalf("Could not remove test file %q: %v", p, err)
					}
				})
			}

			names, err := lokocfgFilenames(test.path)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(names, test.wantFilenames) {
				t.Fatalf("Unexpected filenames: got %v, want %v", names, test.wantFilenames)
			}
		})
	}
}
