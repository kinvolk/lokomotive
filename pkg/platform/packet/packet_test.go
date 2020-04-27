package packet

import (
	"sort"
	"testing"
)

func TestConfigNewconfig(t *testing.T) {
	c := NewConfig()

	if c.OSArch != "amd64" {
		t.Fatalf("Expected default Arch to be `amd64`, got: %s", c.OSArch)
	}

	if c.ControllerType != "baremetal_0" {
		t.Fatalf("Expected default controller Type to be `baremetal_0`, got: %s", c.ControllerType)
	}

	if c.ClusterDomainSuffix != "cluster.local" {
		t.Fatalf("Expected default ClusterDomainSufix as `cluster.local`, got: %s", c.ClusterDomainSuffix)
	}

	if c.CertsValidityPeriodHours != 8760 {
		t.Fatalf("Expected default CertsValidityPeriodHours as `8760`, got: %d", c.CertsValidityPeriodHours)
	}

	if !c.EnableAggregation {
		t.Fatalf("Expected default EnableAggreation to be true, got: %t", c.EnableAggregation)
	}

	if c.ControllerCount != 1 {
		t.Fatalf("Expected default ControllerCount to be 1, got: %d", c.ControllerCount)
	}

	if c.OSChannel != "stable" {
		t.Fatalf("Expected default channel as `stable`, got: %s", c.OSChannel)
	}

	if c.OSVersion != "current" {
		t.Fatalf("Expected default version as `current`, got: %s", c.OSVersion)
	}

	if c.NetworkMTU != 1480 {
		t.Fatalf("Expected default NetworkMTU as `1480`, got: %d", c.NetworkMTU)
	}

	if c.PodCIDR != "10.2.0.0/16" {
		t.Fatalf("Expected default PodCIDR as `10.2.0.0/16`, got: %s", c.PodCIDR)
	}

	if c.ServiceCIDR != "10.3.0.0/16" {
		t.Fatalf("Expected default ServiceCIDR as `10.3.0.0/16`, got: %s", c.ServiceCIDR)
	}

	if c.EnableReporting {
		t.Fatalf("Expected default EnableReporting as false, got: %t", c.EnableReporting)
	}
}

func TestConfigValidateSuccess(t *testing.T) {
	p := &config{
		AssetDir:                 "test-asset-dir",
		AuthToken:                "test-token",
		ProjectID:                "test-project-id",
		Facility:                 "ams1",
		OSArch:                   "amd64",
		OSVersion:                "current",
		OSChannel:                "stable",
		ClusterName:              "test",
		SSHPubKeys:               []string{"ssh-key"},
		PodCIDR:                  "10.2.0.0/16",
		ServiceCIDR:              "10.3.0.0/16",
		NodePrivateCIDR:          "10.0.0.0/16",
		ManagementCIDRs:          []string{"0.0.0.0/16"},
		ControllerType:           "baremetal_0",
		ClusterDomainSuffix:      "test.local",
		ControllerCount:          3,
		CertsValidityPeriodHours: 1000,
		NetworkMTU:               2000,
		WorkerPools: []workerPool{
			{
				Name:  "pool",
				Count: 3,
			},
		},
	}

	diags := p.Validate()
	if diags.HasErrors() {
		for _, diag := range diags {
			t.Error(diag)
		}

		t.Fatalf("Expected no errors in validating configuration, got: %s", diags.Error())
	}
}

func TestConfigValidateFail(t *testing.T) {
	p := &config{
		AuthToken:                "test-token",
		AssetDir:                 "",
		ProjectID:                "",
		Facility:                 "",
		OSArch:                   "qwe",
		OSVersion:                "current",
		OSChannel:                "asd",
		ClusterName:              "test",
		CertsValidityPeriodHours: 1000,
		NetworkMTU:               2000,
		SSHPubKeys:               []string{"ssh-key"},
		PodCIDR:                  "C.2.0.0/16",
		ServiceCIDR:              "10.X.0.0/16",
		NodePrivateCIDR:          "10.0.0.0/16",
		ManagementCIDRs:          []string{"0.0.0.0/16"},
		ControllerType:           "",
		ClusterDomainSuffix:      "test.local",
		ControllerCount:          1,
		WorkerPools: []workerPool{
			{
				Name:  "pool",
				Count: 3,
			},
		},
	}

	diags := p.Validate()
	if !diags.HasErrors() {
		t.Fatalf("Expected errors in validating configuration")
	}
}

func TestRenderSuccess(t *testing.T) {
	p := &config{
		AssetDir:                 "test-asset-dir",
		AuthToken:                "test-token",
		ProjectID:                "test-project-id",
		Facility:                 "ams1",
		OSArch:                   "amd64",
		OSVersion:                "current",
		OSChannel:                "stable",
		ClusterName:              "test",
		CertsValidityPeriodHours: 1000,
		NetworkMTU:               2000,
		SSHPubKeys:               []string{"ssh-key"},
		PodCIDR:                  "10.2.0.0/16",
		ServiceCIDR:              "10.3.0.0/16",
		NodePrivateCIDR:          "10.0.0.0/16",
		ManagementCIDRs:          []string{"0.0.0.0/16"},
		ControllerType:           "baremetal_0",
		ClusterDomainSuffix:      "test.local",
		ControllerCount:          3,
		WorkerPools: []workerPool{
			{
				Name:  "pool",
				Count: 3,
			},
		},
	}

	renderedTemplate, err := p.Render()
	if err != nil {
		t.Fatalf("Expected render to succeed, got: %v", err)
	}

	if renderedTemplate == "" {
		t.Fatalf("Expected rendered string to be non-empty")
	}
}

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

//nolint: funlen
func TestTerraformAddDeps(t *testing.T) {
	type testCase struct {
		// Config to test
		cfg config
		// Expected config after running the test
		exp config
	}

	var cases []testCase

	base := config{
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

	// Test WorkerPool w/o res depends on 1 WorkerPool with res
	test := base
	test.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "dummy"}
	exp := test
	exp.WorkerPools[1].NodesDependOn = []string{poolTarget("1", "worker_nodes_ids")}

	cases = append(cases, testCase{test, exp})

	// Test 2 WorkerPools w/o res depend on 2 WorkerPool with rest
	test = base
	test.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "dummy"}
	test.WorkerPools = append(test.WorkerPools, workerPool{Name: "3"})
	exp = test
	exp.WorkerPools[1].NodesDependOn = []string{poolTarget("1", "worker_nodes_ids")}

	cases = append(cases, testCase{test, exp})

	// Test 1 WorkerPools w/o res depend on 2 WorkerPool with res
	test = base
	test.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "dummy"}
	test.WorkerPools[1].ReservationIDs = map[string]string{"worker-0": "dummy"}
	test.WorkerPools = append(test.WorkerPools, workerPool{Name: "3"})
	exp = test
	exp.WorkerPools[2].NodesDependOn = []string{
		poolTarget("1", "worker_nodes_ids"),
		poolTarget("2", "worker_nodes_ids"),
	}

	cases = append(cases, testCase{test, exp})

	// Test 2 WorkerPools w/o res depend on controllers
	test = base
	test.ReservationIDs = map[string]string{"controller-0": "dummy"}
	exp = test
	exp.WorkerPools[0].NodesDependOn = []string{clusterTarget("1", "device_ids")}
	exp.WorkerPools[1].NodesDependOn = []string{clusterTarget("1", "device_ids")}

	cases = append(cases, testCase{test, exp})

	// Test 1 WorkerPools w/o res depends on controllers and WorkerPool with
	// res
	test = base
	test.ReservationIDs = map[string]string{"controller-0": "dummy"}
	test.WorkerPools[0].ReservationIDs = map[string]string{"worker-0": "dummy"}
	exp = test
	exp.WorkerPools[1].NodesDependOn = []string{clusterTarget("1", "device_ids"), poolTarget("1", "device_ids")}

	cases = append(cases, testCase{test, exp})

	for tcIdx, tc := range cases {
		test := tc.cfg
		exp := tc.exp

		test.terraformAddDeps()

		for i, w := range test.WorkerPools {
			ret := w.NodesDependOn
			expRet := exp.WorkerPools[i].NodesDependOn

			if equal := cmpSliceString(ret, expRet); !equal {
				t.Errorf("In test %v, expected %v, got %v", tcIdx, expRet, ret)
			}
		}
	}
}

func cmpSliceString(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for index, elem := range a {
		if elem != b[index] {
			return false
		}
	}

	return true
}
