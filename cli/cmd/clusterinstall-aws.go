package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/install/aws"
)

var awsCfg = aws.NewConfig()

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Install Lokomotive cluster on AWS provider",
	Run:   runAWS,
}

func init() {
	clusterInstallCmd.AddCommand(awsCmd)

	awsCmd.Flags().StringVar(&awsCfg.AssetDir, "assets", "", "Path to directory where generated Kubernetes configs will be stored")

	awsCmd.Flags().StringVar(&awsCfg.CredsPath, "creds", "", "Path to AWS credentials file")
	awsCmd.MarkFlagRequired("creds")

	awsCmd.Flags().StringVar(&awsCfg.ClusterName, "cluster-name", "", "Name of the cluster to be created")
	awsCmd.MarkFlagRequired("cluster-name")

	awsCmd.Flags().StringVar(&awsCfg.DNSZone, "dns-zone", "", "DNS Zone for the cluster to be created")
	awsCmd.MarkFlagRequired("dns-zone")

	awsCmd.Flags().StringVar(&awsCfg.DNSZoneID, "dns-zone-id", "", "DNS Zone ID of the cluster to be created")
	awsCmd.MarkFlagRequired("dns-zone-id")

	awsCmd.Flags().StringVar(&awsCfg.SSHPubKey, "ssh-public-key", os.ExpandEnv("$HOME/.ssh/id_rsa.pub"), "Path to ssh public key")
}

func runAWS(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl install aws",
		"args":    args,
	})

	if awsCfg.AssetDir == "" {
		clusterIden := fmt.Sprintf("%s-%s", awsCfg.ClusterName, awsCfg.DNSZone)
		awsCfg.AssetDir = filepath.Join(os.ExpandEnv("$HOME"), ".lokoctl", clusterIden)
	}

	if err := aws.Install(awsCfg); err != nil {
		ctxLogger.Fatalf("error installing cluster on aws: %v", err)
	}

	fmt.Printf("\nYour configurations are stored in %s\n", awsCfg.AssetDir)
}
