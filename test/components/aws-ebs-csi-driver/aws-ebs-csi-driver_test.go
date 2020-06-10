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

// +build aws aws_edge
// +build e2e

// nolint:testpackage
package awsebscsidriver

import (
	"fmt"
	"testing"
	"time"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

const namespace = "kube-system"

func TestCSIDriverDeployments(t *testing.T) {
	client := testutil.CreateKubeClient(t)

	testCases := []struct {
		deployment string
	}{
		{
			deployment: "ebs-csi-controller",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(fmt.Sprintf("aws-ebs-csi-driver deployment:%s", test.deployment), func(t *testing.T) {
			t.Parallel()
			testutil.WaitForDeployment(t, client, namespace, test.deployment, time.Second*5, time.Minute*5)
		})
	}
}
