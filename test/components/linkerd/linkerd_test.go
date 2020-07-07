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

package linkerd_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/yaml"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

const (
	namespace = "linkerd"
)

func TestLinkerdDeployment(t *testing.T) {
	t.Parallel()

	deployments := []string{
		"linkerd-controller",
		"linkerd-destination",
		"linkerd-grafana",
		"linkerd-identity",
		"linkerd-prometheus",
		"linkerd-proxy-injector",
		"linkerd-sp-validator",
		"linkerd-tap",
		"linkerd-web",
	}

	client := testutil.CreateKubeClient(t)

	for _, d := range deployments {
		d := d
		t.Run(d, func(t *testing.T) {
			t.Parallel()

			testutil.WaitForDeployment(t, client, namespace, d, retryInterval, timeout)
		})
	}
}

const (
	// nolint: lll
	podConfig = `
apiVersion: v1
kind: Pod
metadata:
  generateName: linkerd-check-test-
spec:
  restartPolicy: Never
  initContainers:
  - image: fedora
    name: download-kubectl
    command:
    - bash
    args:
    - -c
    - 'curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x ./kubectl && cp ./kubectl /data/'
    volumeMounts:
    - name: download-dir
      mountPath: /data
  - image: fedora
    name: download-linkerd
    command:
    - bash
    args:
    - -c
    - 'curl -L https://run.linkerd.io/install | sh && cp -L /root/.linkerd2/bin/linkerd /data/linkerd'
    volumeMounts:
    - name: download-dir
      mountPath: /data
  containers:
  - image: fedora
    name: linkerd-cli
    env:
    - name: KUBECONFIG
      value: /root/.kube/config
    command:
    - bash
    args:
    - -c
    - 'mv /data/* /usr/local/bin/ && kubectl version -o yaml && linkerd version && linkerd check'
    volumeMounts:
    - name: kubeconfig
      mountPath: /root/.kube
    - name: download-dir
      mountPath: /data
  volumes:
  - name: kubeconfig
    secret:
      secretName: kubeconfig
  - name: download-dir
    emptyDir: {}
`

	namespaceConfig = `
apiVersion: v1
kind: Namespace
metadata:
  generateName: linkerd-check-test-
`

	retryInterval = 5 * time.Second
	timeout       = 9 * time.Minute
)

// nolint: funlen
func TestLinkerdCheck(t *testing.T) {
	t.Parallel()

	client := testutil.CreateKubeClient(t).CoreV1()
	nsclient := client.Namespaces()

	// Parsing
	ns := &corev1.Namespace{}
	if err := yaml.Unmarshal([]byte(namespaceConfig), ns); err != nil {
		t.Fatalf("failed to unmarshal namespace manifest: %v", err)
	}

	pod := &corev1.Pod{}
	if err := yaml.Unmarshal([]byte(podConfig), pod); err != nil {
		t.Fatalf("failed to unmarshal pod manifest: %v", err)
	}

	// Creation
	ns, err := nsclient.Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed creating Namespace: %v", err)
	}

	t.Cleanup(func() {
		if err := nsclient.Delete(context.TODO(), ns.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed removing Namespace: %v", err)
		}
	})

	_, err = client.Secrets(ns.Name).Create(context.TODO(), getKubeconfigSecret(t), metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed creating ConfigMap: %v", err)
	}

	podsClient := client.Pods(ns.Name)

	// Retry pod creation. This might fail if Linkerd is not ready yet and some requests might fail.
	if err := wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		pod, err = podsClient.Create(context.TODO(), pod, metav1.CreateOptions{})
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
		p, err := podsClient.Get(context.TODO(), pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		phase = p.Status.Phase
		if phase == corev1.PodSucceeded || phase == corev1.PodFailed {
			return true, nil
		}

		return false, nil
	}); err != nil {
		t.Errorf("error while waiting for the pod: %v", err)
	}

	// Since pod failed print the logs
	if phase == corev1.PodFailed {
		t.Error("pod failed with following error:")

		podLogs, err := podsClient.GetLogs(pod.Name, &corev1.PodLogOptions{}).Stream(context.TODO())
		if err != nil {
			t.Fatalf("failed to get logs from pod: %v", err)
		}
		defer podLogs.Close() // nolint: errcheck

		buf := new(bytes.Buffer)
		if _, err = io.Copy(buf, podLogs); err != nil {
			t.Fatalf("failed to copy logs into buffer: %v", err)
		}

		t.Fatal(buf.String())
	}
}

func getKubeconfigSecret(t *testing.T) *corev1.Secret {
	d, err := ioutil.ReadFile(testutil.KubeconfigPath(t))
	if err != nil {
		t.Fatalf("failed to read kubeconfig: %v", err)
	}

	kubeconfigData := string(d)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "kubeconfig"},
		StringData: map[string]string{
			"config": kubeconfigData,
		},
	}
}
