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

var componentApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Deploy or update a component",
	Long: `Deploy or update a component.
Deploys a component if not yet present, otherwise updates it.
When run with no arguments, all components listed in the configuration are applied.`,
	Run: runApply,
}

func init() {
	componentCmd.AddCommand(componentApplyCmd)
}

func runApply(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl component apply",
		"args":    args,
	})

	l, _ := initialize(ctxLogger)
	for _, name := range args {
		_, err := components.Get(name)
		if err != nil {
			ctxLogger.Fatalf("Unsupported component, got: %v", err)
		}
	}

	//Install components mentioned as arguments,if no arguments provided
	//install all configured components
	l.ApplyComponents(args)
}
