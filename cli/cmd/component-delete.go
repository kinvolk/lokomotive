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
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/config"
)

var componentDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an installed component",
	Long: `Delete a component.
When run with no arguments, all components listed in the configuration are deleted.`,
	Run: runDelete,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return components.ListNames(), cobra.ShellCompDirectiveNoFileComp
	},
}

var deleteNamespace bool

//nolint:gochecknoinits
func init() {
	componentCmd.AddCommand(componentDeleteCmd)
	pf := componentDeleteCmd.PersistentFlags()
	addKubeconfigFileFlag(pf)
	pf.BoolVarP(&deleteNamespace, "delete-namespace", "", false, "Delete namespace with component")
	pf.BoolVarP(&confirm, "confirm", "", false, "Delete component without asking for confirmation")
	pf.BoolVarP(&debug, "debug", "", false, "Print debug messages")
}

func runDelete(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component delete",
		"args":    args,
	})

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	options := componentDeleteOptions{
		confirm:         confirm,
		deleteNamespace: deleteNamespace,
		kubeconfigPath:  kubeconfigFlag,
		configPath:      viper.GetString("lokocfg"),
		valuesPath:      viper.GetString("lokocfg-vars"),
	}

	if err := componentDelete(contextLogger, args, options); err != nil {
		contextLogger.Fatalf("Deleting components failed: %v", err)
	}
}

type componentDeleteOptions struct {
	confirm         bool
	deleteNamespace bool
	kubeconfigPath  string
	configPath      string
	valuesPath      string
}

// componentDelete implements 'lokoctl component delete' separated from CLI
// dependencies.
func componentDelete(contextLogger *log.Entry, componentsList []string, options componentDeleteOptions) error {
	lokoConfig, diags := config.LoadConfig(options.configPath, options.valuesPath)
	if diags.HasErrors() {
		return diags
	}

	componentsToDelete := selectComponentNames(componentsList, *lokoConfig.RootConfig)

	componentObjects, err := componentNamesToObjects(componentsToDelete)
	if err != nil {
		return fmt.Errorf("getting component objects: %v", err)
	}

	confirmationMessage := fmt.Sprintf(
		"The following components will be deleted:\n\t%s\n\nAre you sure you want to proceed?",
		strings.Join(componentsToDelete, "\n\t"),
	)

	if !options.confirm && !askForConfirmation(confirmationMessage) {
		contextLogger.Info("Components deletion cancelled.")

		return nil
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

	if err := deleteComponents(kubeconfig, componentObjects, options.deleteNamespace); err != nil {
		return fmt.Errorf("deleting components: %w", err)
	}

	return nil
}

// selectComponentNames returns list of components to operate on. If explicit list is empty,
// it returns components defined in the configuration.
func selectComponentNames(list []string, lokomotiveConfig config.RootConfig) []string {
	if len(list) != 0 {
		return list
	}

	for _, component := range lokomotiveConfig.Components {
		list = append(list, component.Name)
	}

	return list
}

// componentNamesToObjects converts list of component names to list of component objects.
// If some component does not exist, error is returned.
func componentNamesToObjects(componentNames []string) ([]components.Component, error) {
	c := []components.Component{}

	for _, componentName := range componentNames {
		component, err := components.Get(componentName)
		if err != nil {
			return nil, fmt.Errorf("getting component %q: %w", componentName, err)
		}

		c = append(c, component)
	}

	return c, nil
}

func deleteComponents(kubeconfig []byte, componentObjects []components.Component, deleteNamespace bool) error {
	for _, compObj := range componentObjects {
		fmt.Printf("Deleting component '%s'...\n", compObj.Metadata().Name)

		if err := util.UninstallComponent(compObj, kubeconfig, deleteNamespace); err != nil {
			return fmt.Errorf("uninstalling component %q: %w", compObj.Metadata().Name, err)
		}

		fmt.Printf("Successfully deleted component %q!\n", compObj.Metadata().Name)
	}

	// Add a line to distinguish between info logs and errors, if any.
	fmt.Println()

	return nil
}
