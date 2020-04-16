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
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl cluster destroy",
		"args":    args,
	})

	l, options := initialize(ctxLogger)
	if err := l.Destroy(options); err != nil {
		ctxLogger.Fatalf("Error destroying cluster: %v", err)
	}

	ctxLogger.Println("Cluster destroyed successfully")
	ctxLogger.Println("You can safely remove the assets directory now")
}
