// +build packet
// +build e2e

package metallb

import (
	"testing"
	"time"

	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

func TestMetalLBDeployment(t *testing.T) {
	namespace := "metallb-system"

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	t.Run("speaker daemonset", func(t *testing.T) {
		t.Parallel()
		daemonset := "speaker"

		testutil.WaitForDaemonSet(t, client, namespace, daemonset, time.Second*5, time.Minute*5)
	})

	t.Run("controller deployment", func(t *testing.T) {
		t.Parallel()
		deployment := "controller"
		replicas := 1

		testutil.WaitForDeployment(t, client, namespace, deployment, replicas, time.Second*5, time.Minute*5)
		t.Logf("Found required replicas: %d", replicas)
	})
}
