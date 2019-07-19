package install

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/kinvolk/lokoctl/pkg/lokomotive"
	"github.com/kinvolk/lokoctl/pkg/util/retryutil"
)

const (
	// Max number of retries when waiting for cluster to become available.
	clusterPingRetries = 18
	// Number of seconds to wait between retires when waiting for cluster to become available.
	clusterPingRetryInterval = 10
	// Max number of retries when waiting for nodes to become ready.
	nodeReadinessRetries = 18
	// Number of seconds to wait between retires when waiting for nodes to become ready.
	nodeReadinessRetryInterval = 10
)

// Verify health and readiness of the cluster.
func Verify(cl *lokomotive.Cluster) error {
	fmt.Println("\nNow checking health and readiness of the cluster nodes ...")

	// Wait for cluster to become available
	err := retryutil.Retry(clusterPingRetryInterval*time.Second, clusterPingRetries, cl.Ping)
	if err != nil {
		return errors.Wrapf(err, "failed to ping cluster for readiness")
	}

	var ns *lokomotive.NodeStatus
	var nsErr error
	err = retryutil.Retry(nodeReadinessRetryInterval*time.Second, nodeReadinessRetries, func() (bool, error) {
		// Store the original error because Retry would stop too early if we forward it
		// and anyway overrides the error in case of timeout.
		ns, nsErr = cl.GetNodeStatus()
		if nsErr != nil {
			// To continue retrying, we don't set the error here.
			return false, nil
		}
		return ns.Ready(), nil // Retry if not ready
	})
	if nsErr != nil {
		return errors.Wrapf(nsErr, "error determining node status within the allowed time")
	}
	if err != nil {
		return fmt.Errorf("not all nodes became ready within the allowed time")
	}
	ns.PrettyPrint()

	fmt.Println("\nSuccess - cluster is healthy and nodes are ready!")

	return nil
}
