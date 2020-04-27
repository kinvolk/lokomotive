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

package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"

	backendpkg "github.com/kinvolk/lokomotive/pkg/backend"
	"github.com/kinvolk/lokomotive/pkg/backend/local"
	"github.com/kinvolk/lokomotive/pkg/components"
	lokomotiveconfig "github.com/kinvolk/lokomotive/pkg/lokomotive/config"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/util"
)

// HCLLoader represents loading the HCL configuration provided by the user
type HCLLoader struct {
	ConfigPath    string
	VariablesPath string
}

// Load loads the HCL files provided by the user and parses them into an
// instance of LokomoticeConfig
func (c *HCLLoader) Load() (*lokomotiveconfig.LokomotiveConfig, hcl.Diagnostics) {
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

	return ParseToLokomotiveConfig(hclConfig)
}

// ParseToLokomotiveConfig parses the Config instance to LokomotiveConfig
//nolint:funlen
func ParseToLokomotiveConfig(hclConfig *HCLConfig) (*lokomotiveconfig.LokomotiveConfig, hcl.Diagnostics) {
	var diagnostics hcl.Diagnostics

	// load platform configuration
	platform, diags := loadPlatformConfiguration(hclConfig)
	diagnostics = append(diagnostics, diags...)

	// load backend configuration
	backend, diags := loadBackendConfiguration(hclConfig)
	diagnostics = append(diagnostics, diags...)

	// load components configuration
	configuredComponents, diags := LoadConfiguredComponents(hclConfig)
	diagnostics = append(diagnostics, diags...)

	if diagnostics.HasErrors() {
		return nil, diagnostics
	}

	lokomotiveconfig := &lokomotiveconfig.LokomotiveConfig{
		Platform:   platform,
		Backend:    backend,
		Components: configuredComponents,
	}

	return lokomotiveconfig, hcl.Diagnostics{}
}

func loadPlatformConfiguration(hclConfig *HCLConfig) (platform.Platform, hcl.Diagnostics) {
	// if platform block is not present, return no error
	if hclConfig.ClusterConfig.Cluster == nil {
		return nil, hcl.Diagnostics{}
	}

	platformName := hclConfig.ClusterConfig.Cluster.Name
	// return error if platform not registered
	platform, err := platform.GetPlatform(platformName)
	if err != nil {
		diag := &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("unsupported platform : %v", err),
		}

		return nil, hcl.Diagnostics{diag}
	}

	diags := gohcl.DecodeBody(hclConfig.ClusterConfig.Cluster.Config, hclConfig.EvalContext, platform)
	if diags.HasErrors() {
		return nil, diags
	}

	return platform, hcl.Diagnostics{}
}

// loadBackendConfuguration loads backend configuration
func loadBackendConfiguration(hclConfig *HCLConfig) (backendpkg.Backend, hcl.Diagnostics) {
	var backend backendpkg.Backend

	var err error

	if hclConfig.ClusterConfig.Backend != nil {
		backend, err = backendpkg.GetBackend(hclConfig.ClusterConfig.Backend.Name)
		if err != nil {
			diag := &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("unsupported component: %v", err),
			}

			return nil, hcl.Diagnostics{diag}
		}

		diags := gohcl.DecodeBody(hclConfig.ClusterConfig.Backend.Config, hclConfig.EvalContext, backend)
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

	for _, c := range hclConfig.ClusterConfig.Components {
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
