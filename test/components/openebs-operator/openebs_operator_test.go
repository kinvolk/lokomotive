// +build aws packet
// +build e2e

package openebsoperator

import (
	"testing"
	"time"

	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

func TestOpenEBSOperatorDeployment(t *testing.T) {
	namespace := "openebs"

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	deployments := []string{
		"openebs-provisioner",
		"openebs-localpv-provisioner",
		"openebs-admission-server",
		"openebs-ndm-operator",
		"openebs-snapshot-operator",
		"maya-apiserver",
	}

	for _, deployment := range deployments {
		t.Run("deployment", func(t *testing.T) {
			t.Parallel()
			replicas := 1

			testutil.WaitForDeployment(t, client, namespace, deployment, replicas, time.Second*5, time.Minute*5)
			t.Logf("Required replicas: %d", replicas)
		})
	}
}
