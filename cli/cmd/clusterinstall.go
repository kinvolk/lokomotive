package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/install"
	"github.com/kinvolk/lokoctl/pkg/k8sutil"
	"github.com/kinvolk/lokoctl/pkg/lokomotive"
)

var clusterInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Use for installing Lokomotive on various providers",
}

func init() {
	rootCmd.AddCommand(clusterInstallCmd)
}

func verifyInstall(kubeConfigPath string) error {
	client, err := k8sutil.NewClientset(kubeConfigPath)
	if err != nil {
		return errors.Wrapf(err, "failed to set up clientset")
	}

	cluster, err := lokomotive.NewCluster(client)
	if err != nil {
		return errors.Wrapf(err, "failed to set up cluster client")
	}

	return install.Verify(cluster)
}
