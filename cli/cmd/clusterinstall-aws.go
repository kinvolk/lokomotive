package cmd

import (
	"fmt"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/config"
	"github.com/kinvolk/lokoctl/pkg/install/aws"
)

var awsCfg = aws.NewConfig()

var awsCmd = &cobra.Command{
	Use:               "aws",
	Short:             "Install Lokomotive cluster on AWS provider",
	Run:               runAWS,
	PersistentPreRunE: clusterInstallChecks,
}

func init() {
	clusterInstallCmd.AddCommand(awsCmd)
}

func runAWS(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl install aws",
		"args":    args,
	})

	lokoConfig, diags := config.LoadConfig("")
	if len(diags) > 0 {
		ctxLogger.Fatal(diags)
	}

	if lokoConfig.RootConfig.Cluster == nil {
		ctxLogger.Fatal("No cluster configured")
	}

	clusterConfigBody := &lokoConfig.RootConfig.Cluster.Config
	if diags := awsCfg.LoadConfig(clusterConfigBody, lokoConfig.EvalContext); len(diags) > 0 {
		ctxLogger.Fatal(diags)
	}

	if err := aws.Install(awsCfg); err != nil {
		ctxLogger.Fatalf("error installing cluster on Packet: %v", err)
	}

	fmt.Printf("\nYour configurations are stored in %s\n", awsCfg.AssetDir)

	kubeconfigPath := path.Join(awsCfg.AssetDir, "auth", "kubeconfig")
	if err := verifyInstall(kubeconfigPath); err != nil {
		ctxLogger.Fatalf("Verify cluster installation on Packet: %v", err)
	}

	var componentsToInstall []string
	for _, component := range lokoConfig.RootConfig.Components {
		componentsToInstall = append(componentsToInstall, component.Name)
	}

	if len(componentsToInstall) > 0 {
		installComponents(lokoConfig, kubeconfigPath, componentsToInstall...)
	}
}
