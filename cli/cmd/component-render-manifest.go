package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/components"
	"github.com/kinvolk/lokoctl/pkg/config"
)

var componentRenderCmd = &cobra.Command{
	Use:   "render-manifest",
	Short: "Render and print manifests for a component",
	Run:   runComponentRender,
}

func init() {
	componentCmd.AddCommand(componentRenderCmd)
}

func runComponentRender(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component render-manifest",
		"args":    args,
	})

	componentName := args[0]
	component, err := components.Get(componentName)
	if err != nil {
		contextLogger.Fatal(err)
	}

	lokoConfig, diags := config.LoadConfig("")
	if len(diags) > 0 {
		contextLogger.Fatal(diags)
	}

	componentConfigBody := lokoConfig.LoadComponentConfigBody(componentName)

	if diags := component.LoadConfig(componentConfigBody, lokoConfig.EvalContext); len(diags) > 0 {
		contextLogger.Fatal(diags)
	}

	manifests, err := component.RenderManifests()
	if err != nil {
		contextLogger.Fatal(err)
	}

	for filename, manifest := range manifests {
		fmt.Printf("---\n# %s\n%s", filename, manifest)
	}
}
