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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokomotive/pkg/config"
	"github.com/kinvolk/lokomotive/pkg/lokomotive"
	"github.com/spf13/viper"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage a cluster",
}

func init() {
	RootCmd.AddCommand(clusterCmd)
}

func initialize(ctxLogger *logrus.Entry) (lokomotive.Manager, *lokomotive.Options) {
	// get lokocfg files and lokocfg vars path
	lokocfgPath := viper.GetString("lokocfg")
	variablesPath := viper.GetString("lokocfg-vars")
	// HCLLoader loads the user configuration in lokocfg files into concrete
	// LokomotiveConfig struct which is to be passed around for further operations
	hclLoader := &config.HCLLoader{
		ConfigPath:    lokocfgPath,
		VariablesPath: variablesPath,
	}

	cfg, diags := hclLoader.Load()
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			ctxLogger.Error(diagnostic.Error())
		}

		ctxLogger.Fatal("Errors found while loading configuration")
	}

	options := &lokomotive.Options{
		Verbose:         verbose,
		SkipComponents:  skipComponents,
		UpgradeKubelets: upgradeKubelets,
		Confirm:         confirm,
	}

	lokomotive, diags := lokomotive.NewLokomotive(ctxLogger, cfg, options)
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			ctxLogger.Error(diagnostic.Error())
		}

		ctxLogger.Fatal("Errors found while initializing Lokomotive")
	}

	return lokomotive, options
}
