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

package lokomotive

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Cluster struct {
	KubeClient    *kubernetes.Clientset
	ExpectedNodes int
}

func NewCluster(client *kubernetes.Clientset, expectedNodes int) (*Cluster, error) {
	return &Cluster{KubeClient: client, ExpectedNodes: expectedNodes}, nil
}

func (cl *Cluster) Health() ([]v1.ComponentStatus, error) {
	cs, err := cl.KubeClient.CoreV1().ComponentStatuses().List(meta_v1.ListOptions{})
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
	n, err := cl.KubeClient.CoreV1().Nodes().List(meta_v1.ListOptions{})
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
	_, err := cl.KubeClient.CoreV1().Nodes().List(meta_v1.ListOptions{})
	if err != nil {
		return false, nil
	}
	return true, nil
}
