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
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/pkg/errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/kinvolk/lokomotive/pkg/assets"
	"github.com/kinvolk/lokomotive/pkg/components"
	"github.com/kinvolk/lokomotive/pkg/k8sutil"
)

// LoadChartFromAssets takes in an asset location and returns a Helm
// Chart object or an error.
func LoadChartFromAssets(location string) (*chart.Chart, error) {
	tmpDir, err := ioutil.TempDir("", "lokoctl-chart-")
	if err != nil {
		return nil, errors.Wrap(err, "creating temporary dir")
	}
	defer os.RemoveAll(tmpDir)

	// Rendered files could contain secret data, only allow the
	// current user but not others
	walk := assets.CopyingWalker(tmpDir, 0700)
	if err := assets.Assets.WalkFiles(location, walk); err != nil {
		return nil, errors.Wrap(err, "walking assets")
	}

	return loader.Load(tmpDir)
}

// RenderChart renders a Helm chart with the given name, namespace and values
// and either returns a map of manifest files or an error.
func RenderChart(helmChart *chart.Chart, name, namespace, values string) (map[string]string, error) {
	actionConfig := new(action.Configuration)

	install := action.NewInstall(actionConfig)

	if err := helmChart.Validate(); err != nil {
		return nil, fmt.Errorf("chart is invalid: %w", err)
	}

	install.ReleaseName = name
	install.Namespace = namespace
	install.DryRun = true
	install.ClientOnly = true // avoid contacting k8s api server

	result := make(map[string]interface{})

	if err := yaml.Unmarshal([]byte(strings.TrimSpace(values)), &result); err != nil {
		return nil, fmt.Errorf("failed decoding values: %w", err)
	}

	template, err := install.Run(helmChart, result)
	if err != nil {
		return nil, fmt.Errorf("installing chart failed: %w", err)
	}

	ret := SplitManifests(template.Manifest)

	// Include hooks
	for _, m := range template.Hooks {
		ret[m.Path] = m.Manifest
	}

	// CRD are not rendered, so do this manually
	for _, crd := range helmChart.CRDs() {
		ret[crd.Name] = string(crd.Data)
	}

	return filterOutUnusedFiles(ret), nil
}

var sep = regexp.MustCompile("(?:^|\\s*\n)---\\s*")

// SplitManifests splits a YAML manifest into smaller YAML manifest.
// Name is read from the Source: <filename> annotation in
// https://github.com/helm/helm/blob/456eb7f4118a635427bd43daa3d7aabf29304f13/pkg/action/install.go#L468.
func SplitManifests(bigFile string) map[string]string {
	var name string
	res := map[string]string{}
	// Making sure that any extra whitespace in YAML stream doesn't interfere in splitting documents correctly.
	bigFileTmp := strings.TrimSpace(bigFile)
	docs := sep.Split(bigFileTmp, -1)
	for _, d := range docs {
		if d == "" {
			continue
		}

		d = strings.TrimSpace(d)

		fmt.Sscanf(d, "# Source: %s\n", &name)
		if _, ok := res[name]; ok {
			res[name] += "---\n" + d + "\n"
		} else {
			res[name] = d + "\n"
		}
	}
	return res
}

var regexpLeadingTabsAndNewlines = regexp.MustCompile("^(?:[\t ]*(?:\r?\n|\r))+")

// filterOutUnusedFiles removes all files from the map that are either
// unused (not needed for the installation on Kubernetes) or empty.
func filterOutUnusedFiles(files map[string]string) map[string]string {
	ret := make(map[string]string)
	for filename, fileContent := range files {
		// We are only interested in Kubernetes manifests here that typically
		// have a yaml, yml or json suffix. Ignore all other files.
		if !(strings.HasSuffix(filename, "yaml") ||
			strings.HasSuffix(filename, "yml") ||
			strings.HasSuffix(filename, "json")) {
			continue
		}

		// The Helm charts that are rendered may be empty according to the
		// conditionals in the templates and with the used values. Thus
		// check if the file contains more than emptiness.
		fileContent = regexpLeadingTabsAndNewlines.ReplaceAllString(fileContent, "")
		if len(fileContent) == 0 {
			continue
		}
		ret[filename] = fileContent
	}
	return ret
}

// chartFromComponent creates Helm chart object in memory for given component and makes
// sure it is valid.
func chartFromComponent(c components.Component) (*chart.Chart, error) {
	m, err := c.RenderManifests()
	if err != nil {
		return nil, fmt.Errorf("rendering manifests failed: %w", err)
	}

	ch, err := chartFromManifests(c.Metadata().Name, m)
	if err != nil {
		return nil, fmt.Errorf("creating chart from manifests failed: %w", err)
	}

	if err := ch.Validate(); err != nil {
		return nil, fmt.Errorf("created chart from manifests is invalid: %w", err)
	}

	return ch, nil
}

// chartFromManifests creates Helm chart object in memory from given manifests.
func chartFromManifests(name string, manifests map[string]string) (*chart.Chart, error) {
	ch := &chart.Chart{
		Metadata: &chart.Metadata{
			APIVersion: chart.APIVersionV2,
			Name:       name,
			// TODO Remove hardcode version, which is installed.
			Version: "0.1.0",
		},
	}

	crds := ""

	for p, m := range manifests {
		manifestMap := map[string]string{}
		manifestMap[p] = m
		parsedManifest, err := k8sutil.LoadManifests(manifestMap)
		if err != nil {
			return nil, err
		}

		manifestsRaw := ""

		for _, pm := range parsedManifest {
			// CRDs will be installed separately.
			if pm.Kind() == "CustomResourceDefinition" {
				crds = fmt.Sprintf("%s\n---\n%s", crds, pm.Raw())
				continue
			}

			// Drop Namespace resource as we take care of its creation at another level and we don't want resources to collide.
			// TODO: Remove only the namespace in which the chart is installed.
			if pm.Kind() == "Namespace" {
				continue
			}

			manifestsRaw = fmt.Sprintf("%s\n---\n%s", manifestsRaw, pm.Raw())
		}

		f := &chart.File{
			Data: []byte(manifestsRaw),
			Name: p,
		}

		// Apply rendered manifests to Manifests slice, which does not run through the rendering engine
		// again when the chart is being installed. This is required, as some charts use complex escaping
		// syntax, which breaks if the templates are evaluated twice. This, for example, breaks
		// the prometheus-operator chart.
		ch.Manifests = append(ch.Manifests, f)
	}

	// If we collected any CRDs, put them in the special file in the dedicated crds/ directory.
	if crds != "" {
		f := &chart.File{
			Data: []byte(crds),
			Name: "crds/crds.yaml",
		}

		ch.Files = append(ch.Files, f)
	}

	return ch, nil
}
