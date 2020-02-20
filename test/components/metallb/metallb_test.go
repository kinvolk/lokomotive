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

// +build packet
// +build e2e

package metallb

import (
	"testing"
	"time"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestMetalLBDeployment(t *testing.T) {
	namespace := "metallb-system"

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	t.Run("speaker daemonset", func(t *testing.T) {
		t.Parallel()
		daemonset := "speaker"

		testutil.WaitForDaemonSet(t, client, namespace, daemonset, time.Second*5, time.Minute*5)
	})

	t.Run("controller deployment", func(t *testing.T) {
		t.Parallel()
		deployment := "controller"
		replicas := 1

		testutil.WaitForDeployment(t, client, namespace, deployment, replicas, time.Second*5, time.Minute*5)
		t.Logf("Found required replicas: %d", replicas)
	})
}
