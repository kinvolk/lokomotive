package destroy

import (
	"path/filepath"

	"github.com/kinvolk/lokoctl/pkg/terraform"
	"github.com/kinvolk/lokoctl/pkg/util"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

// TODO: Consider this for remote storage.
// ExecuteTerraformDestroy executes terraform destroy -auto-approve to delete the cluster.
// Currently this assumes that the terraform state file is stored locally.
func ExecuteTerraformDestroy(assetDirectory string) error {
	assetDir, err := homedir.Expand(assetDirectory)
	if err != nil {
		return err
	}

	terraformRootDir := filepath.Join(assetDir, "terraform")
	terraformStateFile := filepath.Join(terraformRootDir, "terraform.tfstate")

	// Check if terraform state file exists
	// This check is performed in case there is a misconfigured config file
	// pointing to a different assests directory.
	pathExists, err := util.PathExists(terraformStateFile)
	if err != nil {
		return errors.Wrapf(err, "failed to stat path %q: %v", terraformStateFile, err)
	}

	if !pathExists {
		return errors.Errorf("terraform state file not found at directory at %q - aborting", terraformRootDir)
	}

	return terraform.Destroy(terraformRootDir)
}
