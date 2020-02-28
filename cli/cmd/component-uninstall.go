package cmd

import (
	"fmt"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util/helmutil"
	"github.com/kinvolk/lokomotive/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var confirmUninstall bool

var componentUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall a component",
	Run:   runUninstall,
}

func init() {
	componentCmd.AddCommand(componentUninstallCmd)
	pf := componentUninstallCmd.PersistentFlags()
	pf.BoolVarP(&confirmUninstall, "confirm", "", false, "Conform component uninstall")
}

func runUninstall(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component uninstall",
		"args":    args,
	})

	lokoConfig, diags := getLokoConfig()
	if diags.HasErrors() {
		contextLogger.Fatal(diags)
	}

	var componentsToUninstall []string
	if len(args) > 0 {
		componentsToUninstall = append(componentsToUninstall, args...)
	}

	kubeconfig, err := getKubeconfig()
	if err != nil {
		contextLogger.Fatal("Error in finding kubeconfig file: %s", err)
	}

	if err := uninstallComponents(lokoConfig, kubeconfig, componentsToUninstall...); err != nil {
		contextLogger.Fatal(err)
	}
}

func uninstallComponents(lokoConfig *config.Config, kubeconfig string, componentsToUninstall ...string) error {
	for _, componentName := range componentsToUninstall {
		fmt.Printf("Uninstalling component '%s'...\n", componentName)

		component, err := components.Get(componentName)
		if err != nil {
			return err
		}

		if err := helmutil.UninstallRelease(componentName, component, kubeconfig); err != nil {
			return err
		}
	}

	return nil
}
