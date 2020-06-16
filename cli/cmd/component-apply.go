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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/config"
)

var componentApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Deploy or update a component",
	Long: `Deploy or update a component.
Deploys a component if not yet present, otherwise updates it.
When run with no arguments, all components listed in the configuration are applied.`,
	Run: runApply,
}

func init() {
	componentCmd.AddCommand(componentApplyCmd)
}

func runApply(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component apply",
		"args":    args,
	})

	lokoConfig, diags := getLokoConfig()
	if len(diags) > 0 {
		contextLogger.Fatal(diags)
	}

	var componentsToApply []string
	if len(args) > 0 {
		componentsToApply = append(componentsToApply, args...)
	} else {
		for _, component := range lokoConfig.RootConfig.Components {
			componentsToApply = append(componentsToApply, component.Name)
		}
	}

	kubeconfig, err := getKubeconfig()
	if err != nil {
		contextLogger.Fatalf("Error in finding kubeconfig file: %s", err)
	}

	kubeconfigContent, err := ioutil.ReadFile(kubeconfig) // #nosec G304
	if err != nil {
		contextLogger.Fatalf("Failed to read kubeconfig file: %q: %v", kubeconfig, err)
	}

	if err := applyComponents(lokoConfig, kubeconfigContent, componentsToApply...); err != nil {
		contextLogger.Fatal(err)
	}
}

func applyComponents(lokoConfig *config.Config, kubeconfig []byte, componentNames ...string) error {
	for _, componentName := range componentNames {
		fmt.Printf("Applying component '%s'...\n", componentName)

		component, err := components.Get(componentName)
		if err != nil {
			return err
		}

		componentConfigBody := lokoConfig.LoadComponentConfigBody(componentName)

		if diags := component.LoadConfig(componentConfigBody, lokoConfig.EvalContext); len(diags) > 0 {
			fmt.Printf("%v\n", diags)
			return diags
		}

		if err := util.InstallComponent(component, kubeconfig); err != nil {
			return err
		}

		fmt.Printf("Successfully applied component '%s' configuration!\n", componentName)
	}
	return nil
}
