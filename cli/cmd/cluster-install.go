package cmd

import (
	"fmt"
	"path"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kinvolk/lokoctl/pkg/config"
	"github.com/kinvolk/lokoctl/pkg/install"
	"github.com/kinvolk/lokoctl/pkg/k8sutil"
	"github.com/kinvolk/lokoctl/pkg/lokomotive"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Install Lokomotive cluster and components",
}

var clusterInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Lokomotive cluster with components",
	Run:   runClusterInstall,
}

func init() {
	rootCmd.AddCommand(clusterCmd)
	clusterCmd.AddCommand(clusterInstallCmd)
}

func runClusterInstall(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl cluster install",
		"args":    args,
	})

	lokoConfig, diags := config.LoadConfig(viper.GetString("lokocfg"), viper.GetString("lokocfg-vars"))
	if len(diags) > 0 {
		ctxLogger.Fatal(diags)
	}

	p, diags := getConfiguredPlatform(lokoConfig)
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			ctxLogger.Error(diagnostic.Summary)
		}
		ctxLogger.Fatal("Errors found while loading cluster configuration")
	}
	if p == nil {
		ctxLogger.Fatal("No cluster configured")
	}

	if err := p.Install(); err != nil {
		ctxLogger.Fatalf("error installing cluster: %v", err)
	}

	assetDir := p.GetAssetDir()

	fmt.Printf("\nYour configurations are stored in %s\n", assetDir)

	kubeconfigPath := path.Join(assetDir, "auth", "kubeconfig")
	if err := verifyInstall(kubeconfigPath); err != nil {
		ctxLogger.Fatalf("Verify cluster installation: %v", err)
	}

	var componentsToInstall []string
	for _, component := range lokoConfig.RootConfig.Components {
		componentsToInstall = append(componentsToInstall, component.Name)
	}

	if len(componentsToInstall) > 0 {
		installComponents(lokoConfig, kubeconfigPath, componentsToInstall...)
	}
}

func verifyInstall(kubeConfigPath string) error {
	client, err := k8sutil.NewClientset(kubeConfigPath)
	if err != nil {
		return errors.Wrapf(err, "failed to set up clientset")
	}

	cluster, err := lokomotive.NewCluster(client)
	if err != nil {
		return errors.Wrapf(err, "failed to set up cluster client")
	}

	return install.Verify(cluster)
}
