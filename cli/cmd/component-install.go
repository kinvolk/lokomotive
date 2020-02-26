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
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/config"
)

var componentInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a component",
	Run:   runInstall,
}

func init() {
	componentCmd.AddCommand(componentInstallCmd)
}

func runInstall(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component install",
		"args":    args,
	})

	lokoConfig, diags := getLokoConfig()
	if len(diags) > 0 {
		contextLogger.Fatal(diags)
	}

	var componentsToInstall []string
	if len(args) > 0 {
		componentsToInstall = append(componentsToInstall, args...)
	} else {
		for _, component := range lokoConfig.RootConfig.Components {
			componentsToInstall = append(componentsToInstall, component.Name)
		}
	}

	kubeconfig, err := getKubeconfig()
	if err != nil {
		contextLogger.Fatalf("Error in finding kubeconfig file: %s", err)
	}
	if err := installComponents(lokoConfig, kubeconfig, componentsToInstall...); err != nil {
		contextLogger.Fatal(err)
	}
}

func installComponents(lokoConfig *config.Config, kubeconfig string, componentNames ...string) error {
	for _, componentName := range componentNames {
		fmt.Printf("Installing component '%s'...\n", componentName)

		component, err := components.Get(componentName)
		if err != nil {
			return err
		}

		componentConfigBody := lokoConfig.LoadComponentConfigBody(componentName)

		if diags := component.LoadConfig(componentConfigBody, lokoConfig.EvalContext); len(diags) > 0 {
			fmt.Printf("%v\n", diags)
			return diags
		}

		if err := util.InstallComponent(componentName, component, kubeconfig); err != nil {
			return err
		}

		fmt.Printf("Successfully installed component '%s'!\n", componentName)
	}
	return nil
}
