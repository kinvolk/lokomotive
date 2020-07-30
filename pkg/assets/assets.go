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

package assets

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/alertmanager/pkg/modtimevfs"
	"github.com/shurcooL/httpfs/union"
	"github.com/shurcooL/vfsgen"
)

const (
	// ControlPlaneSource is the asset source directory for control plane charts.
	ControlPlaneSource = "/charts/control-plane"
	// ComponentsSource is the asset source directory for components.
	ComponentsSource = "/charts/components"
	// TerraformModulesSource is the asset source directory for Terraform modules.
	TerraformModulesSource = "/lokomotive-kubernetes"
)

type WalkFunc func(fileName string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error

type AssetsIface interface {
	// WalkFiles calls cb for every regular file within path.
	//
	// Usually, fileName passed to the cb will be relative to
	// path. But in case of error, it is possible that it will
	// not.be relative. Also, in case of error, fileInfo or r may
	// be nil.
	WalkFiles(path string, cb WalkFunc) error
}

var Assets AssetsIface

func init() {
	Assets = newEmbeddedAssets()
	if value, found := os.LookupEnv("LOKOCTL_USE_FS_ASSETS"); found {
		Assets = newFsAssets(value)
	}
}

// Generate function wraps vfsgen.Generate function.
// Additionally to vfsgen.Generate, it also takes map of directories,
// where key represents path in the assets and a value represents path
// to the assets directory in local filesystem (which should be relative).
//
// This function also resets modification time for every file, so creating a new copy
// of code does not trigger changes in all asset files.
func Generate(fileName string, packageName string, variableName string, dirs map[string]string) error {
	ufs := make(map[string]http.FileSystem)
	for assetsPath, fsPath := range dirs {
		ufs[assetsPath] = http.Dir(fsPath)
	}
	u := union.New(ufs)
	fs := modtimevfs.New(u, time.Unix(1, 0))
	return vfsgen.Generate(fs, vfsgen.Options{
		Filename:     fileName,
		PackageName:  packageName,
		VariableName: variableName,
	})
}

// Extract recursively extracts the assets at src into the directory dst. If dst doesn't exist, the
// directory is created including any missing parents.
//
// The assets are read either from data embedded in the binary or from the filesystem, depending on
// whether the LOKOCTL_USE_FS_ASSETS environment variable is set.
func Extract(src, dst string) error {
	walk := CopyingWalker(dst, 0755)
	if err := Assets.WalkFiles(src, walk); err != nil {
		return fmt.Errorf("failed to walk assets: %v", err)
	}

	return nil
}
