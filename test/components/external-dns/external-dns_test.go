// +build aws packet
// +build e2e

package externaldns

import (
	"fmt"
	testutil "github.com/kinvolk/lokoctl/test/components/util"
	"testing"
	"time"
)

const namespace = "external-dns"

func TestExternalDNSDeployments(t *testing.T) {
	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	t.Run(fmt.Sprintf("deployment"), func(t *testing.T) {
		t.Parallel()
		deployment := "external-dns"
		replicas := 3
		testutil.WaitForDeployment(t, client, namespace, deployment, replicas, time.Second*5, time.Minute*5)
		t.Logf("Found required replicas: %d", replicas)
	})
}
