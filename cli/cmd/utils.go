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

	"github.com/hashicorp/hcl/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/config"
	"github.com/kinvolk/lokomotive/pkg/platform"
)

const (
	kubeconfigEnvVariable = "KUBECONFIG"
	defaultKubeconfigPath = "~/.kube/config"
)

// getConfiguredBackend loads a backend from the given configuration file.
func getConfiguredBackend(lokoConfig *config.Config) (backend.Backend, hcl.Diagnostics) {
	if lokoConfig.RootConfig.Backend == nil {
		// No backend defined and no configuration error
		return nil, hcl.Diagnostics{}
	}

	backend, err := backend.GetBackend(lokoConfig.RootConfig.Backend.Name)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
		}
		return nil, hcl.Diagnostics{diag}
	}

	return backend, backend.LoadConfig(&lokoConfig.RootConfig.Backend.Config, lokoConfig.EvalContext)
}

// getConfiguredPlatform loads a platform from the given configuration file.
func getConfiguredPlatform() (platform.Platform, hcl.Diagnostics) {
	lokoConfig, diags := getLokoConfig()
	if diags.HasErrors() {
		return nil, diags
	}

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
	cfg, diags := getConfiguredPlatform()
	if diags.HasErrors() {
		return "", fmt.Errorf("cannot load config: %s", diags)
	}
	if cfg == nil {
		// No cluster defined and no configuration error
		return "", nil
	}

	return cfg.Meta().AssetDir, nil
}

func getKubeconfig() ([]byte, error) {
	path, err := getKubeconfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed getting kubeconfig path: %w", err)
	}

	if expandedPath, err := homedir.Expand(path); err == nil {
		path = expandedPath
	}

	// homedir.Expand is too restrictive for the ~ prefix,
	// i.e., it errors on "~somepath" which is a valid path,
	// so just read from the original path.
	return ioutil.ReadFile(path) // #nosec G304
}

// getKubeconfig finds the kubeconfig to be used. The precedence is the following:
// - --kubeconfig-file flag OR KUBECONFIG_FILE environment variable (the latter
// is a side-effect of cobra/viper and should NOT be documented because it's
// confusing).
// - Asset directory from cluster configuration.
// - KUBECONFIG environment variable.
// - ~/.kube/config path, which is the default for kubectl.
func getKubeconfigPath() (string, error) {
	assetKubeconfig, err := assetsKubeconfigPath()
	if err != nil {
		return "", fmt.Errorf("reading kubeconfig path from configuration failed: %w", err)
	}

	paths := []string{
		viper.GetString(kubeconfigFlag),
		assetKubeconfig,
		os.Getenv(kubeconfigEnvVariable),
		defaultKubeconfigPath,
	}

	for _, path := range paths {
		if path != "" {
			return path, nil
		}
	}

	return "", nil
}

// assetsKubeconfigPath reads the lokocfg configuration and returns
// the kubeconfig path defined in it.
//
// If no configuration is defined, empty string is returned.
func assetsKubeconfigPath() (string, error) {
	assetDir, err := getAssetDir()
	if err != nil {
		return "", err
	}

	if assetDir != "" {
		return assetsKubeconfig(assetDir), nil
	}

	return "", nil
}

func assetsKubeconfig(assetDir string) string {
	return filepath.Join(assetDir, "cluster-assets", "auth", "kubeconfig")
}

func getLokoConfig() (*config.Config, hcl.Diagnostics) {
	return config.LoadConfig(viper.GetString("lokocfg"), viper.GetString("lokocfg-vars"))
}

// askForConfirmation asks the user to confirm an action.
// It prints the message and then asks the user to type "yes" or "no".
// If the user types "yes" the function returns true, otherwise it returns
// false.
func askForConfirmation(message string) bool {
	var input string
	fmt.Printf("%s [type \"yes\" to continue]: ", message)
	fmt.Scanln(&input)
	return input == "yes"
}
