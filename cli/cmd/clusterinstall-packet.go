package cmd

import (
	"fmt"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/config"
	"github.com/kinvolk/lokoctl/pkg/install/packet"
)

var packetCfg = packet.NewConfig()

var packetCmd = &cobra.Command{
	Use:               "packet",
	Short:             "Install Lokomotive cluster on Packet",
	Run:               runPacket,
	PersistentPreRunE: clusterInstallChecks,
}

func init() {
	clusterInstallCmd.AddCommand(packetCmd)
}

func runPacket(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl install packet",
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
	if diags := packetCfg.LoadConfig(clusterConfigBody, lokoConfig.EvalContext); len(diags) > 0 {
		ctxLogger.Fatal(diags)
	}

	if packetCfg.AuthToken == "" {
		ctxLogger.Fatal("no Packet API token given")
	}

	if err := packet.Install(packetCfg); err != nil {
		ctxLogger.Fatalf("error installing cluster on Packet: %v", err)
	}

	fmt.Printf("\nYour configurations are stored in %s\n", packetCfg.AssetDir)

	kubeconfigPath := path.Join(packetCfg.AssetDir, "auth", "kubeconfig")
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
