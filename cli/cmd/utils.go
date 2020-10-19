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
	log "github.com/sirupsen/logrus"

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
func getConfiguredPlatform(lokoConfig *config.Config, require bool) (platform.Platform, hcl.Diagnostics) {
	if lokoConfig.RootConfig.Cluster == nil && !require {
		// No cluster defined and no configuration error
		return nil, hcl.Diagnostics{}
	}

	if lokoConfig.RootConfig.Cluster == nil && require {
		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "no platform configured",
			},
		}
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

type kubeconfigGetter struct {
	platformRequired bool
	path             string
}

// getKubeconfig finds the right kubeconfig file to use for an action and returns it's content.
//
// If platform is required and user do not have it configured, an error is returned.
func (kg kubeconfigGetter) getKubeconfig(contextLogger *log.Entry, lokoConfig *config.Config) ([]byte, error) {
	sources, err := kg.getKubeconfigSource(contextLogger, lokoConfig)
	if err != nil {
		return nil, fmt.Errorf("selecting kubeconfig source: %w", err)
	}

	// If no sources has been returned, it means we should read from Terraform state.
	if len(sources) == 0 {
		return readKubeconfigFromTerraformState(contextLogger)
	}

	// Select first non-empty source and read it.
	for _, source := range sources {
		if source != "" {
			return expandAndRead(source)
		}
	}

	// This should never occur, since we always fallback to ~/.kube/config.
	return nil, fmt.Errorf("no kubeconfig source found")
}

// getKubeconfigSource defines how we select which kubeconfig file to use. If source slice is empty, it means
// kubeconfig from Terraform state should be used.
//
// If multiple sources are returned, first non-empty should be used.
//
// The precedence is the following:
//
// - If platform configuration is not required, --kubeconfig-file or KUBECONFIG_FILE environment variable
// 	 always takes precedence.
//
// - Kubeconfig in the assets directory.
//
// - If platform is configured, kubeconfig from the Terraform state.
//
// - kubeconfig from KUBECONFIG environment variable.
//
// - kubeconfig from ~/.kube/config file.
//
func (kg kubeconfigGetter) getKubeconfigSource(contextLogger *log.Entry, lokoConfig *config.Config) ([]string, error) { //nolint:lll
	// Always try reading platform configuration.
	p, diags := getConfiguredPlatform(lokoConfig, kg.platformRequired)
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			contextLogger.Error(diagnostic.Error())
		}

		return nil, fmt.Errorf("loading cluster configuration")
	}

	if kg.path != "" {
		return []string{kg.path}, nil
	}

	// If platform is not configured and not required, fallback to global kubeconfig files.
	if p == nil {
		return []string{
			os.Getenv(kubeconfigEnvVariable),
			defaultKubeconfigPath,
		}, nil
	}

	// Next, try reading kubeconfig file from assets directory.
	kubeconfigPath := assetsKubeconfig(p.Meta().AssetDir)

	kubeconfig, err := expandAndRead(kubeconfigPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading kubeconfig file %q: %w", kubeconfigPath, err)
	}

	if len(kubeconfig) != 0 {
		return []string{kubeconfigPath}, nil
	}

	// If reading from assets gave no result and platform is defined, let's indicate, that kubeconfig
	// should be read from the Terraform state, by returning empty source slice.
	return []string{}, nil
}

func assetsKubeconfig(assetDir string) string {
	return filepath.Join(assetDir, "cluster-assets", "auth", "kubeconfig")
}

// readKubeconfigFromTerraformState initializes Terraform and
// reads content of cluster kubeconfig file from the Terraform.
func readKubeconfigFromTerraformState(contextLogger *log.Entry) ([]byte, error) {
	contextLogger.Warn("Kubeconfig file not found in assets directory, pulling kubeconfig from " +
		"Terraform state, this might be slow. Run 'lokoctl cluster apply' to fix it.")

	// TODO: Add Terraform verbose support back.
	cc := clusterConfig{}

	c, err := cc.initialize(contextLogger)
	if err != nil {
		return nil, fmt.Errorf("initializing: %w", err)
	}

	kubeconfig := ""

	if err := c.terraformExecutor.Output(kubeconfigTerraformOutputKey, &kubeconfig); err != nil {
		return nil, fmt.Errorf("reading kubeconfig file content from Terraform state: %w", err)
	}

	return []byte(kubeconfig), nil
}

// expandAndRead optimistically tries to expand ~ in given path and then reads
// the entire content of the file and returns it to the user.
func expandAndRead(path string) ([]byte, error) {
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
