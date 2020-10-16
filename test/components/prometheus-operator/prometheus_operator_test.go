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

// +build aws aws_edge packet aks
// +build e2e

package prometheusoperator

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

const (
	namespace         = "monitoring"
	grafanaDeployment = "prometheus-operator-grafana"
)

func TestPrometheusOperatorDeployment(t *testing.T) {
	client := testutil.CreateKubeClient(t)

	deployments := []string{
		"prometheus-operator-operator",
		"prometheus-operator-kube-state-metrics",
		grafanaDeployment,
	}

	for _, deployment := range deployments {
		deployment := deployment
		t.Run("deployment", func(t *testing.T) {
			t.Parallel()

			testutil.WaitForDeployment(t, client, namespace, deployment, testutil.RetryInterval, testutil.TimeoutSlow)
		})
	}

	statefulSets := []string{
		"alertmanager-prometheus-operator-alertmanager",
		"prometheus-prometheus-operator-prometheus",
	}

	for _, statefulset := range statefulSets {
		statefulset := statefulset
		t.Run(fmt.Sprintf("statefulset %s", statefulset), func(t *testing.T) {
			t.Parallel()
			replicas := 1

			testutil.WaitForStatefulSet(t, client, namespace, statefulset, replicas, testutil.RetryInterval, testutil.TimeoutSlow) //nolint:lll
		})
	}

	testutil.WaitForDaemonSet(t, client, namespace, "prometheus-operator-prometheus-node-exporter", testutil.RetryInterval, testutil.TimeoutSlow) //nolint:lll
}

//nolint:funlen
func TestGrafanaLoadsEnvVars(t *testing.T) {
	kubeconfig := testutil.KubeconfigPath(t)

	t.Logf("using KUBECONFIG=%s", kubeconfig)

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatalf("failed building rest client: %v", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("failed creating new clientset: %v", err)
	}

	// We will wait until the Grafana Pods are up and running so we don't have to reimplement wait logic again.
	testutil.WaitForDeployment(t, client, namespace, grafanaDeployment, testutil.RetryInterval, testutil.TimeoutSlow)

	// Get grafana deployment object so that we can get the corresponding pod.
	deploy, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), grafanaDeployment, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			t.Fatalf("deployment %s not found", grafanaDeployment)
		}

		t.Fatalf("error looking up for deployment %s: %v", grafanaDeployment, err)
	}

	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deploy.Spec.Selector),
	})
	if err != nil {
		t.Fatalf("could not list pods for the deployment %q: %v", grafanaDeployment, err)
	}

	if len(podList.Items) == 0 {
		t.Fatalf("grafana pods not found")
	}

	// Exec into the pod.
	pod := podList.Items[0]
	containerName := "grafana"
	searchEnvVar := "LOKOMOTIVE_VERY_SECRET_PASSWORD"

	req := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(namespace).
		SubResource("exec").Param("container", containerName)

	req.VersionedParams(&v1.PodExecOptions{
		Command:   []string{"env"},
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
		Container: containerName,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		t.Fatalf("could not exec: %v", err)
	}

	var stdout, stderr bytes.Buffer

	if err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	}); err != nil {
		t.Fatalf("exec stream failed: %v", err)
	}

	containerErr := strings.TrimSpace(stderr.String())
	if len(containerErr) > 0 {
		t.Fatalf("error from container: %v", containerErr)
	}

	containerOutput := strings.TrimSpace(stdout.String())
	if !strings.Contains(containerOutput, searchEnvVar) {
		t.Fatalf("required env var %q not found in following env vars:\n\n%s\n", searchEnvVar, containerOutput)
	}
}

func TestPrometheusOperatorPVC(t *testing.T) {
	pvcs := []string{
		"data-alertmanager-prometheus-operator-alertmanager-0",
		"data-prometheus-prometheus-operator-prometheus-0",
	}

	client := testutil.CreateKubeClient(t)

	for _, pvc := range pvcs {
		pvc := pvc
		t.Run(pvc, func(t *testing.T) {
			t.Parallel()
			testutil.WaitForPVCToBeBound(t, client, namespace, pvc, testutil.RetryInterval, testutil.TimeoutSlow)
		})
	}
}
