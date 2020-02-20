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

package tools

import (
	"os"
	"os/exec"

	k8serrs "k8s.io/apimachinery/pkg/util/errors"
)

func isBinaryInPath(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return err
	}
	return nil
}

// InstallerBinaries is used from sub-command `install` to check if all the
// needed binaries are in required place
func InstallerBinaries() error {
	var errs []error

	if err := isBinaryInPath("terraform"); err != nil {
		errs = append(errs, err)
	}

	// see if the terraform plugin is in path
	plugin := "$HOME/.terraform.d/plugins/terraform-provider-ct"
	if _, err := os.Stat(os.ExpandEnv(plugin)); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return k8serrs.NewAggregate(errs)
	}
	return nil
}
