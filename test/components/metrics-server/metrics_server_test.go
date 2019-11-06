// +build aws packet
// +build e2e

package metricsserver

import (
	"testing"
	"time"

	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

func TestMetricsServerDeployment(t *testing.T) {
	namespace := "kube-system"

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	t.Run("deployment", func(t *testing.T) {
		t.Parallel()
		deployment := "metrics-server"
		replicas := 1

		testutil.WaitForDeployment(t, client, namespace, deployment, replicas, time.Second*5, time.Minute*5)
		t.Logf("Found required replicas: %d", replicas)
	})
}
