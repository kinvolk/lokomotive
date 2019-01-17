package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/util/tools"
)

// doesKubeconfigExist checks if the kubeconfig provided by user exists
func doesKubeconfigExist(*cobra.Command, []string) error {
	var err error
	kubeconfig := viper.GetString("kubeconfig")
	if _, err = os.Stat(kubeconfig); os.IsNotExist(err) {
		return fmt.Errorf("Kubeconfig %q not found", kubeconfig)
	}
	return err
}

func validateComponentCmdArgs(cmd *cobra.Command, args []string) error {
	if err := doesKubeconfigExist(cmd, args); err != nil {
		return err
	}

	// check if the component name is given
	if len(args) == 0 {
		return fmt.Errorf("Component name missing from command. " +
			"See 'lokoctl component list' for available components.")
	}

	// The given component should exist
	if _, err := components.Get(args[0]); err != nil {
		return fmt.Errorf("%q %s. "+
			"See 'lokoctl component list' for available components.",
			args[0], err)
	}

	return nil
}

func componentAnswersFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&answers, "answers", "a", "", "Provide answers file to customize component behavior")
}

func clusterInstallChecks(*cobra.Command, []string) error {
	return tools.InstallerBinaries()
}
