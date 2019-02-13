package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

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

	packetCmd.Flags().StringVar(&packetCfg.AssetDir, "assets", "", "Path to directory where generated Kubernetes configs will be stored")

	packetCmd.Flags().StringVar(&packetCfg.AWSRegion, "aws-region", "", "AWS region to use for Route53")
	packetCmd.MarkFlagRequired("aws-region")

	packetCmd.Flags().StringVar(&packetCfg.ClusterName, "cluster-name", "", "Name of the cluster to be created")
	packetCmd.MarkFlagRequired("cluster-name")

	packetCmd.Flags().IntVar(&packetCfg.ControllerCount, "controller-count", 1, "Number of controller nodes")

	packetCmd.Flags().StringVar(&packetCfg.ControllerType, "controller-type", "baremetal_0", "Packet server type for controllers")

	packetCmd.Flags().StringVar(&packetCfg.AWSCredsPath, "aws-creds", "", "Path to AWS credentials file")
	packetCmd.MarkFlagRequired("creds")

	packetCmd.Flags().StringVar(&packetCfg.DNSZone, "dns-zone", "", "DNS Zone for the cluster to be created")
	packetCmd.MarkFlagRequired("dns-zone")

	packetCmd.Flags().StringVar(&packetCfg.DNSZoneID, "dns-zone-id", "", "DNS Zone ID of the cluster to be created")
	packetCmd.MarkFlagRequired("dns-zone-id")

	packetCmd.Flags().StringVar(&packetCfg.Facility, "facility", "", "Packet facility to deploy the cluster in (e.g. ams1)")
	packetCmd.MarkFlagRequired("facility")

	packetCmd.Flags().StringVar(&packetCfg.ProjectID, "project-id", "", "Packet project ID (e.g. 405efe9c-cce9-4c71-87c1-949c290b27dc)")
	packetCmd.MarkFlagRequired("project-id")

	packetCmd.Flags().StringVar(&packetCfg.SSHPubKey, "ssh-public-key", os.ExpandEnv("$HOME/.ssh/id_rsa.pub"), "Path to ssh public key")

	packetCmd.Flags().IntVar(&packetCfg.WorkerCount, "worker-count", 2, "Number of worker nodes")

	packetCmd.Flags().StringVar(&packetCfg.WorkerType, "worker-type", "baremetal_0", "Packet server type for workers")
}

func runPacket(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl install packet",
		"args":    args,
	})

	// Set Packet auth token.
	token := os.Getenv("PACKET_AUTH_TOKEN")
	if token == "" {
		ctxLogger.Fatal("PACKET_AUTH_TOKEN environment variable must be set")
	}
	packetCfg.AuthToken = token

	if packetCfg.AssetDir == "" {
		clusterIden := fmt.Sprintf("%s-%s", packetCfg.ClusterName, packetCfg.DNSZone)
		packetCfg.AssetDir = filepath.Join(os.ExpandEnv("$HOME"), ".lokoctl", clusterIden)
	}

	if err := packet.Install(packetCfg); err != nil {
		ctxLogger.Fatalf("error installing cluster on Packet: %v", err)
	}

	fmt.Printf("\nYour configurations are stored in %s\n", packetCfg.AssetDir)

	kubeconfigPath := path.Join(packetCfg.AssetDir, "auth", "kubeconfig")
	if err := verifyInstall(kubeconfigPath); err != nil {
		ctxLogger.Fatalf("Verify cluster installation on Packet: %v", err)
	}
}
