package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "lokoctl",
	Short: "Command line tool to interact with a Lokomotive Kubernetes cluster",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(cobraInit)
}

func cobraInit() {
	viper.AutomaticEnv()
}
