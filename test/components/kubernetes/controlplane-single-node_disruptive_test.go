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

// +build packet baremetal
// +build disruptivee2e

package kubernetes_test

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestControlplaneComponentsDeploymentsCanBeGracefullyUpdated(t *testing.T) {
	client := testutil.CreateKubeClient(t)
	deployClient := client.AppsV1().Deployments(namespace)

	for c, waitTime := range components() {
		c := c
		waitTime := waitTime

		t.Run(c, func(t *testing.T) {
			deploy, err := deployClient.Get(context.TODO(), c, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("Getting Deployment %q: %v", c, err)
			}

			// Use current time to have different value on each test run.
			deploy.Spec.Template.Annotations["update-test"] = time.Now().String()

			if _, err := deployClient.Update(context.TODO(), deploy, metav1.UpdateOptions{}); err != nil {
				t.Fatalf("Updating Deployment %q: %v", c, err)
			}

			// Wait a bit to let Kubernetes trigger pod updates.
			time.Sleep(waitTime)

			testutil.WaitForDeployment(t, client, namespace, c, testutil.RetryInterval, testutil.Timeout)
		})
	}
}
