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
	"github.com/spf13/viper"

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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return components.ListNames(), cobra.ShellCompDirectiveNoFileComp
	},
}

var debug bool

//nolint:gochecknoinits
func init() {
	componentCmd.AddCommand(componentApplyCmd)
	pf := componentApplyCmd.PersistentFlags()
	addKubeconfigFileFlag(pf)
	pf.BoolVarP(&debug, "debug", "", false, "Print debug messages")
}

func runApply(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component apply",
		"args":    args,
	})

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	options := componentApplyOptions{
		kubeconfigPath: kubeconfigFlag,
		configPath:     viper.GetString("lokocfg"),
		valuesPath:     viper.GetString("lokocfg-vars"),
	}

	if err := componentApply(contextLogger, args, options); err != nil {
		contextLogger.Fatalf("Applying components failed: %v", err)
	}
}

type componentApplyOptions struct {
	kubeconfigPath string
	configPath     string
	valuesPath     string
}

// componentApply implements 'lokoctl component apply' separated from CLI
// dependencies.
func componentApply(contextLogger *log.Entry, componentsList []string, options componentApplyOptions) error {
	lokoConfig, diags := config.LoadConfig(options.configPath, options.valuesPath)
	if diags.HasErrors() {
		return diags
	}

	componentObjects, err := componentNamesToObjects(selectComponentNames(componentsList, *lokoConfig.RootConfig))
	if err != nil {
		return fmt.Errorf("getting component objects: %w", err)
	}

	kg := kubeconfigGetter{
		platformRequired: false,
		path:             options.kubeconfigPath,
	}

	kubeconfig, err := kg.getKubeconfig(contextLogger, lokoConfig)
	if err != nil {
		contextLogger.Debugf("Error in finding kubeconfig file: %s", err)

		return fmt.Errorf("suitable kubeconfig file not found. Did you run 'lokoctl cluster apply' ?")
	}

	if err := applyComponents(lokoConfig, kubeconfig, componentObjects); err != nil {
		return fmt.Errorf("applying components: %w", err)
	}

	return nil
}

// applyComponents reads the configuration of given components and applies them to the cluster pointer
// by given kubeconfig file content.
func applyComponents(lokoConfig *config.Config, kubeconfig []byte, componentObjects []components.Component) error {
	for _, component := range componentObjects {
		componentName := component.Metadata().Name
		fmt.Printf("Applying component '%s'...\n", componentName)

		componentConfigBody := lokoConfig.LoadComponentConfigBody(componentName)

		if diags := component.LoadConfig(componentConfigBody, lokoConfig.EvalContext); diags.HasErrors() {
			fmt.Printf("%v\n", diags)
			return diags
		}

		if err := util.InstallComponent(component, kubeconfig); err != nil {
			return fmt.Errorf("installing component %q: %w", componentName, err)
		}

		fmt.Printf("Successfully applied component '%s' configuration!\n", componentName)
	}

	return nil
}
