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

func TestGetter(t *testing.T) {
	g, err := k8sutil.NewGetter([]byte(validKubeconfig))
	if err != nil {
		t.Fatalf("Creating getter from valid kubeconfig should succeed, got: %v", err)
	}

	if _, err := g.ToDiscoveryClient(); err != nil {
		t.Errorf("Turning getter into discovery client should succeed, got: %v", err)
	}

	if _, err := g.ToRESTMapper(); err != nil {
		t.Errorf("Turning getter into REST mapper should succeed, got: %v", err)
	}

	if c := g.ToRawKubeConfigLoader(); c == nil {
		t.Errorf("Turning getter into RawKubeConfigLoader should succeed")
	}
}

func TestGetterInvalidKubeconfig(t *testing.T) {
	if _, err := k8sutil.NewGetter([]byte("foo")); err == nil {
		t.Fatalf("Creating getter from invalid kubeconfig should fail")
	}
}
