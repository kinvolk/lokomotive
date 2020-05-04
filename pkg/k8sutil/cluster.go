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

package k8sutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

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

type Cluster struct {
	KubeClient    *kubernetes.Clientset
	ExpectedNodes int
}

func NewCluster(client *kubernetes.Clientset, expectedNodes int) (*Cluster, error) {
	return &Cluster{KubeClient: client, ExpectedNodes: expectedNodes}, nil
}

// ComponentsStatus returns the status of Kubernetes cluster components.
func (cl *Cluster) ComponentsStatus() ([]v1.ComponentStatus, error) {
	cs, err := cl.KubeClient.CoreV1().ComponentStatuses().List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// For now we only show the status of etcd.
	var etcdComponents []v1.ComponentStatus

	for _, item := range cs.Items {
		if strings.HasPrefix(item.Name, "etcd") {
			etcdComponents = append(etcdComponents, item)
		}
	}

	return etcdComponents, nil
}

// NodeStatus represents the status of all nodes of a cluster.
type NodeStatus struct {
	nodeConditions map[string][]v1.NodeCondition
	expectedNodes  int
}

// GetNodeStatus returns the status for all running nodes or an error.
func (cl *Cluster) GetNodeStatus() (*NodeStatus, error) {
	n, err := cl.KubeClient.CoreV1().Nodes().List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodeConditions := make(map[string][]v1.NodeCondition)

	for _, node := range n.Items {
		nodeConditions[node.Name] = node.Status.Conditions
	}
	return &NodeStatus{
		nodeConditions: nodeConditions,
		expectedNodes:  cl.ExpectedNodes,
	}, nil
}

// Verify verifies health and readiness of the cluster.
func (cl *Cluster) Verify() error {
	fmt.Println("\nNow checking health and readiness of the cluster nodes ...")

	// Wait for cluster to become available.
	err := retryutil.Retry(clusterPingRetryInterval*time.Second, clusterPingRetries, cl.Ping)
	if err != nil {
		return fmt.Errorf("failed to ping cluster for readiness: %v", err)
	}

	var ns *NodeStatus

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
		return fmt.Errorf("error determining node status within the allowed time: %v", nsErr)
	}

	if err != nil {
		return fmt.Errorf("not all nodes became ready within the allowed time: %v", err)
	}

	ns.PrettyPrint()

	fmt.Println("\nSuccess - cluster is healthy and nodes are ready!")

	return nil
}

// Ready checks if all nodes are ready and returns false otherwise.
func (ns *NodeStatus) Ready() bool {
	if len(ns.nodeConditions) < ns.expectedNodes {
		return false
	}

	for _, conditions := range ns.nodeConditions {
		for _, condition := range conditions {
			if condition.Type == "Ready" && condition.Status != v1.ConditionTrue {
				return false
			}
		}
	}

	return true
}

// PrettyPrint prints Node statuses in a pretty way
func (ns *NodeStatus) PrettyPrint() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	// Print the header.
	fmt.Fprintln(w, "\nNode\tReady\tReason\tMessage\t")

	// An empty line between header and the body.
	fmt.Fprintln(w, "\t\t\t\t")

	for node, conditions := range ns.nodeConditions {
		for _, condition := range conditions {
			if condition.Type == "Ready" {
				line := fmt.Sprintf(
					"%s\t%s\t%s\t%s\t",
					node, condition.Status, condition.Reason, condition.Message,
				)
				fmt.Fprintln(w, line)
			}
		}
	}
	if len(ns.nodeConditions) < ns.expectedNodes {
		line := fmt.Sprintf("%d nodes are missing", ns.expectedNodes-len(ns.nodeConditions))
		fmt.Fprintln(w, line)
	}

	w.Flush()
}

// Ping Cluster to know when its endpoint can be used
func (cl *Cluster) Ping() (bool, error) {
	_, err := cl.KubeClient.CoreV1().Nodes().List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return false, nil
	}
	return true, nil
}
