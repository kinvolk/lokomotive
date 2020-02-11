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
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/backend/local"
	"github.com/kinvolk/lokoctl/pkg/install"
	"github.com/kinvolk/lokoctl/pkg/k8sutil"
	"github.com/kinvolk/lokoctl/pkg/lokomotive"
	"github.com/kinvolk/lokoctl/pkg/terraform"
)

var quiet bool

var clusterInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Lokomotive cluster with components",
	Run:   runClusterInstall,
}

func init() {
	clusterCmd.AddCommand(clusterInstallCmd)
	pf := clusterInstallCmd.PersistentFlags()
	pf.BoolVarP(&quiet, "quiet", "q", false, "Suppress the output from Terraform")
}

func runClusterInstall(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl cluster install",
		"args":    args,
	})

	lokoConfig, diags := getLokoConfig()
	if len(diags) > 0 {
		ctxLogger.Fatal(diags)
	}

	p, diags := getConfiguredPlatform()
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			ctxLogger.Error(diagnostic.Error())
		}
		ctxLogger.Fatal("Errors found while loading cluster configuration")
	}

	if p == nil {
		ctxLogger.Fatal("No cluster configured")
	}

	// Get the configured backend for the cluster. Backend types currently supported: local, s3.
	b, diags := getConfiguredBackend(lokoConfig)
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			ctxLogger.Error(diagnostic.Error())
		}
		ctxLogger.Fatal("Errors found while loading cluster configuration")
	}

	// Use a local backend if no backend is configured.
	if b == nil {
		b = local.NewLocalBackend()
	}

	assetDir, err := homedir.Expand(p.GetAssetDir())
	if err != nil {
		ctxLogger.Fatalf("error expanding path: %v", err)
	}

	// Validate backend configuration.
	if err = b.Validate(); err != nil {
		ctxLogger.Fatalf("Failed to validate backend configuration: %v", err)
	}

	// Render backend configuration.
	renderedBackend, err := b.Render()
	if err != nil {
		ctxLogger.Fatalf("Failed to render backend configuration file: %v", err)
	}

	// Configure Terraform directory, module and backend.
	if err = terraform.Configure(assetDir, renderedBackend); err != nil {
		ctxLogger.Fatalf("Failed to configure terraform : %v", err)
	}

	conf := terraform.Config{
		WorkingDir: terraform.GetTerraformRootDir(assetDir),
		Quiet:      quiet,
	}

	ex, err := terraform.NewExecutor(conf)
	if err != nil {
		ctxLogger.Fatalf("Failed to create terraform executor: %v", err)
	}

	if err := p.Install(ex); err != nil {
		ctxLogger.Fatalf("error installing cluster: %v", err)
	}

	fmt.Printf("\nYour configurations are stored in %s\n", assetDir)

	kubeconfigPath := path.Join(assetDir, "cluster-assets", "auth", "kubeconfig")
	if err := verifyInstall(kubeconfigPath, p.GetExpectedNodes()); err != nil {
		ctxLogger.Fatalf("Verify cluster installation: %v", err)
	}

	var componentsToInstall []string
	for _, component := range lokoConfig.RootConfig.Components {
		componentsToInstall = append(componentsToInstall, component.Name)
	}

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
