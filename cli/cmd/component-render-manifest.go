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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/config"
)

var componentRenderCmd = &cobra.Command{
	Use:   "render-manifest",
	Short: "Print the manifests for a component",
	Run:   runComponentRender,
}

func init() {
	componentCmd.AddCommand(componentRenderCmd)
}

func runComponentRender(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component render-manifest",
		"args":    args,
	})

	lokoConfig, diags := getLokoConfig()
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			contextLogger.Error(diagnostic.Error())
		}
		contextLogger.Fatal("Errors found while loading configuration")
	}

	componentsToRender := selectComponentNames(args, *lokoConfig.RootConfig)

	if err := renderComponentManifests(lokoConfig, componentsToRender...); err != nil {
		contextLogger.Fatal(err)
	}
}

func renderComponentManifests(lokoConfig *config.Config, componentNames ...string) error {
	for _, componentName := range componentNames {
		contextLogger := log.WithFields(log.Fields{
			"component": componentName,
		})

		component, err := components.Get(componentName)
		if err != nil {
			return err
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
			return err
		}

		fmt.Printf("# manifests for component %s\n", componentName)
		for filename, manifest := range manifests {
			fmt.Printf("\n---\n# %s\n%s", filename, manifest)
		}
	}
	return nil
}
