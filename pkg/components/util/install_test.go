package util_test

import (
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
	if _, err := util.HelmActionConfig("foo", []byte(validKubeconfig)); err != nil {
		t.Fatalf("creating helm action config from valid kubeconfig file should succeed, got: %v", err)
	}
}

func TestHelmActionConfigFromInvalidKubeconfigFile(t *testing.T) {
	if _, err := util.HelmActionConfig("foo", []byte("foo")); err == nil {
		t.Fatalf("creating helm action config from invalid kubeconfig file should fail")
	}
}
