package util

import (
	"time"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/k8sutil"
)

func Install(c components.Component, kubeconfig string) error {
	renderedFiles, err := c.RenderManifests()
	if err != nil {
		return err
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{},
	)

	return k8sutil.CreateAssets(clientConfig, renderedFiles, 1*time.Minute)
}
