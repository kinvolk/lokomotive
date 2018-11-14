package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/components"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Interact with components of a Lokomotive cluster",
}

var installCmd = &cobra.Command{
	Use:               "install",
	Short:             "Install a component",
	Run:               runInstall,
	PersistentPreRunE: isKubeconfigSet,
}

var (
	namespace string
)

func init() {
	rootCmd.AddCommand(componentCmd)

	componentCmd.AddCommand(installCmd)
	componentCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (required)")

	installCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "namespace where the component will be installed")
}

func runInstall(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command":   "lokoctl components install",
		"namespace": namespace,
		"args":      args,
	})

	if len(args) == 0 {
		contextLogger.Fatalf("Component name missing from command. Must be one of: %q", components.List())
	}

	c, err := components.Get(args[0])
	if err != nil {
		contextLogger.Fatalf("No such component %q: %q. Must be one of: %q", args[0], err, components.List())
	}

	installOpts := &components.InstallOptions{
		Namespace: namespace,
	}

	if err = c.Install(kubeconfig, installOpts); err != nil {
		contextLogger.Fatalf("Installation of component %q failed: %q", c.Name, err)
	}
}
