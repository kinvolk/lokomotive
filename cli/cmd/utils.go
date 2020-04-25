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

	"github.com/hashicorp/hcl/v2"
	"github.com/spf13/viper"

	"github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/config"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/util"
)

// getConfiguredBackend loads a backend from the given configuration file.
func getConfiguredBackend(lokoConfig *config.HCLConfig) (backend.Backend, hcl.Diagnostics) {
	if lokoConfig.ClusterConfig.Backend == nil {
		// No backend defined and no configuration error
		return nil, hcl.Diagnostics{}
	}

	backend, err := backend.GetBackend(lokoConfig.ClusterConfig.Backend.Name)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
		}
		return nil, hcl.Diagnostics{diag}
	}

	return backend, backend.LoadConfig(&lokoConfig.ClusterConfig.Backend.Config, lokoConfig.EvalContext)
}

// getConfiguredPlatform loads a platform from the given configuration file.
func getConfiguredPlatform() (platform.Platform, hcl.Diagnostics) {
	lokoConfig, diags := getLokoConfig()
	if diags.HasErrors() {
		return nil, diags
	}

	if lokoConfig.ClusterConfig.Cluster == nil {
		// No cluster defined and no configuration error
		return nil, hcl.Diagnostics{}
	}

	platform, err := platform.GetPlatform(lokoConfig.ClusterConfig.Cluster.Name)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  err.Error(),
		}
		return nil, hcl.Diagnostics{diag}
	}

	return platform, platform.LoadConfig(&lokoConfig.ClusterConfig.Cluster.Config, lokoConfig.EvalContext)
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

	return cfg.GetAssetDir(), nil
}

func getLokoConfig() (*config.HCLConfig, hcl.Diagnostics) {
	lokocfgFiles, diags := config.LoadHCLFiles(viper.GetString("lokocfg"), "lokocfg")
	if diags.HasErrors() {
		return nil, diags
	}

	exists, err := util.PathExists(viper.GetString("lokocfg-vars"))
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("error checking variables file path : %v", err),
		}

		return nil, hcl.Diagnostics{diag}
	}

	varFiles := map[string][]byte{}

	if exists {
		varFiles, diags = config.LoadHCLFiles(viper.GetString("lokocfg-vars"), "vars")
		if diags.HasErrors() {
			return nil, diags
		}
	}

	return config.ParseHCLFiles(lokocfgFiles, varFiles)
}
