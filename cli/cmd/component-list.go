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
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokomotive/cli/cmd/cluster"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available components",
	Run:   runList,
}

func init() {
	componentCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component list",
	})

	if len(args) != 0 {
		contextLogger.Fatalf("Unknown argument provided for list")
	}

	fmt.Println("Available components:")

	comps := cluster.AvailableComponents()
	sort.Strings(comps)
	for _, name := range comps {
		fmt.Println("\t", name)
	}
}
