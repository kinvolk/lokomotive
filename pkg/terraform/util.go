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

package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kinvolk/lokomotive/pkg/assets"
)

const backendFileName = "backend.tf"

// GetTerraformRootDir gets the Terraform directory path.
func GetTerraformRootDir(assetDir string) string {
	return filepath.Join(assetDir, "terraform")
}

// Configure creates Terraform directories and modules as well as a Terraform backend file if
// provided by the user.
func Configure(assetDir, renderedBackend string) error {
	if err := prepareTerraformDirectoryAndModules(assetDir); err != nil {
		return fmt.Errorf("creating Terraform directories: %w", err)
	}

	// Create backend file only if the backend rendered string isn't empty.
	if len(strings.TrimSpace(renderedBackend)) <= 0 {
		return nil
	}

	if err := createTerraformBackendFile(assetDir, renderedBackend); err != nil {
		return fmt.Errorf("creating backend configuration file: %w", err)
	}

	return nil
}

// prepareTerraformDirectoryAndModules creates a Terraform directory and downloads required modules.
func prepareTerraformDirectoryAndModules(assetDir string) error {
	terraformModuleDir := filepath.Join(assetDir, "terraform-modules")
	if err := assets.Extract(assets.TerraformModulesSource, terraformModuleDir); err != nil {
		return err
	}

	// Ensure Terraform root directory exists.
	p := filepath.Join(assetDir, "terraform")
	if err := os.MkdirAll(p, 0755); err != nil {
		return fmt.Errorf("creating Terraform root directory at %q: %w", p, err)
	}

	return nil
}

// createTerraformBackendFile creates the Terraform backend configuration file.
func createTerraformBackendFile(assetDir, data string) error {
	backendString := fmt.Sprintf("terraform {%s}\n", data)
	terraformRootDir := GetTerraformRootDir(assetDir)
	path := filepath.Join(terraformRootDir, backendFileName)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", path, err)
	}
	defer f.Close()

	if _, err = f.WriteString(backendString); err != nil {
		return fmt.Errorf("writing to backend file %q: %w", path, err)
	}

	if err = f.Sync(); err != nil {
		return fmt.Errorf("flushing data to file %q: %w", path, err)
	}

	return nil
}
