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

package baremetal

import (
	"io/ioutil"
	"os"
	"testing"
)

// createTerraformConfigFile() test.
func TestCreateTerraformConfigFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "lokoctl-tests-")
	if err != nil {
		t.Fatalf("creating tmp dir should succeed, got: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to remove temp dir %q: %v", tmpDir, err)
		}
	})

	c := &config{}

	if err := createTerraformConfigFile(c, tmpDir); err != nil {
		t.Fatalf("creating Terraform config files should succeed, got: %v", err)
	}
}

func validConfig() *config {
	return &config{}
}

func TestConfigurationIsInvalidWhen(t *testing.T) {
	cases := map[string]func(c *config){
		"conntrack_max_per_core_is_negative": func(c *config) {
			c.ConntrackMaxPerCore = -1
		},
		"clc_snippets_key_is_empty": func(c *config) {
			c.CLCSnippets = map[string][]string{
				"": {"clc_snippet_1", "clc_snippet_2"},
			}
		},
		"clc_snippets_value_is_empty": func(c *config) {
			c.CLCSnippets = map[string][]string{
				"node1": {""},
			}
		},
		"at_least_one_clc_snippets_value_is_empty": func(c *config) {
			c.CLCSnippets = map[string][]string{
				"node1": {"clc_snippet_1", "", "clc_snippet_3"},
			}
		},
	}

	for n, c := range cases {
		c := c

		t.Run(n, func(t *testing.T) {
			config := validConfig()

			c(config)

			if d := config.checkValidConfig(); !d.HasErrors() {
				t.Fatalf("Validating configuration did not return expected error")
			}
		})
	}
}

func TestConfigurationIsValidWhen(t *testing.T) {
	cases := map[string]func(c *config){
		"all_required_fields_are_set": func(c *config) {},
		"conntrack_max_per_core_is_a_positive_value": func(c *config) {
			c.ConntrackMaxPerCore = 10
		},
		"clc_snippets_has_both_key_and_value_populated": func(c *config) {
			c.CLCSnippets = map[string][]string{
				"node1": {"clc_snippet_1", "clc_snippet_2"},
			}
		},
	}

	for n, c := range cases {
		c := c

		t.Run(n, func(t *testing.T) {
			config := validConfig()

			c(config)

			if d := config.checkValidConfig(); d.HasErrors() {
				t.Fatalf("Validating configuration returned expected error: %v", d)
			}
		})
	}
}
