package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/config"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a component",
	Run:   runInstall,
}

func init() {
	componentCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component install",
		"args":    args,
	})

	lokoConfig, diags := config.LoadConfig("")
	if len(diags) > 0 {
		contextLogger.Fatal(diags)
	}

	if err := installComponents(lokoConfig, viper.GetString("kubeconfig"), args[0]); err != nil {
		contextLogger.Fatal(err)
	}
}

func installComponents(lokoConfig *config.Config, kubeconfig string, componentNames ...string) error {
	for _, componentName := range componentNames {
		component, err := components.Get(componentName)
		if err != nil {
			return err
		}

		componentConfigBody := lokoConfig.LoadComponentConfigBody(componentName)

		if diags := component.LoadConfig(componentConfigBody, lokoConfig.EvalContext); len(diags) > 0 {
			fmt.Printf("%v\n", diags)
			return diags
		}

		if err := component.Install(kubeconfig); err != nil {
			return err
		}
	}
	return nil
}
