package components

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	k8serrs "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"

	"github.com/kinvolk/lokoctl/pkg/tar"
)

// findTabsAndNewlines checks if the line starts with tab and newline or just newline.
var findTabsAndNewlines = regexp.MustCompile("^(?:[\t ]*(?:\r?\n|\r))+")

type component struct {
	Name string

	archiveName string
}

func newComponent(archiveName string) *component {
	return &component{
		Name:        strings.TrimSuffix(archiveName, ".tar.gz"),
		archiveName: archiveName,
	}
}

func (c *component) String() string {
	return c.Name
}

// Install extracts the helm chart from binary, renders it as Kubernetes configs
// and then installs it one by one
func (c *component) Install(kubeconfig string, opts *InstallOptions) error {
	chart, err := c.loadHelmChart()
	if err != nil {
		return err
	}

	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			Name:      "lokoctl",
			Namespace: opts.Namespace,
		},
	}

	renderedFiles, err := renderutil.Render(chart, chart.Values, renderOpts)
	if err != nil {
		return errors.Wrap(err, "failed to render chart")
	}

	return orderedInstall(kubeconfig, renderedFiles)
}

// loadHelmChart extracts the chart that is stored in binary as tar into a
// temporary directory and reads it into memory using helm libraries and returns
// the helm chart object
func (c *component) loadHelmChart() (*chart.Chart, error) {
	tarFile, err := Asset("manifests/" + c.archiveName)
	if err != nil {
		return nil, err
	}

	dir, err := ioutil.TempDir("", "lokoctl")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	tarFileReader := bytes.NewReader(tarFile)
	if err := tar.Untar(tarFileReader, dir); err != nil {
		return nil, errors.Wrapf(err,
			"failed to extract archive %s at %s", c.archiveName, dir)
	}

	chartPath := path.Join(dir, "manifests", c.Name)

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

// InstallOptions is a way of passing the data from cmd line to code here.
type InstallOptions struct {
	AnswersFile string
}

func init() {
	chartComponents, err := AssetDir("manifests")
	if err != nil {
		log.WithFields(log.Fields{
			"command": "lokoctl component",
		}).Fatalf("Failed to retrive assets: %q", err)
	}

	for _, name := range chartComponents {
		components = append(components, newComponent(name))
	}
}

// orderedInstall takes kubernetes config as map of [filenames: filecontents]
// and then feeds it to Kubernetes in following order:
// 1. Namespaces
// 2. CRDs
// 3. Rest of the Kubernetes config types
// TODO: @surajssd, currently the code execs and calls the kubectl binary
// directly to install those manifests to the kubernetes cluster, change it to
// use the libraries from client-go.
func orderedInstall(kubeconfig string, files map[string]string) error {
	namespaces := make(map[string]string)
	crds := make(map[string]string)
	others := make(map[string]string)

	// segregate the configs according to their types
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

		objKind, err := findKind(fileContent)
		if err != nil {
			return errors.Wrap(
				err, "failed to parse Kubernetes object kind from template")
		}

		switch objKind {
		case "Namespace":
			namespaces[filename] = fileContent
		case "CustomResourceDefinition":
			crds[filename] = fileContent
		default:
			others[filename] = fileContent
		}
	}

	kubeconfigArg := fmt.Sprintf("--kubeconfig=%s", kubeconfig)
	args := []string{kubeconfigArg, "apply", "-f", "-"}
	var errs []error

	installer := func(files map[string]string) {
		for _, fileContent := range files {
			if err := execKubectlApply(fileContent, args); err != nil {
				errs = append(errs, err)
			}
		}
	}

	// First install namespaces followed by crds and then anything else.
	installer(namespaces)
	installer(crds)
	installer(others)

	if len(errs) > 0 {
		return k8serrs.NewAggregate(errs)
	}
	return nil
}

// findKind parses a Kubernetes manifest and returns its 'kind'
func findKind(manifest string) (string, error) {
	k := struct {
		Kind string `json:"kind"`
	}{}
	if err := yaml.Unmarshal([]byte(manifest), &k); err != nil {
		return "", err
	}
	return k.Kind, nil
}

// execKubectlApply takes kubernetes manifest as data and kubectl command line
// args to apply configuration to the cluster.
func execKubectlApply(data string, args []string) error {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return errors.Wrap(err, "kubectl not found in PATH")
	}

	cmd := exec.Command("kubectl", args...)

	// Read from STDIN.
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "can't get STDIN pipe for kubectl")
	}

	// Write to STDIN.
	go func() {
		defer stdin.Close()
		if _, err := io.WriteString(stdin, data); err != nil {
			fmt.Printf("can't write to STDIN pipe for kubectl %v\n", err)
		}
	}()

	// Execute the actual command, out will have the kubectl output errored or
	// passed. The err just stores what exit code the process failed with.
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Display the error to the user that was returned by kubectl.
		fmt.Printf("%s", string(out))
		return errors.Wrap(err, "failed to execute command")
	}

	// Display the output to the user that was returned by kubectl.
	fmt.Printf("%s", string(out))
	return nil
}
