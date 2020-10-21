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

package gangway

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestGangwayDeployment(t *testing.T) {
	namespace := "gangway"

	client := testutil.CreateKubeClient(t)

	t.Run("deployment", func(t *testing.T) {
		t.Parallel()
		deployment := "gangway"

		testutil.WaitForDeployment(t, client, namespace, deployment, testutil.RetryInterval, testutil.Timeout)
	})
}

func TestGangwayServiceAccount(t *testing.T) {
	namespace := "gangway"
	deployment := "gangway"
	expectedServiceAccountName := "gangway"

	client := testutil.CreateKubeClient(t)

	testutil.WaitForDeployment(t, client, namespace, deployment, testutil.RetryInterval, testutil.Timeout)

	deploy, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), deployment, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Couldn't find gangway deployment")
	}

	if deploy.Spec.Template.Spec.ServiceAccountName != expectedServiceAccountName {
		t.Fatalf("Expected serviceAccountName %q, got: %q",
			deploy.Spec.Template.Spec.ServiceAccountName,
			deploy.Spec.Template.Spec.ServiceAccountName)
	}
}
