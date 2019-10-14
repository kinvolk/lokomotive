package cmd

import (
	"github.com/spf13/cobra"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage a Lokomotive cluster",
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
