package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/components"
)

var componentRenderCmd = &cobra.Command{
	Use:               "render-manifest",
	Short:             "Render a component manifests",
	Run:               runComponentRender,
	PersistentPreRunE: validateComponentCmdArgs,
}

func init() {
	componentCmd.AddCommand(componentRenderCmd)
	componentAnswersFlag(componentRenderCmd)
}

func runComponentRender(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component render-manifest",
		"args":    args,
	})

	c, err := components.Get(args[0])
	if err != nil {
		contextLogger.Fatal(err)
	}

	installOpts := &components.InstallOptions{
		AnswersFile: answers,
	}

	if err := c.RenderManifests(installOpts); err != nil {
		contextLogger.Fatal(err)
	}
}
