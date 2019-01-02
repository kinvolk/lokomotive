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

	err = retryutil.Retry(nodeReadinessRetryInterval*time.Second, nodeReadinessRetries, cl.NodesReady)
	if err != nil {
		if retryutil.IsRetryFailure(err) {
			return fmt.Errorf("not all nodes became ready within the allowed time")
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
