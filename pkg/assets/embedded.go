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

//go:generate go run assets_generate.go

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shurcooL/httpfs/vfsutil"
)

type embeddedAssets struct {
	fs http.FileSystem
}

func newEmbeddedAssets() *embeddedAssets {
	return &embeddedAssets{
		fs: vfsgenAssets,
	}
}

func (a *embeddedAssets) WalkFiles(location string, cb walkFunc) error {
	return vfsutil.WalkFiles(a.fs, location, func(filePath string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			return cb(filePath, fileInfo, r, err)
		}
		if fileInfo.IsDir() {
			return nil
		}
		relPath, relErr := filepath.Rel(location, filePath)
		if relErr != nil {
			return cb(relPath, fileInfo, nil, relErr)
		}
		return cb(relPath, fileInfo, r, err)
	})
}
