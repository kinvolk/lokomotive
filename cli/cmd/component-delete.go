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

	"github.com/kinvolk/lokomotive/pkg/components"
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
	pf.BoolVarP(&deleteNamespace, "delete-namespace", "", false, "Delete namespace with component.")
}

func runDelete(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl component delete",
		"args":    args,
	})

	l, options := initialize(ctxLogger)
	for _, name := range args {
		_, err := components.Get(name)
		if err != nil {
			ctxLogger.Fatalf("Unsupported component, got: %v", err)
		}
	}

	options.DeleteNamespace = deleteNamespace

	l.DeleteComponents(args, options)
}
