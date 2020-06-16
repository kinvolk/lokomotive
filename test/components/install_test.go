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

// +build aws aws_edge packet
// +build e2e

package components

import (
	"testing"

	"github.com/hashicorp/hcl/v2"

	"github.com/kinvolk/lokomotive/pkg/components"
	_ "github.com/kinvolk/lokomotive/pkg/components/flatcar-linux-update-operator"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestInstallIdempotent(t *testing.T) {
	configHCL := `
component "flatcar-linux-update-operator" {}
  `

	n := "flatcar-linux-update-operator"

	c, err := components.Get(n)
	if err != nil {
		t.Fatalf("failed getting component: %v", err)
	}

	body, diagnostics := util.GetComponentBody(configHCL, n)
	if diagnostics != nil {
		t.Fatalf("Error getting component body: %v", diagnostics)
	}

	diagnostics = c.LoadConfig(body, &hcl.EvalContext{})
	if diagnostics.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diagnostics)
	}

	k := testutil.Kubeconfig(t)

	if err := util.InstallComponent(c, k); err != nil {
		t.Fatalf("Installing component as release should succeed, got: %v", err)
	}

	if err := util.InstallComponent(c, k); err != nil {
		t.Fatalf("Installing component twice as release should succeed, got: %v", err)
	}
}
