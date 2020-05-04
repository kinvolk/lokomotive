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

	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/util/walkers"
	"github.com/pkg/errors"
)

const backendFileName = "backend.tf"
const clusterFileName = "cluster.tf"

// Configure creates Terraform directories and modules as well as a Terraform backend file if
// provided by the user.
func Configure(assetDir, renderedBackend, renderedPlatform string) error {
	if err := prepareTerraformDirectoryAndModules(assetDir); err != nil {
		return fmt.Errorf("failed to create required terraform directory: %w", err)
	}

	terraformRootDir := GetTerraformRootDir(assetDir)
	// Create backend file only if the backend rendered string isn't empty.
	if len(strings.TrimSpace(renderedBackend)) > 0 {
		data := fmt.Sprintf("terraform {%s}\n", renderedBackend)
		file := filepath.Join(terraformRootDir, backendFileName)

		if err := createTerraformFile(file, data); err != nil {
			return fmt.Errorf("failed to create backend configuration file: %w", err)
		}
	}

	file := filepath.Join(terraformRootDir, clusterFileName)
	if err := createTerraformFile(file, renderedPlatform); err != nil {
		return fmt.Errorf("failed to create cluster configuration file: %w", err)
	}

	return nil
}

// createTerraformFile creates the terraform file with the contents provided.
func createTerraformFile(path, data string) error {
	var err error

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file '%q', got: %w", path, err)
	}

	defer func() {
		if ferr := f.Close(); ferr != nil {
			err = ferr
		}
	}()

	if _, err = f.WriteString(data); err != nil {
		return fmt.Errorf("failed to write content to file '%q', got: %w", path, err)
	}

	if err = f.Sync(); err != nil {
		return fmt.Errorf("failed to flush data to file '%q', got: %w", path, err)
	}

	return err
}

// prepareTerraformDirectoryAndModules creates a Terraform directory and downloads required modules.
func prepareTerraformDirectoryAndModules(assetDir string) error {
	terraformModuleDir := filepath.Join(assetDir, "lokomotive-kubernetes")
	if err := prepareLokomotiveTerraformModuleAt(terraformModuleDir); err != nil {
		return err
	}

	terraformRootDir := filepath.Join(assetDir, "terraform")
	if err := prepareTerraformRootDir(terraformRootDir); err != nil {
		return err
	}

	return nil
}

// GetTerraformRootDir gets the Terraform directory path.
func GetTerraformRootDir(assetDir string) string {
	return filepath.Join(assetDir, "terraform")
}

// prepareTerraformRootDir creates a directory named path including all
// required parents.
// An error is returned if the directory already exists.
func prepareTerraformRootDir(path string) error {
	if err := os.MkdirAll(path, 0750); err != nil {
		return fmt.Errorf("failed to create terraform assets directory at '%s', got: %w", path, err)
	}

	return nil
}

// prepareLokomotiveTerraformModuleAt creates a directory named path
// including all required parents and puts the Lokomotive Kubernetes
// terraform module sources into path.
// An error is returned if the directory already exists.
//
// The terraform sources are loaded either from data embedded in the
// lokoctl binary or from the filesystem, depending on whether the
// LOKOCTL_USE_FS_ASSETS environment variable was specified.
func prepareLokomotiveTerraformModuleAt(path string) error {
	walk := walkers.CopyingWalker(path, 0750)
	if err := assets.Assets.WalkFiles("/lokomotive-kubernetes", walk); err != nil {
		return errors.Wrap(err, "failed to walk assets")
	}

	return nil
}
