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

// +build aws packet
// +build poste2e

package monitoring

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func testComponentsPrometheusMetrics(t *testing.T, v1api v1.API) {
	testCases := []struct {
		componentName string
		query         string
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
		},
		{
			componentName: "kube-controller-manager",
			query:         "workqueue_work_duration_seconds_bucket",
		},
		{
			componentName: "kube-proxy",
			query:         "kubeproxy_sync_proxy_rules_duration_seconds_bucket",
		},
		{
			componentName: "kubelet",
			query:         "kubelet_running_pod_count",
		},
		{
			componentName: "calico-felix",
			query:         "felix_active_local_endpoints",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("prometheus-%s", tc.componentName), func(t *testing.T) {
			t.Logf("querying %q", tc.query)

			const contextTimeout = 10

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeout*time.Second)
			defer cancel()

			results, warnings, err := v1api.Query(ctx, tc.query, time.Now())
			if err != nil {
				t.Errorf("error querying Prometheus: %v", err)
				return
			}

			if len(warnings) > 0 {
				t.Logf("warnings: %v", warnings)
			}
			t.Logf("found %d results for %s", len(strings.Split(results.String(), "\n")), tc.query)
		})
	}
}
