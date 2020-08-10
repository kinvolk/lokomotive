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

package install

import (
	"os"

	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/util"
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
		return nil
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
	walk := assets.CopyingWalker(path, 0755)
	if err := assets.Assets.WalkFiles("/lokomotive-kubernetes", walk); err != nil {
		return errors.Wrap(err, "failed to walk assets")
	}
	return nil
}
