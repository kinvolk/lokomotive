package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// Register a platform by adding an anonymous import
	_ "github.com/kinvolk/lokoctl/pkg/install/aws"
	_ "github.com/kinvolk/lokoctl/pkg/install/baremetal"
	_ "github.com/kinvolk/lokoctl/pkg/install/packet"
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

	// add kubeconfig flag
	rootCmd.PersistentFlags().String(
		"kubeconfig",
		os.ExpandEnv("$HOME/.kube/config"),
		"Path to kubeconfig file")
	viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))

	rootCmd.PersistentFlags().String("lokocfg", "./", "Path to lokocfg directory or file")
	viper.BindPFlag("lokocfg", rootCmd.PersistentFlags().Lookup("lokocfg"))
	rootCmd.PersistentFlags().String("lokocfg-vars", "./lokocfg.vars", "Path to lokocfg.vars file")
	viper.BindPFlag("lokocfg-vars", rootCmd.PersistentFlags().Lookup("lokocfg-vars"))
}

func cobraInit() {
	viper.AutomaticEnv()
}
