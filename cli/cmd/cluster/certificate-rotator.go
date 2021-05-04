package cluster

import (
	"context"
	"errors"
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

// CertificateRotateOptions contains the options for the RotateCertificates function.
type CertificateRotateOptions struct {
	Confirm    bool
	Verbose    bool
	ConfigPath string
	ValuesPath string
}

// RotateCertificates replaces all certificates in a cluster.
// due to the nature of it running as a lokoctl command it should be idempotency.
func RotateCertificates(contextLogger *log.Entry, options CertificateRotateOptions) error {
	cc := clusterConfig{
		verbose:    options.Verbose,
		configPath: options.ConfigPath,
		valuesPath: options.ValuesPath,
	}

	c, err := cc.initialize(contextLogger)
	if err != nil {
		return fmt.Errorf("initializing: %w", err)
	}

	if err := canRotate(c); err != nil {
		return fmt.Errorf("cannot rotate cluster: %w", err)
	}

	// Tainting certificates so they get rotated.
	if err := c.taintCertificates(); err != nil {
		return fmt.Errorf("tainting certificate resources: %w", err)
	}

	// Apply the Terraform changes to replace tainted resources.
	// We run without parallel to make sure only one etcd can break before receiving an error.
	if err := c.platform.ApplyWithoutParallel(&c.terraformExecutor); err != nil {
		return fmt.Errorf("applying platform: %w", err)
	}

	return rotateControlPlaneCerts(contextLogger, cc)
}

func canRotate(c *cluster) error {
	exists, err := clusterExists(c.terraformExecutor)
	if err != nil {
		return fmt.Errorf("checking if cluster exists: %w", err)
	}

	if !exists {
		return errors.New("cannot rotate certificates on a non-existent cluster")
	}

	if c.platform.Meta().Managed {
		// TODO: do we want to error here?
		return errors.New("cannot rotate certificates on a managed cluster")
	}

	return nil
}

func rotateControlPlaneCerts(contextLogger *log.Entry, cc clusterConfig) error {
	c, err := cc.initialize(contextLogger)
	if err != nil {
		return fmt.Errorf("initializing: %w", err)
	}

	kg := kubeconfigGetter{
		platformRequired: true,
		clusterConfig:    cc,
	}

	kubeconfig, err := kg.getKubeconfig(contextLogger, c.lokomotiveConfig)
	if err != nil {
		return fmt.Errorf("getting kubeconfig: %v", err)
	}

	contextLogger.Log(log.InfoLevel, "Applying a controlplane update with the new CA")

	if err := c.upgradeControlPlane(contextLogger, kubeconfig); err != nil {
		return fmt.Errorf("running controlplane upgrade: %v", err)
	}

	cs, err := k8sutil.NewClientset(kubeconfig)
	if err != nil {
		return fmt.Errorf("creating clientset from kubeconfig: %w", err)
	}

	newCACert, err := c.readKubernetesCAFromTerraformOutput()
	if err != nil {
		return fmt.Errorf("reading Kubernetes CA certificate from Terraform output: %w", err)
	}

	crConfig := certificateRotatorConfig{
		clientSet:            cs,
		newCACert:            newCACert,
		logger:               contextLogger,
		daemonSetsToRestart:  c.platform.Meta().DaemonSets,
		deploymentsToRestart: c.platform.Meta().Deployments,
	}

	cr := certificateRotator{
		config: crConfig,
	}
	if cr.validate() != nil {
		return fmt.Errorf("preparing certificate rotator: %w", err)
	}

	// rotate() restarts control plane deployments and daemonsets again. We
	// need to it again because the first rollout is partial.
	if err := cr.rotate(); err != nil {
		return fmt.Errorf("rotating certificates: %w", err)
	}

	return nil
}

func (cr *certificateRotator) validate() error {
	if cr.config.clientSet == nil {
		return fmt.Errorf("clientSet can't be nil")
	}

	if cr.config.newCACert == "" {
		return fmt.Errorf("new CA certificate can't be empty")
	}

	return nil
}

// rotate will wait for service accounts to be signed by the new CA and restart all system
// DaemonSets and Deployments using the CA certificate.
func (cr *certificateRotator) rotate() error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(k8sutil.RolloutTimeout))
	defer cancel()

	cr.config.logger.Printf("Waiting for all service account tokens on the cluster to be updated...")

	if err := cr.waitForUpdatedServiceAccountTokens(ctx); err != nil {
		return fmt.Errorf("waiting for all service account tokens to be updated: %w", err)
	}

	cr.config.logger.Printf("All service account tokens has been updated with new Kubernetes CA certificate")

	for _, daemonSet := range cr.config.daemonSetsToRestart {
		cr.config.logger.Printf("Restarting DaemonSet %s/%s to pick up new Kubernetes CA Certificate",
			daemonSet.Namespace, daemonSet.Name)

		if err := k8sutil.RolloutDaemonSet(cr.config.clientSet.AppsV1().DaemonSets(daemonSet.Namespace), daemonSet.Name); err != nil {
			return fmt.Errorf("restarting DaemonSet %s/%s: %w", daemonSet.Namespace, daemonSet.Name, err)
		}
	}

	for _, deployment := range cr.config.deploymentsToRestart {
		cr.config.logger.Printf("Restarting Deployment %s/%s to pick up new Kubernetes CA Certificate",
			deployment.Namespace, deployment.Name)

		if err := k8sutil.RolloutDeployment(cr.config.clientSet.AppsV1().Deployments(deployment.Namespace), deployment.Name); err != nil {
			return fmt.Errorf("restarting Deployment %s/%s: %w", deployment.Namespace, deployment.Name, err)
		}
	}

	return nil
}

func (cr *certificateRotator) waitForUpdatedServiceAccountTokens(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context exceeded while checking if all service account tokens include new CA certificate: %w", ctx.Err())
		case <-time.After(time.Second):
			allUpToDate, err := cr.allServiceAccountTokensIncludeNewCA()
			if err != nil {
				return fmt.Errorf("checking if all service account tokens include new CA certificate: %w", err)
			}

			if allUpToDate {
				cr.config.logger.Printf("All service account tokens are up to date and have new CA certificate")

				return nil
			}
		}
	}
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
