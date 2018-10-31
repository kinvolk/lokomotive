package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var kubeconfig string

var rootCmd = &cobra.Command{
	Use:   "lokoctl",
	Short: "Command line tool to interact with a Lokomotive Kubernetes cluster",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error in executing lokoctl: %q", err)
	}
}
