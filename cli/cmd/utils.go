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

	"github.com/hashicorp/hcl/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/config"
	"github.com/kinvolk/lokomotive/pkg/platform"
)

const (
	kubeconfigEnvVariable        = "KUBECONFIG"
	defaultKubeconfigPath        = "~/.kube/config"
	kubeconfigTerraformOutputKey = "kubeconfig"
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
func getConfiguredPlatform(lokoConfig *config.Config) (platform.Platform, hcl.Diagnostics) {
	if lokoConfig.RootConfig.Cluster == nil {
		// No cluster defined by user.
		return nil, nil
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

// kubeconfig discovers the kubeconfig file to be used for cluster/component operations and returns
// its contents.
//
// The following order of precedence is used:
// - User-specified kubeconfig from --kubeconfig-file flag or KUBECONFIG_FILE env var (via Viper).
// - Lokomotive kubeconfig from Terraform state.
// - Global kubeconfig from KUBECONFIG env var.
// - Global kubeconfig from ~/.kube/config.
func kubeconfig(contextLogger *logrus.Entry, lokoConfig *config.Config) ([]byte, error) {
	// User-specified kubeconfig.
	for _, k := range viper.AllKeys() {
		if k != kubeconfigFlag {
			fmt.Printf("Skipping flag %s\n", k)
			continue
		}

		if path := viper.GetString(kubeconfigFlag); path != "" {
			return readKubeconfigFromPath(path)
		}

		fmt.Println("Oh no")
	}

	// Lokomotive kubeconfig from Terraform state.
	if k, err := readKubeconfigFromTerraformState(contextLogger); err == nil {
		return k, nil
	}

	// Global kubeconfig from KUBECONFIG env var.
	if e := os.Getenv(kubeconfigEnvVariable); e != "" {
		return readKubeconfigFromPath(e)
	}

	// Global kubeconfig from default path on disk.
	return readKubeconfigFromPath(defaultKubeconfigPath)
}

func getLokoConfig() (*config.Config, hcl.Diagnostics) {
	return config.LoadConfig(viper.GetString("lokocfg"), viper.GetString("lokocfg-vars"))
}

// readKubeconfigFromTerraformState initializes Terraform and
// reads content of cluster kubeconfig file from the Terraform.
func readKubeconfigFromTerraformState(contextLogger *logrus.Entry) ([]byte, error) {
	contextLogger.Warn("Reading kubeconfig from Terraform state. This could take a while.")

	ex, _, _, _ := initialize(contextLogger) //nolint:dogsled

	kubeconfig := ""

	if err := ex.Output(kubeconfigTerraformOutputKey, &kubeconfig); err != nil {
		return nil, fmt.Errorf("reading kubeconfig from Terraform state: %w", err)
	}

	return []byte(kubeconfig), nil
}

// readKubeconfigFromPath optimistically tries to expand ~ in the given path and then reads the entire
// contents of the file and returns them.
func readKubeconfigFromPath(path string) ([]byte, error) {
	if expandedPath, err := homedir.Expand(path); err == nil {
		path = expandedPath
	}

	// homedir.Expand is too restrictive for the ~ prefix,
	// i.e., it errors on "~somepath" which is a valid path,
	// so just read from the original path.
	return ioutil.ReadFile(path) // #nosec G304
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
