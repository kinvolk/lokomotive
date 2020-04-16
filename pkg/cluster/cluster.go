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

// Package cluster handles the operations related to the Lokomotive cluster.
package cluster

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/hashicorp/hcl/v2"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/pkg/cluster/config"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/install"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	lokomotivepkg "github.com/kinvolk/lokomotive/pkg/lokomotive"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

// lokomotive manages the Lokomotive cluster related operations such as Apply,
// Destroy ,Health etc.
type lokomotive struct {
	Platform config.Platform
	Config   *config.LokomotiveConfig
	Executor *terraform.Executor
}

// NewLokomotive returns the an new lokomotive Instance
func NewLokomotive(cfg *config.LokomotiveConfig, options *Options) (Manager, hcl.Diagnostics) {
	// Set metadata
	if cfg.Platform != nil {
		cfg.Platform.SetMetadata(cfg.Metadata)
	}
	//validate configuration
	if diags := cfg.Validate(); diags.HasErrors() {
		return nil, diags
	}
	// Initialize Terraform Executor
	ex, err := terraform.InitializeExecutor(cfg.Metadata.AssetDir, options.Verbose)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to initialize Terraform executor: %v", err),
		}

		return nil, hcl.Diagnostics{diag}
	}

	return &lokomotive{
		Config:   cfg,
		Platform: cfg.Platform,
		Executor: ex,
	}, hcl.Diagnostics{}
}

// Apply creates the Lokomotive cluster
func (l *lokomotive) Apply(options *Options) error {
	if l.Config.Platform == nil {
		return fmt.Errorf("this operation is not permitted as 'platform' block is not configured")
	}

	assetDir := l.Config.Metadata.AssetDir
	if err := l.initializeTerraform(assetDir); err != nil {
		return err
	}
	// check if cluster exists
	exists, err := l.clusterExists()
	if err != nil {
		return err
	}
	// If cluster exists, reconcile the cluster state upon confirmation
	if exists && !options.Confirm {
		// TODO: We could plan to a file and use it when installing.
		if err := l.Executor.Plan(); err != nil {
			return fmt.Errorf("failed to reconcile cluster state: %v", err)
		}

		confirmation, err := askForConfirmation("Do you want to proceed with cluster apply?")
		if err != nil {
			return fmt.Errorf("error reading input: %v", err)
		}

		if confirmation {
			fmt.Println("Cluster apply cancelled")
			return nil
		}
	}

	if err := l.Platform.Apply(l.Executor); err != nil {
		return err
	}
	// Verify cluster creation
	kubeconfigPath := assetsKubeconfig(assetDir)
	if err := verifyCluster(kubeconfigPath, l.Platform.GetExpectedNodes(l.Config)); err != nil {
		return fmt.Errorf("error in verifying cluster: %v", err)
	}

	fmt.Printf("\nYour configurations are stored in %s\n", assetDir)

	// Do controlplane upgrades only if cluster already exists.
	if exists {
		fmt.Printf("\nEnsuring that cluster controlplane is up to date.\n")

		return l.updateControlPlane(options.UpgradeKubelets)
	}

	if !options.SkipComponents {
		// install all configured components
		componentsToApply := []string{}
		for name := range l.Config.Components {
			componentsToApply = append(componentsToApply, name)
		}

		return l.ApplyComponents(componentsToApply)
	}

	return nil
}

// ApplyComponents installs the components that are configured
func (l *lokomotive) ApplyComponents(args []string) error {
	componentsToApply := map[string]components.Component{}

	for _, name := range args {
		component, ok := l.Config.Components[name]
		if !ok {
			return fmt.Errorf("could not find configuration for the `%s` component to apply", name)
		}

		componentsToApply[name] = component
	}
	// Apply all components if length of args is zero.
	if len(args) == 0 {
		componentsToApply = l.Config.Components
	}

	dir := ""

	if l.Config.Metadata != nil {
		dir = l.Config.Metadata.AssetDir
	}

	kubeconfig := getKubeconfig(dir)

	for name, component := range componentsToApply {
		if err := util.InstallComponent(name, component, kubeconfig); err != nil {
			return err
		}

		fmt.Printf("Successfully applied component '%s' configuration!\n", name)
	}

	return nil
}

// RenderComponents renders the component manifests
func (l *lokomotive) RenderComponents(args []string) error {
	componentsToRender := map[string]components.Component{}

	for _, name := range args {
		component, ok := l.Config.Components[name]
		if !ok {
			return fmt.Errorf("could not find configuration for the `%s` component to render", name)
		}

		componentsToRender[name] = component
	}
	// Render all components if length of args is zero.
	if len(args) == 0 {
		componentsToRender = l.Config.Components
	}

	for name, component := range componentsToRender {
		manifests, err := component.RenderManifests()
		if err != nil {
			return err
		}

		fmt.Printf("# manifests for component %s\n", name)

		for filename, manifest := range manifests {
			fmt.Printf("\n---\n# %s\n%s", filename, manifest)
		}
	}

	return nil
}

// Destroy destroys the Lokomotive cluster
func (l *lokomotive) Destroy(options *Options) error {
	if l.Config.Platform == nil {
		return fmt.Errorf("this operation is not permitted as 'platform' block is not configured")
	}

	if err := l.initializeTerraform(l.Config.Metadata.AssetDir); err != nil {
		return err
	}

	exists, err := l.clusterExists()
	if err != nil {
		return fmt.Errorf("failed to check if the cluster exists: %q", err)
	}

	if !exists {
		fmt.Println("Cluster already destroyed, nothing to do")

		return nil
	}

	if !options.Confirm {
		warningStr := "WARNING: This action cannot be undone. Do you really want to destroy the cluster?"

		confirmation, err := askForConfirmation(warningStr)
		if err != nil {
			return fmt.Errorf("error reading input: %v", err)
		}

		if !confirmation {
			fmt.Println("Cluster destroy canceled")
			return nil
		}
	}

	return l.Platform.Destroy(l.Executor)
}

//nolint:funlen
// Health gets the health of the Lokomotive cluster.
func (l *lokomotive) Health() error {
	if l.Config.Platform == nil {
		return fmt.Errorf("this operation is not permitted as 'platform' block is not configured")
	}

	assetDir := l.Config.Metadata.AssetDir
	if err := doesKubeconfigExist(assetDir); err != nil {
		return err
	}

	exists, err := l.clusterExists()
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("no cluster found")
	}

	kubeconfig := getKubeconfig(assetDir)

	client, err := k8sutil.NewClientset(kubeconfig)
	if err != nil {
		return fmt.Errorf("error in setting up Kubernetes client: %q", err)
	}

	cluster, err := lokomotivepkg.NewCluster(client, l.Platform.GetExpectedNodes(l.Config))
	if err != nil {
		return fmt.Errorf("error in creating new Lokomotive cluster client: %q", err)
	}

	ns, err := cluster.GetNodeStatus()
	if err != nil {
		return fmt.Errorf("error getting node status: %q", err)
	}

	ns.PrettyPrint()

	if !ns.Ready() {
		return fmt.Errorf("the cluster is not completely ready")
	}

	components, err := cluster.Health()
	if err != nil {
		return fmt.Errorf("error in getting Lokomotive cluster health: %q", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	// Print the header.
	fmt.Fprintln(w, "Name\tStatus\tMessage\tError\t")

	// An empty line between header and the body.
	fmt.Fprintln(w, "\t\t\t\t")

	for _, component := range components {
		// The client-go library defines only one `ComponenetConditionType` at the moment,
		// which is `ComponentHealthy`. However, iterating over the list keeps this from
		// breaking in case client-go adds another `ComponentConditionType`.
		for _, condition := range component.Conditions {
			line := fmt.Sprintf(
				"%s\t%s\t%s\t%s\t",
				component.Name, condition.Status, condition.Message, condition.Error,
			)

			fmt.Fprintln(w, line)
		}

		if err := w.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func verifyCluster(kubeconfigPath string, expectedNodes int) error {
	client, err := k8sutil.NewClientset(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to set up clientset: %v", err)
	}

	cluster, err := lokomotivepkg.NewCluster(client, expectedNodes)
	if err != nil {
		return fmt.Errorf("failed to set up cluster client: %v", err)
	}

	return install.Verify(cluster)
}

func (l *lokomotive) clusterExists() (bool, error) {
	o := map[string]interface{}{}

	if err := l.Executor.Output("", &o); err != nil {
		return false, fmt.Errorf("failed to check if cluster exists: %v", err)
	}

	return len(o) != 0, nil
}

func (l *lokomotive) updateControlPlane(upgradeKubelets bool) error {
	//Update control plane
	releases := []string{"pod-checkpointer", "kube-apiserver", "kubernetes", "calico"}
	if upgradeKubelets {
		releases = append(releases, "kubelet")
	}

	for _, component := range releases {
		if err := l.upgradeComponent(component); err != nil {
			return fmt.Errorf("failed to update control plane component: %q", err)
		}
	}

	return nil
}

func (l *lokomotive) initializeTerraform(assetDir string) error {
	// Render backend configuration.
	renderedBackend, err := l.Config.Backend.Render()
	if err != nil {
		return fmt.Errorf("failed to render backend: %q", err)
	}
	// render platform configuration.
	renderedPlatform, err := l.Config.Platform.Render(l.Config)
	if err != nil {
		return fmt.Errorf("failed to render platform: %q", err)
	}

	// Configure Terraform directory, module and cluster and backend files.
	if err := terraform.Configure(assetDir, renderedBackend, renderedPlatform); err != nil {
		return fmt.Errorf("failed to configure Terraform directory: %q", err)
	}
	// Initialize Terraform
	if err := l.Executor.Init(); err != nil {
		return fmt.Errorf("failed to initialize terraform: %v", err)
	}

	return nil
}

func (l *lokomotive) getControlplaneChart(name string) (*chart.Chart, error) {
	chartPath := "/lokomotive-kubernetes/bootkube/resources/charts"

	helmChart, err := loader.Load(filepath.Join(l.Config.Metadata.AssetDir, chartPath, name))
	if err != nil {
		return nil, fmt.Errorf("loading chart from assets failed: %w", err)
	}

	if err := helmChart.Validate(); err != nil {
		return nil, fmt.Errorf("chart is invalid: %w", err)
	}

	return helmChart, nil
}

func (l *lokomotive) getControlplaneValues(name string) (map[string]interface{}, error) {
	valuesRaw := ""
	if err := l.Executor.Output(fmt.Sprintf("%s_values", name), &valuesRaw); err != nil {
		return nil, fmt.Errorf("failed to get controlplane component values.yaml from Terraform: %w", err)
	}

	values := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(valuesRaw), &values); err != nil {
		return nil, fmt.Errorf("failed to parse values.yaml for controlplane component: %w", err)
	}

	return values, nil
}

func (l *lokomotive) upgradeComponent(component string) error {
	// Get kubeconfig
	kubeconfigPath := getKubeconfig(l.Config.Metadata.AssetDir)

	actionConfig, err := util.HelmActionConfig("kube-system", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed initializing helm: %v", err)
	}

	helmChart, err := l.getControlplaneChart(component)
	if err != nil {
		return fmt.Errorf("loading chart from assets failed: %v", err)
	}

	values, err := l.getControlplaneValues(component)
	if err != nil {
		return fmt.Errorf("failed to get kubernetes values.yaml from Terraform: %v", err)
	}

	exists, err := util.ReleaseExists(*actionConfig, component)
	if err != nil {
		return fmt.Errorf("failed checking if controlplane component is '%s' installed: %v", component, err)
	}

	if !exists {
		fmt.Printf("controlplane component '%s' is missing, reinstalling...", component)

		install := action.NewInstall(actionConfig)
		install.ReleaseName = component
		install.Namespace = "kube-system"
		install.Atomic = true

		if _, err := install.Run(helmChart, map[string]interface{}{}); err != nil {
			fmt.Println("Failed!")

			return fmt.Errorf("installing controlplane component '%s' failed: %v", component, err)
		}

		fmt.Println("Done.")
	}

	update := action.NewUpgrade(actionConfig)

	update.Atomic = true

	fmt.Printf("Ensuring controlplane component '%s' is up to date... ", component)

	if _, err := update.Run(component, helmChart, values); err != nil {
		fmt.Println("Failed!")

		return fmt.Errorf("updating chart failed: %v", err)
	}

	fmt.Println("Done.")

	return nil
}
