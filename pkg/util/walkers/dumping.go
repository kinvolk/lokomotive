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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/assets"
)

func DumpingWalker(contentsMap map[string]string, allowedExts ...string) assets.WalkFunc {
	var extsMap map[string]struct{}

	if len(allowedExts) > 0 {
		extsMap = make(map[string]struct{}, len(allowedExts))
		for _, ext := range allowedExts {
			extsMap[ext] = struct{}{}
		}
	}
	return func(fileName string, fileInfo os.FileInfo, r io.ReadSeeker, err error) error {
		if err != nil {
			return errors.Wrapf(err, "error during walking at %q", fileName)
		}

		if extsMap != nil {
			if _, ok := extsMap[filepath.Ext(fileName)]; !ok {
				return nil
			}
		}

		contents, err := ioutil.ReadAll(r)
		if err != nil {
			return errors.Wrapf(err, "failed to read %q", fileName)
		}

		contentsMap[fileName] = string(contents)
		return nil
	}
}
