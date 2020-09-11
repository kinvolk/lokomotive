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

package util

import (
	"strings"
	"testing"

	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

func TestRenderChartBadValues(t *testing.T) {
	c := "cert-manager"
	values := "malformed\t"

	helmChart, err := components.Chart(c)
	if err != nil {
		t.Fatalf("Loading chart from assets should succeed, got: %v", err)
	}

	if _, err := RenderChart(helmChart, c, c, values); err == nil {
		t.Fatalf("Rendering chart with malformed values should fail")
	}
}

func TestChartFromManifests(t *testing.T) {
	tc := []struct {
		metadata  components.Metadata
		manifests map[string]string
		err       bool
	}{
		{
			components.Metadata{
				Name: "foo",
			},
			map[string]string{
				"foo.yaml": "bar",
			},
			true,
		},
		{
			components.Metadata{
				Name: "foo",
			},
			map[string]string{
				"foo.yaml": "---\nfoo: bar",
			},
			false,
		},
	}

	for _, c := range tc {
		c := c

		t.Run("", func(t *testing.T) {
			chart, err := chartFromManifests(c.metadata, c.manifests)
			if c.err && err == nil {
				t.Fatalf("Expected error, got nil")
			}

			if !c.err && err != nil {
				t.Fatalf("Didn't expect error, got: %v", err)
			}

			if c.err && err != nil {
				return
			}

			if err := chart.Validate(); err != nil {
				t.Fatalf("Generated chart should be valid, got: %v", err)
			}
		})
	}
}

func TestChartFromManifestsRemoveNamespace(t *testing.T) {
	manifests := map[string]string{
		"namespace.yaml": `
apiVersion: v1
kind: Namespace
metadata:
  name: foo
`,
	}

	m := components.Metadata{
		Name: "foo",
		Namespace: k8sutil.Namespace{
			Name: "foo",
		},
	}

	chart, err := chartFromManifests(m, manifests)
	if err != nil {
		t.Fatalf("Chart should be created, got: %v", err)
	}

	if len(chart.Templates) != 1 {
		t.Fatalf("Manifest file with the namespace should still be added, as it may contain other objects")
	}

	if len(chart.Templates[0].Data) != 0 {
		t.Fatalf("Namespace object should be removed from chart")
	}
}

func TestChartFromManifestsRemoveNamespaceRetainObject(t *testing.T) {
	manifests := map[string]string{
		"objects.yaml": `
apiVersion: v1
kind: Namespace
metadata:
  name: foo
---
apiVersion: v1
kind: Pod
metadata:
  name: bar
`,
	}

	m := components.Metadata{
		Name: "foo",
		Namespace: k8sutil.Namespace{
			Name: "foo",
		},
	}

	chart, err := chartFromManifests(m, manifests)
	if err != nil {
		t.Fatalf("Chart should be created, got: %v", err)
	}

	if len(chart.Templates) != 1 {
		t.Fatalf("templates should include exactly one object")
	}

	if len(chart.Templates[0].Data) == 0 {
		t.Fatalf("Other objects should be retained in the file containing Namespace object")
	}
}

func TestChartFromManifestsRemoveOnlyReleaseNamespace(t *testing.T) {
	manifests := map[string]string{
		"objects.yaml": `
apiVersion: v1
kind: Namespace
metadata:
  name: foo
`,
	}

	m := components.Metadata{
		Name: "foo",
		Namespace: k8sutil.Namespace{
			Name: "bar",
		},
	}

	chart, err := chartFromManifests(m, manifests)
	if err != nil {
		t.Fatalf("Chart should be created, got: %v", err)
	}

	if len(chart.Templates[0].Data) == 0 {
		t.Fatalf("Only Namespace object with matching namespace name should be filtered")
	}
}

func TestChartFromManifestsMoveCRDs(t *testing.T) {
	manifests := map[string]string{
		"crd.yaml": `
kind: CustomResourceDefinition
metadata:
  name: foo
`,
	}

	m := components.Metadata{
		Name: "foo",
	}

	chart, err := chartFromManifests(m, manifests)
	if err != nil {
		t.Fatalf("Chart should be created, got: %v", err)
	}

	if len(chart.Templates) != 1 {
		t.Fatalf("Manifest file with the CRDs should still be added, as it may contain other objects")
	}

	if len(chart.Templates[0].Data) != 0 {
		t.Fatalf("CRD object should be removed from the manifests file")
	}

	if len(chart.Files) != 1 {
		t.Fatalf("CRD object should be added to Files field")
	}

	if len(chart.Files[0].Data) == 0 {
		t.Fatalf("CRD object in unmanaged files shouldn't be empty")
	}

	if strings.Split(chart.Files[0].Name, "/")[0] != "crds" {
		t.Fatalf("CRD object should be added to file in 'crds' directory")
	}
}
