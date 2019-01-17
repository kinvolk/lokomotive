package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/install/baremetal"
)

var baremetalCfg = baremetal.NewConfig()

var baremetalCmd = &cobra.Command{
	Use:               "baremetal",
	Short:             "Install Lokomotive cluster in a bare-metal environment",
	Run:               runBaremetal,
	PersistentPreRunE: clusterInstallChecks,
}

func init() {
	clusterInstallCmd.AddCommand(baremetalCmd)

	baremetalCmd.Flags().StringVar(&baremetalCfg.AssetDir, "assets", "", "Path to directory where generated assets and Kubernetes configs will be stored (contains secrets)")

	baremetalCmd.Flags().StringVar(&baremetalCfg.CachedInstall, "cached-install", "false", "Whether Flatcar Linux should PXE boot and install from matchbox /assets cache. Note that the Flatcar image must have been downloaded into Matchbox assets.")

	baremetalCmd.Flags().StringVar(&baremetalCfg.ClusterName, "cluster-name", "", "Name of the cluster to be created")
	baremetalCmd.MarkFlagRequired("cluster-name")

	baremetalCmd.Flags().StringSliceVar(&baremetalCfg.ControllerDomains, "controller-domains", []string{}, "Ordered list of controller FQDNs seperated by comma")
	baremetalCmd.MarkFlagRequired("controller-domain")

	baremetalCmd.Flags().StringSliceVar(&baremetalCfg.ControllerMacs, "controller-macs", []string{}, "Ordered list of controller identifying MAC addresses seperated by comma")
	baremetalCmd.MarkFlagRequired("controller-mac")

	baremetalCmd.Flags().StringSliceVar(&baremetalCfg.ControllerNames, "controller-names", []string{}, "Ordered list of controller short names separated by comma")
	baremetalCmd.MarkFlagRequired("controller-name")

	baremetalCmd.Flags().StringVar(&baremetalCfg.K8sDomainName, "k8s-domain-name", "", "FQDN resolving to the controller(s) nodes. Workers and kubectl will communicate with this endpoint")
	baremetalCmd.MarkFlagRequired("k8s-domain-name")

	baremetalCmd.Flags().StringVar(&baremetalCfg.MatchboxCAPath, "matchbox-ca", "", "Path to Matchbox CA file")
	baremetalCmd.MarkFlagRequired("matchbox-ca")

	baremetalCmd.Flags().StringVar(&baremetalCfg.MatchboxClientCertPath, "matchbox-client-cert", "", "Path to Matchbox client certificate")
	baremetalCmd.MarkFlagRequired("matchbox-client-cert")

	baremetalCmd.Flags().StringVar(&baremetalCfg.MatchboxClientKeyPath, "matchbox-client-key", "", "Path to Matchbox client key")
	baremetalCmd.MarkFlagRequired("matchbox-client-key")

	baremetalCmd.Flags().StringVar(&baremetalCfg.MatchboxEndpoint, "matchbox-endpoint", "", "Matchbox endpoint without http protocol e.g matchbox.example.com:8081")
	baremetalCmd.MarkFlagRequired("matchbox-endpoint")

	baremetalCmd.Flags().StringVar(&baremetalCfg.MatchboxHTTPEndpoint, "matchbox-http-endpoint", "", "Matchbox HTTP read-only endpoint e.g http://matchbox.example.com:8080")
	baremetalCmd.MarkFlagRequired("matchbox-http-endpoint")

	baremetalCmd.Flags().StringVar(&baremetalCfg.OSChannel, "os-channel", "flatcar-stable", "Channel for Flatcar Linux")

	baremetalCmd.Flags().StringVar(&baremetalCfg.OSVersion, "os-version", "current", "Version for Flatcar Linux to PXE and install")

	baremetalCmd.Flags().StringVar(&baremetalCfg.SSHPubKeyPath, "ssh-public-key", os.ExpandEnv("$HOME/.ssh/id_rsa.pub"), "path to ssh public key")

	baremetalCmd.Flags().StringSliceVar(&baremetalCfg.WorkerNames, "worker-names", []string{}, "Ordered list of worker short names seperated by comma")
	baremetalCmd.MarkFlagRequired("worker-names")

	baremetalCmd.Flags().StringSliceVar(&baremetalCfg.WorkerMacs, "worker-macs", []string{}, "Ordered list of worker identifying MAC addresses seperated by comma")
	baremetalCmd.MarkFlagRequired("worker-macs")

	baremetalCmd.Flags().StringSliceVar(&baremetalCfg.WorkerDomains, "worker-domains", []string{}, "Ordered list of worker FQDNs seperated by comma")
	baremetalCmd.MarkFlagRequired("worker-domains")
}

func runBaremetal(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl install baremetal",
		"args":    args,
	})

	if baremetalCfg.AssetDir == "" {
		baremetalCfg.AssetDir = filepath.Join(os.ExpandEnv("$HOME"), ".lokoctl", baremetalCfg.ClusterName)
	}

	if err := baremetal.Install(baremetalCfg); err != nil {
		ctxLogger.Fatalf("error installing cluster on baremetal: %v", err)
	}

	fmt.Printf("\nYour configurations are stored in %s\n", baremetalCfg.AssetDir)

	kubeconfigPath := path.Join(baremetalCfg.AssetDir, "auth", "kubeconfig")
	if err := verifyInstall(kubeconfigPath); err != nil {
		ctxLogger.Fatalf("Verify cluster installation on baremetal: %v", err)
	}
}
