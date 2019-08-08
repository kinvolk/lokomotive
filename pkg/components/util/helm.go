package util

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"

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

	return chartutil.Load(tmpDir)
}

// RenderChart renders a Helm chart with the given chartConfig and releaseOptions
// and either returns an ordered map of chart files or an error.
func RenderChart(helmChart *chart.Chart, chartConfig *chart.Config, releaseOptions *chartutil.ReleaseOptions) (map[string]string, error) {
	renderOpts := renderutil.Options{
		ReleaseOptions: *releaseOptions,
	}
	renderedFiles, err := renderutil.Render(helmChart, chartConfig, renderOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render chart")
	}
	return filterOutUnusedFiles(renderedFiles), nil
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
