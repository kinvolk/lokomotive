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
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/config"
)

type ComponentDeleteOptions struct {
	Confirm         bool
	DeleteNamespace bool
	KubeconfigPath  string
	ConfigPath      string
	ValuesPath      string
}

// ComponentApply implements 'lokoctl component delete' separated from CLI
// dependencies.
func ComponentDelete(contextLogger *log.Entry, componentsList []string, options ComponentDeleteOptions) error {
	lokoConfig, diags := config.LoadConfig(options.ConfigPath, options.ValuesPath)
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

	if !options.Confirm && !askForConfirmation(confirmationMessage) {
		contextLogger.Info("Components deletion cancelled.")

		return nil
	}

	kg := kubeconfigGetter{
		platformRequired: false,
		path:             options.KubeconfigPath,
	}

	kubeconfig, err := kg.getKubeconfig(contextLogger, lokoConfig)
	if err != nil {
		contextLogger.Debugf("Error in finding kubeconfig file: %s", err)

		return fmt.Errorf("suitable kubeconfig file not found. Did you run 'lokoctl cluster apply' ?")
	}

	if err := deleteComponents(kubeconfig, componentObjects, options.DeleteNamespace); err != nil {
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
