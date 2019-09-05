package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kinvolk/lokoctl/pkg/config"
	"github.com/kinvolk/lokoctl/pkg/platform"
	"github.com/kinvolk/lokoctl/pkg/util/tools"
)

// getConfiguredPlatform loads a platform from the given configuration file.
func getConfiguredPlatform(lokoConfig *config.Config) (platform.Platform, hcl.Diagnostics) {
	if lokoConfig.RootConfig.Cluster == nil {
		// No cluster defined and no configuration error
		return nil, hcl.Diagnostics{}
	}

	platform, err := platform.GetPlatform(lokoConfig.RootConfig.Cluster.Name)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
		}
		return nil, hcl.Diagnostics{diag}
	}

	return platform, platform.LoadConfig(&lokoConfig.RootConfig.Cluster.Config, lokoConfig.EvalContext)
}

// getAssetDir extracts the asset path from the cluster configuration.
// It is empty if there is no cluster defined. An error is returned if the
// cluster configuration has problems.
func getAssetDir() (string, error) {
	lokoConfig, diags := config.LoadConfig(viper.GetString("lokocfg"), viper.GetString("lokocfg-vars"))
	if diags.HasErrors() {
		return "", fmt.Errorf("cannot load config: %s", diags)
	}

	cfg, diags := getConfiguredPlatform(lokoConfig)
	if diags.HasErrors() {
		return "", fmt.Errorf("cannot load config: %s", diags)
	}
	if cfg == nil {
		// No cluster defined and no configuration error
		return "", nil
	}

	return cfg.GetAssetDir(), nil
}

// getKubeconfig finds the kubeconfig to be used. Precedence takes a specified
// flag or environment variable. Then the asset directory of the cluster is searched
// and finally the global default value is used. This cannot be done in Viper
// because we need the other values from Viper to find the asset directory.
func getKubeconfig() (string, error) {
	expand := func(path string) string {
		expandedPath, err := homedir.Expand(path)
		if err != nil {
			// homedir.Expand is too restrictive for the ~ prefix,
			// i.e., it errors on "~somepath" which is a valid path,
			// so just return the original path.
			return path
		}
		return expandedPath
	}

	kubeconfig := viper.GetString("kubeconfig")
	if kubeconfig != "" {
		return expand(kubeconfig), nil
	}

	assetDir, err := getAssetDir()
	if err != nil {
		return "", err
	}
	if assetDir == "" {
		return expand("~/.kube/config"), nil
	}

	return expand(filepath.Join(assetDir, "cluster-assets", "auth", "kubeconfig")), nil
}

// doesKubeconfigExist checks if the kubeconfig provided by user exists
func doesKubeconfigExist(*cobra.Command, []string) error {
	var err error
	kubeconfig, err := getKubeconfig()
	if err != nil {
		return err
	}
	if _, err = os.Stat(kubeconfig); os.IsNotExist(err) {
		return fmt.Errorf("Kubeconfig %q not found", kubeconfig)
	}
	return err
}

func clusterInstallChecks(*cobra.Command, []string) error {
	return tools.InstallerBinaries()
}
