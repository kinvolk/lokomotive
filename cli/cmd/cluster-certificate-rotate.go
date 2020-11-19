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

var clusterCertificateRotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate certificates of a cluster",
	Long: `Rotate certificates of a cluster.
Rotate will replace all certificates inside a cluster with new ones.
This can be used to renew all certificates with a longer validity.`,
	Run: runClusterCertificateRotate,
}

func init() {
	clusterCertificateCmd.AddCommand(clusterCertificateRotateCmd)

	pf := clusterCertificateRotateCmd.PersistentFlags()
	// TODO: check these
	pf.BoolVarP(&confirm, "confirm", "", false, "Upgrade cluster without asking for confirmation")
	pf.BoolVarP(&verbose, "verbose", "v", false, "Show output from Terraform")

	pf.BoolVarP(&skipPreUpdateHealthCheck, "skip-pre-update-health-check", "", false,
		"Skip ensuring that cluster is healthy before updating (not recommended)")
}

func runClusterCertificateRotate(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl cluster certificate rotate",
		"args":    args,
	})

	options := cluster.CertificateRotateOptions{
		Confirm:    confirm,
		Verbose:    verbose,
		ConfigPath: viper.GetString("lokocfg"),
		ValuesPath: viper.GetString("lokocfg-vars"),
	}

	if err := cluster.RotateCertificates(contextLogger, options); err != nil {
		contextLogger.Fatalf("Rotating Certificates failed: %v", err)
	}
}
