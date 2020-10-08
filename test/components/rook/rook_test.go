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

package rook

import (
	"testing"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestRookDeployment(t *testing.T) {
	namespace := "rook"

	client := testutil.CreateKubeClient(t)

	t.Run("deployment", func(t *testing.T) {
		t.Parallel()
		deployment := "rook-ceph-operator"

		testutil.WaitForDeployment(t, client, namespace, deployment, testutil.RetryInterval, testutil.Timeout)
	})
}
