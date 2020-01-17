// +build aws packet
// +build e2e

package certmanager

import (
	"fmt"
	"testing"
	"time"

	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

const namespace = "cert-manager"

func TestCertManagerDeployments(t *testing.T) {
	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	testCases := []struct {
		deployment string
		replicas   int
	}{
		{
			deployment: "cert-manager",
			replicas:   1,
		},
		{
			deployment: "cert-manager-cainjector",
			replicas:   1,
		},
		{
			deployment: "cert-manager-webhook",
			replicas:   1,
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("cert-manager deployment:%s", test.deployment), func(t *testing.T) {
			t.Parallel()
			testutil.WaitForDeployment(t, client, namespace, test.deployment, test.replicas, time.Second*5, time.Minute*5)
			t.Logf("Found required replicas: %d", test.replicas)
		})
	}
}
