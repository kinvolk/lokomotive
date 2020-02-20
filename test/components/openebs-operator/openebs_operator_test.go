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

// +build aws packet
// +build e2e

package openebsoperator

import (
	"testing"
	"time"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestOpenEBSOperatorDeployment(t *testing.T) {
	namespace := "openebs"

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	deployments := []string{
		"openebs-provisioner",
		"openebs-localpv-provisioner",
		"openebs-admission-server",
		"openebs-ndm-operator",
		"openebs-snapshot-operator",
		"maya-apiserver",
	}

	for _, deployment := range deployments {
		t.Run("deployment", func(t *testing.T) {
			t.Parallel()

			testutil.WaitForDeployment(t, client, namespace, deployment, time.Second*5, time.Minute*5)
		})
	}
}
