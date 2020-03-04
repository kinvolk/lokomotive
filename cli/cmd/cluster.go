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
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/kube"
	"sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/backend/local"
	"github.com/kinvolk/lokomotive/pkg/config"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage a Lokomotive cluster",
}

func init() {
	RootCmd.AddCommand(clusterCmd)
}

// initialize does common initialization actions between cluster operations
// and returns created objects to the caller for further use.
func initialize(ctxLogger *logrus.Entry) (*terraform.Executor, platform.Platform, *config.Config, string) {
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

	ex := initializeTerraform(ctxLogger, p, b)

	return ex, p, lokoConfig, assetDir
}

// initializeTerraform initialized Terraform directory using given backend and platform
// and returns configured executor.
func initializeTerraform(ctxLogger *logrus.Entry, p platform.Platform, b backend.Backend) *terraform.Executor {
	assetDir, err := homedir.Expand(p.GetAssetDir())
	if err != nil {
		ctxLogger.Fatalf("error expanding path: %v", err)
	}

	// Render backend configuration.
	renderedBackend, err := b.Render()
	if err != nil {
		ctxLogger.Fatalf("Failed to render backend configuration file: %v", err)
	}

	// Configure Terraform directory, module and backend.
	if err := terraform.Configure(assetDir, renderedBackend); err != nil {
		ctxLogger.Fatalf("Failed to configure Terraform : %v", err)
	}

	conf := terraform.Config{
		WorkingDir: terraform.GetTerraformRootDir(assetDir),
		Quiet:      quiet,
	}

	ex, err := terraform.NewExecutor(conf)
	if err != nil {
		ctxLogger.Fatalf("Failed to create Terraform executor: %v", err)
	}

	if err := p.Initialize(ex); err != nil {
		ctxLogger.Fatalf("Failed to initialize Platform: %v", err)
	}

	if err := ex.Init(); err != nil {
		ctxLogger.Fatalf("Failed to initialize Terraform: %v", err)
	}

	return ex
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

func upgradeControlplaneComponent(component string, kubeconfigPath string, assetDir string, ctxLogger *logrus.Entry, ex *terraform.Executor) {
	ctxLogger = ctxLogger.WithFields(logrus.Fields{
		"action":    "controlplane-upgrade",
		"component": component,
	})

	actionConfig := &action.Configuration{}

	if err := actionConfig.Init(kube.GetConfig(kubeconfigPath, "", "kube-system"), "kube-system", "secret", func(format string, v ...interface{}) {}); err != nil {
		ctxLogger.Fatalf("Failed initializing helm: %v", err)
	}

	helmChart, err := loader.Load(filepath.Join(assetDir, "/lokomotive-kubernetes/bootkube/resources/charts", component))
	if err != nil {
		ctxLogger.Fatalf("Loading chart from assets failed: %v", err)
	}

	if err := helmChart.Validate(); err != nil {
		ctxLogger.Fatalf("chart is invalid: %v", err)
	}

	valuesRaw := ""
	if err := ex.Output(fmt.Sprintf("%s_values", component), &valuesRaw); err != nil {
		ctxLogger.Fatalf("Failed to get kubernetes values.yaml from Terraform: %v", err)
	}

	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(valuesRaw), &values); err != nil {
		ctxLogger.Fatalf("Failed to parse values.yaml for kubernetes: %v", err)
	}

	update := action.NewUpgrade(actionConfig)

	if _, err := update.Run(component, helmChart, values); err != nil {
		ctxLogger.Fatalf("updating chart failed: %v", err)
	}
}
