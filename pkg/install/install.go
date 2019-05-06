package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gobuffalo/packd"
	packr "github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"

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
// The terraform sources are loaded from a packr2 box: depending on how
// the binary was built, that means loaded from the binary (`make build`)
// or from disk (`make build-slim`).
func PrepareLokomotiveTerraformModuleAt(path string, provider string) error {
	pathExists, err := util.PathExists(path)
	if err != nil {
		return errors.Wrapf(err, "failed to stat path %q: %v", path, err)
	}
	if pathExists {
		return fmt.Errorf("directory at %q exists already - aborting", path)
	}
	// String formatting needs to be done inline, otherwise packr complains
	box := packr.New(fmt.Sprintf("lokomotive-kubernetes/%s", provider), fmt.Sprintf("../../lokomotive-kubernetes/%s", provider))
	walk := func(fileName string, file packd.File) error {
		fileInfo, err := file.FileInfo()
		if err != nil {
			return errors.Wrap(err, "failed to extract file info")
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

		if _, err := io.Copy(targetFile, file); err != nil {
			return errors.Wrap(err, "failed to write file")
		}
		return nil
	}

	if err := box.Walk(walk); err != nil {
		return errors.Wrap(err, "failed to walk box")
	}
	return nil
}
