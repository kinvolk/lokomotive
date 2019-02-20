package cmd

import (
	"github.com/spf13/cobra"

	// Register a component by adding an anonymous import
	_ "github.com/kinvolk/lokoctl/pkg/components/cert-manager"
	_ "github.com/kinvolk/lokoctl/pkg/components/ingress-nginx"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Interact with components of a Lokomotive cluster",
}

func init() {
	rootCmd.AddCommand(componentCmd)
}
