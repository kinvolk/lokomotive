// +build aws packet
// +build e2e

package fluo

import (
	"testing"
	"time"

	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

func TestUpdateAgentDaemonset(t *testing.T) {
	t.Parallel()
	namespace := "reboot-coordinator"
	daemonset := "flatcar-linux-update-agent"

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	testutil.WaitForDaemonSet(t, client, namespace, daemonset, time.Second*5, time.Minute*5)
	t.Logf("Found required replicas")
}

func TestUpdateOperatorDeployment(t *testing.T) {
	t.Parallel()
	namespace := "reboot-coordinator"
	deployment := "flatcar-linux-update-operator"
	replicas := 1

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	testutil.WaitForDeployment(t, client, namespace, deployment, replicas, time.Second*5, time.Minute*5)
	t.Logf("Found required replicas: %d", replicas)
}
