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

//+build e2e

package terraform_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/kinvolk/lokomotive/pkg/terraform"
)

func executor(t *testing.T) *terraform.Executor {
	tmpDir, err := ioutil.TempDir("", "lokoctl-tests-")
	if err != nil {
		t.Fatalf("Creating tmp dir should succeed, got: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Removing directory %q: %v", tmpDir, err)
		}
	})

	conf := terraform.Config{
		Verbose:    false,
		WorkingDir: tmpDir,
	}

	ex, err := terraform.NewExecutor(conf)
	if err != nil {
		t.Fatalf("Creating new executor should succeed, got: %v", err)
	}

	return ex
}

func TestExecuteCheckErrors(t *testing.T) {
	ex := executor(t)

	if err := ex.Apply(nil); err == nil {
		t.Fatalf("Applying on empty directory should fail")
	}
}

func TestOutputIncludeKeyInError(t *testing.T) {
	ex := executor(t)

	k := "foo"
	o := ""

	err := ex.Output(k, &o)
	if err == nil {
		t.Fatalf("Output should fail on non existing installation")
	}

	if !strings.Contains(err.Error(), k) {
		t.Fatalf("Error message should contain key, got: %v", err)
	}
}
