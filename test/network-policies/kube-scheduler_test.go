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

package networkpolicies //nolint:testpackage

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

const (
	retryInterval     = time.Second * 5
	timeout           = time.Minute * 10
	kubeSchedulerPort = "10251"

	namespaceManifest = `
apiVersion: v1
kind: Namespace
metadata:
  generateName: kube-scheduler-network-policy-test-
`

	podManifest = `
apiVersion: v1
kind: Pod
metadata:
  generateName: kube-scheduler-network-policy-test-
spec:
  restartPolicy: Never
  containers:
  - image: alpine
    name: kubeschedulertest
    command:
    - wget
    args:
    - -O
    - /dev/null
    - --timeout
    - "5"
`
)

func TestDenyLivenessProbeFromOtherPodsNotHavingLabel(t *testing.T) {
	t.Parallel()

	client := testutil.CreateKubeClient(t)
	// kube-system namespace object.
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
		},
	}
	// Unmarshal Pod manifest.
	pod := &corev1.Pod{}
	if err := yaml.Unmarshal([]byte(podManifest), pod); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	// Get the IP address of the kube-scheduler pod.
	kubeSchedulerPodIP := getKubeSchedulerPodIP(t, client)
	kubeSchedulerAddress := fmt.Sprintf("%s:%s/healthz", kubeSchedulerPodIP, kubeSchedulerPort)
	// Add the kube-scheduler pod IP address to the pod command.
	pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, kubeSchedulerAddress)

	podsclient := client.CoreV1().Pods(ns.ObjectMeta.Name)

	// Create pod.
	pod, err := podsclient.Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}

	phase := WaitForPodAndGetPodPhase(t, client, ns, pod)
	if phase == corev1.PodSucceeded {
		t.Fatalf("Expected kube-scheduler readiness probe to fail: %q", kubeSchedulerAddress)
	}
}
func TestAllowLivenessProbeFromControlPlanePods(t *testing.T) {
	t.Parallel()

	client := testutil.CreateKubeClient(t)
	// kube-system namespace object.
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
		},
	}
	// Unmarshal Pod manifest.
	pod := &corev1.Pod{}
	if err := yaml.Unmarshal([]byte(podManifest), pod); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}
	// Add metrics label to the pod.
	pod.ObjectMeta.Labels = map[string]string{
		"tier": "control-plane",
	}

	// Get the IP address of the kube-scheduler pod.
	kubeSchedulerPodIP := getKubeSchedulerPodIP(t, client)
	kubeSchedulerAddress := fmt.Sprintf("%s:%s/healthz", kubeSchedulerPodIP, kubeSchedulerPort)
	// Add the kube-scheduler pod IP address to the pod command.
	pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, kubeSchedulerAddress)

	podsclient := client.CoreV1().Pods(ns.ObjectMeta.Name)

	// Create pod.
	pod, err := podsclient.Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}

	phase := WaitForPodAndGetPodPhase(t, client, ns, pod)
	if phase != corev1.PodSucceeded {
		t.Fatalf("Expected kube-scheduler liveness probe to succeeded: %q", kubeSchedulerAddress)
	}
}

func TestAllowMetricsFromPodAndNamespaceWithRequiredLabels(t *testing.T) {
	t.Parallel()

	client := testutil.CreateKubeClient(t)
	nsclient := client.CoreV1().Namespaces()
	// Unmarshal namespace manifest.
	ns := &corev1.Namespace{}
	if err := yaml.Unmarshal([]byte(namespaceManifest), ns); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}
	// Add metrics label to the namespace.
	ns.ObjectMeta.Labels = map[string]string{
		"lokomotive.kinvolk.io/scrape-metrics": "true",
	}
	// Create namespace.
	ns, err := nsclient.Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create namespace: %v", err)
	}
	// Unmarshal Pod manifest.
	pod := &corev1.Pod{}
	if err := yaml.Unmarshal([]byte(podManifest), pod); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}
	// Add metrics label to the pod.
	pod.ObjectMeta.Labels = map[string]string{
		"lokomotive.kinvolk.io/scrape-metrics": "true",
	}

	// Get the IP address of the kube-scheduler pod.
	kubeSchedulerPodIP := getKubeSchedulerPodIP(t, client)
	kubeSchedulerAddress := fmt.Sprintf("%s:%s/metrics", kubeSchedulerPodIP, kubeSchedulerPort)
	// Add the kube-scheduler pod IP address to the pod command.
	pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, kubeSchedulerAddress)

	podsclient := client.CoreV1().Pods(ns.ObjectMeta.Name)

	// Create pod.
	pod, err = podsclient.Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}

	phase := WaitForPodAndGetPodPhase(t, client, ns, pod)
	if phase != corev1.PodSucceeded {
		t.Fatalf("Expected scraping kube-scheduler metrics to succeeded: %q", kubeSchedulerAddress)
	}

	t.Cleanup(func() {
		if err := nsclient.Delete(context.TODO(), ns.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed removing Namespace: %v", err)
		}
	})
}

func TestDenyMetricsFromNamespaceHavingRequiredLabelsButNotPod(t *testing.T) {
	t.Parallel()
	client := testutil.CreateKubeClient(t)
	nsclient := client.CoreV1().Namespaces()
	// Unmarshal namespace manifest.
	ns := &corev1.Namespace{}
	if err := yaml.Unmarshal([]byte(namespaceManifest), ns); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}
	// Add metrics label to the namespace.
	ns.ObjectMeta.Labels = map[string]string{
		"lokomotive.kinvolk.io/scrape-metrics": "true",
	}
	// Create namespace.
	ns, err := nsclient.Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create namespace: %v", err)
	}
	// Unmarshal Pod manifest.
	pod := &corev1.Pod{}
	if err := yaml.Unmarshal([]byte(podManifest), pod); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}

	// Get the IP address of the kube-scheduler pod.
	kubeSchedulerPodIP := getKubeSchedulerPodIP(t, client)
	kubeSchedulerAddress := fmt.Sprintf("%s:%s/metrics", kubeSchedulerPodIP, kubeSchedulerPort)
	// Add the kube-scheduler pod IP address to the pod command.
	pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, kubeSchedulerAddress)

	podsclient := client.CoreV1().Pods(ns.ObjectMeta.Name)

	// Create pod.
	pod, err = podsclient.Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}

	phase := WaitForPodAndGetPodPhase(t, client, ns, pod)
	if phase == corev1.PodSucceeded {
		t.Fatalf("Expected scraping kube-scheduler metrics to fail: %q", kubeSchedulerAddress)
	}

	t.Cleanup(func() {
		if err := nsclient.Delete(context.TODO(), ns.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed removing Namespace: %v", err)
		}
	})
}

func TestDenyMetricsFromPodHavingRequiredLabelsButNotNamespace(t *testing.T) {
	t.Parallel()
	client := testutil.CreateKubeClient(t)
	nsclient := client.CoreV1().Namespaces()
	// Unmarshal namespace manifest.
	ns := &corev1.Namespace{}
	if err := yaml.Unmarshal([]byte(namespaceManifest), ns); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}
	// Create namespace.
	ns, err := nsclient.Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		t.Fatalf("failed to create namespace: %v", err)
	}
	// Unmarshal Pod manifest.
	pod := &corev1.Pod{}
	if err := yaml.Unmarshal([]byte(podManifest), pod); err != nil {
		t.Fatalf("failed unmarshaling manifest: %v", err)
	}
	// Add metrics label to the pod.
	pod.ObjectMeta.Labels = map[string]string{
		"lokomotive.kinvolk.io/scrape-metrics": "true",
	}

	// Get the IP address of the kube-scheduler pod.
	kubeSchedulerPodIP := getKubeSchedulerPodIP(t, client)
	kubeSchedulerAddress := fmt.Sprintf("%s:%s/metrics", kubeSchedulerPodIP, kubeSchedulerPort)
	// Add the kube-scheduler pod IP address to the pod command.
	pod.Spec.Containers[0].Args = append(pod.Spec.Containers[0].Args, kubeSchedulerAddress)

	podsclient := client.CoreV1().Pods(ns.ObjectMeta.Name)

	// Create pod.
	pod, err = podsclient.Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create pod: %v", err)
	}

	phase := WaitForPodAndGetPodPhase(t, client, ns, pod)

	if phase == corev1.PodSucceeded {
		t.Fatalf("Expected scraping kube-scheduler metrics to fail: %q", kubeSchedulerAddress)
	}

	t.Cleanup(func() {
		if err := nsclient.Delete(context.TODO(), ns.ObjectMeta.Name, metav1.DeleteOptions{}); err != nil {
			t.Logf("failed removing Namespace: %v", err)
		}
	})
}

func WaitForPodAndGetPodPhase(
	t *testing.T,
	client kubernetes.Interface,
	ns *corev1.Namespace,
	pod *corev1.Pod,
) corev1.PodPhase {
	podsclient := client.CoreV1().Pods(ns.ObjectMeta.Name)

	phase := corev1.PodUnknown

	if err := wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		p, err := podsclient.Get(context.TODO(), pod.ObjectMeta.Name, metav1.GetOptions{})
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

	return phase
}

func getKubeSchedulerPodIP(t *testing.T, client kubernetes.Interface) string {
	namespace := "kube-system"
	deploymentName := "kube-scheduler"
	// Get kube-scheduler deployment object so that we can get the corresponding pod.
	deploy, err := client.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			t.Fatalf("deployment not found: %v", err)
		}

		t.Fatalf("error looking up for kube-scheduler deployment %v", err)
	}

	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deploy.Spec.Selector),
	})
	if err != nil {
		t.Fatalf("could not list pods for kube-scheduler deployment: %v", err)
	}

	if len(podList.Items) == 0 {
		t.Fatalf("kube-scheduler pods not found")
	}

	// Return kube-scheduler pod IP address
	return podList.Items[0].Status.PodIP
}
