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
)

var confirm bool

var clusterDestroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy a cluster",
	Run:   runClusterDestroy,
}

func init() {
	clusterCmd.AddCommand(clusterDestroyCmd)
	pf := clusterDestroyCmd.PersistentFlags()
	pf.BoolVarP(&confirm, "confirm", "", false, "Destroy cluster without asking for confirmation")
	pf.BoolVarP(&verbose, "verbose", "v", false, "Show output from Terraform")
}

func runClusterDestroy(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl cluster destroy",
		"args":    args,
	})

	options := clusterDestroyOptions{
		confirm:    confirm,
		verbose:    verbose,
		configPath: viper.GetString("lokocfg"),
		valuesPath: viper.GetString("lokocfg-vars"),
	}

	if err := clusterDestroy(contextLogger, options); err != nil {
		contextLogger.Fatalf("Destroying cluster: %v", err)
	}
}

type clusterDestroyOptions struct {
	confirm    bool
	verbose    bool
	configPath string
	valuesPath string
}

func clusterDestroy(contextLogger *log.Entry, options clusterDestroyOptions) error {
	cc := clusterConfig{
		verbose:    options.verbose,
		configPath: options.configPath,
		valuesPath: options.valuesPath,
	}

	c, err := cc.initialize(contextLogger)
	if err != nil {
		return fmt.Errorf("initializing: %w", err)
	}

	exists, err := clusterExists(c.terraformExecutor)
	if err != nil {
		return fmt.Errorf("checking if cluster exists: %w", err)
	}

	if !exists {
		contextLogger.Println("Cluster already destroyed, nothing to do")

		return nil
	}

	if !options.confirm {
		confirmation := askForConfirmation("WARNING: This action cannot be undone. Do you really want to destroy the cluster?")
		if !confirmation {
			contextLogger.Println("Cluster destroy canceled")

			return nil
		}
	}

	if err := c.platform.Destroy(&c.terraformExecutor); err != nil {
		return fmt.Errorf("destroying cluster: %v", err)
	}

	contextLogger.Println("Cluster destroyed successfully")
	contextLogger.Println("You can safely remove the assets directory now")

	return nil
}
