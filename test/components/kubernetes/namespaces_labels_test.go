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

// +build aws aws_edge packet aks baremetal
// +build e2e

package kubernetes_test

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kinvolk/lokomotive/internal"
	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestAllNamespacesHaveNameLabels(t *testing.T) {
	client := testutil.CreateKubeClient(t)

	namespaces, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing namespaces: %v", err)
	}

	for _, namespace := range namespaces.Items {
		name := namespace.ObjectMeta.Name
		labels := namespace.ObjectMeta.Labels

		t.Logf("testing namespace %q", name)
		// Do not consider the namespaces generated during tests.
		// i.e with "test-" prefix.
		if testutil.IsTestNamespace(name) {
			continue
		}

		// AKS creates this namespace which we don't label it hence ignore it.
		if testutil.IsPlatformSupported(t, []testutil.Platform{testutil.PlatformAKS}) && name == "calico-system" {
			continue
		}

		if name != labels[internal.NamespaceLabelKey] {
			t.Fatalf("expected %q, got: %q", name, labels[internal.NamespaceLabelKey])
		}
	}
}
