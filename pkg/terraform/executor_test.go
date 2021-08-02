// Copyright 2021 The Lokomotive Authors
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

package terraform_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/kinvolk/lokomotive/pkg/terraform"
)

//nolint:funlen
func TestVersionConstraint(t *testing.T) {
	cases := map[string]struct {
		output      string
		expectError bool
	}{
		"valid": {
			output: `{
  "terraform_version": "0.13.10",
  "terraform_revision": "",
  "provider_selections": {},
  "terraform_outdated": false
}
`,
		},
		"outdated": {
			output: `{
  "terraform_version": "0.11.0",
  "terraform_revision": "",
  "provider_selections": {},
  "terraform_outdated": false
}
`,
			expectError: true,
		},
		"unsupported": {
			output: `{
  "terraform_version": "0.14.5",
  "terraform_revision": "",
  "provider_selections": {},
  "terraform_outdated": false
}
`,
			expectError: true,
		},
		"not JSON": {
			output: `Terraform v0.13.11

Your version of Terraform is out of date! The latest version
is 0.13.3. You can update by downloading from https://www.terraform.io/downloads.html`,
			expectError: true,
		},
		"empty version field": {
			output: `{
  "terraform_version": "",
  "terraform_revision": "",
  "provider_selections": {},
  "terraform_outdated": false
}
`,
			expectError: true,
		},
	}

	for n, c := range cases {
		c := c

		t.Run(n, func(t *testing.T) {
			tmpDir, err := ioutil.TempDir("", "lokoctl-tests-")
			if err != nil {
				t.Fatalf("Creating tmp dir should succeed, got: %v", err)
			}

			t.Cleanup(func() {
				if err := os.RemoveAll(tmpDir); err != nil {
					t.Logf("Removing directory %q: %v", tmpDir, err)
				}
			})

			v := []byte(fmt.Sprintf(`#!/bin/sh
		cat <<EOF
%s
EOF
		`, c.output))

			path := filepath.Join(tmpDir, "terraform")

			// #nosec G306 // File must be executable to pretend it's a Terraform binary.
			if err := ioutil.WriteFile(path, v, 0o700); err != nil {
				t.Fatalf("Writing file %q: %v", path, err)
			}

			if err := os.Setenv("PATH", fmt.Sprintf("%s:%s", tmpDir, os.Getenv("PATH"))); err != nil {
				t.Fatalf("Overriding PATH variable for testing: %v", err)
			}

			conf := terraform.Config{
				Verbose:    false,
				WorkingDir: tmpDir,
			}

			_, err = terraform.NewExecutor(conf)

			if !c.expectError && err != nil {
				t.Fatalf("Creating new executor should succeed, got: %v", err)
			}

			if c.expectError && err == nil {
				t.Fatalf("Creating new executor should fail")
			}
		})
	}
}
