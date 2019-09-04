package k8sutil

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Adapted from https://github.com/kubernetes-incubator/bootkube/blob/83d32756c6b02c26cab1de3f03b57f06ae4339a7/pkg/bootkube/create.go

const (
	crdRolloutDuration = 1 * time.Second
	crdRolloutTimeout  = 2 * time.Minute
)

// createAssets takes manifest files as map of [filenames: filecontents]
// and then feeds it to Kubernetes in following order:
// 1. Namespaces
// 2. CRDs
// 3. Rest of the Kubernetes config types
// It also takes a k8s ClientConfig and a timeout duration determining wait time for API server
// It returns an error if any.
func CreateAssets(config clientcmd.ClientConfig, manifestFiles map[string]string, timeout time.Duration) error {
	c, err := config.ClientConfig()
	if err != nil {
		return err
	}
	creater, err := newCreater(c)
	if err != nil {
		return err
	}

	m, err := loadManifests(manifestFiles)
	if err != nil {
		return errors.Wrapf(err, "error loading manifests")
	}

	upFn := func() (bool, error) {
		if err := apiTest(config); err != nil {
			fmt.Printf("Unable to determine api-server readiness: %v\n", err)
			return false, nil
		}
		return true, nil
	}

	fmt.Println("Waiting for api-server...")
	if err := wait.PollImmediate(5*time.Second, timeout, upFn); err != nil {
		return errors.Wrapf(err, "API Server is not ready")
	}

	fmt.Println("Creating assets...")
	if ok := creater.createManifests(m); !ok {
		return fmt.Errorf("some assets could not be created")
	}

	return nil
}

func apiTest(c clientcmd.ClientConfig) error {
	config, err := c.ClientConfig()
	if err != nil {
		return err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// API Server is responding
	healthStatus := 0
	client.Discovery().RESTClient().Get().AbsPath("/healthz").Do().StatusCode(&healthStatus)
	if healthStatus != http.StatusOK {
		return fmt.Errorf("API Server http status: %d", healthStatus)
	}

	// System namespace has been created
	_, err = client.CoreV1().Namespaces().Get("kube-system", metav1.GetOptions{})
	return err
}

type manifest struct {
	kind       string
	apiVersion string
	namespace  string
	name       string
	raw        []byte

	filepath string
}

func (m manifest) String() string {
	if m.namespace == "" {
		return fmt.Sprintf("%s %s %s", m.filepath, m.kind, m.name)
	}
	return fmt.Sprintf("%s %s %s/%s", m.filepath, m.kind, m.namespace, m.name)
}

type creater struct {
	client *rest.RESTClient

	// mapper maps resource kinds ("ConfigMap") with their pluralized URL
	// path ("configmaps") using the discovery APIs.
	mapper *resourceMapper
}

func newCreater(c *rest.Config) (*creater, error) {
	c.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	client, err := rest.UnversionedRESTClientFor(c)
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(c)
	if err != nil {
		return nil, err
	}

	return &creater{
		mapper: newResourceMapper(discoveryClient),
		client: client,
	}, nil
}

func (c *creater) createManifests(manifests []manifest) (ok bool) {
	ok = true
	// Bootkube used to create manifests in named order ("01-foo" before "02-foo").
	// Maintain this behavior for everything except CRDs and NSs, which have strict ordering
	// that we should always respect.
	sort.Slice(manifests, func(i, j int) bool {
		return manifests[i].filepath < manifests[j].filepath
	})

	var namespaces, crds, other []manifest
	for _, m := range manifests {
		if m.kind == "CustomResourceDefinition" && strings.HasPrefix(m.apiVersion, "apiextensions.k8s.io/") {
			crds = append(crds, m)
		} else if m.kind == "Namespace" && m.apiVersion == "v1" {
			namespaces = append(namespaces, m)
		} else {
			other = append(other, m)
		}
	}

	create := func(m manifest) error {
		if err := c.create(m); err != nil {
			ok = false
			fmt.Printf("Failed creating %s: %v\n", m, err)
			return err
		}
		fmt.Printf("Created %s\n", m)
		return nil
	}

	// Create all namespaces first
	for _, m := range namespaces {
		if err := create(m); err != nil {
			return false
		}
	}

	// Create the custom resource definition before creating the actual custom resources.
	for _, m := range crds {
		if err := create(m); err != nil {
			return false
		}
	}

	// Wait until the API server registers the CRDs. Until then it's not safe to create the
	// manifests for those custom resources.
	for _, crd := range crds {
		if err := c.waitForCRD(crd); err != nil {
			ok = false
			fmt.Printf("Failed waiting for %s: %v\n", crd, err)
			return false
		}
	}

	for _, m := range other {
		if err := create(m); err != nil {
			return false
		}
	}
	return ok
}

// waitForCRD blocks until the API server begins serving the custom resource this
// manifest defines. This is determined by listing the custom resource in a loop.
func (c *creater) waitForCRD(m manifest) error {
	var crd apiextensionsv1beta1.CustomResourceDefinition
	if err := json.Unmarshal(m.raw, &crd); err != nil {
		return errors.Wrapf(err, "failed to unmarshal manifest")
	}

	// get first served version
	firstVer := ""
	if len(crd.Spec.Versions) > 0 {
		for _, v := range crd.Spec.Versions {
			if v.Served {
				firstVer = v.Name
				break
			}
		}
	} else {
		firstVer = crd.Spec.Version
	}
	if len(firstVer) == 0 {
		return fmt.Errorf("expected at least one served version")
	}

	return wait.PollImmediate(crdRolloutDuration, crdRolloutTimeout, func() (bool, error) {
		// get all resources, giving a 200 result with empty list on success, 404 before the CRD is active.
		namespaceLessURI := allCustomResourcesURI(schema.GroupVersionResource{Group: crd.Spec.Group, Version: firstVer, Resource: crd.Spec.Names.Plural})
		res := c.client.Get().RequestURI(namespaceLessURI).Do()
		if res.Error() != nil {
			if k8serrors.IsNotFound(res.Error()) {
				return false, nil
			}
			return false, res.Error()
		}
		return true, nil
	})
}

// allCustomResourcesURI returns the URI for the CRD resource without a namespace, listing
// all objects of that GroupVersionResource.
func allCustomResourcesURI(gvr schema.GroupVersionResource) string {
	return fmt.Sprintf("/apis/%s/%s/%s",
		strings.ToLower(gvr.Group),
		strings.ToLower(gvr.Version),
		strings.ToLower(gvr.Resource),
	)
}

func (c *creater) create(m manifest) error {
	info, err := c.mapper.resourceInfo(m.apiVersion, m.kind)
	if err != nil {
		return fmt.Errorf("discovery failed: %v", err)
	}

	return c.client.Post().
		AbsPath(m.urlPath(info.Name, info.Namespaced)).
		Body(m.raw).
		SetHeader("Content-Type", "application/json").
		Do().Error()
}

func (m manifest) urlPath(plural string, namespaced bool) string {
	u := "/apis"
	if m.apiVersion == "v1" {
		u = "/api"
	}
	u = u + "/" + m.apiVersion
	// NOTE(ericchiang): Some of our non-namespaced manifests have a "namespace" field.
	// Since kubectl create accepts this, also accept this.
	if m.namespace != "" && namespaced {
		u = u + "/namespaces/" + m.namespace
	}
	return u + "/" + plural
}

// loadManifests parses a map of Kubernetes manifest.
func loadManifests(files map[string]string) ([]manifest, error) {
	var manifests []manifest
	for path, fileContent := range files {
		r := strings.NewReader(fileContent)
		ms, err := parseManifests(r)
		if err != nil {
			return nil, errors.Wrapf(err, "error parsing file %s:", path)
		}
		manifests = append(manifests, ms...)
	}
	return manifests, nil
}

// parseManifests parses a YAML or JSON document that may contain one or more
// kubernetes resources.
func parseManifests(r io.Reader) ([]manifest, error) {
	reader := yaml.NewYAMLReader(bufio.NewReader(r))
	var manifests []manifest
	for {
		yamlManifest, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				return manifests, nil
			}
			return nil, err
		}
		yamlManifest = bytes.TrimSpace(yamlManifest)
		if len(yamlManifest) == 0 {
			continue
		}

		jsonManifest, err := yaml.ToJSON(yamlManifest)
		if err != nil {
			return nil, fmt.Errorf("invalid manifest: %v", err)
		}
		m, err := parseJSONManifest(jsonManifest)
		if err != nil {
			return nil, fmt.Errorf("parse manifest: %v", err)
		}
		manifests = append(manifests, m...)
	}
}

// parseJSONManifest parses a single JSON Kubernetes resource.
func parseJSONManifest(data []byte) ([]manifest, error) {
	if string(data) == "null" {
		return nil, nil
	}
	var m struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Metadata   struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, errors.Wrapf(err, "failed to parse manifest")
	}

	// We continue if the object we received was a *List kind. Otherwise if a
	// single object is received we just return from here.
	if !strings.HasSuffix(m.Kind, "List") {
		return []manifest{{
			kind:       m.Kind,
			apiVersion: m.APIVersion,
			namespace:  m.Metadata.Namespace,
			name:       m.Metadata.Name,
			raw:        data,
		}}, nil
	}

	// We parse the list of items and extract one object at a time
	var mList struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Metadata   struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(data, &mList); err != nil {
		return nil, errors.Wrapf(err, "failed to parse manifest list")
	}
	var manifests []manifest
	for _, item := range mList.Items {
		// make a recursive call, since this is a single object it will be
		// parsed and returned to us
		mn, err := parseJSONManifest(item)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, mn...)
	}
	return manifests, nil
}

func newResourceMapper(d discovery.DiscoveryInterface) *resourceMapper {
	return &resourceMapper{d, sync.Mutex{}, make(map[string]*metav1.APIResourceList)}
}

// resourceMapper uses the Kubernetes discovery APIs to map a resource Kind to its pluralized
// name to construct a URL path. For example, "ClusterRole" would be converted to "clusterroles".
//
// A note from where this code was drafted from:
// Could not get discovery.DeferredDiscoveryRESTMapper working. This implements the same logic.
//
// To Do: Maybe we can try using discovery.DeferredDiscoveryRESTMapper to see if there is still an issue.
type resourceMapper struct {
	discoveryClient discovery.DiscoveryInterface

	mu    sync.Mutex
	cache map[string]*metav1.APIResourceList
}

// resourceInfo uses the API server discovery APIs to determine the resource definition
// of a given Kind.
func (m *resourceMapper) resourceInfo(groupVersion, kind string) (*metav1.APIResource, error) {
	m.mu.Lock()
	l, ok := m.cache[groupVersion]
	m.mu.Unlock()

	if ok {
		// Check cache.
		for _, r := range l.APIResources {
			if r.Kind == kind {
				return &r, nil
			}
		}
	}

	l, err := m.discoveryClient.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to discover group version %s", groupVersion)
	}

	m.mu.Lock()
	m.cache[groupVersion] = l
	m.mu.Unlock()

	for _, r := range l.APIResources {
		if r.Kind == kind {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("resource %s %s not found", groupVersion, kind)
}
