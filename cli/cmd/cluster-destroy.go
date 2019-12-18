package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var confirm bool

var clusterDestroyCmd = &cobra.Command{
	Use:     "destroy",
	Short:   "Destroy Lokomotive cluster",
	Run:     runClusterDestroy,
	PreRunE: checkForDeleteConfirmation,
}

func init() {
	clusterCmd.AddCommand(clusterDestroyCmd)
	pf := clusterDestroyCmd.PersistentFlags()
	pf.BoolVarP(&confirm, "confirm", "", false, "Confirm cluster removal")
}

func checkForDeleteConfirmation(cmd *cobra.Command, args []string) error {
	if !confirm {
		return fmt.Errorf("PERMANENT LOSS OF DATA. ACTION CANNOT BE UNDONE\n" +
			"If you are sure you want to destroy the cluster, execute `cluster destroy --confirm` to continue\n",
		)
	}

	return nil
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

	if err := p.Destroy(); err != nil {
		ctxLogger.Fatalf("error destroying cluster: %v", err)
	}

	ctxLogger.Println("Cluster destroyed successfully")
	ctxLogger.Println("You can safely remove the assets directory now")
}
