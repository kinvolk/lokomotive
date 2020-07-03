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

// +build aws aks aws_edge packet
// +build e2e

package istio_test

import (
	"testing"
	"time"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

const (
	retryInterval = 3 * time.Second
	timeout       = 7 * time.Minute
)

func TestIstioDeployments(t *testing.T) {
	deployments := []struct {
		Namespace  string
		Deployment string
	}{
		{
			Namespace:  "istio-operator",
			Deployment: "istio-operator",
		},
		{
			Namespace:  "istio-system",
			Deployment: "istiod",
		},
	}

	client := testutil.CreateKubeClient(t)

	for _, d := range deployments {
		d := d
		t.Run(d.Deployment, func(t *testing.T) {
			t.Parallel()

			testutil.WaitForDeployment(t, client, d.Namespace, d.Deployment, retryInterval, timeout)
		})
	}
}
