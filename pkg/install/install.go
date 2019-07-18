package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/assets"
	"github.com/kinvolk/lokoctl/pkg/util"
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
// The terraform sources are loaded from data embedded in the lokoctl
// binary.
func PrepareLokomotiveTerraformModuleAt(path string) error {
	pathExists, err := util.PathExists(path)
	if err != nil {
		return errors.Wrapf(err, "failed to stat path %q: %v", path, err)
	}
	if pathExists {
		return fmt.Errorf("directory at %q exists already - aborting", path)
	}
	walk := func(fileName string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			return errors.Wrapf(err, "error during walking at %q", fileName)
		}

		fileName = filepath.Join(path, fileName)

		if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
			return errors.Wrap(err, "failed to create dir")
		}

		targetFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, fileInfo.Mode())
		if err != nil {
			return errors.Wrap(err, "failed to open target file")
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, r); err != nil {
			return errors.Wrap(err, "failed to write file")
		}
		return nil
	}

	if err := assets.Assets.WalkFiles("/lokomotive-kubernetes", walk); err != nil {
		return errors.Wrap(err, "failed to walk assets")
	}
	return nil
}
