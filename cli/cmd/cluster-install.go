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

package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokomotive/pkg/install"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"github.com/kinvolk/lokomotive/pkg/lokomotive"
)

var (
	verbose         bool
	skipComponents  bool
	upgradeKubelets bool
)

var clusterInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Lokomotive cluster with components",
	Run:   runClusterInstall,
}

func init() {
	clusterCmd.AddCommand(clusterInstallCmd)
	pf := clusterInstallCmd.PersistentFlags()
	pf.BoolVarP(&confirm, "confirm", "", false, "Upgrade cluster without asking for confirmation")
	pf.BoolVarP(&verbose, "verbose", "v", false, "Show output from Terraform")
	pf.BoolVarP(&skipComponents, "skip-components", "", false, "Skip component installation")
	pf.BoolVarP(&upgradeKubelets, "upgrade-kubelets", "", false, "Experimentally upgrade self-hosted kubelets")
}

func runClusterInstall(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl cluster install",
		"args":    args,
	})

	ex, p, lokoConfig, assetDir := initialize(ctxLogger)

	exists := clusterExists(ctxLogger, ex)
	if exists && !confirm {
		// TODO: We could plan to a file and use it when installing.
		if err := ex.Plan(); err != nil {
			ctxLogger.Fatalf("Failed to reconsile cluster state: %v", err)
		}

		if !askForConfirmation("Do you want to proceed with cluster install?") {
			ctxLogger.Println("Cluster install cancelled")

			return
		}
	}

	if err := p.Install(ex); err != nil {
		ctxLogger.Fatalf("error installing cluster: %v", err)
	}

	fmt.Printf("\nYour configurations are stored in %s\n", assetDir)

	kubeconfigPath := assetsKubeconfig(assetDir)
	if err := verifyInstall(kubeconfigPath, p.GetExpectedNodes()); err != nil {
		ctxLogger.Fatalf("Verify cluster installation: %v", err)
	}

	// Do controlplane upgrades only if cluster already exists.
	if exists {
		fmt.Printf("\nEnsuring that cluster controlplane is up to date.\n")

		cu := controlplaneUpdater{
			kubeconfigPath: kubeconfigPath,
			assetDir:       assetDir,
			ctxLogger:      *ctxLogger,
			ex:             *ex,
		}

		releases := []string{"pod-checkpointer", "kube-apiserver", "kubernetes", "calico"}

		if upgradeKubelets {
			releases = append(releases, "kubelet")
		}

		for _, c := range releases {
			cu.upgradeComponent(c)
		}
	}

	if skipComponents {
		return
	}

	var componentsToInstall []string
	for _, component := range lokoConfig.RootConfig.Components {
		componentsToInstall = append(componentsToInstall, component.Name)
	}

	ctxLogger.Println("Installing components")

	if len(componentsToInstall) > 0 {
		if err := installComponents(lokoConfig, kubeconfigPath, componentsToInstall...); err != nil {
			ctxLogger.Fatalf("Installing components failed: %v", err)
		}
	}
}

func verifyInstall(kubeconfigPath string, expectedNodes int) error {
	client, err := k8sutil.NewClientset(kubeconfigPath)
	if err != nil {
		return errors.Wrapf(err, "failed to set up clientset")
	}

	cluster, err := lokomotive.NewCluster(client, expectedNodes)
	if err != nil {
		return errors.Wrapf(err, "failed to set up cluster client")
	}

	return install.Verify(cluster)
}
