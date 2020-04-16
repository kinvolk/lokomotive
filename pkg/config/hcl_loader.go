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

// Package config loads the user provided HCL configuration and
// parses it into LokomotiveConfig for further processing
package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	backendpkg "github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/backend/local"
	"github.com/kinvolk/lokomotive/pkg/cluster/config"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/util"
)

// HCLLoader represents loading the HCL configuration provided by the user
type HCLLoader struct {
	ConfigPath    string
	VariablesPath string
}

// Load loads the HCL files provided by the user and parses them into an
// instance of LokomoticeConfig
func (c *HCLLoader) Load() (*config.LokomotiveConfig, hcl.Diagnostics) {
	configFiles, diags := loadHCLFiles(c.ConfigPath, "lokocfg")
	if diags.HasErrors() {
		return nil, diags
	}

	exists, err := util.PathExists(c.VariablesPath)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("error checking variables file path : %v", err),
		}

		return nil, hcl.Diagnostics{diag}
	}

	variablesFile := map[string][]byte{}

	if exists {
		variablesFile, diags = loadHCLFiles(c.VariablesPath, "vars")
		if diags.HasErrors() {
			return nil, diags
		}
	}

	hclConfig, diags := ParseHCLFiles(configFiles, variablesFile)
	if diags.HasErrors() {
		return nil, diags
	}

	if diags := ValidateHCLConfig(hclConfig); diags.HasErrors() {
		return nil, diags
	}

	return ParseToLokomotiveConfig(hclConfig)
}

// ValidateHCLConfig validates the hcl config blocks.
// This checks that other blocks are not present if platform
// block is not present.
func ValidateHCLConfig(hclConfig *HCLConfig) hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if hclConfig.Config.Cluster != nil && hclConfig.Config.Platform == nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "found 'cluster' block but no 'platform' block",
		})
	}

	if hclConfig.Config.Metadata != nil && hclConfig.Config.Platform == nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "found 'metadata' block but no 'platform' block",
		})
	}

	if hclConfig.Config.Controller != nil && hclConfig.
		Config.Platform == nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "found 'controller' block but no 'platform' block",
		})
	}

	if hclConfig.Config.Flatcar != nil && hclConfig.Config.Platform == nil {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "found 'flatcar' block but no 'platform' block",
		})
	}

	return diagnostics
}

// ParseToLokomotiveConfig parses the Config instance to LokomotiveConfig
//nolint:funlen
func ParseToLokomotiveConfig(hclConfig *HCLConfig) (*config.LokomotiveConfig, hcl.Diagnostics) {
	var diagnostics hcl.Diagnostics
	// load platform configuration
	platform, diags := loadPlatformConfiguration(hclConfig)
	diagnostics = append(diagnostics, diags...)
	//	if diags.HasErrors() {
	//		return nil, diags
	//	}
	// load cluster configuration
	cluster, diags := loadClusterConfiguration(hclConfig)
	diagnostics = append(diagnostics, diags...)
	//if diags.HasErrors() {
	//	return nil, diags
	//}
	// load network configuration
	network, diags := loadNetworkConfiguration(hclConfig)
	diagnostics = append(diagnostics, diags...)
	//if diags.HasErrors() {
	//	return nil, diags
	//}
	// load metadata configuration
	metadata, diags := loadMetadataConfiguration(hclConfig)
	diagnostics = append(diagnostics, diags...)
	//if diags.HasErrors() {
	//	return nil, diags
	//}
	// load controller configuration
	controller, diags := loadControllerConfiguration(hclConfig)
	diagnostics = append(diagnostics, diags...)
	//if diags.HasErrors() {
	//	return nil, diags
	//}
	// load flatcar configuration
	flatcar, diags := loadFlatcarConfiguration(hclConfig)
	diagnostics = append(diagnostics, diags...)
	//if diags.HasErrors() {
	//	return nil, diags
	//}
	// load backend configuration
	backend, diags := loadBackendConfiguration(hclConfig)
	diagnostics = append(diagnostics, diags...)
	//	if diags.HasErrors() {
	//		return nil, diags
	//	}
	// load components configuration
	configuredComponents, diags := LoadConfiguredComponents(hclConfig)
	diagnostics = append(diagnostics, diags...)
	//	if diags.HasErrors() {
	//		return nil, diags
	//	}

	if diagnostics.HasErrors() {
		return nil, diagnostics
	}

	lokomotiveconfig := &config.LokomotiveConfig{
		Cluster:    cluster,
		Flatcar:    flatcar,
		Metadata:   metadata,
		Network:    network,
		Controller: controller,
		Platform:   platform,
		Backend:    backend,
		Components: configuredComponents,
	}

	return lokomotiveconfig, hcl.Diagnostics{}
}

func loadPlatformConfiguration(hclConfig *HCLConfig) (config.Platform, hcl.Diagnostics) {
	// if platform block is not present, return no error
	if hclConfig.Config.Platform == nil {
		return nil, hcl.Diagnostics{}
	}

	platformName := hclConfig.Config.Platform.Name
	// return error if platform not registered
	platform, err := config.GetPlatform(platformName)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("unsupported platform : %v", err),
		}

		return nil, hcl.Diagnostics{diag}
	}

	diags := gohcl.DecodeBody(hclConfig.Config.Platform.Config, hclConfig.EvalContext, platform)
	if diags.HasErrors() {
		return nil, diags
	}

	return platform, hcl.Diagnostics{}
}

func loadClusterConfiguration(hclConfig *HCLConfig) (*config.ClusterConfig, hcl.Diagnostics) {
	cluster := config.DefaultClusterConfig()

	if hclConfig.Config.Cluster != nil {
		diags := gohcl.DecodeBody(hclConfig.Config.Cluster.Config, hclConfig.EvalContext, cluster)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	return cluster, hcl.Diagnostics{}
}

func loadNetworkConfiguration(hclConfig *HCLConfig) (*config.NetworkConfig, hcl.Diagnostics) {
	network := config.DefaultNetworkConfig()

	if hclConfig.Config.Network != nil {
		diags := gohcl.DecodeBody(hclConfig.Config.Network.Config, hclConfig.EvalContext, network)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	return network, hcl.Diagnostics{}
}

// loadMetadataConfiguration loads metadata configuration
func loadMetadataConfiguration(hclConfig *HCLConfig) (*config.Metadata, hcl.Diagnostics) {
	metadata := &config.Metadata{}

	if hclConfig.Config.Metadata != nil {
		diags := gohcl.DecodeBody(hclConfig.Config.Metadata.Config, hclConfig.EvalContext, metadata)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	return metadata, hcl.Diagnostics{}
}

// loadControllerConfiguration loads controller configuration
func loadControllerConfiguration(hclConfig *HCLConfig) (*config.ControllerConfig, hcl.Diagnostics) {
	controller := config.DefaultControllerConfig()

	if hclConfig.Config.Controller != nil {
		diags := gohcl.DecodeBody(hclConfig.Config.Controller.Config, hclConfig.EvalContext, controller)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	return controller, hcl.Diagnostics{}
}

// loadFlatcarConfiguration loads flatcar configuration
func loadFlatcarConfiguration(hclConfig *HCLConfig) (*config.FlatcarConfig, hcl.Diagnostics) {
	flatcar := config.DefaultFlatcarConfig()

	if hclConfig.Config.Flatcar != nil {
		diags := gohcl.DecodeBody(hclConfig.Config.Flatcar.Config, hclConfig.EvalContext, flatcar)
		if diags.HasErrors() {
			return nil, diags
		}
	}

	return flatcar, hcl.Diagnostics{}
}

// loadBackendConfuguration loads backend configuration
func loadBackendConfiguration(hclConfig *HCLConfig) (backendpkg.Backend, hcl.Diagnostics) {
	var backend backendpkg.Backend

	var err error

	if hclConfig.Config.Backend != nil {
		backend, err = backendpkg.GetBackend(hclConfig.Config.Backend.Name)
		if err != nil {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("unsupported component: %v", err),
			}

			return nil, hcl.Diagnostics{diag}
		}

		diags := gohcl.DecodeBody(hclConfig.Config.Backend.Config, hclConfig.EvalContext, backend)
		if diags.HasErrors() {
			return nil, diags
		}
	}
	// Use a local backend if no backend is configured
	if backend == nil {
		backend = local.NewLocalBackend()
	}

	return backend, hcl.Diagnostics{}
}

// LoadConfiguredComponents loads components configuration
func LoadConfiguredComponents(hclConfig *HCLConfig) (map[string]components.Component, hcl.Diagnostics) {
	configuredComponents := make(map[string]components.Component)

	for _, c := range hclConfig.Config.Components {
		component, err := components.Get(c.Name)
		if err != nil {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("unsupported component: %v", err),
			}

			return nil, hcl.Diagnostics{diag}
		}

		componentConfigBody := hclConfig.LoadComponentConfigBody(c.Name)
		if diags := component.LoadConfig(componentConfigBody, hclConfig.EvalContext); diags.HasErrors() {
			return nil, diags
		}

		configuredComponents[c.Name] = component
	}

	return configuredComponents, hcl.Diagnostics{}
}
