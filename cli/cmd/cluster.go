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
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/backend/local"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/config"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
	"github.com/kinvolk/lokomotive/pkg/util/retryutil"
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

func getControlplaneChart(assetDir string, name string) (*chart.Chart, error) {
	helmChart, err := loader.Load(filepath.Join(assetDir, "/lokomotive-kubernetes/bootkube/resources/charts", name))
	if err != nil {
		return nil, fmt.Errorf("loading chart from assets failed: %w", err)
	}

	if err := helmChart.Validate(); err != nil {
		return nil, fmt.Errorf("chart is invalid: %w", err)
	}

	return helmChart, nil
}

func getControlplaneValues(ex *terraform.Executor, name string) (map[string]interface{}, error) {
	valuesRaw := ""
	if err := ex.Output(fmt.Sprintf("%s_values", name), &valuesRaw); err != nil {
		return nil, fmt.Errorf("failed to get values.yaml from Terraform: %w", err)
	}

	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(valuesRaw), &values); err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml for '%s': %w", name, err)
	}

	return values, nil
}

const (
	upgradeRetryInterval = 10 * time.Second
)

func upgradeKubeAPIServer(kubeconfigPath string, assetDir string, ex *terraform.Executor) error {
	var err error

	_ = retryutil.Retry(upgradeRetryInterval, 60, func() (bool, error) {
		if err = upgradeControlplaneComponent("kube-apiserver", kubeconfigPath, assetDir, ex); err != nil {
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		return fmt.Errorf("failed upgrading kube-apiserver: %w", err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed building Kubernetes client config: %w", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed building Kubernetes client: %w", err)
	}

	_ = retryutil.Retry(upgradeRetryInterval, 60, func() (bool, error) {
		var r bool
		if r, err = k8sutil.DaemonSetReady(client, "kube-system", "kube-apiserver"); err != nil {
			return false, nil
		}

		return r, nil
	})

	return nil
}

func upgradeControlplaneComponent(component string, kubeconfigPath string, assetDir string, ex *terraform.Executor) error {
	actionConfig, err := util.HelmActionConfig("kube-system", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed initializing helm: %w", err)
	}

	helmChart, err := getControlplaneChart(assetDir, component)
	if err != nil {
		return fmt.Errorf("loading chart from assets failed: %w", err)
	}

	values, err := getControlplaneValues(ex, component)
	if err != nil {
		return fmt.Errorf("failed to get values.yaml from Terraform: %w", err)
	}

	return installOrUpdateControlplaneComponent(component, *helmChart, values, *actionConfig)
}

func installOrUpdateControlplaneComponent(component string, helmChart chart.Chart, values map[string]interface{}, actionConfig action.Configuration) error {
	exists, err := util.ReleaseExists(actionConfig, component)
	if err != nil {
		return fmt.Errorf("failed checking if controlplane component is installed: %w", err)
	}

	if !exists {
		install := action.NewInstall(&actionConfig)
		install.ReleaseName = component
		install.Namespace = "kube-system"
		install.Atomic = true

		if _, err := install.Run(&helmChart, map[string]interface{}{}); err != nil {
			return fmt.Errorf("installing controlplane component failed: %w", err)
		}
	}

	update := action.NewUpgrade(&actionConfig)

	update.Atomic = true

	if _, err := update.Run(component, &helmChart, values); err != nil {
		return fmt.Errorf("updating chart failed: %w", err)
	}

	return nil
}
