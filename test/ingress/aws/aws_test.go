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

// +build aws
// +build e2e

package aws

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kinvolk/lokomotive/pkg/util/retryutil"
	testutil "github.com/kinvolk/lokomotive/test/components/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	retryIntervalSeconds = 5
	maxRetries           = 60
)

func TestAWSIngress(t *testing.T) {
	client := testutil.CreateKubeClient(t)

	i, err := client.NetworkingV1beta1().Ingresses("httpbin").Get("httpbin", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting httpbin ingress: %v", err)
	}

	h := i.Spec.Rules[0].Host

	err = retryutil.Retry(retryIntervalSeconds*time.Second, maxRetries, func() (bool, error) {
		resp, err := http.Get(fmt.Sprintf("https://%s/get", h))
		if err != nil {
			t.Logf("got an HTTP error: %v", err)
			return false, nil
		}

		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("closing HTTP response body: %v", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			t.Logf("got a non-OK HTTP status: %d", resp.StatusCode)
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		t.Fatal("could not get a successful HTTP response in time")
	}
}
