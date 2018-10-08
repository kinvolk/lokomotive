package component

import "k8s.io/client-go/kubernetes"

type Interface interface {
	Name() string
	Install(*kubernetes.Clientset, string) error
}
