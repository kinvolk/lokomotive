package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var confirm bool

var clusterDestroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy Lokomotive cluster",
	Run:   runClusterDestroy,
}

func init() {
	clusterCmd.AddCommand(clusterDestroyCmd)
	pf := clusterDestroyCmd.PersistentFlags()
	pf.BoolVarP(&confirm, "confirm", "", false, "Destroy cluster without asking for confirmation")
}

func runClusterDestroy(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl cluster destroy",
		"args":    args,
	})

	p, diags := getConfiguredPlatform()
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			ctxLogger.Error(diagnostic.Summary)
		}
		ctxLogger.Fatal("Errors found while loading cluster configuration")
	}

	if p == nil {
		ctxLogger.Fatal("No cluster configured")
	}

	if !confirm {
		confirmation := askForConfirmation("WARNING: This action cannot be undone. Do you really want to destroy the cluster?")
		if !confirmation {
			ctxLogger.Println("Cluster destroy canceled")
			return
		}
	}

	if err := p.Destroy(); err != nil {
		ctxLogger.Fatalf("error destroying cluster: %v", err)
	}

	ctxLogger.Println("Cluster destroyed successfully")
	ctxLogger.Println("You can safely remove the assets directory now")
}
