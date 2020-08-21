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
	"github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/backend/local"
	"github.com/kinvolk/lokomotive/pkg/backend/s3"
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

//nolint:funlen
func initialize(ctxLogger *logrus.Entry) (*config.Config, platform.Cluster, *terraform.Executor) {
	// Read cluster config from HCL files.
	cp := viper.GetString("lokocfg")
	vp := viper.GetString("lokocfg-vars")
	cc, diags := config.LoadConfig(cp, vp)
	if len(diags) > 0 {
		ctxLogger.Fatal(diags)
	}

	if cc.RootConfig.Cluster == nil {
		// No `cluster` block specified in the configuration.
		ctxLogger.Fatal("No cluster configured")
	}

	// Construct a Cluster.
	c := createCluster(ctxLogger, cc)

	assetDir, err := homedir.Expand(c.AssetDir())
	if err != nil {
		ctxLogger.Fatalf("Error expanding path: %v", err)
	}

	// Write Terraform modules to disk.
	terraformModuleDir := filepath.Join(assetDir, "terraform-modules")
	if err := assets.Extract(assets.TerraformModulesSource, terraformModuleDir); err != nil {
		ctxLogger.Fatalf("Writing Terraform files to disk: %v", err)
	}

	// Create Terraform root directory.
	terraformRootDir := filepath.Join(assetDir, "terraform")
	if err := os.MkdirAll(terraformRootDir, 0750); err != nil {
		ctxLogger.Fatalf("Creating Terraform root directory at %q: %v", terraformRootDir, err)
	}

	// Render backend configuration.
	var renderedBackend string

	if cc.RootConfig.Backend != nil {
		b := createBackend(ctxLogger, cc)

		renderedBackend = b.String()
	}

	// Create backend file if the backend configuration isn't empty.
	if len(strings.TrimSpace(renderedBackend)) > 0 {
		path := filepath.Join(terraformRootDir, "backend.tf")
		if err := ioutil.WriteFile(path, []byte(renderedBackend), 0600); err != nil {
			ctxLogger.Fatalf("Failed to write backend file %q to disk: %v", path, err)
		}
	}

	// Write control plane chart files to disk.
	for _, chart := range c.ControlPlaneCharts() {
		src := filepath.Join(assets.ControlPlaneSource, chart.Name)
		dst := filepath.Join(assetDir, "cluster-assets", "charts", "kube-system", chart.Name)

		if err := assets.Extract(src, dst); err != nil {
			ctxLogger.Fatalf("Failed to extract charts: %v", err)
		}
	}

	// Write Terraform root module to disk.
	path := filepath.Join(terraformRootDir, "cluster.tf")
	if err := ioutil.WriteFile(path, []byte(c.TerraformRootModule()), 0600); err != nil {
		ctxLogger.Fatalf("Failed to write Terraform root module %q to disk: %v", path, err)
	}

	// Construct Terraform executor.
	ex, err := terraform.NewExecutor(terraform.Config{
		WorkingDir: filepath.Join(assetDir, "terraform"),
		Verbose:    verbose,
	})
	if err != nil {
		ctxLogger.Fatalf("Failed to create Terraform executor: %v", err)
	}

	// Execute `terraform init`.
	if err := ex.Init(); err != nil {
		ctxLogger.Fatalf("Failed to initialize Terraform: %v", err)
	}

	return cc, c, ex
}

// clusterExists determines if cluster has already been created by getting all
// outputs from the Terraform. If there is any output defined, it means 'terraform apply'
// run at least once.
func clusterExists(ctxLogger *logrus.Entry, ex *terraform.Executor) bool {
	o := map[string]interface{}{}

	if err := ex.Output("", &o); err != nil {
		ctxLogger.Fatalf("Failed to check if cluster exists: %v", err)
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

// createBackend constructs a Backend based on the provided cluster config and returns a pointer to
// it. If a backend with the provided name doesn't exist, an error is returned.
func createBackend(logger *logrus.Entry, config *config.Config) backend.Backend {
	bn := config.RootConfig.Backend.Name

	switch bn {
	case backend.Local:
		bc, diags := local.NewConfig(&config.RootConfig.Backend.Config, config.EvalContext)
		if diags.HasErrors() {
			for _, diagnostic := range diags {
				logger.Error(diagnostic.Error())
			}

			logger.Fatal("Errors found while loading backend configuration")
		}

		b, err := local.NewBackend(bc)
		if err != nil {
			logger.Fatalf("Error constructing backend: %v", err)
		}

		return b
	case backend.S3:
		bc, diags := s3.NewConfig(&config.RootConfig.Backend.Config, config.EvalContext)
		if diags.HasErrors() {
			for _, diagnostic := range diags {
				logger.Error(diagnostic.Error())
			}

			logger.Fatal("Errors found while loading backend configuration")
		}

		b, err := s3.NewBackend(bc)
		if err != nil {
			logger.Fatalf("Error constructing backend: %v", err)
		}

		return b
	}

	logger.Fatalf("Unknown backend %q", bn)

	return nil
}

type controlplaneUpdater struct {
	kubeconfig []byte
	assetDir   string
	ctxLogger  logrus.Entry
	ex         terraform.Executor
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
	p := filepath.Join(c.assetDir, "cluster-assets", "charts", "kube-system", name+".yaml")

	v, err := ioutil.ReadFile(p) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read Helm values file %q: %w", p, err)
	}

	values := map[string]interface{}{}
	if err := yaml.Unmarshal(v, &values); err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml for controlplane component: %w", err)
	}

	return values, nil
}

func (c controlplaneUpdater) upgradeComponent(component, namespace string) {
	ctxLogger := c.ctxLogger.WithFields(logrus.Fields{
		"action":    "controlplane-upgrade",
		"component": component,
	})

	actionConfig, err := util.HelmActionConfig(namespace, c.kubeconfig)
	if err != nil {
		ctxLogger.Fatalf("Failed initializing helm: %v", err)
	}

	helmChart, err := c.getControlplaneChart(component)
	if err != nil {
		ctxLogger.Fatalf("Loading chart from assets failed: %v", err)
	}

	values, err := c.getControlplaneValues(component)
	if err != nil {
		ctxLogger.Fatalf("Failed to get kubernetes values.yaml from Terraform: %v", err)
	}

	exists, err := util.ReleaseExists(*actionConfig, component)
	if err != nil {
		ctxLogger.Fatalf("Failed checking if controlplane component is installed: %v", err)
	}

	if !exists {
		fmt.Printf("Controlplane component '%s' is missing, reinstalling...", component)

		install := action.NewInstall(actionConfig)
		install.ReleaseName = component
		install.Namespace = namespace
		install.Atomic = true

		if _, err := install.Run(helmChart, values); err != nil {
			fmt.Println("Failed!")

			ctxLogger.Fatalf("Installing controlplane component failed: %v", err)
		}

		fmt.Println("Done.")
	}

	update := action.NewUpgrade(actionConfig)

	update.Atomic = true

	fmt.Printf("Ensuring controlplane component '%s' is up to date... ", component)

	if _, err := update.Run(component, helmChart, values); err != nil {
		fmt.Println("Failed!")

		ctxLogger.Fatalf("Updating chart failed: %v", err)
	}

	fmt.Println("Done.")
}
