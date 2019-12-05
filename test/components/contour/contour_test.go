// +build aws packet
// +build e2e

package contour

import (
	"testing"
	"time"

	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

func TestEnvoyDaemonset(t *testing.T) {
	t.Parallel()
	namespace := "projectcontour"
	daemonset := "envoy"
	// is equal to no of worker nodes
	replicas := 2

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	testutil.WaitForDaemonSet(t, client, namespace, daemonset, replicas, time.Second*5, time.Minute*5)
	t.Logf("Found required replicas: %d", replicas)
}
