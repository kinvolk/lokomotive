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

// +build aws aws_edge baremetal packet
// +build e2e

package system // nolint:testpackage

import (
	"context"
	"crypto/sha512"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

const (
	// sshKeyEnv is an environment variable from which we take the SSH key
	// to be used in testing.
	sshKeyEnv = "PUB_KEY"
)

// authorizedKeysSHA512 takes a list of public SSH keys as an argument and calculates their
// SHA512, thus mimicking the behavior of the following command: `sha512sum /home/core/.ssh/authorized_keys`.
// The format of authorized_keys should match what the 'update-ssh-keys' command generates on Flatcar.
func authorizedKeysSHA512(keys []string) string {
	plain := ""

	for _, key := range keys {
		// Trim the key before appending to make sure it does not contain
		// a newline.
		plain += fmt.Sprintf("%s\n", strings.TrimSpace(key))
	}

	// Add one more extra line to the file, this is what Ignition does.
	plain += "\n"

	return fmt.Sprintf("%x", sha512.Sum512([]byte(plain)))
}

// Define manifest as YAML and then unmarshal it to a Go struct so that it is easier to
// write and debug, as it can be copy-pasted to a file and applied manually.
const noExtraSSHKeysOnNodesDSManifest = `apiVersion: apps/v1
kind: DaemonSet
metadata:
  generateName: test-ssh-keys-
spec:
  selector:
    matchLabels:
      name: test-ssh-keys
  template:
    metadata:
      labels:
        name: test-ssh-keys
    spec:
      tolerations:
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      terminationGracePeriodSeconds: 1
      containers:
      - name: test-ssh-keys
        image: ubuntu
        command: ["bash"]
        volumeMounts:
        - name: ssh
          mountPath: /home/core/.ssh
          readOnly: true
      volumes:
      - name: ssh
        hostPath:
          path: /home/core/.ssh
`

func TestNoExtraSSHKeysOnNodes(t *testing.T) {
	t.Parallel()

	key := os.Getenv(sshKeyEnv)
	if key == "" {
		t.Skipf("%q environment variable not set", sshKeyEnv)
	}

	namespace := "kube-system"

	client := testutil.CreateKubeClient(t)

	sum := authorizedKeysSHA512([]string{key})

	ds := &appsv1.DaemonSet{}
	if err := yaml.Unmarshal([]byte(noExtraSSHKeysOnNodesDSManifest), ds); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	// Set the right arguments from the manifest with the desired SHA512 sum.
	ds.Spec.Template.Spec.Containers[0].Args = []string{
		"-c",
		fmt.Sprintf("sha512sum --status -c <(echo %s /home/core/.ssh/authorized_keys) && exec tail -f /dev/null", sum), //nolint:lll
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
