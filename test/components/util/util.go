package util

import (
	"fmt"
	"os"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func CreateKubeClient(t *testing.T) (*kubernetes.Clientset, error) {
	kubeconfig := os.ExpandEnv(os.Getenv("KUBECONFIG"))
	if kubeconfig == "" {
		t.Errorf("env var KUBECONFIG was not set")
	}
	t.Logf("using KUBECONFIG=%s", kubeconfig)

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Errorf("could not build config from KUBECONFIG: %v", err)
	}

	return kubernetes.NewForConfig(config)
}

func WaitForDaemonSet(t *testing.T, client kubernetes.Interface, ns, name string, replicas int, retryInterval, timeout time.Duration) {
	if err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		ds, err := client.AppsV1().DaemonSets(ns).Get(name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				t.Logf("waiting for daemonset %s to be available", name)
				return false, nil
			}
			return false, err
		}

		if int(ds.Status.NumberAvailable) == replicas {
			return true, nil
		}
		t.Logf("daemonset: %s, replicas: %d/%d", name, int(ds.Status.NumberAvailable), replicas)
		return false, nil
	}); err != nil {
		t.Errorf("error while waiting for the daemonset: %v", err)
	}
}

func WaitForDeployment(t *testing.T, client kubernetes.Interface, ns, name string, replicas int, retryInterval, timeout time.Duration) {
	var err error
	var deploy *appsv1.Deployment

	// Check the readiness of the Deployment
	if err = wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		deploy, err = client.AppsV1().Deployments(ns).Get(name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				t.Logf("waiting for deployment %s to be available", name)
				return false, nil
			}
			return false, err
		}

		if int(deploy.Status.AvailableReplicas) == replicas {
			return true, nil
		}
		t.Logf("deployment: %s, replicas: %d/%d", name, int(deploy.Status.AvailableReplicas), replicas)
		return false, nil
	}); err != nil {
		t.Errorf("error while waiting for the deployment: %v", err)
		return
	}

	// Check the readiness of the pods
	labelSet := labels.Set(deploy.Labels)
	if err := wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		pods, err := client.CoreV1().Pods(ns).List(metav1.ListOptions{LabelSelector: labelSet.String()})
		if err != nil {
			return false, err
		}
		pods = filterNonControllerPods(pods)
		// go through each pod in the returned list and check the readiness status of it
		for _, pod := range pods.Items {
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.RestartCount > 10 {
					return false, fmt.Errorf("pod: %s, container %s; pod in CrashLoopBackOff", pod.Name, cs.Name)
				}
				if !cs.Ready {
					t.Logf("pod: %s, container %s; container not ready", pod.Name, cs.Name)
					return false, nil
				}
			}
			t.Logf("pod %s, has all containers in ready state", pod.Name)
		}
		t.Logf("all pods for deployment %s, are in ready state", deploy.Name)
		return true, nil
	}); err != nil {
		t.Errorf("error while waiting for the pods: %v", err)
	}
}

func filterNonControllerPods(pods *corev1.PodList) *corev1.PodList {
	var filteredPods []corev1.Pod

	for _, pod := range pods.Items {
		// The pod that has a controller, has this label
		if _, ok := pod.Labels["pod-template-hash"]; !ok {
			continue
		}
		filteredPods = append(filteredPods, pod)
	}
	pods.Items = filteredPods
	return pods
}
