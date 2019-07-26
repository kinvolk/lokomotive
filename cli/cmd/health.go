package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/kinvolk/lokoctl/pkg/k8sutil"
	"github.com/kinvolk/lokoctl/pkg/lokomotive"
)

var healthCmd = &cobra.Command{
	Use:               "health",
	Short:             "Get the health of a Lokomotive cluster",
	Run:               runHealth,
	PersistentPreRunE: doesKubeconfigExist,
}

func init() {
	rootCmd.AddCommand(healthCmd)
}

func runHealth(cmd *cobra.Command, args []string) {
	contextLogger := log.WithFields(log.Fields{
		"command": "lokoctl health",
		"args":    args,
	})

	kubeconfig, err := getKubeconfig()
	if err != nil {
		contextLogger.Fatalf("Error in finding kubeconfig file: %s", err)
	}
	client, err := k8sutil.NewClientset(kubeconfig)
	if err != nil {
		contextLogger.Fatalf("Error in creating setting up Kubernetes client: %q", err)
	}

	cluster, err := lokomotive.NewCluster(client)
	if err != nil {
		contextLogger.Fatalf("Error in creating new Lokomotive cluster: %q", err)
	}

	components, err := cluster.Health()
	if err != nil {
		contextLogger.Fatalf("Error in getting Lokomotive cluster health: %q", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

	// Print the header.
	fmt.Fprintln(w, "Name\tStatus\tMessage\tError\t")

	// An empty line between header and the body.
	fmt.Fprintln(w, "\t\t\t\t")

	for _, component := range components {

		// The client-go library defines only one `ComponenetConditionType` at the moment,
		// which is `ComponentHealthy`. However, iterating over the list keeps this from
		// breaking in case client-go adds another `ComponentConditionType`.
		for _, condition := range component.Conditions {
			line := fmt.Sprintf(
				"%s\t%s\t%s\t%s\t",
				component.Name, condition.Status, condition.Message, condition.Error,
			)

			fmt.Fprintln(w, line)
		}

		w.Flush()
	}
}
