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

// +build aws aws_edge packet aks
// +build e2e

package nodeproblemdetector_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	daemonSetName = "node-problem-detector"
	namespace     = "kube-system"
)

func TestNodeProblemDetectorDaemonSet(t *testing.T) {
	client := testutil.CreateKubeClient(t)

	testutil.WaitForDaemonSet(t, client, namespace, daemonSetName, testutil.RetryInterval, testutil.TimeoutSlow)
}

//nolint: funlen
func TestNodeProblemDetectorNodeEvent(t *testing.T) {
	t.Parallel()

	client := testutil.CreateKubeClient(t).CoreV1()
	nsclient := client.Namespaces()

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: testutil.TestNamespace("node-problem-detector"),
		},
	}

	privileged := true
	// Pod config.
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "node-problem-detector-test-",
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			Containers: []corev1.Container{
				{
					Name:  "nginx",
					Image: "nginx",
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
					Command: []string{"sh"},
					Args: []string{"-c", "echo 'kernel: BUG: " +
						"unable to handle kernel NULL pointer " +
						"dereference at TESTING_IS_COOL' >> /dev/kmsg"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "kmsg",
							MountPath: "/dev/kmsg",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "kmsg",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/dev/kmsg",
						},
					},
				},
			},
		},
	}

	// Creation of namespace.
	ns, err := nsclient.Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("creating namespace: %v", err)
	}

	t.Cleanup(func() {
		if err := nsclient.Delete(context.TODO(), ns.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed removing namespace: %v", err)
		}
	})

	podsClient := client.Pods(ns.Name)

	// Retry pod creation. This might fail if node-problem-detector is not ready yet and some requests might fail.
	if err := wait.PollImmediate(testutil.RetryInterval, testutil.TimeoutSlow, func() (done bool, err error) {
		pod, err = podsClient.Create(context.TODO(), pod, metav1.CreateOptions{})
		if err != nil {
			t.Logf("retrying pod creation, failed with: %v", err)

			return false, nil
		}

		return true, nil
	}); err != nil {
		t.Fatalf("creating pod: %v", err)
	}

	phase := corev1.PodUnknown

	if err := wait.PollImmediate(testutil.RetryInterval, testutil.TimeoutSlow, func() (done bool, err error) {
		p, err := podsClient.Get(context.TODO(), pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, fmt.Errorf("couldn't get the pod %q: %w", pod.Name, err)
		}

		// Check for event.
		events, err := client.Events("default").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return false, fmt.Errorf("couldn't list the events %w", err)
		}

		phase = p.Status.Phase
		if phase == corev1.PodSucceeded || phase == corev1.PodFailed {
			for _, v := range events.Items {
				if v.Reason == "KernelOops" && strings.Contains(v.Message, "TESTING_IS_COOL") {
					return true, nil
				}
			}
		}

		return false, nil
	}); err != nil {
		t.Errorf("waiting for the pod: %v", err)
	}

	// Since pod failed print the logs.
	if phase == corev1.PodFailed {
		t.Error("pod failed with following error:")

		err := testutil.PrintPodsLogs(t, podsClient, &metav1.LabelSelector{MatchLabels: pod.Labels})
		if err != nil {
			t.Errorf("printing error logs failed: %v", err)
		}
	}
}
