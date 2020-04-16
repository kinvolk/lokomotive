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

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Get the health of a Lokomotive cluster",
	Run:   runHealth,
}

func init() {
	RootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl health",
		"args":    args,
	})

	l, _ := initialize(ctxLogger)

	if err := l.Health(); err != nil {
		ctxLogger.Fatalf("Error retrieving health of the cluster: %q", err)
	}
}
