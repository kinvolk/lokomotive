package terraform

import (
	"github.com/kinvolk/lokoctl/pkg/install"
	"path/filepath"
)

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

func GetTerraformRootDir(assetDir string) string {
	return filepath.Join(assetDir, "terraform")
}
