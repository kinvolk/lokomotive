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
	"io/ioutil"
	"os"
	"path/filepath"
)

func CopyingWalker(path string, newDirPerms os.FileMode) WalkFunc {
	return func(fileName string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			return fmt.Errorf("error while walking at %q: %v", fileName, err)
		}

		fileName = filepath.Join(path, fileName)

		if err := os.MkdirAll(filepath.Dir(fileName), newDirPerms); err != nil {
			return fmt.Errorf("failed to create dir: %v", err)
		}

		return writeFile(fileName, r)
	}
}

func DumpingWalker(contentsMap map[string]string, allowedExts ...string) WalkFunc {
	var extsMap map[string]struct{}

	if len(allowedExts) > 0 {
		extsMap = make(map[string]struct{}, len(allowedExts))
		for _, ext := range allowedExts {
			extsMap[ext] = struct{}{}
		}
	}
	return func(fileName string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			return fmt.Errorf("error while walking at %q: %v", fileName, err)
		}

		if extsMap != nil {
			if _, ok := extsMap[filepath.Ext(fileName)]; !ok {
				return nil
			}
		}

		contents, err := ioutil.ReadAll(r)
		if err != nil {
			return fmt.Errorf("failed to read %q: %v", fileName, err)
		}

		contentsMap[fileName] = string(contents)
		return nil
	}
}

// writeFile writes data from given io.Reader to the file and makes sure, that
// this is the only content stored in the file.
func writeFile(p string, r io.Reader) error {
	// TODO: If we start packing binaries, make sure they have executable bit set.
	f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open target file %s: %w", p, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("failed writing to file %s: %w", p, err)
	}

	return nil
}
