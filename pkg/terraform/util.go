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

package terraform

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/backend"
)

const backendFileName = "backend.tf"

// Configure creates Terraform directories and modules as well as a Terraform backend file if
// provided by the user.
func Configure(assetDir string, b backend.Backend) error {
	if err := PrepareTerraformDirectoryAndModules(assetDir); err != nil {
		return fmt.Errorf("creating Terraform directories: %w", err)
	}

	// Create backend file only if a backend was configured by the user.
	if b == nil {
		return nil
	}

	if err := CreateTerraformBackendFile(assetDir, b.String()); err != nil {
		return fmt.Errorf("creating backend configuration file: %w", err)
	}

	return nil
}

// PrepareTerraformDirectoryAndModules creates a Terraform directory and downloads required modules.
func PrepareTerraformDirectoryAndModules(assetDir string) error {
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

// GetTerraformRootDir gets the Terraform directory path.
func GetTerraformRootDir(assetDir string) string {
	return filepath.Join(assetDir, "terraform")
}

// CreateTerraformBackendFile creates the Terraform backend configuration file.
func CreateTerraformBackendFile(assetDir, data string) error {
	backendString := fmt.Sprintf("terraform {%s}\n", data)
	terraformRootDir := GetTerraformRootDir(assetDir)
	p := filepath.Join(terraformRootDir, backendFileName)

	return ioutil.WriteFile(p, []byte(backendString), 0600)
}
