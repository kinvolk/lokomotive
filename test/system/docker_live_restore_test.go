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

// +build aws aws_edge baremetal equinixmetal
// +build e2e

package system_test

import (
	"context"
	"testing"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// Define manifests as YAML and then unmarshal it to a Go struct so that it is easier to write and
// debug, as it can be copy-pasted to a file and applied manually.
const dockerLiveRestoreOnNodesDSManifest = `apiVersion: apps/v1
kind: DaemonSet
metadata:
  generateName: test-live-restore-
spec:
  selector:
    matchLabels:
      name: test-live-restore
  template:
    metadata:
      labels:
        name: test-live-restore
    spec:
      tolerations:
      - key: ""
        operator: Exists
      terminationGracePeriodSeconds: 1
      initContainers:
      - name: test-live-restore
        image: quay.io/kinvolk/docker:stable
        imagePullPolicy: IfNotPresent
        command:
        - /bin/sh
        args:
        - -c
        - 'docker info -f {{.LiveRestoreEnabled}} | grep true'
        volumeMounts:
        - name: docker-socket
          mountPath: /var/run/docker.sock

      containers:
      - name: wait
        image: quay.io/kinvolk/docker:stable
        imagePullPolicy: IfNotPresent
        command:
        - /bin/sh
        args:
        - -c
        - sleep infinity
      volumes:
      - name: docker-socket
        hostPath:
          path: /var/run/docker.sock
`

const dockerLogOptsOnNodesDSManifest = `apiVersion: apps/v1
kind: DaemonSet
metadata:
  generateName: test-log-opts-
spec:
  selector:
    matchLabels:
      name: test-log-opts
  template:
    metadata:
      labels:
        name: test-log-opts
    spec:
      tolerations:
      - key: ""
        operator: Exists
      terminationGracePeriodSeconds: 1
      containers:
      - name: wait
        image: quay.io/kinvolk/jq:1.6
        imagePullPolicy: IfNotPresent
        command:
        - /bin/sh
        args:
        - -c
        - sleep infinity

      initContainers:
      - name: test-log-opts-1
        image: quay.io/kinvolk/jq:1.6
        imagePullPolicy: IfNotPresent
        command:
        - /bin/sh
        args:
        - -c
        - jq '."log-opts"."max-size"' /etc/docker/daemon.json | grep 100m
        volumeMounts:
        - name: docker-config
          mountPath: /etc/docker/daemon.json
      - name: test-log-opts-2
        image: quay.io/kinvolk/jq:1.6
        imagePullPolicy: IfNotPresent
        command:
        - /bin/sh
        args:
        - -c
        - jq '."log-opts"."max-file"' /etc/docker/daemon.json | grep 3
        volumeMounts:
        - name: docker-config
          mountPath: /etc/docker/daemon.json
      volumes:
      - name: docker-config
        hostPath:
          path: /etc/docker/daemon.json
`

func TestDockerNodeConfig(t *testing.T) {
	tcs := []struct {
		name   string
		config string
	}{
		{"Is_Live_Restore_Enabled", dockerLiveRestoreOnNodesDSManifest},
		{"Compare_Log_Options", dockerLogOptsOnNodesDSManifest},
	}

	namespace := "kube-system"
	client := testutil.CreateKubeClient(t)
	dsClient := client.AppsV1().DaemonSets(namespace)

	for _, tc := range tcs {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ds := &appsv1.DaemonSet{}
			if err := yaml.Unmarshal([]byte(tc.config), ds); err != nil {
				t.Fatalf("Failed unmarshaling manifest: %v", err)
			}

			ds, err := dsClient.Create(context.TODO(), ds, metav1.CreateOptions{})
			if err != nil {
				t.Fatalf("Failed to create DaemonSet: %v", err)
			}

			testutil.WaitForDaemonSet(t, client, namespace, ds.Name, testutil.RetryInterval, testutil.Timeout)

			t.Cleanup(func() {
				if err := dsClient.Delete(context.TODO(), ds.Name, metav1.DeleteOptions{}); err != nil {
					t.Logf("Failed to delete DaemonSet: %v", err)
				}
			})
		})
	}
}
