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
	KubeClient *kubernetes.Clientset
}

func NewCluster(client *kubernetes.Clientset) (*Cluster, error) {
	return &Cluster{KubeClient: client}, nil
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
	}, nil
}

// NodesReady checks if all nodes are ready
// and returns false otherwise.
func (cl *Cluster) NodesReady() (bool, error) {
	ns, err := cl.GetNodeStatus()
	if err != nil {
		return false, err
	}

	for _, conditions := range ns.nodeConditions {
		for _, condition := range conditions {
			if condition.Type == "Ready" && condition.Status != v1.ConditionTrue {
				return false, nil
			}
		}
	}

	return true, nil
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
