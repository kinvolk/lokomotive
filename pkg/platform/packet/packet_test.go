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

package packet

import (
	"reflect"
	"testing"
)

func TestCheckNotEmptyWorkersEmpty(t *testing.T) {
	c := config{}

	if d := c.checkNotEmptyWorkers(); !d.HasErrors() {
		t.Errorf("Expected to fail with empty workers")
	}
}

func TestCheckNotEmptyWorkers(t *testing.T) {
	c := config{WorkerPools: []workerPool{{Name: "test"}}}

	if d := c.checkNotEmptyWorkers(); d.HasErrors() {
		t.Errorf("Should not fail with no duplicated worker pool names")
	}
}

func TestCheckWorkerPoolNamesUniqueDup(t *testing.T) {
	c := config{
		WorkerPools: []workerPool{
			{
				Name: "dup",
			},
			{
				Name: "dup",
			},
		},
	}

	if d := c.checkWorkerPoolNamesUnique(); !d.HasErrors() {
		t.Error("Should fail with duplicated worker pool names")
	}
}

func TestCheckWorkerPoolNamesUniqueNotDup(t *testing.T) {
	c := config{
		WorkerPools: []workerPool{
			{
				Name: "not",
			},
			{
				Name: "dup",
			},
		},
	}

	if d := c.checkWorkerPoolNamesUnique(); d.HasErrors() {
		t.Error("Should work with no duplicated worker pool names")
	}
}

//nolint: funlen
func TestValidateOSVersion(t *testing.T) {
	type testCase struct {
		// Config to test
		cfg config
		// Expected output after running test
		hasError bool
	}

	cases := []testCase{
		{
			cfg: config{
				ClusterName: "c",
				WorkerPools: []workerPool{
					{
						Name:      "1",
						OSVersion: "current",
					},
				},
			},
			hasError: true,
		},
		{
			cfg: config{
				ClusterName: "c",
				OSVersion:   "current",
				WorkerPools: []workerPool{
					{
						Name: "2",
					},
				},
			},
			hasError: true,
		},
		{
			cfg: config{
				ClusterName: "c",
				WorkerPools: []workerPool{
					{
						Name: "3",
					},
				},
			},
			hasError: false,
		},
		{
			cfg: config{
				ClusterName: "c",
				WorkerPools: []workerPool{
					{
						Name:          "4",
						OSVersion:     "current",
						IPXEScriptURL: "https://demo.version",
					},
				},
			},
			hasError: false,
		},
		{
			cfg: config{
				ClusterName:   "c",
				OSVersion:     "current",
				IPXEScriptURL: "https://demo.version",
				WorkerPools: []workerPool{
					{
						Name: "5",
					},
				},
			},
			hasError: false,
		},
	}

	for tcIdx, tc := range cases {
		output := tc.cfg.validateOSVersion()
		if output.HasErrors() != tc.hasError {
			t.Errorf("In test %v, expected %v, got %v", tcIdx+1, tc.hasError, output.HasErrors())
		}
	}
}

func TestCheckResFormatPrefixValidInput(t *testing.T) {
	r := map[string]string{"worker-2": "", "worker-3": ""}

	if d := checkResFormat(r, "", "", "worker"); d.HasErrors() {
		t.Errorf("Validation failed and shouldn't for: %v", r)
	}
}

func TestCheckResFormatPrefixInvalidInput(t *testing.T) {
	r := map[string]string{"broken-1": "", "worker-1-2": "", "worker-a": ""}

	if d := checkResFormat(r, "", "", "worker"); !d.HasErrors() {
		t.Errorf("Should fail with res: %v", r)
	}
}

func TestCheckEachReservationValidInput(t *testing.T) {
	cases := []struct {
		role           nodeRole
		resDefault     string
		reservationIDs map[string]string
	}{
		{
			// Test validates worker config.
			role: worker,
			reservationIDs: map[string]string{
				"worker-2": "",
				"worker-3": "",
			},
		},
		{
			// Test validates controller config.
			role: controller,
			reservationIDs: map[string]string{
				"controller-2": "",
				"controller-3": "",
			},
		},
		{
			// Test works if resDefault is set and no
			// reservationIDs.
			role:       controller,
			resDefault: "next-available",
		},
	}

	for _, tc := range cases {
		if d := checkEachReservation(tc.reservationIDs, tc.resDefault, "", tc.role); d.HasErrors() {
			t.Errorf("Should not fail with valid input: %v", tc)
		}
	}
}

func TestCheckEachReservationInvalidInput(t *testing.T) {
	cases := []struct {
		role           nodeRole
		resDefault     string
		reservationIDs map[string]string
	}{
		{
			// Test if nodeRole is worker, reservation should be
			// "worker-" not "controller".
			role: worker,
			reservationIDs: map[string]string{
				"controller-1": "",
			},
		},
		{
			// Idem previous but vice-versa.
			role: controller,
			reservationIDs: map[string]string{
				"worker-3": "",
			},
		},
		{
			// Test if resDefault is set to next-available,
			// reservationIDs should be empty.
			role:       worker,
			resDefault: "next-available",
			reservationIDs: map[string]string{
				"worker-3": "",
			},
		},
		{
			// Test reservationIDs should never be set to
			// "next-available".
			role: worker,
			reservationIDs: map[string]string{
				"worker-3": "next-available",
			},
		},
	}

	for _, tc := range cases {
		if d := checkEachReservation(tc.reservationIDs, tc.resDefault, "", tc.role); !d.HasErrors() {
			t.Errorf("No error detected in invalid input: %v", tc)
		}
	}
}

func baseConfig() *config {
	return &config{
		ClusterName: "c",
		WorkerPools: []workerPool{
			{
				Name: "1",
			},
			{
				Name: "2",
			},
		},
	}
}

//nolint: funlen
func TestTerraformAddDeps(t *testing.T) {
	cases := map[string]struct {
		configF         func(*config)
		expectedConfigF func(*config)
	}{
		"worker pool without reservation IDs depends on worker pool with reservation ID": {
			func(c *config) {
				c.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "dummy"}
			},
			func(c *config) {
				c.WorkerPools[1].NodesDependOn = []string{poolTarget("1", "worker_nodes_ids")}
			},
		},
		"all worker pools without reservation IDs depends on worker pool with reservation ID": {
			func(c *config) {
				c.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "dummy"}
				c.WorkerPools = append(c.WorkerPools, workerPool{Name: "3"})
			},
			func(c *config) {
				c.WorkerPools[1].NodesDependOn = []string{poolTarget("1", "worker_nodes_ids")}
			},
		},
		"worker pool without reservation IDs depends on all worker pools with reservation IDs": {
			func(c *config) {
				c.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "dummy"}
				c.WorkerPools[1].ReservationIDs = map[string]string{"worker-0": "dummy"}
				c.WorkerPools = append(c.WorkerPools, workerPool{Name: "3"})
			},
			func(c *config) {
				c.WorkerPools[2].NodesDependOn = []string{
					poolTarget("1", "worker_nodes_ids"),
					poolTarget("2", "worker_nodes_ids"),
				}
			},
		},
		"worker pools without reservation IDs depends on controller nodes with reservation IDs": {
			func(c *config) {
				c.ReservationIDs = map[string]string{"controller-0": "dummy"}
			},
			func(c *config) {
				c.WorkerPools[0].NodesDependOn = []string{clusterTarget("1", "device_ids")}
				c.WorkerPools[1].NodesDependOn = []string{clusterTarget("1", "device_ids")}
			},
		},
		"worker pool without reservation IDs depends on controller nodes and worker pools with reservation IDs": {
			func(c *config) {
				c.ReservationIDs = map[string]string{"controller-0": "dummy"}
				c.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "dummy"}
			},
			func(c *config) {
				c.WorkerPools[1].NodesDependOn = []string{clusterTarget("1", "device_ids"), poolTarget("1", "device_ids")}
			},
		},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			// Create copy of base config.
			config := baseConfig()

			// Mutate it.
			c.configF(config)

			// Copy mutated config.
			expectedConfig := config

			// Mutate to expected config.
			c.expectedConfigF(expectedConfig)

			// Add dependencies.
			config.terraformAddDeps()

			for i, workerPool := range config.WorkerPools {
				dependencies := workerPool.NodesDependOn
				expectedDependencies := expectedConfig.WorkerPools[i].NodesDependOn

				if !reflect.DeepEqual(dependencies, expectedDependencies) {
					t.Fatalf("Expected %v, got %v", expectedDependencies, dependencies)
				}
			}
		})
	}
}
