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

// +build aws aws-edge packet aks
// +build e2e

package certmanager

import (
	"fmt"
	"testing"
	"time"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

const namespace = "cert-manager"

func TestCertManagerDeployments(t *testing.T) {
	client := testutil.CreateKubeClient(t)

	testCases := []struct {
		deployment string
	}{
		{
			deployment: "cert-manager",
		},
		{
			deployment: "cert-manager-cainjector",
		},
		{
			deployment: "cert-manager-webhook",
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("cert-manager deployment:%s", test.deployment), func(t *testing.T) {
			t.Parallel()
			testutil.WaitForDeployment(t, client, namespace, test.deployment, time.Second*5, time.Minute*5)
		})
	}
}
