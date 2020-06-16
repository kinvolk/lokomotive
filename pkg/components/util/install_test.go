package util_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/kinvolk/lokomotive/pkg/components/util"
)

const (
	validKubeconfig = `
apiVersion: v1
kind: Config
clusters:
- name: admin
  cluster:
    server: https://nonexistent:6443
users:
- name: admin
  user:
    token: "foo.bar"
current-context: admin
contexts:
- name: admin
  context:
    cluster: admin
    user: admin
`
)

func TestHelmActionConfigFromValidKubeconfigFile(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "lokoctl-tests-")
	if err != nil {
		t.Fatalf("creating tmp file should succeed, got: %v", err)
	}

	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("failed to remove tmp file %q: %v", tmpFile.Name(), err)
		}
	}()

	if _, err := tmpFile.Write([]byte(validKubeconfig)); err != nil {
		t.Fatalf("writing to tmp file %q should succeed, got: %v", tmpFile.Name(), err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("closing tmp file %q should succeed, got: %v", tmpFile.Name(), err)
	}

	if _, err := util.HelmActionConfig("foo", tmpFile.Name()); err != nil {
		t.Fatalf("creating helm action config from valid kubeconfig file should succeed, got: %v", err)
	}
}

func TestHelmActionConfigFromInvalidKubeconfigFile(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "lokoctl-tests-")
	if err != nil {
		t.Fatalf("creating tmp file should succeed, got: %v", err)
	}

	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("failed to remove tmp file %q: %v", tmpFile.Name(), err)
		}
	}()

	if _, err := tmpFile.Write([]byte("foo")); err != nil {
		t.Fatalf("writing to tmp file %q should succeed, got: %v", tmpFile.Name(), err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("closing tmp file %q should succeed, got: %v", tmpFile.Name(), err)
	}

	if _, err := util.HelmActionConfig("foo", tmpFile.Name()); err == nil {
		t.Fatalf("creating helm action config from invalid kubeconfig file should fail")
	}
}
