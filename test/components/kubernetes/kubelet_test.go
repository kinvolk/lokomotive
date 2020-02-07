// +build aws packet
// +build e2e

package kubernetes

import (
	"testing"
	"time"

	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

const retryInterval = time.Second * 5

const timeout = time.Minute * 5

func TestSelfHostedKubeletPods(t *testing.T) {
	t.Parallel()

	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}

	t.Log("got kubernetes client")

	namespace := "kube-system"
	daemonset := "kubelet"

	testutil.WaitForDaemonSet(t, client, namespace, daemonset, retryInterval, timeout)
}
