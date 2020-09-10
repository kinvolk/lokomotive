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
	"os"
	"path/filepath"
	"strings"

	"github.com/kardianos/osext"
)

type fsAssets struct {
	assetsDir string
}

func newFsAssets(dir string) *fsAssets {
	if dir == "" {
		execDir, err := osext.ExecutableFolder()
		if err != nil {
			panic("Unable to get a directory of an executable for assets")
		}
		dir = filepath.Join(execDir, "assets")
	}
	return &fsAssets{
		assetsDir: dir,
	}
}

func (a *fsAssets) WalkFiles(location string, cb walkFunc) error {
	relativeLocation := strings.TrimLeft(location, string(os.PathSeparator))
	assetsLocation := filepath.Join(a.assetsDir, relativeLocation)
	return filepath.Walk(assetsLocation, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return cb(path, info, nil, err)
		}
		if info.IsDir() {
			return nil
		}
		relPath, relErr := filepath.Rel(assetsLocation, path)
		if relErr != nil {
			return cb(relPath, info, nil, relErr)
		}
		file, err := os.Open(path)
		return cb(relPath, info, file, err)
	})
}
