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

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/kinvolk/lokomotive/pkg/config"
)

type kubeconfigSources struct {
	env        string
	configFile string
}

const (
	tmpPattern = "lokoctl-tests-"
)

func prepareKubeconfigSource(t *testing.T, k *kubeconfigSources) (*log.Entry, *config.Config) {
	if k.env == "" {
		// Ensure KUBECONFIG is not set.
		if err := os.Unsetenv(kubeconfigEnvVariable); err != nil {
			t.Fatalf("unsetting %q environment variable: %v", kubeconfigEnvVariable, err)
		}
	}

	if k.env != "" {
		// Ensure KUBECONFIG IS set.
		if err := os.Setenv(kubeconfigEnvVariable, k.env); err != nil {
			t.Fatalf("setting %q environment variable: %v", kubeconfigEnvVariable, err)
		}
	}

	// Ensure there is no lokocfg configuration in working directory.
	tmpDir, err := ioutil.TempDir("", tmpPattern)
	if err != nil {
		t.Fatalf("creating tmp dir: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("removing temp dir %q: %v", tmpDir, err)
		}
	})

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("changing working directory to %q: %v", tmpDir, err)
	}

	if k.configFile != "" {
		path := filepath.Join(tmpDir, "cluster.lokocfg")
		if err := ioutil.WriteFile(path, []byte(k.configFile), 0600); err != nil {
			t.Fatalf("writing file %q: %v", path, err)
		}
	}

	lokoConfig, diags := config.LoadConfig(tmpDir, "")
	if diags.HasErrors() {
		t.Fatalf("getting lokomotive configuration: %v", err)
	}

	return log.WithFields(log.Fields{}), lokoConfig
}

func TestGetKubeconfigBadConfig(t *testing.T) {
	k := &kubeconfigSources{
		configFile: `cluster "packet" {
  asset_dir = "/foo"
}`,
	}

	contextLogger, lokoConfig := prepareKubeconfigSource(t, k)

	kg := kubeconfigGetter{
		platformRequired: false,
	}

	kubeconfig, err := kg.getKubeconfig(contextLogger, lokoConfig)
	if err == nil {
		t.Errorf("getting kubeconfig with bad configuration should fail")
	}

	if kubeconfig != nil {
		t.Fatalf("if getting kubeconfig fails, empty content should be returned")
	}
}

func TestGetKubeconfigNoConfigButRequired(t *testing.T) {
	k := &kubeconfigSources{
		env: "/foo",
	}

	contextLogger, lokoConfig := prepareKubeconfigSource(t, k)

	kg := kubeconfigGetter{
		platformRequired: true,
		path:             "/bar",
	}

	kubeconfig, err := kg.getKubeconfig(contextLogger, lokoConfig)
	if err == nil {
		t.Errorf("getting kubeconfig with no configuration and platform required should fail")
	}

	if kubeconfig != nil {
		t.Fatalf("if getting kubeconfig fails, empty content should be returned")
	}
}

func TestGetKubeconfig(t *testing.T) {
	expectedContent := []byte("foo")

	f, err := ioutil.TempFile("", tmpPattern)
	if err != nil {
		t.Fatalf("creating temp file should succeed, got: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Logf("removing temp file %q: %v", f.Name(), err)
		}
	})

	if err := ioutil.WriteFile(f.Name(), expectedContent, 0600); err != nil {
		t.Fatalf("writing temp file %q should succeed, got: %v", f.Name(), err)
	}

	k := &kubeconfigSources{
		env: f.Name(),
	}

	contextLogger, lokoConfig := prepareKubeconfigSource(t, k)

	kg := kubeconfigGetter{
		platformRequired: false,
	}

	kubeconfig, err := kg.getKubeconfig(contextLogger, lokoConfig)
	if err != nil {
		t.Fatalf("getting kubeconfig: %v", err)
	}

	if !reflect.DeepEqual(kubeconfig, expectedContent) {
		t.Fatalf("expected %q, got %q", expectedContent, kubeconfig)
	}
}

func TestGetKubeconfigSourceFlag(t *testing.T) {
	expectedPath := []string{"/foo"}

	k := &kubeconfigSources{
		configFile: `cluster "packet" {
  asset_dir = "/bad"

  cluster_name      = ""
  controller_count  = 0
  facility          = ""
  management_cidrs  = []
  node_private_cidr = ""
  project_id        = ""
  ssh_pubkeys       = []
  dns {
    provider = ""
    zone     = ""
  }
  worker_pool "foo" {
    count = 0
  }
}`,
		env: "/badpath",
	}

	contextLogger, lokoConfig := prepareKubeconfigSource(t, k)

	kg := kubeconfigGetter{
		platformRequired: true,
		path:             expectedPath[0],
	}

	kubeconfig, err := kg.getKubeconfigSource(contextLogger, lokoConfig)
	if err != nil {
		t.Fatalf("getting kubeconfig: %v", err)
	}

	if !reflect.DeepEqual(kubeconfig, expectedPath) {
		t.Fatalf("expected %v, got %v", expectedPath, kubeconfig)
	}
}

func TestGetKubeconfigSourceConfigFile(t *testing.T) {
	expectedPath := []string{}

	k := &kubeconfigSources{
		configFile: `cluster "packet" {
  asset_dir = "/foo"

  cluster_name      = ""
  controller_count  = 0
  facility          = ""
  management_cidrs  = []
  node_private_cidr = ""
  project_id        = ""
  ssh_pubkeys       = []
  dns {
    provider = ""
    zone     = ""
  }
  worker_pool "foo" {
    count = 0
  }
}`,
		env: "/badpath",
	}

	contextLogger, lokoConfig := prepareKubeconfigSource(t, k)

	kg := kubeconfigGetter{
		platformRequired: true,
	}

	kubeconfig, err := kg.getKubeconfigSource(contextLogger, lokoConfig)
	if err != nil {
		t.Fatalf("getting kubeconfig: %v", err)
	}

	if !reflect.DeepEqual(kubeconfig, expectedPath) {
		t.Fatalf("expected %v, got %v", expectedPath, kubeconfig)
	}
}

func TestGetKubeconfigFromAssetsDir(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", tmpPattern)
	if err != nil {
		t.Fatalf("creating tmp dir: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("removing temp dir %q: %v", tmpDir, err)
		}
	})

	expected := []byte("foo")

	kubeconfigPath := assetsKubeconfig(tmpDir)
	dir, _ := filepath.Split(kubeconfigPath)

	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("creating directory structure %q: %v", dir, err)
	}

	if err := ioutil.WriteFile(kubeconfigPath, expected, 0600); err != nil {
		t.Fatalf("writing file %q: %v", kubeconfigPath, err)
	}

	k := &kubeconfigSources{
		configFile: fmt.Sprintf(`cluster "packet" {
  asset_dir = "%s"

  cluster_name      = ""
  controller_count  = 0
  facility          = ""
  management_cidrs  = []
  node_private_cidr = ""
  project_id        = ""
  ssh_pubkeys       = []
  dns {
    provider = ""
    zone     = ""
  }
  worker_pool "foo" {
    count = 0
  }
}`, tmpDir),
		env: "/badpath",
	}

	contextLogger, lokoConfig := prepareKubeconfigSource(t, k)

	kg := kubeconfigGetter{
		platformRequired: true,
	}

	kubeconfig, err := kg.getKubeconfig(contextLogger, lokoConfig)
	if err != nil {
		t.Fatalf("getting kubeconfig: %v", err)
	}

	if !reflect.DeepEqual(kubeconfig, expected) {
		t.Fatalf("expected %v, got %v", expected, kubeconfig)
	}
}

func TestGetKubeconfigSourceBadConfigFile(t *testing.T) {
	k := &kubeconfigSources{
		configFile: `cluster "packet" {
	asset_dir = "/foo"
}`,
	}

	contextLogger, lokoConfig := prepareKubeconfigSource(t, k)

	kg := kubeconfigGetter{
		platformRequired: true,
	}

	kubeconfig, err := kg.getKubeconfigSource(contextLogger, lokoConfig)
	if err == nil {
		t.Errorf("getting kubeconfig with bad configuration should fail")
	}

	if kubeconfig != nil {
		t.Fatalf("if getting kubeconfig fails, empty path should be returned")
	}
}

func TestGetKubeconfigSourceEnvVariable(t *testing.T) {
	expectedPath := []string{
		"/foo",
		defaultKubeconfigPath,
	}

	k := &kubeconfigSources{
		env: expectedPath[0],
	}

	contextLogger, lokoConfig := prepareKubeconfigSource(t, k)

	kg := kubeconfigGetter{
		platformRequired: false,
	}

	kubeconfig, err := kg.getKubeconfigSource(contextLogger, lokoConfig)
	if err != nil {
		t.Fatalf("getting kubeconfig: %v", err)
	}

	if !reflect.DeepEqual(kubeconfig, expectedPath) {
		t.Fatalf("expected %v, got %v", expectedPath, kubeconfig)
	}
}

func TestGetKubeconfigSourceDefault(t *testing.T) {
	expectedPath := []string{"", defaultKubeconfigPath}

	k := &kubeconfigSources{}

	contextLogger, lokoConfig := prepareKubeconfigSource(t, k)

	kg := kubeconfigGetter{
		platformRequired: false,
	}

	kubeconfig, err := kg.getKubeconfigSource(contextLogger, lokoConfig)
	if err != nil {
		t.Fatalf("getting kubeconfig: %v", err)
	}

	if !reflect.DeepEqual(kubeconfig, expectedPath) {
		t.Fatalf("expected %v, got %v", expectedPath, kubeconfig)
	}
}
