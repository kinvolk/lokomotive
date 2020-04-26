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
package lokomotive

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"sigs.k8s.io/yaml"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"github.com/kinvolk/lokomotive/pkg/lokomotive/config"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

// lokomotive manages the Lokomotive cluster related operations such as Apply,
// Destroy ,Health etc.
type lokomotive struct {
	Logger   *logrus.Entry
	Platform platform.Platform
	Config   *config.LokomotiveConfig
	Executor *terraform.Executor
}

// NewLokomotive returns the an new lokomotive Instance
func NewLokomotive(ctxLogger *logrus.Entry, cfg *config.LokomotiveConfig, options *Options) (Manager, hcl.Diagnostics) {
	// Initialize Terraform Executor
	ex, err := terraform.InitializeExecutor(cfg.Platform.GetAssetDir(), options.Verbose)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("failed to initialize Terraform executor: %v", err),
		}

		return nil, hcl.Diagnostics{diag}
	}

	return &lokomotive{
		Logger:   ctxLogger,
		Config:   cfg,
		Platform: cfg.Platform,
		Executor: ex,
	}, hcl.Diagnostics{}
}

// Apply creates the Lokomotive cluster
func (l *lokomotive) Apply(options *Options) {
	if l.Config.Platform == nil {
		l.Logger.Fatalf("This operation is not permitted as 'platform' block is not configured")
	}

	assetDir := l.Config.Platform.GetAssetDir()
	if err := l.initializeTerraform(assetDir); err != nil {
		l.Logger.Fatalf("Error intializing terraform: %v", err)
	}
	// check if cluster exists
	exists, err := l.clusterExists()
	if err != nil {
		l.Logger.Fatalf("Error checking for existing cluster: %v", err)
	}
	// If cluster exists, reconcile the cluster state upon confirmation
	if exists && !options.Confirm {
		// TODO: We could plan to a file and use it when installing.
		if err := l.Executor.Plan(); err != nil {
			l.Logger.Fatalf("failed to reconcile cluster state: %v", err)
		}

		confirmation, err := askForConfirmation("Do you want to proceed with cluster apply?")
		if err != nil {
			l.Logger.Fatalf("error reading input: %v", err)
		}

		if confirmation {
			l.Logger.Infoln("Cluster apply cancelled")
			return
		}
	}

	if err := l.Platform.Apply(l.Executor); err != nil {
		l.Logger.Fatalf("Error creating Lokomotive cluster: %v", err)
	}
	// Verify cluster creation
	kubeconfigPath := assetsKubeconfig(assetDir)
	if err := verifyCluster(kubeconfigPath, l.Platform.GetExpectedNodes()); err != nil {
		l.Logger.Fatalf("Error in verifying cluster: %v", err)
	}

	l.Logger.Infof("\nYour configurations are stored in %s\n", assetDir)

	// Do controlplane upgrades only if cluster already exists.
	if exists {
		l.Logger.Infof("\nEnsuring that cluster controlplane is up to date.\n")

		if err := l.updateControlPlane(options.UpgradeKubelets); err != nil {
			l.Logger.Fatalf("Error updating Lokomotive control plane: %v", err)
		}
	}

	if !options.SkipComponents {
		// install all configured components
		componentsToApply := []string{}
		for name := range l.Config.Components {
			componentsToApply = append(componentsToApply, name)
		}

		l.ApplyComponents(componentsToApply)
	}
}

// Destroy destroys the Lokomotive cluster
func (l *lokomotive) Destroy(options *Options) {
	if l.Config.Platform == nil {
		l.Logger.Fatalf("this operation is not permitted as 'platform' block is not configured")
	}

	if err := l.initializeTerraform(l.Platform.GetAssetDir()); err != nil {
		l.Logger.Fatalf("Error intializing terraform: %v", err)
	}

	exists, err := l.clusterExists()
	if err != nil {
		l.Logger.Fatalf("Error checking for existing cluster: %v", err)
	}

	if !exists {
		l.Logger.Infoln("Cluster already destroyed, nothing to do")

		return
	}

	if !options.Confirm {
		warningStr := "WARNING: This action cannot be undone. Do you really want to destroy the cluster?"

		confirmation, err := askForConfirmation(warningStr)
		if err != nil {
			l.Logger.Fatalf("error reading input: %v", err)
		}

		if !confirmation {
			l.Logger.Infoln("Cluster destroy canceled")
			return
		}
	}

	if err := l.Platform.Destroy(l.Executor); err != nil {
		l.Logger.Fatalf("Error destroying Lokomotive cluster: %v", err)
	}
}

// ApplyComponents installs the components that are configured
func (l *lokomotive) ApplyComponents(args []string) {
	componentsToApply := map[string]components.Component{}

	for _, name := range args {
		component, ok := l.Config.Components[name]
		if !ok {
			l.Logger.Fatalf("could not find configuration for the `%s` component to apply", name)
		}

		componentsToApply[name] = component
	}
	// Apply all components if length of args is zero.
	if len(args) == 0 {
		componentsToApply = l.Config.Components
	}

	kubeconfig := getKubeconfig(l.Platform.GetAssetDir())

	for name, component := range componentsToApply {
		if err := util.InstallComponent(name, component, kubeconfig); err != nil {
			l.Logger.Fatalf("Error installing component '%s' : %v", name, err)
		}

		l.Logger.Infof("Successfully applied component '%s' configuration!\n", name)
	}
}

// RenderComponents renders the component manifests
func (l *lokomotive) RenderComponents(args []string) {
	componentsToRender := map[string]components.Component{}

	for _, name := range args {
		component, ok := l.Config.Components[name]
		if !ok {
			l.Logger.Fatalf("could not find configuration for the `%s` component to render", name)
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
			l.Logger.Fatalf("Error rendering component '%s' manifests: %v", name, err)
		}

		fmt.Printf("# manifests for component %s\n", name)

		for filename, manifest := range manifests {
			fmt.Printf("\n---\n# %s\n%s", filename, manifest)
		}
	}
}

func (l *lokomotive) DeleteComponents(args []string, options *Options) {
	//TODO: This check brings UI logic in lokomotive package
	// Ideally `askForConfirmation` should be part of the cmd package.
	confirmationStr := fmt.Sprintf("The following components will be deleted:\n\t%s\n\nAre you sure you want to proceed?",
		strings.Join(args, "\n\t"))

	confirmation, err := askForConfirmation(confirmationStr)
	if err != nil {
		l.Logger.Fatalf("error reading input: %v", err)
	}

	if !confirmation {
		l.Logger.Info("Components deletion cancelled.")
		return
	}

	componentsToDelete := map[string]components.Component{}
	// Delete all configured components if length of args is zero.
	if len(args) == 0 {
		componentsToDelete = l.Config.Components
	}

	// Create a map of all component objects to be deleted.
	for _, name := range args {
		c, err := components.Get(name)
		if err != nil {
			l.Logger.Fatalf("Unsupported component, got: %v", err)
		}
		componentsToDelete[name] = c
	}

	kubeconfig := getKubeconfig(l.Platform.GetAssetDir())

	for name, component := range componentsToDelete {
		l.Logger.Infof("Deleting component '%s'...\n", name)
		if err := l.deleteHelmRelease(component, kubeconfig, options.DeleteNamespace); err != nil {
			l.Logger.Fatalf("Error deleting component '%s': %v", name, err)
		}

		l.Logger.Infof("Successfully deleted component %q!\n", name)
	}

	// Add a line to distinguish between info logs and errors, if any.
	l.Logger.Println()
}

//nolint:funlen
// Health gets the health of the Lokomotive cluster.
func (l *lokomotive) Health() {
	if l.Config.Platform == nil {
		l.Logger.Fatalf("this operation is not permitted as 'platform' block is not configured")
	}

	assetDir := l.Platform.GetAssetDir()
	if err := doesKubeconfigExist(assetDir); err != nil {
		l.Logger.Fatalf("Error finding kubeconfig: %v", err)
	}

	exists, err := l.clusterExists()
	if err != nil {
		l.Logger.Fatalf("Error checking for existing cluster: %v", err)
	}

	if !exists {
		l.Logger.Fatalf("no cluster found")
	}

	kubeconfig := getKubeconfig(assetDir)

	client, err := k8sutil.NewClientset(kubeconfig)
	if err != nil {
		l.Logger.Fatalf("error in setting up Kubernetes client: %v", err)
	}

	cluster, err := NewCluster(client, l.Platform.GetExpectedNodes())
	if err != nil {
		l.Logger.Fatalf("error in creating new Lokomotive cluster client: %q", err)
	}

	ns, err := cluster.GetNodeStatus()
	if err != nil {
		l.Logger.Fatalf("error getting node status: %q", err)
	}

	ns.PrettyPrint()

	if !ns.Ready() {
		l.Logger.Fatalf("the cluster is not completely ready")
	}

	components, err := cluster.Health()
	if err != nil {
		l.Logger.Fatalf("error in getting Lokomotive cluster health: %q", err)
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
			l.Logger.Fatalf(err.Error())
		}
	}
}

func (l *lokomotive) initializeTerraform(assetDir string) error {
	// Render backend configuration.
	renderedBackend, err := l.Config.Backend.Render()
	if err != nil {
		return fmt.Errorf("failed to render backend: %q", err)
	}
	// render platform configuration.
	renderedPlatform, err := l.Config.Platform.Render()
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

func (l *lokomotive) clusterExists() (bool, error) {
	o := map[string]interface{}{}

	if err := l.Executor.Output("", &o); err != nil {
		return false, fmt.Errorf("failed to check if cluster exists: %v", err)
	}

	return len(o) != 0, nil
}

func verifyCluster(kubeconfigPath string, expectedNodes int) error {
	client, err := k8sutil.NewClientset(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to set up clientset: %v", err)
	}

	cluster, err := NewCluster(client, expectedNodes)
	if err != nil {
		return fmt.Errorf("failed to set up cluster client: %v", err)
	}

	return Verify(cluster)
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

func (l *lokomotive) getControlplaneChart(name string) (*chart.Chart, error) {
	chartPath := "/lokomotive-kubernetes/bootkube/resources/charts"

	helmChart, err := loader.Load(filepath.Join(l.Platform.GetAssetDir(), chartPath, name))
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
	kubeconfigPath := getKubeconfig(l.Platform.GetAssetDir())

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
		l.Logger.Infof("controlplane component '%s' is missing, reinstalling...", component)

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

func (l *lokomotive) deleteHelmRelease(c components.Component, kubeconfig string, deleteNSBool bool) error {
	name := c.Metadata().Name
	if name == "" {
		// This should never fail in real user usage, if this does that means the component was not
		// created with all the needed information.
		panic(fmt.Errorf("component name is empty"))
	}

	ns := c.Metadata().Namespace
	if ns == "" {
		// This should never fail in real user usage, if this does that means the component was not
		// created with all the needed information.
		panic(fmt.Errorf("component %s namespace is empty", name))
	}

	cfg, err := util.HelmActionConfig(ns, kubeconfig)
	if err != nil {
		return fmt.Errorf("failed preparing helm client: %w", err)
	}

	history := action.NewHistory(cfg)
	// Check if the component's release exists. If it does only then proceed to delete.
	//
	// Note: It is assumed that this call will return error only when the release does not exist.
	// The error check is ignored to make `lokoctl component delete ..` idempotent.
	// We rely on the fact that the 'component name' == 'release name'. Since component's name is
	// hardcoded and unlikely to change release name won't change as well. And they will be
	// consistent if installed by lokoctl. So it is highly unlikely that following call will return
	// any other error than "release not found".
	if _, err := history.Run(name); err == nil {
		uninstall := action.NewUninstall(cfg)

		// Ignore the err when we have deleted the release already or it does not exist for some reason.
		if _, err := uninstall.Run(name); err != nil {
			return err
		}
	}

	if deleteNSBool {
		if err := deleteNS(ns, kubeconfig); err != nil {
			return err
		}
	}

	return nil
}