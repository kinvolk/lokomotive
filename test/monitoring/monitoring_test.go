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
	"fmt"
	"testing"

	testutil "github.com/kinvolk/lokomotive/test/components/util"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func TestPrometheus(t *testing.T) {
	const prometheusPodPort = 9090

	p := &testutil.PortForwardInfo{
		PodName:   "prometheus-prometheus-operator-prometheus-0",
		Namespace: "monitoring",
		PodPort:   prometheusPodPort,
	}

	p.PortForward(t)
	defer p.CloseChan()
	p.WaitUntilForwardingAvailable(t)

	promClient, err := api.NewClient(api.Config{
		Address: fmt.Sprintf("http://127.0.0.1:%d", p.LocalPort),
	})
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	v1api := v1.NewAPI(promClient)

	// Add to the list all the prometheus tests that should be invoked one at a time
	tests := []struct {
		Name string
		Func func(*testing.T, v1.API)
	}{
		{
			Name: "ComponentMetrics",
			Func: testComponentsPrometheusMetrics,
		},
		{
			Name: "ComponentAlerts",
			Func: testComponentAlerts,
		},
		{
			Name: "ScrapeTargetReachability",
			Func: testScrapeTargetRechability,
		},
		{
			Name: "TestGrafanaDefaultPassword",
			Func: testGrafanaDefaultPassword,
		},
	}

	// Invoke the test functions passing them the test object and the prometheus client.
	for _, test := range tests {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			test.Func(t, v1api)
		})
	}
}
