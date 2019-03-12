package util

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gobuffalo/packd"
	packr "github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
)

// LoadChartFromBox takes in a packr Box and returns a Helm Chart object or an error.
func LoadChartFromBox(box *packr.Box) (*chart.Chart, error) {
	tmpDir, err := ioutil.TempDir("", "lokoctl-chart-")
	if err != nil {
		return nil, errors.Wrap(err, "creating temporary dir")
	}
	defer os.RemoveAll(tmpDir)

	walk := func(fileName string, file packd.File) error {
		fileInfo, err := file.FileInfo()
		if err != nil {
			return errors.Wrap(err, "extracting file info")
		}

		fileName = filepath.Join(tmpDir, fileName)

		// Rendered files could contain secret data,
		// only allow the current user but not others
		if err := os.MkdirAll(filepath.Dir(fileName), 0700); err != nil {
			return errors.Wrap(err, "creating dir")
		}

		targetFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, fileInfo.Mode())
		if err != nil {
			return errors.Wrap(err, "opening target file")
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, file); err != nil {
			return errors.Wrap(err, "writing file")
		}
		return nil
	}

	if err := box.WalkPrefix("", walk); err != nil {
		return nil, errors.Wrap(err, "walking box")
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
