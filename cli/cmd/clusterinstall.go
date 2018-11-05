package cmd

import (
	"github.com/spf13/cobra"
)

var clusterInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Use for installing Lokomotive on various providers",
}

func init() {
	rootCmd.AddCommand(clusterInstallCmd)
}
