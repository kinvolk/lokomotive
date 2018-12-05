package install

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/lokomotive"
	"github.com/kinvolk/lokoctl/pkg/util/retryutil"
)

// Verify health and readiness of the cluster.
func Verify(cl *lokomotive.Cluster) error {
	fmt.Println("\nNow checking health and readiness of the cluster nodes ...")

	// Wait for cluster to become available
	err := retryutil.Retry(10*time.Second, 5, cl.Ping)
	if err != nil {
		return errors.Wrapf(err, "failed to ping cluster for readiness")
	}

	err = retryutil.Retry(10*time.Second, 5, cl.NodesReady)
	if err != nil {
		if retryutil.IsRetryFailure(err) {
			return fmt.Errorf("not all nodes became ready within 50 seconds")
		}
		return errors.Wrapf(err, "error determining node readiness")
	}

	ns, err := cl.GetNodeStatus()
	if err != nil {
		return errors.Wrapf(err, "failed to get node status")
	}

	ns.PrettyPrint()

	fmt.Println("\nSuccess - cluster is healthy and nodes are ready!")

	return nil
}
