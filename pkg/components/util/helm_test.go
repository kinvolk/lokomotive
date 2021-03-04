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

package util_test

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"helm.sh/helm/v3/pkg/chart"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/components/util"
)

func TestRenderChartBadValues(t *testing.T) {
	c := "cert-manager"
	values := "malformed\t"

	helmChart, err := components.Chart(c)
	if err != nil {
		t.Fatalf("Loading chart from assets should succeed, got: %v", err)
	}

	if _, err := util.RenderChart(helmChart, c, c, values); err == nil {
		t.Fatalf("Rendering chart with malformed values should fail")
	}
}

//nolint:funlen
func Test_RenderChart_include_multiple_hooks_from_single_file(t *testing.T) {
	values := ""
	chartName := "foo"
	fileName := "secrets.yaml"

	expectedManifest := `---
apiVersion: v1
kind: Secret
metadata:
  name: foo-controller-tls
  annotations:
    "helm.sh/resource-policy": "keep"
    "helm.sh/hook": "pre-install"
    "helm.sh/hook-delete-policy": "before-hook-creation"
type: kubernetes.io/tls
data:
  tls.crt: "[...]"
  tls.key: "[...]"
  ca.crt: "[...]"
---
apiVersion: v1
kind: Secret
metadata:
  name: foo-client-tls
  annotations:
    "helm.sh/resource-policy": "keep"
    "helm.sh/hook": "pre-install"
    "helm.sh/hook-delete-policy": "before-hook-creation"
type: kubernetes.io/tls
data:
  tls.crt: [...]
  tls.key: [...]
  ca.crt: [...]
`

	chart := &chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion: "2.0.0",
			Name:       chartName,
			Version:    "0.1.0",
		},
		Templates: []*chart.File{
			{
				Name: fileName,
				Data: []byte(expectedManifest),
			},
		},
	}

	manifests, err := util.RenderChart(chart, "chartName", "chartNamespace", values)
	if err != nil {
		t.Fatalf("Rendering chart: %v", err)
	}

	expectedKey := filepath.Join(chartName, fileName)

	manifest, ok := manifests[expectedKey]
	if !ok {
		t.Fatalf("Expected manifest not found, got %+v", manifests)
	}

	if diff := cmp.Diff(expectedManifest, manifest); diff != "" {
		t.Fatalf("Unexpected manifest diff: %v", diff)
	}
}
