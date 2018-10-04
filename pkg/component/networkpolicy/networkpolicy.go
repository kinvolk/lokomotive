package networkpolicy

import (
	"github.com/kinvolk/lokoctl/pkg/k8sutil"
	log "github.com/sirupsen/logrus"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	name = "default-network-policy"
)

type DefaultNetworkPolicy struct {
}

func (dnp DefaultNetworkPolicy) Name() string {
	return name
}

func (dnp DefaultNetworkPolicy) Install(clientset *kubernetes.Clientset, namespace string) error {
	contextLogger := log.WithFields(log.Fields{
		"command": "install default-network-policy",
	})

	data, err := Asset("manifests/default-network-policies/deny-metadata-access.yaml")
	if err != nil {
		return err
	}

	contextLogger.Infof("default-network-policy manifest template data: %q", data)

	net, err := k8sutil.GetKubernetesObjectFromTmpl(data, nil)
	if err != nil {
		return err
	}

	if _, err := clientset.NetworkingV1().NetworkPolicies(namespace).Create(net.(*networkingv1.NetworkPolicy)); err != nil {
		return err
	}

	return nil
}

func New() *DefaultNetworkPolicy {
	return &DefaultNetworkPolicy{}
}
