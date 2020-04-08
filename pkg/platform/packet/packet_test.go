package packet

import "testing"

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
		nodeRole       string
		resDefault     string
		reservationIDs map[string]string
	}{
		{
			// Test validates worker config.
			nodeRole: "worker",
			reservationIDs: map[string]string{
				"worker-2": "",
				"worker-3": "",
			},
		},
		{
			// Test validates controller config.
			nodeRole: "controller",
			reservationIDs: map[string]string{
				"controller-2": "",
				"controller-3": "",
			},
		},
		{
			// Test works if resDefault is set and no
			// reservationIDs.
			nodeRole:   "controller",
			resDefault: "next-available",
		},
	}

	for _, tc := range cases {
		if d := checkEachReservation(tc.reservationIDs, tc.resDefault, tc.nodeRole, ""); d.HasErrors() {
			t.Errorf("Should not fail with valid input: %v", tc)
		}
	}
}

func TestCheckEachReservationInvalidInput(t *testing.T) {
	cases := []struct {
		nodeRole       string
		resDefault     string
		reservationIDs map[string]string
	}{
		{
			// Test if nodeRole is worker, reservation should be
			// "worker-" not "controller".
			nodeRole: "worker",
			reservationIDs: map[string]string{
				"controller-1": "",
			},
		},
		{
			// Idem previous but vice-versa.
			nodeRole: "controller",
			reservationIDs: map[string]string{
				"worker-3": "",
			},
		},
		{
			// Test if resDefault is set to next-available,
			// reservationIDs should be empty.
			nodeRole:   "worker",
			resDefault: "next-available",
			reservationIDs: map[string]string{
				"worker-3": "",
			},
		},
		{
			// Test reservationIDs should never be set to
			// "next-available".
			nodeRole: "worker",
			reservationIDs: map[string]string{
				"worker-3": "next-available",
			},
		},
	}

	for _, tc := range cases {
		if d := checkEachReservation(tc.reservationIDs, tc.resDefault, tc.nodeRole, ""); !d.HasErrors() {
			t.Errorf("No error detected in invalid input: %v", tc)
		}
	}
}
