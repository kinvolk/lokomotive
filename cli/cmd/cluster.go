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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/backend/local"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/config"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/platform/aks"
	"github.com/kinvolk/lokomotive/pkg/platform/aws"
	"github.com/kinvolk/lokomotive/pkg/platform/baremetal"
	"github.com/kinvolk/lokomotive/pkg/platform/packet"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage a cluster",
}

func init() {
	RootCmd.AddCommand(clusterCmd)
}

// initialize does common initialization actions between cluster operations
// and returns created objects to the caller for further use.
func initialize(contextLogger *logrus.Entry) (*config.Config, platform.Cluster, *terraform.Executor) {
	// Read cluster config from HCL files.
	cp := viper.GetString("lokocfg")
	vp := viper.GetString("lokocfg-vars")
	cc, diags := config.LoadConfig(cp, vp)
	if diags.HasErrors() {
		contextLogger.Fatal(diags)
	}

	if cc.RootConfig.Cluster == nil {
		// No `cluster` block specified in the configuration.
		contextLogger.Fatal("No cluster configured")
	}

	// Construct a Cluster.
	c := createCluster(contextLogger, cc)

	assetDir, err := homedir.Expand(c.AssetDir())
	if err != nil {
		contextLogger.Fatalf("Error expanding path: %v", err)
	}

	// Write Terraform modules to disk.
	terraformModuleDir := filepath.Join(assetDir, assets.TerraformModulesSource)
	if err := assets.Extract(assets.TerraformModulesSource, terraformModuleDir); err != nil {
		contextLogger.Fatalf("Writing Terraform files to disk: %v", err)
	}

	// Create Terraform root directory.
	terraformRootDir := filepath.Join(assetDir, "terraform")
	if err := os.MkdirAll(terraformRootDir, 0750); err != nil {
		contextLogger.Fatalf("Creating Terraform root directory at %q: %v", terraformRootDir, err)
	}

	// Get the configured backend for the cluster. Backend types currently supported: local, s3.
	b, diags := getConfiguredBackend(cc)
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			contextLogger.Error(diagnostic.Error())
		}

		contextLogger.Fatal("Errors found while loading cluster configuration")
	}

	// Use a local backend if no backend is configured.
	if b == nil {
		b = local.NewLocalBackend()
	}

	// Validate backend configuration.
	if err = b.Validate(); err != nil {
		contextLogger.Fatalf("Validating backend configuration: %v", err)
	}

	// Create backend file if the backend configuration isn't empty.
	rb, err := b.Render()
	if err != nil {
		contextLogger.Fatalf("Rendering backend: %v", err)
	}

	if len(strings.TrimSpace(rb)) > 0 {
		path := filepath.Join(terraformRootDir, "backend.tf")

		// TODO: Refactor backend template handling.
		if err := ioutil.WriteFile(path, []byte(fmt.Sprintf("terraform {%s}\n", rb)), 0600); err != nil {
			contextLogger.Fatalf("Failed to write backend file %q to disk: %v", path, err)
		}
	}

	// Write control plane chart files to disk.
	for _, chart := range c.ControlPlaneCharts() {
		src := filepath.Join(assets.ControlPlaneSource, chart.Name)
		dst := filepath.Join(assetDir, "cluster-assets", "charts", "kube-system", chart.Name)

		if err := assets.Extract(src, dst); err != nil {
			contextLogger.Fatalf("Failed to extract charts: %v", err)
		}
	}

	// Write Terraform root module to disk.
	path := filepath.Join(terraformRootDir, "cluster.tf")
	if err := ioutil.WriteFile(path, []byte(c.TerraformRootModule()), 0600); err != nil {
		contextLogger.Fatalf("Failed to write Terraform root module %q to disk: %v", path, err)
	}

	// Construct Terraform executor.
	ex, err := terraform.NewExecutor(terraform.Config{
		WorkingDir: filepath.Join(assetDir, "terraform"),
		Verbose:    verbose,
	})
	if err != nil {
		contextLogger.Fatalf("Failed to create Terraform executor: %v", err)
	}

	// Execute `terraform init`.
	if err := ex.Init(); err != nil {
		contextLogger.Fatalf("Failed to initialize Terraform: %v", err)
	}

	return cc, c, ex
}

// clusterExists determines if cluster has already been created by getting all
// outputs from the Terraform. If there is any output defined, it means 'terraform apply'
// run at least once.
func clusterExists(contextLogger *logrus.Entry, ex *terraform.Executor) bool {
	o := map[string]interface{}{}

	if err := ex.Output("", &o); err != nil {
		contextLogger.Fatalf("Failed to check if cluster exists: %v", err)
	}

	return len(o) != 0
}

// createCluster constructs a Cluster based on the provided cluster config and returns a pointer to
// it.
//nolint:funlen
func createCluster(logger *logrus.Entry, config *config.Config) platform.Cluster {
	p := config.RootConfig.Cluster.Name

	switch p {
	case platform.Packet:
		pc, diags := packet.NewConfig(&config.RootConfig.Cluster.Config, config.EvalContext)
		if diags.HasErrors() {
			for _, diagnostic := range diags {
				logger.Error(diagnostic.Error())
			}

			logger.Fatal("Errors found while loading cluster configuration")
		}

		c, err := packet.NewCluster(pc)
		if err != nil {
			logger.Fatalf("Error constructing cluster: %v", err)
		}

		return c
	case platform.AKS:
		pc, diags := aks.NewConfig(&config.RootConfig.Cluster.Config, config.EvalContext)
		if diags.HasErrors() {
			for _, diagnostic := range diags {
				logger.Error(diagnostic.Error())
			}

			logger.Fatal("Errors found while loading cluster configuration")
		}

		c, err := aks.NewCluster(pc)
		if err != nil {
			logger.Fatalf("Error constructing cluster: %v", err)
		}

		return c
	case platform.AWS:
		pc, diags := aws.NewConfig(&config.RootConfig.Cluster.Config, config.EvalContext)
		if diags.HasErrors() {
			for _, diagnostic := range diags {
				logger.Error(diagnostic.Error())
			}

			logger.Fatal("Errors found while loading cluster configuration")
		}

		c, err := aws.NewCluster(pc)
		if err != nil {
			logger.Fatalf("Error constructing cluster: %v", err)
		}

		return c
	case platform.BareMetal:
		pc, diags := baremetal.NewConfig(&config.RootConfig.Cluster.Config, config.EvalContext)
		if diags.HasErrors() {
			for _, diagnostic := range diags {
				logger.Error(diagnostic.Error())
			}

			logger.Fatal("Errors found while loading cluster configuration")
		}

		c, err := baremetal.NewCluster(pc)
		if err != nil {
			logger.Fatalf("Error constructing cluster: %v", err)
		}

		return c
	}

	logger.Fatalf("Unknown platform %q", p)

	return nil
}

type controlplaneUpdater struct {
	kubeconfig    []byte
	assetDir      string
	contextLogger logrus.Entry
	ex            terraform.Executor
}

func (c controlplaneUpdater) getControlplaneChart(name string) (*chart.Chart, error) {
	chart, err := platform.ControlPlaneChart(name)
	if err != nil {
		return nil, fmt.Errorf("loading chart from assets failed: %w", err)
	}

	if err := chart.Validate(); err != nil {
		return nil, fmt.Errorf("chart is invalid: %w", err)
	}

	return chart, nil
}

func (c controlplaneUpdater) getControlplaneValues(name string) (map[string]interface{}, error) {
	valuesRaw := ""
	if err := c.ex.Output(fmt.Sprintf("%s_values", name), &valuesRaw); err != nil {
		return nil, fmt.Errorf("failed to get controlplane component values.yaml from Terraform: %w", err)
	}

	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(valuesRaw), &values); err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml for controlplane component: %w", err)
	}

	return values, nil
}

func (c controlplaneUpdater) upgradeComponent(component, namespace string) {
	contextLogger := c.contextLogger.WithFields(logrus.Fields{
		"action":    "controlplane-upgrade",
		"component": component,
	})

	actionConfig, err := util.HelmActionConfig(namespace, c.kubeconfig)
	if err != nil {
		contextLogger.Fatalf("Failed initializing helm: %v", err)
	}

	helmChart, err := c.getControlplaneChart(component)
	if err != nil {
		contextLogger.Fatalf("Loading chart from assets failed: %v", err)
	}

	values, err := c.getControlplaneValues(component)
	if err != nil {
		contextLogger.Fatalf("Failed to get kubernetes values.yaml from Terraform: %v", err)
	}

	exists, err := util.ReleaseExists(*actionConfig, component)
	if err != nil {
		contextLogger.Fatalf("Failed checking if controlplane component is installed: %v", err)
	}

	if !exists {
		fmt.Printf("Controlplane component '%s' is missing, reinstalling...", component)

		install := action.NewInstall(actionConfig)
		install.ReleaseName = component
		install.Namespace = namespace
		install.Atomic = true
		install.CreateNamespace = true

		if _, err := install.Run(helmChart, values); err != nil {
			fmt.Println("Failed!")

			contextLogger.Fatalf("Installing controlplane component failed: %v", err)
		}

		fmt.Println("Done.")
	}

	update := action.NewUpgrade(actionConfig)

	update.Atomic = true

	fmt.Printf("Ensuring controlplane component '%s' is up to date... ", component)

	if _, err := update.Run(component, helmChart, values); err != nil {
		fmt.Println("Failed!")

		contextLogger.Fatalf("Updating chart failed: %v", err)
	}

	fmt.Println("Done.")
}
