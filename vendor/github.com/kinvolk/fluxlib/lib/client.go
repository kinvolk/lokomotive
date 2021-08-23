package lib

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getRestClient(kubeconfig []byte) (*restclient.Config, error) {
	if len(kubeconfig) == 0 {
		client, err := clientcmd.BuildConfigFromFlags("", "")
		if err != nil {
			return nil, fmt.Errorf("neither kubeconfig not provided nor in-cluster config available: %w", err)
		}

		return client, nil
	}

	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("creating client config failed: %w", err)
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("converting client config to rest client config failed: %w", err)
	}

	return restConfig, nil
}

func GetKubernetesClient(kubeconfig []byte, s *runtime.Scheme) (client.Client, error) {
	restConfig, err := getRestClient(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("creating a REST client: %w", err)
	}

	kclient, err := client.New(restConfig, client.Options{Scheme: s})
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes client: %w", err)
	}

	return kclient, err
}
