package components

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gobuffalo/packd"
	packr "github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
)

// findTabsAndNewlines checks if the line starts with tab and newline or just newline.
var findTabsAndNewlines = regexp.MustCompile("^(?:[\t ]*(?:\r?\n|\r))+")

type component struct {
	Name    string
	Answers ComponentChanger

	archiveName string
}

func newComponent(name string, obj ComponentChanger) *component {
	return &component{
		Name:    name,
		Answers: obj,

		archiveName: name + ".tar.gz",
	}
}

func (cmpChart *component) String() string {
	return cmpChart.Name
}

// Install extracts the helm chart from binary, renders it as Kubernetes configs
// and then installs it one by one
func (cmpChart *component) Install(kubeconfig string, opts *InstallOptions) error {
	renderedFiles, err := cmpChart.processChart(opts.AnswersFile, opts.Namespace)
	if err != nil {
		return err
	}

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{},
	)
	return createAssets(config, renderedFiles, 1*time.Minute)
}

func (cmpChart *component) RenderManifests(opts *InstallOptions) error {
	renderedFiles, err := cmpChart.processChart(opts.AnswersFile, opts.Namespace)
	if err != nil {
		return err
	}
	for _, v := range renderedFiles {
		v = strings.TrimSpace(v)
		// Some helm chart templates can have multiple configs starting with
		// yaml delimiter '---' so if the file already has one then don't put
		// our own delimiter.
		if !strings.HasPrefix(v, "---") {
			fmt.Println("---")
		}
		fmt.Println(v)
	}
	return nil
}

func (cmpChart *component) processChart(ansFile, namespace string) (map[string]string, error) {
	ch, err := cmpChart.loadHelmChart()
	if err != nil {
		return nil, err
	}

	// renderutil expects to get 'ReleaseOptions' for rendering. We don't need
	// any of those options now and can pass an empty object for the time being.
	// Since some operators depend on fields like `IsUpgrade` or `IsInstall`
	// https://github.com/helm/helm/blob/82d01cb3124906e97caceb967a09f2941d6a392d/pkg/chartutil/values.go#L356-L357
	// we probably have to make the release options configurable from the
	// answers file in the future.
	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			Name:      cmpChart.Name,
			IsInstall: true,
			Namespace: namespace,
		},
	}

	if ansFile != "" {
		// read answers file
		data, err := ioutil.ReadFile(ansFile)
		if err != nil {
			return nil, err
		}

		values, err := cmpChart.Answers.GetValues(data)
		if err != nil {
			return nil, err
		}
		ch.Values = &chart.Config{Raw: values}
	}

	renderedFiles, err := renderutil.Render(ch, ch.Values, renderOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render chart")
	}
	return cleanConfigs(renderedFiles), nil
}

func cleanConfigs(files map[string]string) map[string]string {
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
		// conditionals in the templates and values provided.
		//
		// Find all ocurrences of tabs and newlines and remove them
		// used for removing all the blank lines from the file.
		fileContent = findTabsAndNewlines.ReplaceAllString(fileContent, "")
		if len(fileContent) == 0 {
			continue
		}
		ret[filename] = fileContent
	}
	return ret
}

// loadHelmChart extracts the chart that is stored in binary as tar into a
// temporary directory and reads it into memory using helm libraries and returns
// the helm chart object
func (cmpChart *component) loadHelmChart() (*chart.Chart, error) {
	b := packr.New("components", "../../manifests/")

	tmpPrefix, err := ioutil.TempDir("", "lokoctl")
	if err != nil {
		return nil, errors.Wrap(err, "creating temporary dir")
	}
	defer os.RemoveAll(tmpPrefix)

	walk := func(fileName string, file packd.File) error {
		// extract the file info for permissions
		fileInfo, err := file.FileInfo()
		if err != nil {
			return errors.Wrap(err, "extracting file info")
		}

		fileName = filepath.Join(tmpPrefix, fileName)

		// make sure that the directory is created before creating file
		if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
			return errors.Wrap(err, "creating dir")
		}

		// write the content into file
		if err := ioutil.WriteFile(fileName, []byte(file.String()), fileInfo.Mode()); err != nil {
			return errors.Wrap(err, "writing file")
		}
		return nil
	}

	if err := b.WalkPrefix(cmpChart.Name, walk); err != nil {
		return nil, errors.Wrap(err, "walking the dir")
	}

	chartPath := filepath.Join(tmpPrefix, cmpChart.Name)
	return chartutil.Load(chartPath)
}

var components []*component

func List() []*component {
	return components
}

func Get(name string) (*component, error) {
	for _, c := range components {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, fmt.Errorf("component not found")
}

// Register is called by individual component packages' init function to
// register it's Answers' object
func Register(name string, obj ComponentChanger) {
	components = append(components, newComponent(name, obj))
}

// InstallOptions is a way of passing the data from cmd line to code here.
type InstallOptions struct {
	AnswersFile string
	Namespace   string
}
