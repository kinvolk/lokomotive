// +build packet
// +build e2e

package prometheusoperator

import (
	"fmt"
	"testing"
	"time"

	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

func TestPrometheusOperatorDeployment(t *testing.T) {
	namespace := "monitoring"

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	deployments := []string{
		"prometheus-operator-operator",
		"prometheus-operator-kube-state-metrics",
		"prometheus-operator-grafana",
	}

	for _, deployment := range deployments {
		t.Run("deployment", func(t *testing.T) {
			t.Parallel()
			replicas := 1

			testutil.WaitForDeployment(t, client, namespace, deployment, replicas, time.Second*5, time.Minute*5)
			t.Logf("Required replicas: %d", replicas)
		})
	}

	statefulSets := []string{
		"alertmanager-prometheus-operator-alertmanager",
		"prometheus-prometheus-operator-prometheus",
	}

	for _, statefulset := range statefulSets {
		t.Run(fmt.Sprintf("statefulset %s", statefulset), func(t *testing.T) {
			t.Parallel()
			replicas := 1

			testutil.WaitForStatefulSet(t, client, namespace, statefulset, replicas, time.Second*5, time.Minute*5)
			t.Logf("Required replicas: %d", replicas)
		})
	}

	testutil.WaitForDaemonSet(t, client, namespace, "prometheus-operator-prometheus-node-exporter", time.Second*5, time.Minute*10)
}
