// Copyright 2022 The Lokomotive Authors
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

// Package azure provides the implenentation of the Platform interface
// for Azure cloud provider.
package azure //nolint:testpackage

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"

	lokoconfig "github.com/kinvolk/lokomotive/pkg/config"
)

const (
	testWorkerCount = 1
)

// createTerraformConfigFile().
func TestCreateTerraformConfigFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "lokoctl-tests-")
	if err != nil {
		t.Fatalf("creating tmp dir should succeed, got: %v", err)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("failed to remove temp dir %q: %v", tmpDir, err)
		}
	}()

	c := &config{
		WorkerPools: []workerPool{
			{
				Name:  "foo",
				Count: testWorkerCount,
			},
		},
	}

	if err := createTerraformConfigFile(c, tmpDir); err != nil {
		t.Fatalf("creating Terraform config files should succeed, got: %v", err)
	}
}

func TestCheckNotEmptyWorkersEmpty(t *testing.T) {
	c := config{}

	if d := c.checkNotEmptyWorkers(); !d.HasErrors() {
		t.Errorf("Expected to fail with empty workers")
	}
}

func TestCreateTerraformConfigFileNonExistingPath(t *testing.T) {
	c := &config{}

	if err := createTerraformConfigFile(c, "/nonexisting"); err == nil {
		t.Fatalf("creating Terraform config files in non-existing path should fail")
	}
}

// Meta().
func TestMeta(t *testing.T) {
	assetDir := "foo"

	moreWorkers := 3

	c := &config{
		AssetDir: assetDir,
		WorkerPools: []workerPool{
			{
				Count: testWorkerCount,
			},
			{
				Count: moreWorkers,
			},
		},
	}

	expectedNodes := 4
	if e := c.Meta().ExpectedNodes; e != expectedNodes {
		t.Errorf("Meta should count workers from all pools. Expected %d, got %d", expectedNodes, e)
	}

	if a := c.Meta().AssetDir; a != assetDir {
		t.Errorf("Meta should return configured asset dir. Expected %q, got %q", assetDir, a)
	}
}

// checkWorkerPoolNamesUnique().
func TestCheckWorkerPoolNamesUniqueDuplicated(t *testing.T) {
	c := &config{
		WorkerPools: []workerPool{
			{
				Name: "foo",
			},
			{
				Name: "foo",
			},
		},
	}

	if d := c.checkWorkerPoolNamesUnique(); !d.HasErrors() {
		t.Fatalf("should return error when worker pools are duplicated")
	}
}

func TestCheckWorkerPoolNamesUnique(t *testing.T) {
	c := &config{
		WorkerPools: []workerPool{
			{
				Name: "foo",
			},
			{
				Name: "bar",
			},
		},
	}

	if d := c.checkWorkerPoolNamesUnique(); d.HasErrors() {
		t.Fatalf("should not return errors when pool names are unique, got: %v", d)
	}
}

func TestNotEmptyWorkersEmpty(t *testing.T) {
	c := &config{}

	if d := c.checkNotEmptyWorkers(); !d.HasErrors() {
		t.Fatalf("should return error when there is no worker pool defined")
	}
}

func TestNotEmptyWorkers(t *testing.T) {
	c := &config{
		WorkerPools: []workerPool{
			{
				Name: "foo",
			},
		},
	}

	if d := c.checkNotEmptyWorkers(); d.HasErrors() {
		t.Fatalf("should not return errors when worker pool is not empty, got: %v", d)
	}
}

func TestCheckWorkerPoolNamesUniqueTest(t *testing.T) {
	c := &config{
		WorkerPools: []workerPool{
			{
				Name: "foo",
			},
			{
				Name: "bar",
			},
		},
	}

	if d := c.checkWorkerPoolNamesUnique(); d.HasErrors() {
		t.Fatalf("should not return errors when pool names are unique, got: %v", d)
	}
}

// loadConfigFromString loads config from string.
func loadConfigFromString(t *testing.T, c string) hcl.Diagnostics {
	p := hclparse.NewParser()

	f, d := p.ParseHCL([]byte(c), "x.lokocfg")
	if d.HasErrors() {
		t.Fatalf("parsing HCL should succeed, got: %v", d)
	}

	configBody := hcl.MergeFiles([]*hcl.File{f})

	var rootConfig lokoconfig.RootConfig

	if d := gohcl.DecodeBody(configBody, nil, &rootConfig); d.HasErrors() {
		t.Fatalf("decoding root config should succeed, got: %v", d)
	}

	cc := &config{}

	return cc.LoadConfig(&rootConfig.Cluster.Config, &hcl.EvalContext{})
}

func TestLoadConfig(t *testing.T) {
	c := `
cluster "azure" {
  asset_dir           = "/fooo"
  ssh_pubkeys         = ["testkey"]
  controller_count    = 1
  cluster_name        = "mycluster"
  dns {
    zone     = "dnszone"
    provider = "route53"
  }
  worker_pool "foo" {
    count   = 1
    ssh_pubkeys = ["foo"]
  }
}
`
	if d := loadConfigFromString(t, c); d.HasErrors() {
		t.Fatalf("valid config should not return error, got: %v", d)
	}
}

func TestLoadConfigEmpty(t *testing.T) {
	c := `
cluster "azure" {}
`

	if d := loadConfigFromString(t, c); !d.HasErrors() {
		t.Fatalf("empty config should return error, got: %v", d)
	}
}

func TestLoadConfigBadHCL(t *testing.T) {
	c := `
cluster "azure" {
  not_defined_field = "doh"
}
`

	if d := loadConfigFromString(t, c); !d.HasErrors() {
		t.Fatalf("invalid HCL should return errors")
	}
}
