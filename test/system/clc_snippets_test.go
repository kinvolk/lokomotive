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

// +build aws aws_edge packet
// +build e2e

//nolint:dupl
package system_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

const (
	fileName = "clc_snippet_hello"
)

// Define manifest as YAML and then unmarshal it to a Go struct so that it is easier to
// write and debug, as it can be copy-pasted to a file and applied manually.
const clcSnippetsAddedToNodesDSManifest = `apiVersion: apps/v1
kind: DaemonSet
metadata:
  generateName: test-clc-snippets-
spec:
  selector:
    matchLabels:
      name: test-clc-snippets
  template:
    metadata:
      labels:
        name: test-clc-snippets
    spec:
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      terminationGracePeriodSeconds: 1
      containers:
      - name: test-clc-snippets
        image: ubuntu
        command: ["bash"]
        volumeMounts:
        - name: opt
          mountPath: /opt
          readOnly: true
      volumes:
      - name: opt
        hostPath:
          path: /opt
`

func TestFileCreatedByCLCSnippetExistsOnNodes(t *testing.T) {
	t.Parallel()

	namespace := "kube-system"

	client := testutil.CreateKubeClient(t)

	ds := &appsv1.DaemonSet{}
	if err := yaml.Unmarshal([]byte(clcSnippetsAddedToNodesDSManifest), ds); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	// Set the right arguments from the manifest with the correct fileName
	ds.Spec.Template.Spec.Containers[0].Args = []string{
		"-c",
		fmt.Sprintf("test -e /opt/%s && exec tail -f /dev/null", fileName),
	}

	ds, err := client.AppsV1().DaemonSets(namespace).Create(context.TODO(), ds, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create DaemonSet: %v", err)
	}

	testutil.WaitForDaemonSet(t, client, namespace, ds.ObjectMeta.Name, time.Second*5, time.Minute*5)

	t.Cleanup(func() {
		if err := client.AppsV1().DaemonSets(namespace).Delete(
			context.TODO(), ds.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed to remove DaemonSet: %v", err)
		}
	})
}
