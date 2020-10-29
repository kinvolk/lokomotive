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

package cluster

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/kinvolk/lokomotive/internal"
	"github.com/kinvolk/lokomotive/pkg/helm"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"github.com/kinvolk/lokomotive/pkg/lokomotive"
	"github.com/kinvolk/lokomotive/pkg/platform"
)

// ApplyOptions defines how cluster apply operation will behave.
type ApplyOptions struct {
	Confirm         bool
	UpgradeKubelets bool
	SkipComponents  bool
	Verbose         bool
	ConfigPath      string
	ValuesPath      string
}

// Apply applies cluster configuration together with components.
//
//nolint:funlen
func Apply(contextLogger *log.Entry, options ApplyOptions) error {
	cc := clusterConfig{
		verbose:    options.Verbose,
		configPath: options.ConfigPath,
		valuesPath: options.ValuesPath,
	}

	c, err := cc.initialize(contextLogger)
	if err != nil {
		return fmt.Errorf("initializing: %w", err)
	}

	exists, err := clusterExists(c.terraformExecutor)
	if err != nil {
		return fmt.Errorf("checking if cluster exists: %w", err)
	}

	if exists && !options.Confirm {
		// TODO: We could plan to a file and use it when installing.
		if err := c.terraformExecutor.Plan(); err != nil {
			return fmt.Errorf("reconciling cluster state: %v", err)
		}

		if !askForConfirmation("Do you want to proceed with cluster apply?") {
			contextLogger.Println("Cluster apply cancelled")

			return nil
		}
	}

	if err := c.platform.Apply(&c.terraformExecutor); err != nil {
		return fmt.Errorf("applying platform: %v", err)
	}

	fmt.Printf("\nYour configurations are stored in %s\n", c.assetDir)

	kg := kubeconfigGetter{
		platformRequired: true,
	}

	kubeconfig, err := kg.getKubeconfig(contextLogger, c.lokomotiveConfig)
	if err != nil {
		return fmt.Errorf("getting kubeconfig: %v", err)
	}

	if err := verifyCluster(kubeconfig, c.platform.Meta().ExpectedNodes); err != nil {
		return fmt.Errorf("verifying cluster: %v", err)
	}

	// Update all the pre installed namespaces with lokomotive specific label.
	// `lokomotive.kinvolk.io/name: <namespace_name>`.
	if err := updateInstalledNamespaces(kubeconfig); err != nil {
		return fmt.Errorf("updating installed namespace: %v", err)
	}

	// Do controlplane upgrades only if cluster already exists and it is not a managed platform.
	if exists && !c.platform.Meta().Managed {
		fmt.Printf("\nEnsuring that cluster controlplane is up to date.\n")

		cu := controlplaneUpdater{
			kubeconfig:    kubeconfig,
			assetDir:      c.assetDir,
			contextLogger: *contextLogger,
			ex:            c.terraformExecutor,
		}

		charts := platform.CommonControlPlaneCharts()

		if options.UpgradeKubelets {
			charts = append(charts, helm.LokomotiveChart{
				Name:      "kubelet",
				Namespace: "kube-system",
			})
		}

		for _, c := range charts {
			if err := cu.upgradeComponent(c.Name, c.Namespace); err != nil {
				return fmt.Errorf("upgrading controlplane component %q: %w", c.Name, err)
			}
		}
	}

	if ph, ok := c.platform.(platform.PlatformWithPostApplyHook); ok {
		if err := ph.PostApplyHook(kubeconfig); err != nil {
			return fmt.Errorf("running platform post install hook: %v", err)
		}
	}

	if options.SkipComponents {
		return nil
	}

	componentObjects, err := componentNamesToObjects(selectComponentNames(nil, *c.lokomotiveConfig.RootConfig))
	if err != nil {
		return fmt.Errorf("getting component objects: %w", err)
	}

	contextLogger.Println("Applying component configuration")

	if err := applyComponents(c.lokomotiveConfig, kubeconfig, componentObjects); err != nil {
		return fmt.Errorf("applying component configuration: %v", err)
	}

	return nil
}

func verifyCluster(kubeconfig []byte, expectedNodes int) error {
	cs, err := k8sutil.NewClientset(kubeconfig)
	if err != nil {
		return fmt.Errorf("creating Kubernetes clientset: %w", err)
	}

	cluster := lokomotive.NewCluster(cs, expectedNodes)

	return cluster.Verify()
}

func updateInstalledNamespaces(kubeconfig []byte) error {
	cs, err := k8sutil.NewClientset(kubeconfig)
	if err != nil {
		return fmt.Errorf("create clientset: %v", err)
	}

	nsclient := cs.CoreV1().Namespaces()

	namespaces, err := k8sutil.ListNamespaces(nsclient)
	if err != nil {
		return fmt.Errorf("getting list of namespaces: %v", err)
	}

	for _, ns := range namespaces.Items {
		ns := k8sutil.Namespace{
			Name: ns.ObjectMeta.Name,
			Labels: map[string]string{
				internal.NamespaceLabelKey: ns.ObjectMeta.Name,
			},
		}

		if err := k8sutil.CreateOrUpdateNamespace(ns, nsclient); err != nil {
			return fmt.Errorf("namespace %q with labels: %v", ns, err)
		}
	}

	return nil
}
