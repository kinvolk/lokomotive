package cmd

import (
	"github.com/spf13/cobra"

	// Register a component by adding an anonymous import
	_ "github.com/kinvolk/lokoctl/pkg/components/calico-hostendpoint-controller"
	_ "github.com/kinvolk/lokoctl/pkg/components/cert-manager"
	_ "github.com/kinvolk/lokoctl/pkg/components/cluster-autoscaler"
	_ "github.com/kinvolk/lokoctl/pkg/components/contour"
	_ "github.com/kinvolk/lokoctl/pkg/components/dex"
	_ "github.com/kinvolk/lokoctl/pkg/components/external-dns"
	_ "github.com/kinvolk/lokoctl/pkg/components/flatcar-linux-update-operator"
	_ "github.com/kinvolk/lokoctl/pkg/components/gangway"
	_ "github.com/kinvolk/lokoctl/pkg/components/httpbin"
	_ "github.com/kinvolk/lokoctl/pkg/components/metallb"
	_ "github.com/kinvolk/lokoctl/pkg/components/metrics-server"
	_ "github.com/kinvolk/lokoctl/pkg/components/openebs-operator"
	_ "github.com/kinvolk/lokoctl/pkg/components/openebs-storage-class"
	_ "github.com/kinvolk/lokoctl/pkg/components/prometheus-operator"
	_ "github.com/kinvolk/lokoctl/pkg/components/rook"
	_ "github.com/kinvolk/lokoctl/pkg/components/rook-ceph"
	_ "github.com/kinvolk/lokoctl/pkg/components/velero"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Install Lokomotive components",
}

func init() {
	rootCmd.AddCommand(componentCmd)
}
