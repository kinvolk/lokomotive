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
// +build disruptivee2e

package components_test

import (
	"testing"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	_ "github.com/kinvolk/lokomotive/pkg/components/web-ui"
	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestDeleteNamespaceMultipleRelease(t *testing.T) {
	n := "web-ui"

	c, err := components.Get(n)
	if err != nil {
		t.Fatalf("failed getting component: %v", err)
	}

	k := testutil.Kubeconfig(t)

	if err := util.UninstallComponent(c, k, true); err == nil {
		t.Fatalf("Deleting component should fail.")
	}
}
