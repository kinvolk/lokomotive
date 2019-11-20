// +build aws packet
// +build e2e

package rook

import (
	"testing"
	"time"

	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

func TestRookDeployment(t *testing.T) {
	namespace := "rook"

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	t.Run("deployment", func(t *testing.T) {
		t.Parallel()
		deployment := "rook-ceph-operator"
		replicas := 1

		testutil.WaitForDeployment(t, client, namespace, deployment, replicas, time.Second*5, time.Minute*5)
		t.Logf("Required replicas: %d", replicas)
	})
}
