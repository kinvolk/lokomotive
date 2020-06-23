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

// nolint:testpackage
package assets

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// Verify assets are successfully extracted from memory to disk.
func TestExtract(t *testing.T) {
	dir, err := ioutil.TempDir("", "lokoctl")
	if err != nil {
		t.Fatalf("Creating temp dir: %v", err)
	}

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("Could not remove temp dir %q", dir)
		}
	}()

	// Use non-existing subdirectories (`/a/b/c`) to verify parent dirs are created.
	err = Extract(TerraformModulesSource, filepath.Join(dir, "a", "b", "c"))
	if err != nil {
		t.Fatalf("Extracting assets: %v", err)
	}
}

// Verify assets are successfully extracted from memory to an existing directory.
func TestExtractToExistingDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "lokoctl")
	if err != nil {
		t.Fatalf("Creating temp dir: %v", err)
	}

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("Could not remove temp dir %q", dir)
		}
	}()

	err = Extract(TerraformModulesSource, dir)
	if err != nil {
		t.Fatalf("Extracting assets: %v", err)
	}
}
