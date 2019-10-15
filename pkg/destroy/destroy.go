package destroy

import (
	"path/filepath"

	"github.com/kinvolk/lokoctl/pkg/terraform"
	"github.com/kinvolk/lokoctl/pkg/util"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

// ExecuteTerraformDestroy executes terraform destroy -auto-approve to delete the cluster.
func ExecuteTerraformDestroy(assetDirectory string) error {
	assetDir, err := homedir.Expand(assetDirectory)
	if err != nil {
		return err
	}

	terraformRootDir := filepath.Join(assetDir, "terraform")

	// Check if terraform dir exists
	pathExists, err := util.PathExists(terraformRootDir)
	if err != nil {
		return errors.Wrapf(err, "failed to stat path %q: %v", terraformRootDir, err)
	}

	if !pathExists {
		return errors.Errorf("terraform state file not found at directory at %q - aborting", terraformRootDir)
	}

	return terraform.Destroy(terraformRootDir)
}
