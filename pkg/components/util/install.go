package util

import (
	"fmt"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/k8sutil"
)

// InstallComponent installs given component using given kubeconfig.
func InstallComponent(_ string, c components.Component, kubeconfig string) error {
	return InstallAsManifests(c, kubeconfig)
}

// InstallAsManifests installs given component by applying manifests directly
// to the kube-apiserver using given kubeconfig.
func InstallAsManifests(c components.Component, kubeconfig string) error {
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

// InstallAsRelease installs a component as a Helm release using a Helm client.
func InstallAsRelease(name string, c components.Component, kubeconfig string) error {
	actionConfig := &action.Configuration{}

	cs, err := k8sutil.NewClientset(kubeconfig)
	if err != nil {
		return err
	}

	// Get the namespace in which the component should be created.
	ns := c.Metadata().Namespace
	if ns == "" {
		return fmt.Errorf("component %s namespace is empty", name)
	}

	// Ensure the namespace in which we create release and resources exists.
	_, err = cs.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	// TODO: Add some logging implementation? We currently just pass an empty function for logging.
	if err := actionConfig.Init(kube.GetConfig(kubeconfig, "", ns), ns, "secret", func(format string, v ...interface{}) {}); err != nil {
		return err
	}

	chart, err := chartFromManifests(name, c)
	if err != nil {
		return err
	}

	if err := chart.Validate(); err != nil {
		return fmt.Errorf("chart is invalid: %w", err)
	}

	install := action.NewInstall(actionConfig)
	install.ReleaseName = name
	install.Namespace = ns

	// Wait for charts to become ready to avoid race conditions.
	// TODO: Make this configurable per component.
	install.Wait = true

	if _, err := install.Run(chart, map[string]interface{}{}); err != nil {
		return fmt.Errorf("installing component '%s' as chart failed: %w", name, err)
	}

	return nil
}
