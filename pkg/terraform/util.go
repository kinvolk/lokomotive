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
	"os"
	"path/filepath"
	"strings"

	"github.com/kinvolk/lokomotive/pkg/install"

	"github.com/pkg/errors"
)

const backendFileName = "backend.tf"
const clusterFileName = "cluster.tf"

// Configure creates Terraform directories and modules as well as a Terraform backend file if
// provided by the user.
func Configure(assetDir, renderedBackend, renderedPlatform string) error {
	if err := PrepareTerraformDirectoryAndModules(assetDir); err != nil {
		return errors.Wrapf(err, "Failed to create required terraform directory")
	}

	// Create backend file only if the backend rendered string isn't empty.
	if len(strings.TrimSpace(renderedBackend)) > 0 {
		if err := CreateTerraformBackendFile(assetDir, renderedBackend); err != nil {
			return errors.Wrapf(err, "Failed to create backend configuration file")
		}
	}

	if err := CreateTerraformClusterFile(assetDir, renderedPlatform); err != nil {
		return errors.Wrapf(err, "Failed to create cluster configuration file")
	}

	return nil
}

// PrepareTerraformDirectoryAndModules creates a Terraform directory and downloads required modules.
func PrepareTerraformDirectoryAndModules(assetDir string) error {
	terraformModuleDir := filepath.Join(assetDir, "lokomotive-kubernetes")
	if err := install.PrepareLokomotiveTerraformModuleAt(terraformModuleDir); err != nil {
		return err
	}

	terraformRootDir := filepath.Join(assetDir, "terraform")
	if err := install.PrepareTerraformRootDir(terraformRootDir); err != nil {
		return err
	}

	return nil
}

// GetTerraformRootDir gets the Terraform directory path.
func GetTerraformRootDir(assetDir string) string {
	return filepath.Join(assetDir, "terraform")
}

// CreateTerraformBackendFile creates the Terraform backend configuration file.
func CreateTerraformBackendFile(assetDir, data string) error {
	var err error

	backendString := fmt.Sprintf("terraform {%s}\n", data)
	terraformRootDir := GetTerraformRootDir(assetDir)
	path := filepath.Join(terraformRootDir, backendFileName)
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %q", path)
	}

	defer func() {
		if ferr := f.Close(); ferr != nil {
			err = ferr
		}
	}()

	if _, err = f.WriteString(backendString); err != nil {
		return errors.Wrapf(err, "failed to write to backend file %q", path)
	}

	if err = f.Sync(); err != nil {
		return errors.Wrapf(err, "failed to flush data to file %q", path)
	}

	return err
}

// CreateTerraformClusterFile creates the Terraform cluster configuration file.
func CreateTerraformClusterFile(assetDir, data string) error {
	var err error

	terraformRootDir := GetTerraformRootDir(assetDir)
	path := filepath.Join(terraformRootDir, clusterFileName)

	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %q", path)
	}

	defer func() {
		if ferr := f.Close(); ferr != nil {
			err = ferr
		}
	}()

	if _, err = f.WriteString(data); err != nil {
		return errors.Wrapf(err, "failed to write to cluster file %q", path)
	}

	if err = f.Sync(); err != nil {
		return errors.Wrapf(err, "failed to flush data to file %q", path)
	}

	return err
}
