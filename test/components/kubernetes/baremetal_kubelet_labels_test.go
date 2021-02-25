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

// +build baremetal
// +build e2e

package kubernetes //nolint:testpackage

import (
	"context"
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestBaremetalNodeHasLabels(t *testing.T) {
	client := testutil.CreateKubeClient(t)

	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Errorf("could not list nodes: %v", err)
	}

	if len(nodes.Items) == 0 {
		t.Fatalf("no worker nodes found")
	}

	for _, node := range nodes.Items {
		labels := node.GetLabels()

		fmt.Println(labels)
		fmt.Println(node.Name)
	}
}
