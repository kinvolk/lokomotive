package terraform

import (
	"fmt"
	"github.com/kinvolk/lokoctl/pkg/install"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const backendFileName = "backend.tf"

// Configure calls supporting menthods to create Teraform directories, modules
// and create terraform backend file if user has provided.
func Configure(assetDir, renderedBackend string) error {
	if err := PrepareTerraformDirectoryAndModules(assetDir); err != nil {
		return errors.Wrapf(err, "Failed to create required terraform directory")
	}

	// Create backend file only if there content in the backend rendered string.
	if len(strings.TrimSpace(renderedBackend)) <= 0 {
		return nil
	}

	if err := CreateTerraformBackendFile(assetDir, renderedBackend); err != nil {
		return errors.Wrapf(err, "Failed to create backend configuration file")
	}

	return nil
}

// PrepareTerraformDirectoryAndModules creates terraform directory and downloads
// required modules.
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

// GetTerraformRootDir gets the terraform directory path
func GetTerraformRootDir(assetDir string) string {
	return filepath.Join(assetDir, "terraform")
}

// CreateTerraformBackendFile creates the terraform backend configuration file
func CreateTerraformBackendFile(assetDir, data string) error {
	backendString := fmt.Sprintf("terraform {%s}\n", data)
	terraformRootDir := GetTerraformRootDir(assetDir)
	path := filepath.Join(terraformRootDir, backendFileName)
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %q", path)
	}
	defer f.Close()

	if _, err = f.WriteString(backendString); err != nil {
		return errors.Wrapf(err, "failed to write to backend file %q", path)
	}

	if err = f.Sync(); err != nil {
		return errors.Wrapf(err, "failed to flush data to file %q", path)
	}

	return nil
}
