package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/components"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available components",
	Run:   runList,
}

func init() {
	componentCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl component list",
	})

	if len(args) != 0 {
		contextLogger.Fatalf("Unknown argument provided for list")
	}

	fmt.Println("Available components:")
	for _, name := range components.ListNames() {
		fmt.Println("\t", name)
	}
}
