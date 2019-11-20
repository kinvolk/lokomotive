// +build aws packet
// +build e2e

package openebsstorageclass

import (
	"testing"
	"time"

	testutil "github.com/kinvolk/lokoctl/test/components/util"
)

func TestOpenEBSStorageClass(t *testing.T) {
	client, err := testutil.CreateKubeClient(t)
	if err != nil {
		t.Errorf("could not create Kubernetes client: %v", err)
	}
	t.Log("got kubernetes client")

	t.Run("storageclass", func(t *testing.T) {
		t.Parallel()
		sc := "openebs-test-sc"

		testutil.WaitForStorageClass(t, client, sc, time.Second*5, time.Minute*5)
	})
}
