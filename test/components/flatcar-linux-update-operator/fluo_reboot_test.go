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

// +build equinixmetal_fluo
// +build disruptivee2e

package fluo_test

import (
	"context"
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

//nolint:funlen
func TestNodeCanReachIdleState(t *testing.T) {
	client := testutil.CreateKubeClient(t)
	nodeLabel := "fluo-test-pool=true"

	// Select a node from the general worker pool.
	nodesList, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
		LabelSelector: nodeLabel,
	})
	if err != nil {
		t.Fatalf("Listing nodes with label %q: %v", nodeLabel, err)
	}

	nodes := nodesList.Items
	if len(nodes) < 1 {
		t.Fatalf("Wanted one or more nodes with label %q, found none.", nodeLabel)
	}

	// Select a node to remove annotation.
	chosenNode := nodes[0]
	t.Logf("Chosen node to reboot: %s", chosenNode.Name)

	// Remove annotation that disables node reboot by FLUO. This test assumes that the node has
	// annotation set.
	annotation := "flatcar-linux-update.v1.flatcar-linux.net/reboot-paused"

	// This annotation should be set on the node during the cluster creation, if it is not set then
	// something changed in the cluster setup process in the CI.
	if _, ok := chosenNode.Annotations[annotation]; ok {
		t.Logf("Annotation %q found on the node, removing.", annotation)

		// Remove the annotation that disables node reboot.
		delete(chosenNode.Annotations, annotation)

		if _, err := client.CoreV1().Nodes().Update(context.Background(), &chosenNode, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("Removing node annotation %q: %v", annotation, err)
		}
	}

	// Wait for the FLUO to add following annotation key value pair to the chosen node. This may
	// involve node reboot if the OS is outdated.
	// "flatcar-linux-update.v1.flatcar-linux.net/status": "UPDATE_STATUS_IDLE"
	statusAnnotationKey := "flatcar-linux-update.v1.flatcar-linux.net/status"
	statusAnnotationVal := "UPDATE_STATUS_IDLE"

	if err := wait.PollImmediate(testutil.RetryInterval, testutil.TimeoutSlow, func() (done bool, err error) {
		node, err := client.CoreV1().Nodes().Get(context.Background(), chosenNode.Name, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("Getting node %q: %v", chosenNode.Name, err)
		}

		val, ok := node.Annotations[statusAnnotationKey]
		if !ok {
			return false, nil
		}

		// Test passed since the annotation is added by the FLUO to the nodes.
		if val == statusAnnotationVal {
			return true, nil
		}

		return false, nil
	}); err != nil {
		t.Fatalf("Waiting for the node to add annotation '%s:%s' %v",
			statusAnnotationKey, statusAnnotationVal, err)
	}
}
