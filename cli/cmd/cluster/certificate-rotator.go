package cluster

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"github.com/kinvolk/lokomotive/pkg/platform"
)

type certificateRotatorConfig struct {
	clientSet            *kubernetes.Clientset
	newCACert            string
	logger               *log.Entry
	daemonSetsToRestart  []platform.Workload
	deploymentsToRestart []platform.Workload
}

type certificateRotator struct {
	config certificateRotatorConfig
}

const (
	// Time to wait between updating DaemonSet/Deployment and start looking if
	// workload has converged. This should account for kube-controller-manager election time,
	// reconciliation period and time spent in reconciliation loop (based on number of workloads
	// in the cluster). 10 seconds might not be enough.
	kubeControllerManagerReconciliationPeriod = 10 * time.Second
)

func newCertificateRotator(config certificateRotatorConfig) (*certificateRotator, error) {
	if config.clientSet == nil {
		return nil, fmt.Errorf("clientSet can't be nil")
	}

	if config.newCACert == "" {
		return nil, fmt.Errorf("new CA certificate can't be empty")
	}

	return &certificateRotator{
		config: config,
	}, nil
}

func (cr *certificateRotator) restartDaemonSetAndWaitToConverge(ns, name string) error {
	dsc := cr.config.clientSet.AppsV1().DaemonSets(ns)

	generation, err := k8sutil.RolloutRestartDaemonSet(dsc, name)
	if err != nil {
		return fmt.Errorf("restarting: %w", err)
	}

	// TODO: make sure this is the right value.
	time.Sleep(kubeControllerManagerReconciliationPeriod)

	options := k8sutil.WaitOptions{
		Generation: generation,
	}

	if err := k8sutil.WaitForDaemonSet(dsc, name, options); err != nil {
		return fmt.Errorf("waiting for DaemonSet to converge: %w", err)
	}

	return nil
}

func (cr *certificateRotator) restartDeploymentAndWaitToConverge(ns, name string) error {
	deployClient := cr.config.clientSet.AppsV1().Deployments(ns)

	generation, err := k8sutil.RolloutRestartDeployment(deployClient, name)
	if err != nil {
		return fmt.Errorf("restarting: %w", err)
	}

	// TODO: make sure this is the right value.
	time.Sleep(kubeControllerManagerReconciliationPeriod)

	options := k8sutil.WaitOptions{
		Generation: generation,
	}

	if err := k8sutil.WaitForDeployment(deployClient, name, options); err != nil {
		return fmt.Errorf("waiting for Deployment to converge: %w", err)
	}

	return nil
}

func (cr *certificateRotator) rotate() error {
	cr.config.logger.Printf("Waiting for all service account tokens on the cluster to be updated...")

	if err := cr.waitForUpdatedServiceAccountTokens(); err != nil {
		return fmt.Errorf("waiting for all service account tokens to be updated: %w", err)
	}

	cr.config.logger.Printf("All service account tokens has been updated with new Kubernetes CA certificate")

	for _, daemonSet := range cr.config.daemonSetsToRestart {
		cr.config.logger.Printf("Restarting DaemonSet %s/%s to pick up new Kubernetes CA Certificate",
			daemonSet.Namespace, daemonSet.Name)

		if err := cr.restartDaemonSetAndWaitToConverge(daemonSet.Namespace, daemonSet.Name); err != nil {
			return fmt.Errorf("restarting DaemonSet %s/%s: %w", daemonSet.Namespace, daemonSet.Name, err)
		}
	}

	for _, deployment := range cr.config.deploymentsToRestart {
		cr.config.logger.Printf("Restarting Deployment %s/%s to pick up new Kubernetes CA Certificate",
			deployment.Namespace, deployment.Name)

		if err := cr.restartDeploymentAndWaitToConverge(deployment.Namespace, deployment.Name); err != nil {
			return fmt.Errorf("restarting Deployment %s/%s: %w", deployment.Namespace, deployment.Name, err)
		}
	}

	return nil
}

func (cr *certificateRotator) waitForUpdatedServiceAccountTokens() error {
	for {
		allUpToDate, err := cr.allServiceAccountTokensIncludeNewCA()
		if err != nil {
			return fmt.Errorf("checking if all service account tokens include new CA certificate: %w", err)
		}

		if allUpToDate {
			cr.config.logger.Printf("all service account tokens are up to date and have new CA certificate")

			break
		}
	}

	return nil
}

func (cr *certificateRotator) allServiceAccountTokensIncludeNewCA() (bool, error) {
	secrets, err := cr.config.clientSet.CoreV1().Secrets("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "type=kubernetes.io/service-account-token",
	})
	if err != nil {
		return false, fmt.Errorf("getting secrets: %v", err)
	}

	allUpToDate := true

	for _, v := range secrets.Items {
		if string(v.Data["ca.crt"]) != cr.config.newCACert {
			allUpToDate = false
		}
	}

	return allUpToDate, nil
}
