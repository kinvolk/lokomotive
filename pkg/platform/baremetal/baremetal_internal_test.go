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

	"github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/backend/local"
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

	c := &config{
		Backend: &backend.Backend{
			Type: "local",
			Config: local.Config{
				Path: "fake",
			},
		},
	}

	if err := createTerraformConfigFile(c, tmpDir); err != nil {
		t.Fatalf("creating Terraform config files should succeed, got: %v", err)
	}
}
