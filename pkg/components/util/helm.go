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

	"github.com/kinvolk/lokoctl/pkg/assets"
	"github.com/kinvolk/lokoctl/pkg/util/walkers"
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
	walk := walkers.CopyingWalker(tmpDir, 0700)
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
	yaml.Unmarshal([]byte(strings.TrimSpace(values)), &result)

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

		// The helm charts that are rendered may be empty according to the
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
