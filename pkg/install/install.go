package install

import (
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/assets"
	"github.com/kinvolk/lokoctl/pkg/util"
	"github.com/kinvolk/lokoctl/pkg/util/walkers"
)

// PrepareTerraformRootDir creates a directory named path including all
// required parents.
// An error is returned if the directory already exists.
func PrepareTerraformRootDir(path string) error {
	pathExists, err := util.PathExists(path)
	if err != nil {
		return errors.Wrapf(err, "failed to stat path %q: %v", path, err)
	}
	if pathExists {
		return fmt.Errorf("terraform assets directory at %q exists already - aborting", path)
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return errors.Wrapf(err, "failed to create terraform assets directory at: %s", path)
	}
	return nil
}

// PrepareLokomotiveTerraformModuleAt creates a directory named path
// including all required parents and puts the Lokomotive Kubernetes
// terraform module sources into path.
// An error is returned if the directory already exists.
//
// The terraform sources are loaded either from data embedded in the
// lokoctl binary or from the filesystem, depending on whether the
// LOKOCTL_USE_FS_ASSETS environment variable was specified.
func PrepareLokomotiveTerraformModuleAt(path string) error {
	pathExists, err := util.PathExists(path)
	if err != nil {
		return errors.Wrapf(err, "failed to stat path %q: %v", path, err)
	}
	if pathExists {
		return fmt.Errorf("directory at %q exists already - aborting", path)
	}
	walk := walkers.CopyingWalker(path, 0755)
	if err := assets.Assets.WalkFiles("/lokomotive-kubernetes", walk); err != nil {
		return errors.Wrap(err, "failed to walk assets")
	}
	return nil
}
