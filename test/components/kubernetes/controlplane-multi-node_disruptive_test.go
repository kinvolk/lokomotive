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
// +build disruptivee2e

package kubernetes_test

import (
	"context"
	"testing"
	"time"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestControlplaneComponentsDaemonSetsCanBeGracefullyUpdated(t *testing.T) {
	client := testutil.CreateKubeClient(t)
	dsClient := client.AppsV1().DaemonSets(namespace)

	components := components()
	components["kube-proxy"] = testutil.RetryInterval

	for c, waitTime := range components {
		c := c
		waitTime := waitTime

		t.Run(c, func(t *testing.T) {
			ds, err := dsClient.Get(context.TODO(), c, metav1.GetOptions{})
			if err != nil {
				t.Fatalf("Getting DaemonSet %q: %v", c, err)
			}

			// Use current time to have different value on each test run.
			ds.Spec.Template.Annotations["update-test"] = time.Now().String()

			if _, err := dsClient.Update(context.TODO(), ds, metav1.UpdateOptions{}); err != nil {
				t.Fatalf("Updating DaemonSet %q: %v", c, err)
			}

			// Wait a bit to let Kubernetes trigger pod updates.
			time.Sleep(waitTime)

			testutil.WaitForDaemonSet(t, client, namespace, c, testutil.RetryInterval, testutil.Timeout)
		})
	}
}
