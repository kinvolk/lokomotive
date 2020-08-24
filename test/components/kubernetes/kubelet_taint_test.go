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

// +build aws
// +build e2e

package kubernetes_test

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestNodeHasTaints(t *testing.T) {
	client := testutil.CreateKubeClient(t)

	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: "testing.io=yes,roleofnode=testing",
	})
	if err != nil {
		t.Errorf("could not list nodes: %v", err)
	}

	if len(nodes.Items) == 0 {
		t.Fatalf("no worker nodes found")
	}

	for _, items := range nodes.Items {
		for _, taint := range items.Spec.Taints {
			taintKey := taint.Key
			taintValue := taint.Value
			taintEffect := taint.Effect

			if taintKey != "nodeType" || taintValue != "storage" || taintEffect != "NoSchedule" {
				t.Fatalf("expected taint: %s, got: %s=%s:%s", "nodeType=storage:NoSchedule", taintKey, taintValue, taintEffect)
			}
		}
	}
}
