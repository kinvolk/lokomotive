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
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	corev1typed "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const (
	// RetryInterval is time for test to retry.
	RetryInterval = time.Second * 5
	// Timeout is time after which tests stops and fails.
	Timeout = time.Minute * 5
	// TimeoutSlow is time after which tests stops and fails.
	TimeoutSlow = time.Minute * 9
)

func KubeconfigPath(t *testing.T) string {
	kubeconfig := os.ExpandEnv(os.Getenv("KUBECONFIG"))

	if kubeconfig == "" {
		t.Fatalf("env var KUBECONFIG was not set")
	}

	return kubeconfig
}

// Kubeconfig returns content of kubeconfig file defined with KUBECONFIG
// environment variable.
func Kubeconfig(t *testing.T) []byte {
	path := KubeconfigPath(t)

	k, err := ioutil.ReadFile(path) // #nosec:G304
	if err != nil {
		t.Fatalf("reading KUBECONFIG file from %q failed: %v", path, err)
	}

	return k
}

// buildKubeConfig reads the environment variable KUBECONFIG and then builds the rest client config
// object which can be either used to create kube client to talk to apiserver or to just read the
// kubeconfig data.
func buildKubeConfig(t *testing.T) *restclient.Config {
	kubeconfig := KubeconfigPath(t)

	t.Logf("using KUBECONFIG=%s", kubeconfig)

	c, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatalf("failed building rest client: %v", err)
	}

	return c
}

// CreateKubeClient returns a kubernetes client reading the KUBECONFIG environment variable.
func CreateKubeClient(t *testing.T) *kubernetes.Clientset {
	cs, err := kubernetes.NewForConfig(buildKubeConfig(t))
	if err != nil {
		t.Fatalf("failed creating new clientset: %v", err)
	}

	return cs
}

// WaitForPVCToBeBound is a helper utility function to test if given PVC reaches Bound phase in given time.
func WaitForPVCToBeBound(
	t *testing.T, client kubernetes.Interface, ns, name string, retryInterval, timeout time.Duration,
) {
	if err := wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		pvc, err := client.CoreV1().PersistentVolumeClaims(ns).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				t.Logf("waiting for PVC %q to be created", name)

				return false, nil
			}

			return false, fmt.Errorf("getting PVC %q: %w", name, err)
		}

		phase := pvc.Status.Phase

		t.Logf("PVC: %q, Phase: %q", name, phase)

		if phase == corev1.ClaimBound {
			return true, nil
		}

		return false, nil
	}); err != nil {
		t.Fatalf("waiting for the PVC to be in Bound phase: %v", err)
	}
}

func WaitForStatefulSet(t *testing.T, client kubernetes.Interface, ns, name string, replicas int, retryInterval, timeout time.Duration) {
	if err := wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		ds, err := client.AppsV1().StatefulSets(ns).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				t.Logf("waiting for statefulset %s to be available", name)
				return false, nil
			}

			return false, fmt.Errorf("getting StatefulSet %q: %w", name, err)
		}

		t.Logf("statefulset: %s, replicas: %d/%d", name, int(ds.Status.ReadyReplicas), replicas)

		if int(ds.Status.ReadyReplicas) == replicas {
			t.Logf("found required replicas")
			return true, nil
		}

		return false, nil
	}); err != nil {
		t.Errorf("error while waiting for the statefulset: %v", err)
	}
}

func WaitForDaemonSet(t *testing.T, client kubernetes.Interface, ns, name string, retryInterval, timeout time.Duration) {
	var ds *appsv1.DaemonSet

	if err := wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		ds, err = client.AppsV1().DaemonSets(ns).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				t.Logf("waiting for daemonset %s to be available", name)
				return false, nil
			}

			return false, fmt.Errorf("getting DaemonSet %q: %w", name, err)
		}
		replicas := ds.Status.DesiredNumberScheduled

		if replicas == 0 {
			t.Logf("no replicas scheduled for daemonset %s", name)

			return false, nil
		}

		t.Logf("daemonset: %s, replicas: %d/%d", name, ds.Status.DesiredNumberScheduled, replicas)
		if ds.Status.NumberReady == replicas {
			t.Logf("found required replicas")
			return true, nil
		}
		return false, nil
	}); err != nil {
		if err := PrintPodsLogs(t, client.CoreV1().Pods(ns), ds.Spec.Selector); err != nil {
			t.Logf("reading pod logs: %v", err)
		}

		t.Errorf("error while waiting for the daemonset: %v", err)
	}
}

func WaitForDeployment(t *testing.T, client kubernetes.Interface, ns, name string, retryInterval, timeout time.Duration) {
	var err error
	var deploy *appsv1.Deployment

	// Check the readiness of the Deployment
	if err = wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		deploy, err = client.AppsV1().Deployments(ns).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				t.Logf("waiting for deployment %s to be available", name)
				return false, nil
			}

			return false, fmt.Errorf("getting Deployment %q: %w", name, err)
		}

		replicas := deploy.Status.Replicas

		if replicas == 0 {
			t.Logf("no replicas scheduled for deployment %s", name)

			return false, nil
		}

		t.Logf("deployment: %s, replicas: %d/%d", name, int(deploy.Status.AvailableReplicas), replicas)

		if deploy.Status.AvailableReplicas == replicas {
			t.Logf("found required replicas")
			return true, nil
		}
		return false, nil
	}); err != nil {
		if err := PrintPodsLogs(t, client.CoreV1().Pods(ns), deploy.Spec.Selector); err != nil {
			t.Logf("reading pod logs: %v", err)
		}

		t.Fatalf("error while waiting for the deployment: %v", err)
	}

	// Check the readiness of the pods
	selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
	if err != nil {
		t.Fatalf("converting label selector to map: %v", err)
	}

	if err := wait.PollImmediate(retryInterval, timeout, func() (done bool, err error) {
		listOptions := metav1.ListOptions{LabelSelector: selector.String()}
		pods, err := client.CoreV1().Pods(ns).List(context.TODO(), listOptions)
		if err != nil {
			return false, fmt.Errorf("getting pods: %w", err)
		}
		pods = filterNonControllerPods(pods)

		// Sanity check.
		if len(pods.Items) == 0 {
			t.Fatalf("checking containers status failed. No pods selected.")
		}

		// go through each pod in the returned list and check the readiness status of it
		for _, pod := range pods.Items {
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.RestartCount > 10 {
					return false, fmt.Errorf("pod: %s, container %s; pod in CrashLoopBackOff", pod.Name, cs.Name)
				}
				if !cs.Ready {
					t.Logf("pod: %s, container %s; container not ready", pod.Name, cs.Name)
					return false, nil
				}
			}
			t.Logf("pod %s, has all containers in ready state", pod.Name)
		}
		t.Logf("all pods for deployment %s, are in ready state", deploy.Name)
		return true, nil
	}); err != nil {
		if err := PrintPodsLogs(t, client.CoreV1().Pods(ns), deploy.Spec.Selector); err != nil {
			t.Logf("reading pod logs: %v", err)
		}

		t.Errorf("error while waiting for the pods: %v", err)
	}
}

// PrintPodsLogs print logs from all pods and containers selected by given selector.
func PrintPodsLogs(t *testing.T, p corev1typed.PodInterface, selector *metav1.LabelSelector) error {
	s, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return fmt.Errorf("converting label selector to map: %w", err)
	}

	pods, err := p.List(context.TODO(), metav1.ListOptions{LabelSelector: s.String()})
	if err != nil {
		return fmt.Errorf("listing pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("no pods selected")
	}

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			podLogOpts := corev1.PodLogOptions{
				Container: container.Name,
			}

			req := p.GetLogs(pod.Name, &podLogOpts)

			podLogs, err := req.Stream(context.TODO())
			if err != nil {
				return fmt.Errorf("opening stream: %w", err)
			}

			defer func() {
				if err := podLogs.Close(); err != nil {
					t.Logf("closing stream: %v\n", err)
				}
			}()

			scanner := bufio.NewScanner(podLogs)
			for scanner.Scan() {
				t.Logf("%s/%s: %s", pod.Name, container.Name, scanner.Text())
			}
		}
	}

	return nil
}

func filterNonControllerPods(pods *corev1.PodList) *corev1.PodList {
	var filteredPods []corev1.Pod

	for _, pod := range pods.Items {
		// The pod that has a controller, has this label
		if _, ok := pod.Labels["pod-template-hash"]; !ok {
			continue
		}
		filteredPods = append(filteredPods, pod)
	}
	pods.Items = filteredPods
	return pods
}

// PortForwardInfo allows user to provide information needed to forward port from a Kubernetes Pod
// to local machine.
type PortForwardInfo struct {
	readyChan     chan struct{}
	stopChan      chan struct{}
	portForwarder *portforward.PortForwarder

	// TODO: Add support providing service name and the API figures out the pod to forward the
	// connection to. Port forwarding works with pods only.
	PodName   string
	Namespace string
	PodPort   int
	LocalPort int
}

// CloseChan closes the stopChan which essentially disables port forwarding. User should call this
// method once they are done using port forwarding. This is generally called using `defer`
// immediately after `PortFoward`.
func (p *PortForwardInfo) CloseChan() {
	// This to guard against the closed channel, if you close the closed channel it panics this
	// piece of code guards against that.
	select {
	case <-p.stopChan:
		return
	default:
	}
	close(p.stopChan)
}

// PortForward initiates the port forward in an asynchronous mode. The user is responsible to stop
// port forwarding by calling `CloseChan` method on the object. Also user should use the helper
// method to wait until port forwarding is enabled by calling `WaitUntilForwardingAvailable`. So
// the user of the this method should always call this API in following sequence for correct
// functionality:
//
// p.PortForward(t)
// defer p.CloseChan()
// p.WaitUntilForwardingAvailable(t)
//
func (p *PortForwardInfo) PortForward(t *testing.T) {
	config := buildKubeConfig(t)

	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		t.Fatalf("could not create round tripper: %v", err)
	}

	serverURL, err := url.Parse(config.Host)
	if err != nil {
		t.Fatalf("could not parse the URL from kubeconfig: %v", err)
	}

	serverURL.Path = fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", p.Namespace, p.PodName)
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, serverURL)

	ports := []string{fmt.Sprintf("0:%d", p.PodPort)}

	out, errOut := new(bytes.Buffer), new(bytes.Buffer)
	p.stopChan, p.readyChan = make(chan struct{}, 1), make(chan struct{}, 1)

	forwarder, err := portforward.New(dialer, ports, p.stopChan, p.readyChan, out, errOut)
	if err != nil {
		p.CloseChan()
		t.Fatalf("could not create forwarder: %v", err)
	}

	p.portForwarder = forwarder

	// This goroutine will print any error or output to stdout.
	go func() {
		for range p.readyChan {
		}

		t.Logf("output of port forwarder:\n%s\n", out.String())

		if len(errOut.String()) != 0 {
			p.CloseChan()
			t.Errorf(errOut.String())
		}
	}()

	go func() {
		if err := forwarder.ForwardPorts(); err != nil { // Locks until stopChan is closed.
			p.CloseChan()
			t.Errorf("could not establish port forwarding: %v", err)
		}
	}()
}

// findLocalPort finds out the local port that was randomly chosen. This is done here because when
// the port forwarding is done the local port is not known upfront. It is chosen randomly and can
// only be found once port forwarding has started.
func (p *PortForwardInfo) findLocalPort(t *testing.T) {
	forwardedPorts, err := p.portForwarder.GetPorts()
	if err != nil {
		t.Fatalf("could not get information about ports: %v", err)
	}

	const noOfForwardedPorts = 1
	if len(forwardedPorts) != noOfForwardedPorts {
		t.Fatalf("number of forwarded ports not 1, currently forwarding for %d ports.", len(forwardedPorts))
	}

	p.LocalPort = int(forwardedPorts[0].Local)
}

// WaitUntilForwardingAvailable is a blocking call which waits until the port-forwarding is made
// available.
func (p *PortForwardInfo) WaitUntilForwardingAvailable(t *testing.T) {
	const portForwardTimeout = 2

	// Wait until port forwarding is available.
	select {
	case <-p.readyChan:
	case <-time.After(portForwardTimeout * time.Minute):
		t.Fatal("timed out waiting for port forwarding")
	}
	p.findLocalPort(t)
}

// Platform is a type tests will use to specify which platform they are supported on.
type Platform string

const (
	// PlatformAWS is for AWS
	PlatformAWS = "aws"

	// PlatformAWSEdge is for AWS with FCL Edge.
	PlatformAWSEdge = "aws_edge"

	// PlatformPacket is for Packet
	PlatformPacket = "packet"

	// PlatformPacketARM is for Packet on ARM
	PlatformPacketARM = "packet_arm"

	// PlatformBaremetal is for Baremetal
	PlatformBaremetal = "baremetal"

	// PlatformAKS is for AKS.
	PlatformAKS = "aks"
)

// IsPlatformSupported takes in the test object and the list of supported platforms. The function
// detects the supported platform from environment variable. And if the platform is available in the
// supported platforms provided then this returns true otherwise false. If the supported platforms
// list is empty it is interpreted as all platforms are supported.
func IsPlatformSupported(t *testing.T, platforms []Platform) bool {
	// This means that all platforms are supported.
	if len(platforms) == 0 {
		return true
	}

	// Find out what platform we are running on.
	p := os.Getenv("PLATFORM")
	if p == "" {
		t.Fatal("env var PLATFORM was not set")
	}

	// The platform should be present in the list of supported platforms.
	for _, platform := range platforms {
		if Platform(p) == platform {
			return true
		}
	}

	return false
}

// TestNamespacePrefix is testing namespace prefix.
const TestNamespacePrefix = "test-"

// IsTestNamespace checks if namespace is a testing namespace or not.
func IsTestNamespace(name string) bool {
	regEx := regexp.MustCompile("^" + TestNamespacePrefix + ".*")

	return regEx.MatchString(name)
}

// TestNamespace creates a test namespace.
func TestNamespace(name string) string {
	return TestNamespacePrefix + name + "-"
}

// IsUserNamespace checks for user namespace.
func IsUserNamespace(ns string) bool {
	switch ns {
	case "kube-system", "kube-public", "kube-node-lease", "default", "lokomotive-system":
		return true
	}

	// Check for testing namespaces.
	return IsTestNamespace(ns)
}
