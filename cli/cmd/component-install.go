package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/kinvolk/lokoctl/pkg/components"
	// This registers the answers object with its corresponding component object
	// in `components` list, every time a new component is added an import needs
	// to be done here
	_ "github.com/kinvolk/lokoctl/pkg/components/cert-manager"
	_ "github.com/kinvolk/lokoctl/pkg/components/network-policies"
	_ "github.com/kinvolk/lokoctl/pkg/components/nginx-ingress"
)

var installCmd = &cobra.Command{
	Use:               "install",
	Short:             "Install a component",
	Run:               runInstall,
	PersistentPreRunE: doesKubeconfigExist,
}

var (
	answers string
)

func init() {
	componentCmd.AddCommand(installCmd)
	installCmd.Flags().StringVarP(&answers, "answers", "a", "", "Provide answers file to customize component behavior")
}

func runInstall(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component install",
		"args":    args,
	})

	if len(args) == 0 {
		contextLogger.Fatalf("Component name missing from command. Must be one of: %q", components.List())
	}

	c, err := components.Get(args[0])
	if err != nil {
		contextLogger.Fatalf("No such component %q: %q. See 'lokoctl component list' for available components", args[0], err)
	}

	installOpts := &components.InstallOptions{
		AnswersFile: answers,
	}

	if err = c.Install(viper.GetString("kubeconfig"), installOpts); err != nil {
		contextLogger.Fatalf("Installation of component %q failed: %q", c.Name, err)
	}
}
