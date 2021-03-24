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

// +build packet baremetal aws
// +build disruptivee2e

package kubernetes_test

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/kinvolk/lokomotive/cli/cmd/cluster"
	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

func TestCertificateRotate(t *testing.T) {
	contextLogger := log.WithFields(log.Fields{
		"command": "TestCertificateRotate",
	})

	oldConfig := testutil.BuildKubeConfig(t)

	oldCerts, err := fetchCertificatesForEndpoint(oldConfig.Host)
	if err != nil {
		t.Fatalf("Calling API endpoint pre-rotation failed: %v", err)
	}

	cfgPath := testutil.LokocfgPath(t)
	dir, _ := path.Split(cfgPath)

	// We will be changing directories for relative paths in lokocfg to function.
	// This will revert it back when this test exits.
	testDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getting current work directory failed: %v", err)
	}
	defer os.Chdir(testDir)

	err = os.Chdir(dir)
	if err != nil {
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

	// test if all kubelets are running
	namespace := "kube-system"
	daemonset := "kubelet"

	testutil.WaitForDaemonSet(t, client, namespace, daemonset, testutil.RetryInterval, testutil.Timeout)

	newCerts, err := fetchCertificatesForEndpoint(oldConfig.Host)
	if err != nil {
		t.Fatalf("Calling API endpoint pre-rotation failed: %v", err)
	}

	// Currently Lokomotive doesn't use intermediate CAs, thus we can safely assume the first one is the leaf cert.
	if oldCerts[0].NotAfter.After(newCerts[0].NotAfter) || oldCerts[0].NotAfter.Equal(newCerts[0].NotAfter) {
		t.Fatalf("Incorrect not after time on new certificate: %v is equal or before %v", newCerts[0].NotAfter, oldCerts[0].NotAfter)
	}
}

func fetchCertificatesForEndpoint(url string) (cert []*x509.Certificate, err error) {
	if !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	if len(resp.TLS.PeerCertificates) < 1 {
		return nil, errors.New("No cortificates offered in TLS handshake")
	}

	return resp.TLS.PeerCertificates, nil
}
