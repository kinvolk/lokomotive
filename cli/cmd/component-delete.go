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

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

var componentDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an installed component",
	Long: `Delete a component.
When run with no arguments, all components listed in the configuration are deleted.`,
	Run: runDelete,
}

var deleteNamespace bool

// nolint:gochecknoinits
func init() {
	componentCmd.AddCommand(componentDeleteCmd)
	pf := componentDeleteCmd.PersistentFlags()
	pf.BoolVarP(&deleteNamespace, "delete-namespace", "", false, "Delete namespace with component")
	pf.BoolVarP(&confirm, "confirm", "", false, "Delete component without asking for confirmation")
}

func runDelete(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component delete",
		"args":    args,
	})

	lokoCfg, diags := getLokoConfig()
	if len(diags) > 0 {
		contextLogger.Fatal(diags)
	}

	componentsToDelete := make([]string, len(args))
	copy(componentsToDelete, args)

	if len(args) == 0 {
		componentsToDelete = make([]string, len(lokoCfg.RootConfig.Components))

		for i, component := range lokoCfg.RootConfig.Components {
			componentsToDelete[i] = component.Name
		}
	}

	componentsObjects := make([]components.Component, len(componentsToDelete))

	for i, componentName := range componentsToDelete {
		compObj, err := components.Get(componentName)
		if err != nil {
			contextLogger.Fatal(err)
		}

		componentsObjects[i] = compObj
	}

	if !confirm && !askForConfirmation(
		fmt.Sprintf(
			"The following components will be deleted:\n\t%s\n\nAre you sure you want to proceed?",
			strings.Join(componentsToDelete, "\n\t"),
		),
	) {
		contextLogger.Info("Components deletion cancelled.")
		return
	}

	kubeconfig, err := getKubeconfig()
	if err != nil {
		contextLogger.Fatalf("Error in finding kubeconfig file: %s", err)
	}

	if err := deleteComponents(kubeconfig, componentsObjects...); err != nil {
		contextLogger.Fatal(err)
	}
}

func deleteComponents(kubeconfig string, componentObjects ...components.Component) error {
	for _, compObj := range componentObjects {
		fmt.Printf("Deleting component '%s'...\n", compObj.Metadata().Name)

		if err := util.UninstallComponent(compObj, kubeconfig, deleteNamespace); err != nil {
			return err
		}

		fmt.Printf("Successfully deleted component %q!\n", compObj.Metadata().Name)
	}

	// Add a line to distinguish between info logs and errors, if any.
	fmt.Println()

	return nil
}
