package util

import (
	"os"
	"testing"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	if err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		ds, err := client.AppsV1().Deployments(ns).Get(name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				t.Logf("waiting for deployment %s to be available", name)
				return false, nil
			}
			return false, err
		}

		if int(ds.Status.AvailableReplicas) == replicas {
			return true, nil
		}
		t.Logf("deployment: %s, replicas: %d/%d", name, int(ds.Status.AvailableReplicas), replicas)
		return false, nil
	}); err != nil {
		t.Errorf("error while waiting for the deployment: %v", err)
	}
}
