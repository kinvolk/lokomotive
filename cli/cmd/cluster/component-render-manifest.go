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

package cluster

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/config"
)

// ComponentRenderManifestOptions controls ComponentRenderManifest() behavior.
type ComponentRenderManifestOptions struct {
	ConfigPath string
	ValuesPath string
}

// ComponentRenderManifest prints selected components manifests.
//
//nolint:lll
func ComponentRenderManifest(contextLogger *log.Entry, componentsList []string, options ComponentRenderManifestOptions) error {
	lokoConfig, diags := config.LoadConfig(options.ConfigPath, options.ValuesPath)
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			contextLogger.Error(diagnostic.Error())
		}

		return diags
	}

	componentsToRender := selectComponentNames(componentsList, *lokoConfig.RootConfig)

	if err := renderComponentManifests(lokoConfig, componentsToRender); err != nil {
		return fmt.Errorf("rendering component manifests: %w", err)
	}

	return nil
}

func renderComponentManifests(lokoConfig *config.Config, componentNames []string) error {
	for _, componentName := range componentNames {
		contextLogger := log.WithFields(log.Fields{
			"component": componentName,
		})

		component, err := components.Get(componentName)
		if err != nil {
			return fmt.Errorf("getting component %q: %w", componentName, err)
		}

		componentConfigBody := lokoConfig.LoadComponentConfigBody(componentName)

		if diags := component.LoadConfig(componentConfigBody, lokoConfig.EvalContext); diags.HasErrors() {
			for _, diagnostic := range diags {
				contextLogger.Error(diagnostic.Error())
			}

			return diags
		}

		manifests, err := component.RenderManifests()
		if err != nil {
			return fmt.Errorf("rendering manifest of component %q: %w", componentName, err)
		}

		fmt.Printf("# manifests for component %s\n", componentName)

		for filename, manifest := range manifests {
			fmt.Printf("\n---\n# %s\n%s", filename, manifest)
		}
	}

	return nil
}
