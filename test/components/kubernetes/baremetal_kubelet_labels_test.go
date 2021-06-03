// Copyright 2021 The Lokomotive Authors
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

// +build baremetal
// +build e2e

package kubernetes_test

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkglabels "k8s.io/apimachinery/pkg/labels"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func Test_Baremetal_NodeSpecificLabels(t *testing.T) {
	// This map is copied from file ci/baremetal/baremetal-cluster.lokocfg.envsubst
	// the field is `node_specific_labels`.
	// The node name keys are constructed as follows: <cluster-name>-<controller/worker>-<index>.
	// The labels include the labels common to every worker node derived from `labels` and
	// labels specific to the node.
	nodeNamesAndExpectedLabelsInCI := map[string]map[string]string{
		"mercury-controller-0": {
			"testkey":                       "testvalue",
			"node.kubernetes.io/master":     "",
			"node.kubernetes.io/controller": "true",
		},
		"mercury-worker-0": {
			"ingressnode": "yes",
			"testing.io":  "yes",
			"roleofnode":  "testing",
		},
		"mercury-worker-1": {
			"storagenode": "yes",
			"testing.io":  "yes",
			"roleofnode":  "testing",
		},
	}

	client := testutil.CreateKubeClient(t)

	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("could not list nodes: %v", err)
	}

	if len(nodes.Items) != len(nodeNamesAndExpectedLabelsInCI) {
		t.Errorf("expected %d worker nodes, got: %d", len(nodeNamesAndExpectedLabelsInCI), len(nodes.Items))
	}

	for nodeName, labelsMap := range nodeNamesAndExpectedLabelsInCI {
		labels := pkglabels.Set(labelsMap).String()

		nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
			LabelSelector: labels,
		})
		if err != nil {
			t.Errorf("expected node with name %q not found: %v", nodeName, err)
		}

		// Currently we assume that the labels added to the node in CI are unique to the worker node.
		// This check confirms that the expected set of labels as per CI configuration were
		// present on the correct node.
		if len(nodes.Items) == 1 && nodes.Items[0].Name != nodeName {
			t.Errorf("expected node %q to have the labels %s", nodes.Items[0].Name, labels)
		}
	}
}
