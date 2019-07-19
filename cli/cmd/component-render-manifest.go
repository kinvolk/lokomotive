package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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

	lokoConfig, diags := config.LoadConfig(viper.GetString("lokocfg"), viper.GetString("lokocfg-vars"))
	if diags.HasErrors() {
		for _, diagnostic := range diags {
			contextLogger.Error(diagnostic.Summary)
		}
		contextLogger.Fatal("Errors found while loading configuration")
	}

	var componentsToRender []string
	if len(args) > 0 {
		componentsToRender = append(componentsToRender, args...)
	} else {
		for _, component := range lokoConfig.RootConfig.Components {
			componentsToRender = append(componentsToRender, component.Name)
		}
	}

	if err := renderComponentManifests(lokoConfig, componentsToRender...); err != nil {
		contextLogger.Fatal(err)
	}
}

func renderComponentManifests(lokoConfig *config.Config, componentNames ...string) error {
	for _, componentName := range componentNames {
		contextLogger := log.WithFields(log.Fields{
			"component": componentName,
		})

		component, err := components.Get(componentName)
		if err != nil {
			return err
		}

		componentConfigBody := lokoConfig.LoadComponentConfigBody(componentName)

		if diags := component.LoadConfig(componentConfigBody, lokoConfig.EvalContext); diags.HasErrors() {
			for _, diagnostic := range diags {
				contextLogger.Error(diagnostic.Summary)
			}
			return diags
		}

		manifests, err := component.RenderManifests()
		if err != nil {
			return err
		}

		fmt.Printf("# manifests for component %s\n", componentName)
		for filename, manifest := range manifests {
			fmt.Printf("\n---\n# %s\n%s", filename, manifest)
		}
	}
	return nil
}
