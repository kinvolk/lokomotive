package cmd

import (
	"github.com/spf13/cobra"

	// Register a component by adding an anonymous import
	_ "github.com/kinvolk/lokoctl/pkg/components/cert-manager"
	_ "github.com/kinvolk/lokoctl/pkg/components/dex"
	_ "github.com/kinvolk/lokoctl/pkg/components/gangway"
	_ "github.com/kinvolk/lokoctl/pkg/components/httpbin"
	_ "github.com/kinvolk/lokoctl/pkg/components/ingress-nginx"
	_ "github.com/kinvolk/lokoctl/pkg/components/metallb"
	_ "github.com/kinvolk/lokoctl/pkg/components/openebs-default-storage-class"
	_ "github.com/kinvolk/lokoctl/pkg/components/openebs-operator"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Install Lokomotive components",
}

func init() {
	rootCmd.AddCommand(componentCmd)
}
