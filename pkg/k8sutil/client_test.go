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

package k8sutil_test

import (
	"testing"

	"github.com/kinvolk/lokomotive/pkg/k8sutil"
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

func TestNewClientset(t *testing.T) {
	if _, err := k8sutil.NewClientset([]byte(validKubeconfig)); err != nil {
		t.Fatalf("Creating clientset from valid kubeconfig should succeed, got: %v", err)
	}
}

func TestNewClientsetInvalidKubeconfig(t *testing.T) {
	if _, err := k8sutil.NewClientset([]byte("foo")); err == nil {
		t.Fatalf("creating clientset from invalid kubeconfig should fail")
	}
}
