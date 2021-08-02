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

// +build aws
// +build e2e

//nolint:dupl
package system_test

import (
	"context"
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

// expectedConntrackMaxCLC is an expected value returned by
// net.netfilter.nf_conntrack_max on nodes with 'conntrack-modified: true' label set.
//
// 50000 is arbitrary number close to 32768*2CPU on AWS machine type we're testing.
// It's good if this value is not divisible by 32768 which is a default value in kube-proxy,
// so it is easy to distinguish.
const expectedConntrackMaxCLC = 50000

// Define manifest as YAML and then unmarshal it to a Go struct so that it is easier to
// write and debug, as it can be copy-pasted to a file and applied manually.
const conntrackMaxCLCDSManifest = `apiVersion: apps/v1
kind: DaemonSet
metadata:
  generateName: test-conntrack-
spec:
  selector:
    matchLabels:
      name: test-conntrack
  template:
    metadata:
      labels:
        name: test-conntrack
    spec:
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      terminationGracePeriodSeconds: 1
      containers:
      - name: test-conntrack
        image: ubuntu
        command: ["bash"]
`

func TestCLCSnippetWithConntrackSettingIsRespected(t *testing.T) {
	t.Parallel()

	namespace := "kube-system"

	client := testutil.CreateKubeClient(t)

	ds := &appsv1.DaemonSet{}
	if err := yaml.Unmarshal([]byte(conntrackMaxCLCDSManifest), ds); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	// Set the right arguments from the manifest with the correct fileName.
	ds.Spec.Template.Spec.Containers[0].Args = []string{
		"-c",
		fmt.Sprintf("test %d -eq $(sysctl net.netfilter.nf_conntrack_max | cut -d' ' -f3) && exec tail -f /dev/null",
			expectedConntrackMaxCLC),
	}
	ds.Spec.Template.Spec.NodeSelector = map[string]string{"conntrack-modified": "true"}

	ds, err := client.AppsV1().DaemonSets(namespace).Create(context.TODO(), ds, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create DaemonSet: %v", err)
	}

	testutil.WaitForDaemonSet(t, client, namespace, ds.ObjectMeta.Name, testutil.RetryInterval, testutil.Timeout)

	t.Cleanup(func() {
		if err := client.AppsV1().DaemonSets(namespace).Delete(
			context.TODO(), ds.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed to remove DaemonSet: %v", err)
		}
	})
}

func TestNodesWithoutCLCSnippetRetainDefaultConntrackMax(t *testing.T) {
	t.Parallel()

	namespace := "kube-system"

	client := testutil.CreateKubeClient(t)

	ds := &appsv1.DaemonSet{}
	if err := yaml.Unmarshal([]byte(conntrackMaxCLCDSManifest), ds); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	// Set the right arguments from the manifest with the correct fileName.
	ds.Spec.Template.Spec.Containers[0].Args = []string{
		"-c",
		fmt.Sprintf("test %d -ne $(sysctl net.netfilter.nf_conntrack_max | cut -d' ' -f3) && exec tail -f /dev/null",
			expectedConntrackMaxCLC),
	}
	ds.Spec.Template.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "conntrack-modified",
								Operator: corev1.NodeSelectorOpDoesNotExist,
							},
						},
					},
				},
			},
		},
	}

	ds, err := client.AppsV1().DaemonSets(namespace).Create(context.TODO(), ds, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create DaemonSet: %v", err)
	}

	testutil.WaitForDaemonSet(t, client, namespace, ds.ObjectMeta.Name, testutil.RetryInterval, testutil.Timeout)

	t.Cleanup(func() {
		if err := client.AppsV1().DaemonSets(namespace).Delete(
			context.TODO(), ds.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed to remove DaemonSet: %v", err)
		}
	})
}
