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

package helmutil

import (
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage/driver"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

// InstallComponent installs given component using given kubeconfig.
func InstallComponent(name string, c components.Component, kubeconfig string) error {
	return InstallAsRelease(name, c, kubeconfig)
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

	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1

	_, err = histClient.Run(name)
	if err != nil && err != driver.ErrReleaseNotFound {
		return fmt.Errorf("failed checking for chart history: %w", err)
	}

	wait := c.Metadata().Helm.Wait

	if err == driver.ErrReleaseNotFound {
		install := action.NewInstall(actionConfig)
		install.ReleaseName = name
		install.Namespace = ns

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
		install.Wait = wait

		if _, err := install.Run(chart, map[string]interface{}{}); err != nil {
			return fmt.Errorf("installing component '%s' as chart failed: %w", name, err)
		}

		return nil
	}

	upgrade := action.NewUpgrade(actionConfig)
	upgrade.Wait = wait

	if _, err := upgrade.Run(name, chart, map[string]interface{}{}); err != nil {
		return fmt.Errorf("updating chart failed: %w", err)
	}

	return nil
}
