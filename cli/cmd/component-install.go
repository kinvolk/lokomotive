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
	_ "github.com/kinvolk/lokoctl/pkg/components/ingress-nginx"
	_ "github.com/kinvolk/lokoctl/pkg/components/network-policies"
)

var installCmd = &cobra.Command{
	Use:               "install",
	Short:             "Install a component",
	Run:               runInstall,
	PersistentPreRunE: validateComponentCmdArgs,
}

var (
	answers   string
	namespace string
)

func init() {
	componentCmd.AddCommand(installCmd)
	componentAnswersFlag(installCmd)
	componentNamespaceFlag(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component install",
		"args":    args,
	})

	c, err := components.Get(args[0])
	if err != nil {
		contextLogger.Fatal(err)
	}

	installOpts := &components.InstallOptions{
		AnswersFile: answers,
		Namespace:   namespace,
	}

	if err := c.Install(viper.GetString("kubeconfig"), installOpts); err != nil {
		contextLogger.Fatalf("Installation of component %q failed: %q", c.Name, err)
	}
}
