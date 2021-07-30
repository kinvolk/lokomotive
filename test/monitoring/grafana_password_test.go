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

// +build aws aws_edge equinixmetal aks
// +build poste2e

package monitoring //nolint:testpackage

import (
	"context"
	"testing"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

// testGrafanaDefaultPassword tests if the Grafana deployment does not expose Grafana with default
// password of `prom-operator`.
func testGrafanaDefaultPassword(t *testing.T, v1api v1.API) {
	client := testutil.CreateKubeClient(t)

	const (
		namespace              = "monitoring"
		secretName             = "prometheus-operator-grafana"
		defaultGrafanaPassword = "prom-operator"
	)

	secret, err := client.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			t.Fatalf("secret not found")
		}

		t.Fatalf("could not get secret: %v", err)
	}

	data, ok := secret.Data["admin-password"]
	if !ok {
		t.Fatalf("password not found in the secret")
	}

	if string(data) == defaultGrafanaPassword {
		t.Fatalf("default password %q provided", defaultGrafanaPassword)
	}
}
