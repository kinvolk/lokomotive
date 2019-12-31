package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// Register platforms by adding an anonymous import.
	_ "github.com/kinvolk/lokoctl/pkg/install/aws"
	_ "github.com/kinvolk/lokoctl/pkg/install/baremetal"
	_ "github.com/kinvolk/lokoctl/pkg/install/packet"

	// Register backends by adding an anonymous import.
	_ "github.com/kinvolk/lokoctl/pkg/backend/local"
	_ "github.com/kinvolk/lokoctl/pkg/backend/s3"
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

	// Add kubeconfig flag.
	rootCmd.PersistentFlags().String(
		"kubeconfig",
		"", // Special empty default, use getKubeconfig()
		"Path to kubeconfig file, taken from the asset dir if not given, and finally falls back to ~/.kube/config")
	viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))

	rootCmd.PersistentFlags().String("lokocfg", "./", "Path to lokocfg directory or file")
	viper.BindPFlag("lokocfg", rootCmd.PersistentFlags().Lookup("lokocfg"))
	rootCmd.PersistentFlags().String("lokocfg-vars", "./lokocfg.vars", "Path to lokocfg.vars file")
	viper.BindPFlag("lokocfg-vars", rootCmd.PersistentFlags().Lookup("lokocfg-vars"))
}

func cobraInit() {
	viper.AutomaticEnv()
}
