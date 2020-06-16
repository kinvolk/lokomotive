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

package util

import (
	"context"
	"fmt"
	"io/ioutil"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage/driver"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

func ensureNamespaceExists(name string, kubeconfigPath string) error {
	kubeconfig, err := ioutil.ReadFile(kubeconfigPath) // #nosec G304
	if err != nil {
		return fmt.Errorf("reading kubeconfig file: %w", err)
	}

	cs, err := k8sutil.NewClientset(kubeconfig)
	if err != nil {
		return fmt.Errorf("creating clientset: %w", err)
	}

	if name == "" {
		return fmt.Errorf("namespace name can't be empty")
	}

	// Ensure the namespace in which we create release and resources exists.
	_, err = cs.CoreV1().Namespaces().Create(context.TODO(), &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

// InstallComponent installs given component using given kubeconfig as a Helm release using a Helm client.
func InstallComponent(c components.Component, kubeconfig string) error {
	name := c.Metadata().Name
	ns := c.Metadata().Namespace

	if err := ensureNamespaceExists(ns, kubeconfig); err != nil {
		return fmt.Errorf("failed ensuring that namespace %q for component %q exists: %w", ns, name, err)
	}

	kubeconfigContent, err := ioutil.ReadFile(kubeconfig) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig file %q: %v", kubeconfig, err)
	}

	actionConfig, err := HelmActionConfig(ns, kubeconfigContent)
	if err != nil {
		return fmt.Errorf("failed preparing helm client: %w", err)
	}

	chart, err := chartFromComponent(c)
	if err != nil {
		return err
	}

	if err := chart.Validate(); err != nil {
		return fmt.Errorf("chart is invalid: %w", err)
	}

	exists, err := ReleaseExists(*actionConfig, name)
	if err != nil {
		return fmt.Errorf("failed checking if component is installed: %w", err)
	}

	wait := c.Metadata().Helm.Wait

	helmAction := &helmAction{
		releaseName:  name,
		chart:        chart,
		actionConfig: actionConfig,
		wait:         wait,
	}

	if !exists {
		return install(helmAction, ns)
	}

	return upgrade(helmAction)
}

type helmAction struct {
	releaseName  string
	chart        *chart.Chart
	actionConfig *action.Configuration
	wait         bool
}

func install(helmAction *helmAction, namespace string) error {
	install := action.NewInstall(helmAction.actionConfig)
	install.ReleaseName = helmAction.releaseName
	install.Namespace = namespace

	// Currently, we install components one-by-one, in the order how they are
	// defined in the configuration and we do not support any dependencies between
	// the components.
	//
	// If it is critical for component to have it's dependencies ready before it is
	// installed, all dependencies should set Wait field to 'true' in components.HelmMetadata
	// struct.
	//
	// The example of such dependency is between prometheus-operator and openebs-storage-class, where
	// both openebs-operator and openebs-storage-class components must be fully functional, before
	// prometheus-operator is deployed, otherwise it won't pick the default storage class.
	install.Wait = helmAction.wait

	if _, err := install.Run(helmAction.chart, map[string]interface{}{}); err != nil {
		return fmt.Errorf("installing release failed: %w", err)
	}

	return nil
}

func upgrade(helmAction *helmAction) error {
	upgrade := action.NewUpgrade(helmAction.actionConfig)
	upgrade.Wait = helmAction.wait
	upgrade.RecreateResources = true

	if _, err := upgrade.Run(helmAction.releaseName, helmAction.chart, map[string]interface{}{}); err != nil {
		return fmt.Errorf("upgrading release failed: %w", err)
	}

	return nil
}

// HelmActionConfig creates initialized Helm action configuration.
func HelmActionConfig(ns string, kubeconfig []byte) (*action.Configuration, error) {
	actionConfig := &action.Configuration{}

	getter, err := k8sutil.NewGetter(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client getter: %v", err)
	}

	// TODO: Add some logging implementation? We currently just pass an empty function for logging.
	logF := func(format string, v ...interface{}) {}

	if err := actionConfig.Init(getter, ns, "secret", logF); err != nil {
		return nil, fmt.Errorf("failed initializing helm: %w", err)
	}

	kc := kube.New(getter)
	kc.Log = logF
	kc.Namespace = ns

	actionConfig.KubeClient = kc

	return actionConfig, nil
}

// ReleaseExists checks if given Helm release exists.
func ReleaseExists(actionConfig action.Configuration, name string) (bool, error) {
	histClient := action.NewHistory(&actionConfig)
	histClient.Max = 1

	_, err := histClient.Run(name)
	if err != nil && err != driver.ErrReleaseNotFound {
		return false, fmt.Errorf("failed checking for chart history: %w", err)
	}

	return err != driver.ErrReleaseNotFound, nil
}

// UninstallComponent uninstalls a component and optionally removes it's namespace.
func UninstallComponent(c components.Component, kubeconfig string, deleteNSBool bool) error {
	name := c.Metadata().Name
	if name == "" {
		// This should never fail in real user usage, if this does that means the component was not
		// created with all the needed information.
		panic(fmt.Errorf("component name is empty"))
	}

	ns := c.Metadata().Namespace
	if ns == "" {
		// This should never fail in real user usage, if this does that means the component was not
		// created with all the needed information.
		panic(fmt.Errorf("component %s namespace is empty", name))
	}

	kubeconfigContent, err := ioutil.ReadFile(kubeconfig) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig file %q: %v", kubeconfig, err)
	}

	cfg, err := HelmActionConfig(ns, kubeconfigContent)
	if err != nil {
		return fmt.Errorf("failed preparing helm client: %w", err)
	}

	history := action.NewHistory(cfg)
	// Check if the component's release exists. If it does only then proceed to delete.
	//
	// Note: It is assumed that this call will return error only when the release does not exist.
	// The error check is ignored to make `lokoctl component delete ..` idempotent.
	// We rely on the fact that the 'component name' == 'release name'. Since component's name is
	// hardcoded and unlikely to change release name won't change as well. And they will be
	// consistent if installed by lokoctl. So it is highly unlikely that following call will return
	// any other error than "release not found".
	if _, err := history.Run(name); err == nil {
		uninstall := action.NewUninstall(cfg)

		// Ignore the err when we have deleted the release already or it does not exist for some reason.
		if _, err := uninstall.Run(name); err != nil {
			return err
		}
	}

	if deleteNSBool {
		if err := deleteNS(ns, kubeconfig); err != nil {
			return err
		}
	}

	return nil
}

func deleteNS(ns string, kubeconfig string) error {
	kubeconfigContent, err := ioutil.ReadFile(kubeconfig) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig file: %v", err)
	}

	cs, err := k8sutil.NewClientset(kubeconfigContent)
	if err != nil {
		return err
	}

	// Delete the manually created namespace which was not created by helm.
	if err = cs.CoreV1().Namespaces().Delete(context.TODO(), ns, metav1.DeleteOptions{}); err != nil {
		// Ignore error when the namespace does not exist.
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	return nil
}
