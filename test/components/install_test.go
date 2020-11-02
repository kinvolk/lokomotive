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

	fluo "github.com/kinvolk/lokomotive/pkg/components/flatcar-linux-update-operator"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestInstallIdempotent(t *testing.T) {
	c := fluo.NewConfig()

	k := testutil.Kubeconfig(t)

	if err := util.InstallComponent(c, k); err != nil {
		t.Fatalf("Installing component as release should succeed, got: %v", err)
	}

	if err := util.InstallComponent(c, k); err != nil {
		t.Fatalf("Installing component twice as release should succeed, got: %v", err)
	}
}
