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

// +build packet_fluo
// +build e2e

package rook_test

import (
	"context"
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

const namespace = "rook"

func TestRookCephDeployment(t *testing.T) {
	t.Parallel()

	client := testutil.CreateKubeClient(t)
	testCases := []struct {
		deployment string
	}{
		{"csi-cephfsplugin-provisioner"},
		{"csi-rbdplugin-provisioner"},
		{"rook-ceph-mgr-a"},
		{"rook-ceph-mon-a"},
		{"rook-ceph-mon-b"},
		{"rook-ceph-mon-c"},
		{"rook-ceph-osd-0"},
		{"rook-ceph-osd-1"},
		{"rook-ceph-osd-2"},
		{"rook-ceph-osd-3"},
		{"rook-ceph-osd-4"},
		{"rook-ceph-osd-5"},
		{"rook-ceph-osd-6"},
		{"rook-ceph-osd-7"},
		{"rook-ceph-osd-8"},
		{"rook-ceph-tools"},
	}

	for _, test := range testCases {
		test := test
		t.Run(fmt.Sprintf("rook-ceph deployment:%s", test.deployment), func(t *testing.T) {
			t.Parallel()
			testutil.WaitForDeployment(t, client, namespace, test.deployment, testutil.RetryInterval, testutil.Timeout)
		})
	}
}

// TestRookCephCrashCollector tests the crash collector deployments. This is a separate function
// because the names of the deployments depend upon the node names. So this test first extracts the
// deployment names based on the labels and then verifies if the pods are up.
func TestRookCephCrashCollector(t *testing.T) {
	t.Parallel()

	client := testutil.CreateKubeClient(t)
	testCases := []string{}

	if err := wait.PollImmediate(testutil.RetryInterval, testutil.Timeout, func() (done bool, err error) {
		items, err := client.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: "crashcollector=crash",
		})
		if err != nil {
			return false, fmt.Errorf("Listing crashcollector deployments: %w", err)
		}

		// If we get zero deployments listed that means, the pods are not up yet. It takes some time
		// for them to be up. So we try listing them again.
		if len(items.Items) == 0 {
			t.Log("No deployments with label 'crashcollector=crash.")

			return false, nil
		}

		for _, itm := range items.Items {
			testCases = append(testCases, itm.Name)
		}

		return true, nil
	}); err != nil {
		t.Fatalf("Finding names of crashcollector deployments: %v", err)
	}

	for _, test := range testCases {
		test := test
		t.Run(fmt.Sprintf("rook-ceph crashcollector deployment:%s", test), func(t *testing.T) {
			t.Parallel()
			testutil.WaitForDeployment(t, client, namespace, test, testutil.RetryInterval, testutil.Timeout)
		})
	}
}

func TestRookCephDaemonset(t *testing.T) {
	t.Parallel()

	client := testutil.CreateKubeClient(t)
	testCases := []struct {
		daemonset string
	}{
		{"csi-cephfsplugin"},
		{"csi-rbdplugin"},
		{"rook-discover"},
	}

	for _, test := range testCases {
		test := test
		t.Run(fmt.Sprintf("rook-ceph daemonset:%s", test.daemonset), func(t *testing.T) {
			t.Parallel()
			testutil.WaitForDaemonSet(t, client, namespace, test.daemonset, testutil.RetryInterval, testutil.Timeout)
		})
	}
}
