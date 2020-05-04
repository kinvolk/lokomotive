// Copyright 2020 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package install

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"github.com/kinvolk/lokomotive/pkg/util/retryutil"
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
func Verify(cl *k8sutil.Cluster) error {
	fmt.Println("\nNow checking health and readiness of the cluster nodes ...")

	// Wait for cluster to become available
	err := retryutil.Retry(clusterPingRetryInterval*time.Second, clusterPingRetries, cl.Ping)
	if err != nil {
		return errors.Wrapf(err, "failed to ping cluster for readiness")
	}

	var ns *k8sutil.NodeStatus
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
