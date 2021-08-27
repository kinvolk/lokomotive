// Copyright 2021 The Lokomotive Authors
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

// +build packet baremetal
// +build disruptivee2e

package kubernetes_test

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/kinvolk/lokomotive/cli/cmd/cluster"
	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestCertificateRotate(t *testing.T) { //nolint:funlen
	contextLogger := log.WithFields(log.Fields{
		"command": "TestCertificateRotate",
	})

	oldConfig := testutil.BuildKubeConfig(t)

	oldCerts := fetchCertificatesForEndpoint(t, oldConfig.Host)

	cfgPath := testutil.LokocfgPath(t)
	dir, _ := path.Split(cfgPath)

	// We will be changing directories for relative paths in lokocfg to function.
	// This will revert it back when this test exits.
	testDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getting current work directory failed: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(testDir); err != nil {
			t.Fatalf("Failed to change back %q: %v", testDir, err)
		}
	})

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Changing work directory failed: %v", err)
	}

	options := cluster.CertificateRotateOptions{
		Confirm:    true,
		Verbose:    true,
		ConfigPath: cfgPath,
		ValuesPath: path.Join(dir, "lokocfg.vars"),
	}

	if err := cluster.RotateCertificates(contextLogger, options); err != nil {
		t.Fatalf("Rotating Certificates failed: %v", err)
	}

	newConfig := testutil.BuildKubeConfig(t)

	if bytes.Equal(oldConfig.CAData, newConfig.CAData) {
		t.Fatal("CA Certificate is identical after rotation")
	}

	client := testutil.CreateKubeClient(t)

	// Test if all kubelets are running.
	namespace := "kube-system"
	daemonset := "kubelet"

	t.Run("leaves_all_self_hosted_kubelets_running", func(t *testing.T) {
		testutil.WaitForDaemonSet(t, client, namespace, daemonset, testutil.RetryInterval, testutil.Timeout)

		newCerts := fetchCertificatesForEndpoint(t, oldConfig.Host)

		// Currently Lokomotive doesn't use intermediate CAs, thus we can safely assume the first one is the leaf cert.
		if oldCerts[0].NotAfter.After(newCerts[0].NotAfter) || oldCerts[0].NotAfter.Equal(newCerts[0].NotAfter) {
			t.Fatalf("Incorrect not after time on new certificate: %v is equal or before %v",
				newCerts[0].NotAfter, oldCerts[0].NotAfter)
		}
	})
}

func fetchCertificatesForEndpoint(t *testing.T, url string) []*x509.Certificate {
	if !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	/* #nosec G402*/
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url) //nolint:noctx
	if err != nil {
		t.Fatalf("Getting url %q: %v", url, err)
	}

	t.Cleanup(func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("Failed to close response body: %v", err)
		}
	})

	if len(resp.TLS.PeerCertificates) < 1 {
		t.Fatalf("No certificates offered in TLS handshake")
	}

	return resp.TLS.PeerCertificates
}
