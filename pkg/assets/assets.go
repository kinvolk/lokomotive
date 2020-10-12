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

// Package assets handles Lokomotive assets. Operations such as storing files as binary data and
// extracting files from memory to disk belong in this package.
package assets

import (
	"fmt"
	"io"
	"os"

	// Currently pkg/assets/assets_generate.go file is ignored from regular
	// builds and it's only used when running 'go generate'.
	//
	// Because of that, external dependencies of assets_generate.go are not
	// normally added to go.mod and go.sum files, unless you run 'go generate'.
	// But then 'go mod tidy' removes them again.
	//
	// Those anonymous imports ensure, that those external dependencies
	// will always be pulled into go.mod file.
	//
	// The alternative would be to refactor the function to actually use the
	// dependencies here, but then we have an exported function, which
	// shouldn't really be exported.
	//
	// Also see https://github.com/golang/go/issues/25922 and https://github.com/golang/go/issues/29516
	// for more details about this issue.
	_ "github.com/prometheus/alertmanager/pkg/modtimevfs"
	_ "github.com/shurcooL/httpfs/union"
	_ "github.com/shurcooL/vfsgen"
)

const (
	// ControlPlaneSource is the asset source directory for control plane charts.
	ControlPlaneSource = "/charts/control-plane"
	// ComponentsSource is the asset source directory for components.
	ComponentsSource = "/charts/components"
	// TerraformModulesSource is the asset source directory for Terraform modules.
	TerraformModulesSource = "/terraform-modules"
)

type walkFunc func(fileName string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error

type assetsIface interface {
	// WalkFiles calls cb for every regular file within path.
	//
	// Usually, fileName passed to the cb will be relative to
	// path. But in case of error, it is possible that it will
	// not.be relative. Also, in case of error, fileInfo or r may
	// be nil.
	WalkFiles(path string, cb walkFunc) error
}

func get() assetsIface {
	if value, found := os.LookupEnv("LOKOCTL_USE_FS_ASSETS"); found {
		return newFsAssets(value)
	}

	return newEmbeddedAssets()
}

// Extract recursively extracts the assets at src into the directory dst. If dst doesn't exist, the
// directory is created including any missing parents.
//
// The assets are read either from data embedded in the binary or from the filesystem, depending on
// whether the LOKOCTL_USE_FS_ASSETS environment variable is set.
func Extract(src, dst string) error {
	walk := copyingWalker(dst, 0700)
	if err := get().WalkFiles(src, walk); err != nil {
		return fmt.Errorf("failed to walk assets: %v", err)
	}

	return nil
}
