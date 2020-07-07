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

// +build packet aws aws_edge
// +build e2e

package calico_test

import (
	"context"
	"os"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/yaml"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

const (
	// Define manifest as YAML and then unmarshal it to Go struct, so it is easier to
	// write and debug, as it can be copy-pasted to a YAML file and applied manually.
	metadataAccessPodManifest = `
apiVersion: v1
kind: Pod
metadata:
  generateName: metadata-access-test-
spec:
  restartPolicy: Never
  containers:
  - image: alpine
    name: foo
    command:
    - wget
    args:
    - -O
    - /dev/null
    - --timeout
    - "5"
`

	retryInterval = 1 * time.Second
	timeout       = 5 * time.Minute
)

func TestNoMetadataAccessRandomPod(t *testing.T) { //nolint:funlen
	t.Parallel()

	client := testutil.CreateKubeClient(t).CoreV1()
	nsclient := client.Namespaces()

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: testutil.TestNamespace("metadata-access"),
		},
	}

	p := &corev1.Pod{}
	if err := yaml.Unmarshal([]byte(metadataAccessPodManifest), p); err != nil {
		t.Fatalf("failed to unmarshal pod manifest: %v", err)
	}

	ns, err := nsclient.Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed creating Namespace: %v", err)
	}

	metadataAddress := ""
	platform := os.Getenv("PLATFORM")

	switch platform {
	case testutil.PlatformPacket, testutil.PlatformPacketARM:
		metadataAddress = "https://metadata.packet.net/metadata"
	case testutil.PlatformAWS, testutil.PlatformAWSEdge:
		metadataAddress = "http://169.254.169.254/latest/meta-data"
	}

	if metadataAddress == "" {
		t.Fatalf("Platform %q not supported", platform)
	}

	// Set the right address to query, which is platform dependent.
	p.Spec.Containers[0].Args = append(p.Spec.Containers[0].Args, metadataAddress)

	podsclient := client.Pods(ns.ObjectMeta.Name)

	// Retry pod creation. This might fail if Linkerd is not ready yet and some requests might fail.
	if err := wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		p, err = podsclient.Create(context.TODO(), p, metav1.CreateOptions{})
		if err != nil {
			t.Logf("retrying pod creation, failed with: %v", err)

			return false, nil
		}

		return true, nil
	}); err != nil {
		t.Fatalf("error while trying to create the pod: %v", err)
	}

	phase := corev1.PodUnknown

	if err := wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		p, err := podsclient.Get(context.TODO(), p.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		phase = p.Status.Phase

		if phase == corev1.PodFailed || phase == corev1.PodSucceeded {
			return true, nil
		}

		return false, nil
	}); err != nil {
		t.Errorf("error while waiting for the pod: %v", err)
	}

	if phase == corev1.PodSucceeded {
		t.Fatalf("Pods should not have access to %q", metadataAddress)
	}

	t.Cleanup(func() {
		if err := nsclient.Delete(context.TODO(), ns.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed removing Namespace: %v", err)
		}
	})
}
