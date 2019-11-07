// +build aws packet
// +build e2e

package coredns

import (
	"testing"
	"time"

	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

func TestCoreDNSDeployment(t *testing.T) {
	t.Parallel()
	namespace := "kube-system"
	deployment := "coredns"
	replicas := 2

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	testutil.WaitForDeployment(t, client, namespace, deployment, replicas, time.Second*5, time.Minute*5)
}
