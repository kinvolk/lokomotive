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

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/backend/local"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/config"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

// cluster is a temporary helper struct to aggregate objects which are used
// for managing the cluster and components.
type cluster struct {
	terraformExecutor terraform.Executor
	platform          platform.Platform
	lokomotiveConfig  *config.Config
	assetDir          string
}

type clusterConfig struct {
	verbose    bool
	configPath string
	valuesPath string
}

// initialize does common initialization actions between cluster operations
// and returns created objects to the caller for further use.
func (cc clusterConfig) initialize(contextLogger *log.Entry) (*cluster, error) {
	lokoConfig, diags := config.LoadConfig(cc.configPath, cc.valuesPath)
	if diags.HasErrors() {
		return nil, diags
	}

	p, diags := getConfiguredPlatform(lokoConfig, true)
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			contextLogger.Error(diagnostic.Error())
		}

		return nil, fmt.Errorf("loading platform configuration")
	}

	// Get the configured backend for the cluster. Backend types currently supported: local, s3.
	b, diags := getConfiguredBackend(lokoConfig)
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			contextLogger.Error(diagnostic.Error())
		}

		return nil, fmt.Errorf("loading backend configuration")
	}

	// Use a local backend if no backend is configured.
	if b == nil {
		b = local.NewLocalBackend()
	}

	assetDir, err := homedir.Expand(p.Meta().AssetDir)
	if err != nil {
		return nil, fmt.Errorf("expanding path %q: %v", p.Meta().AssetDir, err)
	}

	// Validate backend configuration.
	if err = b.Validate(); err != nil {
		return nil, fmt.Errorf("validating backend configuration: %v", err)
	}

	ex, err := cc.initializeTerraform(p, b)
	if err != nil {
		return nil, fmt.Errorf("initializing Terraform: %w", err)
	}

	return &cluster{
		terraformExecutor: *ex,
		platform:          p,
		lokomotiveConfig:  lokoConfig,
		assetDir:          assetDir,
	}, nil
}

// initializeTerraform initialized Terraform directory using given backend and platform
// and returns configured executor.
func (cc clusterConfig) initializeTerraform(p platform.Platform, b backend.Backend) (*terraform.Executor, error) {
	assetDir, err := homedir.Expand(p.Meta().AssetDir)
	if err != nil {
		return nil, fmt.Errorf("expanding path %q: %w", p.Meta().AssetDir, err)
	}

	// Render backend configuration.
	renderedBackend, err := b.Render()
	if err != nil {
		return nil, fmt.Errorf("rendering backend configuration: %w", err)
	}

	// Configure Terraform directory, module and backend.
	if err := terraform.Configure(assetDir, renderedBackend); err != nil {
		return nil, fmt.Errorf("configuring Terraform: %w", err)
	}

	conf := terraform.Config{
		WorkingDir: terraform.GetTerraformRootDir(assetDir),
		Verbose:    cc.verbose,
	}

	ex, err := terraform.NewExecutor(conf)
	if err != nil {
		return nil, fmt.Errorf("creating Terraform executor: %w", err)
	}

	if err := p.Initialize(ex); err != nil {
		return nil, fmt.Errorf("initializing Platform: %w", err)
	}

	if err := ex.Init(); err != nil {
		return nil, fmt.Errorf("running 'terraform init': %w", err)
	}

	return ex, nil
}

// clusterExists determines if cluster has already been created by getting all
// outputs from the Terraform. If there is any output defined, it means 'terraform apply'
// run at least once.
func clusterExists(ex terraform.Executor) (bool, error) {
	o := map[string]interface{}{}

	if err := ex.Output("", &o); err != nil {
		return false, fmt.Errorf("getting Terraform output: %w", err)
	}

	return len(o) != 0, nil
}

type controlplaneUpdater struct {
	kubeconfig    []byte
	assetDir      string
	contextLogger log.Entry
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

func (c controlplaneUpdater) upgradeComponent(component, namespace string) error {
	actionConfig, err := util.HelmActionConfig(namespace, c.kubeconfig)
	if err != nil {
		return fmt.Errorf("initializing Helm action: %w", err)
	}

	helmChart, err := c.getControlplaneChart(component)
	if err != nil {
		return fmt.Errorf("loading chart from assets: %w", err)
	}

	values, err := c.getControlplaneValues(component)
	if err != nil {
		return fmt.Errorf("getting chart values from Terraform: %w", err)
	}

	exists, err := util.ReleaseExists(*actionConfig, component)
	if err != nil {
		return fmt.Errorf("checking if controlplane component is installed: %w", err)
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

			return fmt.Errorf("installing controlplane component: %w", err)
		}

		fmt.Println("Done.")
	}

	update := action.NewUpgrade(actionConfig)

	update.Atomic = true

	fmt.Printf("Ensuring controlplane component '%s' is up to date... ", component)

	if _, err := update.Run(component, helmChart, values); err != nil {
		fmt.Println("Failed!")

		return fmt.Errorf("updating controlplane component: %w", err)
	}

	fmt.Println("Done.")

	return nil
}
