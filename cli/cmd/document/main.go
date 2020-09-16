package main

import (
	"os"

	"github.com/kinvolk/lokomotive/cli/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
)

var documentCommand = &cobra.Command{
	Use:   "go run cli/cmd/document/main.go [path]",
	Short: "Generate reference documentation for lokoctl CLI",
	Args:  cobra.ExactArgs(1),
	Run:   runDocument,
}

func main() {
	Execute()
}

func Execute() {
	if err := documentCommand.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(viper.AutomaticEnv)
}

func runDocument(docCmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "go run cli/cmd/document/main.go",
		"args":    args,
	})

	err := doc.GenMarkdownTree(cmd.RootCmd, args[0])
	if err != nil {
		contextLogger.Fatalf("Failed to generate markdown documentation: %v", err)
	}

	contextLogger.Printf("Markdown documentation written to %s\n", args[0])
}
