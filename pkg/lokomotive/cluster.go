package lokomotive

import (
	"strings"

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
