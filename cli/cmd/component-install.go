package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/components/util"
	"github.com/kinvolk/lokoctl/pkg/config"
)

var componentInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a component",
	Run:   runInstall,
}

func init() {
	componentCmd.AddCommand(componentInstallCmd)
}

func runInstall(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component install",
		"args":    args,
	})

	lokoConfig, diags := getLokoConfig()
	if len(diags) > 0 {
		contextLogger.Fatal(diags)
	}

	var componentsToInstall []string
	if len(args) > 0 {
		componentsToInstall = append(componentsToInstall, args...)
	} else {
		for _, component := range lokoConfig.RootConfig.Components {
			componentsToInstall = append(componentsToInstall, component.Name)
		}
	}

	kubeconfig, err := getKubeconfig()
	if err != nil {
		contextLogger.Fatalf("Error in finding kubeconfig file: %s", err)
	}
	if err := installComponents(lokoConfig, kubeconfig, componentsToInstall...); err != nil {
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

		if err := util.InstallComponent(component, kubeconfig); err != nil {
			return err
		}
	}
	return nil
}
