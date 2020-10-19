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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kinvolk/lokomotive/cli/cmd/cluster"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Get the health of a cluster",
	Run:   runHealth,
}

//nolint:gochecknoinits
func init() {
	RootCmd.AddCommand(healthCmd)
	pf := healthCmd.PersistentFlags()
	pf.BoolVarP(&debug, "debug", "", false, "Print debug messages")
}

func runHealth(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl health",
		"args":    args,
	})

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	options := cluster.HealthOptions{
		ConfigPath: viper.GetString("lokocfg"),
		ValuesPath: viper.GetString("lokocfg-vars"),
	}

	if err := cluster.Health(contextLogger, options); err != nil {
		contextLogger.Fatalf("Checking cluster health failed: %v", err)
	}
}
