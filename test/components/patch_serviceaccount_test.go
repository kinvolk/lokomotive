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

// +build aws aws_edge equinixmetal
// +build e2e

package components_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

const (
	contextTimeout = 10
)

func TestDisableAutomountServiceAccountToken(t *testing.T) {
	client := testutil.CreateKubeClient(t)

	ns, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatal("listing namespaces: %w", err)
	}

	for _, namespace := range ns.Items {
		if !testutil.IsUserNamespace(namespace.Name) {
			name := namespace.Name
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if err := wait.PollImmediate(
					testutil.RetryInterval, testutil.Timeout, checkDefaultServiceAccountPatch(client, name),
				); err != nil {
					t.Fatalf("%v", err)
				}
			})
		}
	}
}

func checkDefaultServiceAccountPatch(client kubernetes.Interface, ns string) wait.ConditionFunc {
	return func() (done bool, err error) {
		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout*time.Second)
		defer cancel()

		sa, err := client.CoreV1().ServiceAccounts(ns).Get(ctx, "default", metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("getting service account: %v", err)
		}

		automountServiceAccountToken := *sa.AutomountServiceAccountToken

		if automountServiceAccountToken {
			return false, fmt.Errorf("failed for namespace %q. Expected %v got %v", ns, false, automountServiceAccountToken)
		}

		return true, nil
	}
}
