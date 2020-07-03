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

// +build aws aws_edge packet aks
// +build poste2e

package monitoring

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	testutil "github.com/kinvolk/lokomotive/test/components/util"
)

//nolint:funlen
func testComponentsPrometheusMetrics(t *testing.T, v1api v1.API) {
	selfHostedPlatforms := []testutil.Platform{
		testutil.PlatformPacket,
		testutil.PlatformAWS,
	}

	testCases := []struct {
		componentName string
		query         string
		platforms     []testutil.Platform
	}{
		{
			componentName: "kube-apiserver",
			query:         "apiserver_request_total",
		},
		{
			componentName: "coredns",
			query:         "coredns_build_info",
		},
		{
			componentName: "kube-scheduler",
			query:         "scheduler_schedule_attempts_total",
			platforms:     selfHostedPlatforms,
		},
		{
			componentName: "kube-controller-manager",
			query:         "workqueue_work_duration_seconds_bucket",
			platforms:     selfHostedPlatforms,
		},
		{
			componentName: "kube-proxy",
			query:         "kubeproxy_sync_proxy_rules_duration_seconds_bucket",
			platforms:     selfHostedPlatforms,
		},
		{
			componentName: "kubelet",
			query:         "kubelet_running_pod_count",
			platforms:     selfHostedPlatforms,
		},
		{
			componentName: "etcd",
			query:         "etcd_server_has_leader",
			platforms:     selfHostedPlatforms,
		},
		{
			componentName: "metallb",
			query:         "metallb_bgp_session_up",
			platforms:     []testutil.Platform{testutil.PlatformPacket},
		},
		{
			componentName: "contour",
			query:         "contour_dagrebuild_timestamp",
			platforms:     []testutil.Platform{testutil.PlatformPacket, testutil.PlatformAWS, testutil.PlatformAKS},
		},
		{
			componentName: "cert-manager",
			query:         "certmanager_controller_sync_call_count",
			platforms:     []testutil.Platform{testutil.PlatformPacket, testutil.PlatformAWS, testutil.PlatformAKS},
		},
		{
			componentName: "experimental-istio-operator",
			query:         "pilot_k8s_reg_events",
			platforms:     []testutil.Platform{testutil.PlatformPacket, testutil.PlatformAWS, testutil.PlatformAKS},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("prometheus-%s", tc.componentName), func(t *testing.T) {
			if !testutil.IsPlatformSupported(t, tc.platforms) {
				t.Skip()
			}

			t.Parallel()

			t.Logf("querying %q", tc.query)

			if err := wait.PollImmediate(retryInterval, timeout, getMetricRetryFunc(t, v1api, tc.query)); err != nil {
				t.Errorf("%v", err)
			}
		})
	}
}

// getMetricRetryFunc returns a function which can be passed to wait.PollImmediate
// checking if a given Prometheus query returns any result.
func getMetricRetryFunc(t *testing.T, v1api v1.API, query string) wait.ConditionFunc {
	return func() (done bool, err error) {
		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout*time.Second)
		defer cancel()

		results, warnings, err := v1api.Query(ctx, query, time.Now())
		if err != nil {
			t.Logf("error querying Prometheus for metric %q: %v", query, err)

			return false, nil
		}

		if len(warnings) > 0 {
			t.Logf("warnings: %v", warnings)
		}

		if len(results.String()) == 0 {
			t.Logf("no metrics found for query %q", query)

			return false, nil
		}

		t.Logf("found %d results for %q", len(strings.Split(results.String(), "\n")), query)

		return true, nil
	}
}
