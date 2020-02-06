// +build aws packet
// +build e2e

package kubernetes

import (
	"testing"

	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

func TestControllerManagerDeployment(t *testing.T) {
	t.Parallel()

	namespace := "kube-system"
	deployment := "kube-controller-manager"
	replicas := 2

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}

	t.Log("got kubernetes client")

	testutil.WaitForDeployment(t, client, namespace, deployment, replicas, retryInterval, timeout)
}
