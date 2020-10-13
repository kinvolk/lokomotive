// Copyright 2020 The Lokomotive Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package aks is a Platform implementation for creating a Kubernetes cluster using
// Azure AKS.
package aks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/mitchellh/go-homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	v1 "k8s.io/client-go/kubernetes/typed/storage/v1"

	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"github.com/kinvolk/lokomotive/pkg/platform"
	"github.com/kinvolk/lokomotive/pkg/terraform"
)

// workerPool defines "worker_pool" block.
type workerPool struct {
	// Label field.
	Name string `hcl:"name,label"`

	// Block properties.
	Count  int               `hcl:"count,optional"`
	VMSize string            `hcl:"vm_size,optional"`
	Labels map[string]string `hcl:"labels,optional"`
	Taints []string          `hcl:"taints,optional"`
}

// config defines "cluster" block for AKS.
type config struct {
	AssetDir    string            `hcl:"asset_dir,optional"`
	ClusterName string            `hcl:"cluster_name,optional"`
	Tags        map[string]string `hcl:"tags,optional"`

	// Azure specific.
	TenantID       string `hcl:"tenant_id,optional"`
	SubscriptionID string `hcl:"subscription_id,optional"`
	ClientID       string `hcl:"client_id,optional"`
	ClientSecret   string `hcl:"client_secret,optional"`

	Location string `hcl:"location,optional"`

	// ApplicationName for created service principal.
	ApplicationName string `hcl:"application_name,optional"`

	ResourceGroupName   string `hcl:"resource_group_name,optional"`
	ManageResourceGroup bool   `hcl:"manage_resource_group,optional"`

	WorkerPools []workerPool `hcl:"worker_pool,block"`

	KubernetesVersion string
}

const (
	name = "aks"

	// Environment variables used to load sensitive parts of the configuration.
	clientIDEnv       = "LOKOMOTIVE_AKS_CLIENT_ID"
	clientSecretEnv   = "LOKOMOTIVE_AKS_CLIENT_SECRET" // #nosec G101
	subscriptionIDEnv = "LOKOMOTIVE_AKS_SUBSCRIPTION_ID"
	tenantIDEnv       = "LOKOMOTIVE_AKS_TENANT_ID"

	kubernetesVersion = "1.18.8"
)

// init registers AKS as a platform.
func init() { //nolint:gochecknoinits
	c := &config{
		Location:            "West Europe",
		ManageResourceGroup: true,
		KubernetesVersion:   kubernetesVersion,
	}

	platform.Register(name, c)
}

// LoadConfig loads configuration values into the config struct from given HCL configuration.
func (c *config) LoadConfig(configBody *hcl.Body, evalContext *hcl.EvalContext) hcl.Diagnostics {
	if configBody == nil {
		emptyConfig := hcl.EmptyBody()
		configBody = &emptyConfig
	}

	if d := gohcl.DecodeBody(*configBody, evalContext, c); d.HasErrors() {
		return d
	}

	return c.checkValidConfig()
}

// checkValidConfig validates cluster configuration.
func (c *config) checkValidConfig() hcl.Diagnostics {
	var d hcl.Diagnostics

	d = append(d, c.checkNotEmptyWorkers()...)
	d = append(d, c.checkWorkerPoolNamesUnique()...)
	d = append(d, c.checkWorkerPools()...)
	d = append(d, c.checkCredentials()...)
	d = append(d, c.checkRequiredFields()...)

	return d
}

// checkWorkerPools validates all configured worker pool fields.
func (c *config) checkWorkerPools() hcl.Diagnostics {
	var d hcl.Diagnostics

	for _, w := range c.WorkerPools {
		if w.VMSize == "" {
			d = append(d, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("pool %q: VMSize field can't be empty", w.Name),
			})
		}

		if w.Count <= 0 {
			d = append(d, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("pool %q: count must be bigger than 0", w.Name),
			})
		}
	}

	return d
}

// checkRequiredFields checks if that all required fields are populated in the top level configuration.
func (c *config) checkRequiredFields() hcl.Diagnostics {
	var d hcl.Diagnostics

	if c.SubscriptionID == "" && os.Getenv(subscriptionIDEnv) == "" {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "cannot find the Azure subscription ID",
			Detail: fmt.Sprintf("%q field is empty and %q environment variable "+
				"is not defined. At least one of these should be defined",
				"SubscriptionID", subscriptionIDEnv),
		})
	}

	if c.TenantID == "" && os.Getenv(tenantIDEnv) == "" {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "cannot find the Azure client ID",
			Detail: fmt.Sprintf("%q field is empty and %q environment variable "+
				"is not defined. At least one of these should be defined", "TenantID", tenantIDEnv),
		})
	}

	f := map[string]string{
		"AssetDir":          c.AssetDir,
		"ClusterName":       c.ClusterName,
		"ResourceGroupName": c.ResourceGroupName,
	}

	for k, v := range f {
		if v == "" {
			d = append(d, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("field %q can't be empty", k),
			})
		}
	}

	return d
}

// checkCredentials checks if credentials are correctly defined.
func (c *config) checkCredentials() hcl.Diagnostics {
	var d hcl.Diagnostics

	// If the application name is defined, we assume that we work as a highly privileged
	// account which has permissions to create new Azure AD application, so Client ID
	// and Client Secret fields are not needed.
	if c.ApplicationName != "" {
		if c.ClientID != "" {
			d = append(d, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "ClientID and ApplicationName are mutually exclusive",
			})
		}

		if c.ClientSecret != "" {
			d = append(d, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "ClientSecret and ApplicationName are mutually exclusive",
			})
		}

		return d
	}

	if c.ClientSecret == "" && os.Getenv(clientSecretEnv) == "" {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "cannot find the Azure client secret",
			Detail: fmt.Sprintf("%q field is empty and %q environment variable "+
				"is not defined. At least one of these should be defined", "ClientSecret", clientSecretEnv),
		})
	}

	if c.ClientID == "" && os.Getenv(clientIDEnv) == "" {
		d = append(d, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "cannot find the Azure client ID",
			Detail: fmt.Sprintf("%q field is empty and %q environment variable is "+
				"not defined. At least one of these should be defined", "ClientID", clientIDEnv),
		})
	}

	return d
}

// checkNotEmptyWorkers checks if the cluster has at least 1 node pool defined.
func (c *config) checkNotEmptyWorkers() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	if len(c.WorkerPools) == 0 {
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "At least one worker pool must be defined",
			Detail:   "Make sure to define at least one worker pool block in your cluster block",
		})
	}

	return diagnostics
}

// checkWorkerPoolNamesUnique verifies that all worker pool names are unique.
func (c *config) checkWorkerPoolNamesUnique() hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	dup := make(map[string]bool)

	for _, w := range c.WorkerPools {
		if !dup[w.Name] {
			dup[w.Name] = true
			continue
		}

		// It is duplicated.
		diagnostics = append(diagnostics, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Worker pool names should be unique",
			Detail:   fmt.Sprintf("Worker pool '%v' is duplicated", w.Name),
		})
	}

	return diagnostics
}

// Meta is part of Platform interface and returns common information about the platform configuration.
func (c *config) Meta() platform.Meta {
	nodes := 0
	for _, workerpool := range c.WorkerPools {
		nodes += workerpool.Count
	}

	return platform.Meta{
		AssetDir:      c.AssetDir,
		ExpectedNodes: nodes,
		Managed:       true,
	}
}

// Apply creates AKS infrastructure via Terraform.
func (c *config) Apply(ex *terraform.Executor) error {
	if err := c.Initialize(ex); err != nil {
		return err
	}

	return ex.Apply()
}

// Destroy destroys AKS infrastructure via Terraform.
func (c *config) Destroy(ex *terraform.Executor) error {
	if err := c.Initialize(ex); err != nil {
		return err
	}

	return ex.Destroy()
}

// Initialize creates Terrafrom files required for AKS.
func (c *config) Initialize(ex *terraform.Executor) error {
	assetDir, err := homedir.Expand(c.AssetDir)
	if err != nil {
		return err
	}

	terraformRootDir := terraform.GetTerraformRootDir(assetDir)

	return createTerraformConfigFile(c, terraformRootDir)
}

const (
	retryInterval = 5 * time.Second
	timeout       = 10 * time.Minute
)

// PostApplyHook implements platform.PlatformWithPostApplyHook interface and defines hooks
// which should be executed after AKS cluster is created.
func (c *config) PostApplyHook(kubeconfig []byte) error {
	client, err := k8sutil.NewClientset(kubeconfig)
	if err != nil {
		return fmt.Errorf("creating clientset from kubeconfig: %w", err)
	}

	return waitForDefaultStorageClass(client.StorageV1().StorageClasses())
}

// waitForDefaultStorageClass waits until the default storage class appears on a given cluster.
// If it doesn't appear within a defined time range, an error is returned.
func waitForDefaultStorageClass(sci v1.StorageClassInterface) error {
	defaultStorageClassAnnotation := "storageclass.kubernetes.io/is-default-class"

	if err := wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		scs, err := sci.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return false, fmt.Errorf("getting storage classes: %w", err)
		}

		for _, sc := range scs.Items {
			if v, ok := sc.ObjectMeta.Annotations[defaultStorageClassAnnotation]; ok && v == "true" {
				return true, nil
			}
		}

		return false, nil
	}); err != nil {
		return fmt.Errorf("waiting for the default storage class to be configured: %w", err)
	}

	return nil
}

// createTerraformConfigFiles create Terraform config files in given directory.
func createTerraformConfigFile(cfg *config, terraformRootDir string) error {
	t := template.Must(template.New("t").Parse(terraformConfigTmpl))

	path := filepath.Join(terraformRootDir, "cluster.tf")

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", path, err)
	}

	platform.AppendVersionTag(&cfg.Tags)

	if cfg.ClientSecret == "" {
		cfg.ClientSecret = os.Getenv(clientSecretEnv)
	}

	if cfg.SubscriptionID == "" {
		cfg.SubscriptionID = os.Getenv(subscriptionIDEnv)
	}

	if cfg.ClientID == "" {
		cfg.ClientID = os.Getenv(clientIDEnv)
	}

	if cfg.TenantID == "" {
		cfg.TenantID = os.Getenv(tenantIDEnv)
	}

	if err := t.Execute(f, cfg); err != nil {
		return fmt.Errorf("failed to write template to file %q: %w", path, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("failed closing file %q: %w", path, err)
	}

	return nil
}
