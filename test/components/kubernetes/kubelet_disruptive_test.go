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
// +build disruptive-e2e

package kubernetes

import (
	"testing"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestSelfHostedKubeletLabels(t *testing.T) {
	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	// List all the nodes and then delete a node that is not controller.
	nodes, err := client.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: "node.kubernetes.io/node=",
	})
	if err != nil {
		t.Errorf("could not list nodes: %v", err)
	}
	if len(nodes.Items) == 0 {
		t.Fatalf("no worker nodes found")
	}
	chosenNode := nodes.Items[0].Name

	// Delete the chosen node.
	if err = client.CoreV1().Nodes().Delete(chosenNode, &metav1.DeleteOptions{}); err != nil {
		t.Errorf("could not delete the node %s: %v", chosenNode, err)
	}

	retryInterval := time.Second * 5
	timeout := time.Minute * 5
	// Wait for the node to come up.
	if err = wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		node, err := client.CoreV1().Nodes().Get(chosenNode, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				t.Logf("waiting for node %s to be available", chosenNode)
				return false, nil
			}
			return false, err
		}

		// Match the expected labels to the labels found on the node.
		expectedLabels := map[string]string{
			"testing.io": "yes",
			"roleofnode": "testing",
		}
		for k, v := range expectedLabels {
			if node.Labels[k] != v {
				t.Errorf("label %q:%q not found on node %q", k, v, chosenNode)
			}
		}

		return true, nil
	}); err != nil {
		t.Errorf("error waiting for the node %s: %v", chosenNode, err)
	}
}
