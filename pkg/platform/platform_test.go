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

package platform_test

import (
	"reflect"
	"testing"

	"github.com/kinvolk/lokomotive/pkg/platform"
)

func TestAppendVersionTagUninitializedMap(t *testing.T) {
	var f map[string]string

	platform.AppendVersionTag(&f)

	if len(f) == 0 {
		t.Fatalf("should append version tag to uninitialized map")
	}
}

func TestAppendVersionTagIgnoreNil(t *testing.T) {
	platform.AppendVersionTag(nil)
}

func TestAppendVersionTag(t *testing.T) {
	f := map[string]string{
		"foo": "bar",
	}

	platform.AppendVersionTag(&f)

	if len(f) != 2 {
		t.Fatalf("should append version tag to existing map")
	}
}

func TestCommonControlPlaneChartsOrder(t *testing.T) {
	expectedOrder := []string{
		"bootstrap-secrets",
		"pod-checkpointer",
		"kube-apiserver",
		"kubernetes",
		"calico",
		"lokomotive",
		"kubelet",
	}

	commonControlPlaneCharts := platform.CommonControlPlaneCharts(true)

	actualOrder := []string{}

	for _, v := range commonControlPlaneCharts {
		actualOrder = append(actualOrder, v.Name)
	}

	if !reflect.DeepEqual(actualOrder, expectedOrder) {
		t.Fatalf("expected order: %s, got: %s", expectedOrder, actualOrder)
	}
}
