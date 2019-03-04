package cmd

import (
	"fmt"
	"path"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/config"
	"github.com/kinvolk/lokoctl/pkg/install"
	"github.com/kinvolk/lokoctl/pkg/install/aws"
	"github.com/kinvolk/lokoctl/pkg/install/baremetal"
	"github.com/kinvolk/lokoctl/pkg/install/packet"
	"github.com/kinvolk/lokoctl/pkg/k8sutil"
	"github.com/kinvolk/lokoctl/pkg/lokomotive"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Install Lokomotive cluster and components",
}

var clusterInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Lokomotive cluster with components",
	Run:   runClusterInstall,
}

func init() {
	rootCmd.AddCommand(clusterCmd)
	clusterCmd.AddCommand(clusterInstallCmd)
}

func runClusterInstall(cmd *cobra.Command, args []string) {
	ctxLogger := log.WithFields(log.Fields{
		"command": "lokoctl cluster install",
		"args":    args,
	})

	lokoConfig, diags := config.LoadConfig("")
	if len(diags) > 0 {
		ctxLogger.Fatal(diags)
	}

	if lokoConfig.RootConfig.Cluster == nil {
		ctxLogger.Fatal("No cluster configured")
	}

	clusterType := lokoConfig.RootConfig.Cluster.Name
	clusterConfigBody := &lokoConfig.RootConfig.Cluster.Config

	var assetDir string

	switch clusterType {
	case "aws":
		awsCfg := aws.NewConfig()
		if diags := awsCfg.LoadConfig(clusterConfigBody, lokoConfig.EvalContext); len(diags) > 0 {
			ctxLogger.Fatal(diags)
		}
		if err := aws.Install(awsCfg); err != nil {
			ctxLogger.Fatalf("error installing cluster on AWS: %v", err)
		}
		assetDir = awsCfg.AssetDir
	case "bare-metal":
		baremetalCfg := baremetal.NewConfig()
		if diags := baremetalCfg.LoadConfig(clusterConfigBody, lokoConfig.EvalContext); len(diags) > 0 {
			ctxLogger.Fatal(diags)
		}
		if err := baremetal.Install(baremetalCfg); err != nil {
			ctxLogger.Fatalf("error installing cluster on bare-metal: %v", err)
		}
		assetDir = baremetalCfg.AssetDir
	case "packet":
		packetCfg := packet.NewConfig()
		if diags := packetCfg.LoadConfig(clusterConfigBody, lokoConfig.EvalContext); len(diags) > 0 {
			ctxLogger.Fatal(diags)
		}
		if packetCfg.AuthToken == "" {
			ctxLogger.Fatal("no Packet API token given")
		}
		if err := packet.Install(packetCfg); err != nil {
			ctxLogger.Fatalf("error installing cluster on Packet: %v", err)
		}
		assetDir = packetCfg.AssetDir
	default:
		ctxLogger.Fatalf("Cluster type %q is unknown", clusterType)
	}

	fmt.Printf("\nYour configurations are stored in %s\n", assetDir)

	kubeconfigPath := path.Join(assetDir, "auth", "kubeconfig")
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

func verifyInstall(kubeConfigPath string) error {
	client, err := k8sutil.NewClientset(kubeConfigPath)
	if err != nil {
		return errors.Wrapf(err, "failed to set up clientset")
	}

	cluster, err := lokomotive.NewCluster(client)
	if err != nil {
		return errors.Wrapf(err, "failed to set up cluster client")
	}

	return install.Verify(cluster)
}
