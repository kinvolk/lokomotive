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
	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/config"
)

type ComponentApplyOptions struct {
	KubeconfigPath string
	ConfigPath     string
	ValuesPath     string
}

// ComponentApply implements 'lokoctl component apply' separated from CLI
// dependencies.
func ComponentApply(contextLogger *log.Entry, componentsList []string, options ComponentApplyOptions) error {
	lokoConfig, diags := config.LoadConfig(options.ConfigPath, options.ValuesPath)
	if diags.HasErrors() {
		return diags
	}

	componentObjects, err := componentNamesToObjects(selectComponentNames(componentsList, *lokoConfig.RootConfig))
	if err != nil {
		return fmt.Errorf("getting component objects: %w", err)
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
