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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

type kubeconfigSources struct {
	flag       string
	env        string
	configFile string
}

const (
	tmpPattern = "lokoctl-tests-"
)

func prepareKubeconfigSource(t *testing.T, k *kubeconfigSources) {
	// Ensure viper flag is NOT empty.
	viper.Set(kubeconfigFlag, k.flag)

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
}

func TestGetKubeconfigBadConfig(t *testing.T) {
	k := &kubeconfigSources{
		configFile: `cluster "packet" {
  asset_dir = "/foo"
}`,
	}

	prepareKubeconfigSource(t, k)

	kubeconfig, err := getKubeconfig()
	if err == nil {
		t.Errorf("getting kubeconfig with bad configuration should fail")
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

	if err := ioutil.WriteFile(f.Name(), expectedContent, 0600); err != nil {
		t.Fatalf("writing temp file %q should succeed, got: %v", f.Name(), err)
	}

	k := &kubeconfigSources{
		env: f.Name(),
	}

	prepareKubeconfigSource(t, k)

	kubeconfig, err := getKubeconfig()
	if err != nil {
		t.Fatalf("getting kubeconfig: %v", err)
	}

	if !reflect.DeepEqual(kubeconfig, expectedContent) {
		t.Fatalf("expected %q, got %q", expectedContent, kubeconfig)
	}
}

func TestGetKubeconfigPathFlag(t *testing.T) {
	expectedPath := "/foo"

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
		flag: expectedPath,
		env:  "/badpath",
	}

	prepareKubeconfigSource(t, k)

	kubeconfig, err := getKubeconfigPath()
	if err != nil {
		t.Fatalf("getting kubeconfig: %v", err)
	}

	if kubeconfig != expectedPath {
		t.Fatalf("expected %q, got %q", expectedPath, kubeconfig)
	}
}

func TestGetKubeconfigPathConfigFile(t *testing.T) {
	expectedPath := assetsKubeconfig("/foo")

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

	prepareKubeconfigSource(t, k)

	kubeconfig, err := getKubeconfigPath()
	if err != nil {
		t.Fatalf("getting kubeconfig: %v", err)
	}

	if kubeconfig != expectedPath {
		t.Fatalf("expected %q, got %q", expectedPath, kubeconfig)
	}
}

func TestGetKubeconfigPathBadConfigFile(t *testing.T) {
	expectedPath := ""

	k := &kubeconfigSources{
		configFile: `cluster "packet" {
	asset_dir = "/foo"
}`,
	}

	prepareKubeconfigSource(t, k)

	kubeconfig, err := getKubeconfigPath()
	if err == nil {
		t.Errorf("getting kubeconfig with bad configuration should fail")
	}

	if kubeconfig != expectedPath {
		t.Fatalf("if getting kubeconfig fails, empty path should be returned")
	}
}

func TestGetKubeconfigPathEnvVariable(t *testing.T) {
	expectedPath := "/foo"

	k := &kubeconfigSources{
		env: expectedPath,
	}

	prepareKubeconfigSource(t, k)

	kubeconfig, err := getKubeconfigPath()
	if err != nil {
		t.Fatalf("getting kubeconfig: %v", err)
	}

	if kubeconfig != expectedPath {
		t.Fatalf("expected %q, got %q", expectedPath, kubeconfig)
	}
}

func TestGetKubeconfigPathDefault(t *testing.T) {
	expectedPath := defaultKubeconfigPath

	k := &kubeconfigSources{}

	prepareKubeconfigSource(t, k)

	kubeconfig, err := getKubeconfigPath()
	if err != nil {
		t.Fatalf("getting kubeconfig: %v", err)
	}

	if kubeconfig != expectedPath {
		t.Fatalf("expected %q, got %q", expectedPath, kubeconfig)
	}
}
