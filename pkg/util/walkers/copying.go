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

package walkers

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/assets"
)

func CopyingWalker(path string, newDirPerms os.FileMode) assets.WalkFunc {
	return func(fileName string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			return errors.Wrapf(err, "error during walking at %q", fileName)
		}

		fileName = filepath.Join(path, fileName)

		if err := os.MkdirAll(filepath.Dir(fileName), newDirPerms); err != nil {
			return errors.Wrap(err, "failed to create dir")
		}

		// TODO: If we start packing binaries, make sure they have executable bit set.
		targetFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return errors.Wrap(err, "failed to open target file")
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, r); err != nil {
			return errors.Wrap(err, "failed to write file")
		}
		return nil
	}
}
