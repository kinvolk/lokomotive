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
package clusterautoscaler

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/packethost/packngo"
)

func TestGetClusterWorkersFilterByFacility(t *testing.T) {
	device := packngo.Device{
		DeviceRaw: packngo.DeviceRaw{
			Hostname: "worker",
			Facility: &packngo.Facility{
				Code: "bar",
			},
		},
	}

	d := []packngo.Device{
		device,
		device,
		{
			DeviceRaw: packngo.DeviceRaw{
				Hostname: "worker",
				Facility: &packngo.Facility{
					Code: "doh",
				},
			},
		},
	}

	if w := getClusterWorkers("", device.Facility.Code, d); len(w) != 2 {
		t.Fatalf("got workers from other facility: %+v", w)
	}
}

func TestGetClusterWorkersFilterByCluster(t *testing.T) {
	clusterName := "baz"
	facility := &packngo.Facility{
		Code: "bar",
	}

	device := packngo.Device{
		DeviceRaw: packngo.DeviceRaw{
			Hostname: fmt.Sprintf("%s-worker", clusterName),
			Facility: facility,
		},
	}

	d := []packngo.Device{
		device,
		device,
		{
			DeviceRaw: packngo.DeviceRaw{
				Hostname: "doh-worker",
				Facility: facility,
			},
		},
	}

	if w := getClusterWorkers(clusterName, device.Facility.Code, d); len(w) != 2 {
		t.Fatalf("got workers from other cluster: %+v", w)
	}
}

func TestGetClusterWorkersFilterNonWorkers(t *testing.T) {
	facility := &packngo.Facility{
		Code: "bar",
	}

	device := packngo.Device{
		DeviceRaw: packngo.DeviceRaw{
			Hostname: "worker",
			Facility: facility,
		},
	}

	d := []packngo.Device{
		device,
		device,
		{
			DeviceRaw: packngo.DeviceRaw{
				Hostname: "controller",
				Facility: facility,
			},
		},
	}

	if w := getClusterWorkers("", device.Facility.Code, d); len(w) != 2 {
		t.Fatalf("got devices which are not workers: %+v", w)
	}
}

// Packet API seems to have a bug, that when you use Project API key for querying devices,
// it returns duplicated entries. However IDs of the devices should be the same, so it should
// be safe to relax the check on those, as if there would actually be 2 different devices with
// the same name, they should have different IDs.
func TestFindDuplicatedDevicesDuplicateHostnamesAndIDs(t *testing.T) {
	d := []packngo.Device{
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "bar",
				Hostname: "foo",
			},
		},
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "bar",
				Hostname: "foo",
			},
		},
	}

	if w := findDuplicatedDevices(d); len(w) != 0 {
		t.Fatalf("two devices with the same hostname and same IDs should not be treated as duplicates")
	}
}

func TestFindDuplicatedDevicesDuplicateHostnamesUniqueIDs(t *testing.T) {
	d := []packngo.Device{
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "baz",
				Hostname: "foo",
			},
		},
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "bar",
				Hostname: "foo",
			},
		},
	}

	if w := findDuplicatedDevices(d); len(w) == 0 {
		t.Fatalf("two devices with the same hostname but different IDs should be treated as duplicates")
	}
}

func TestDeviceHostnames(t *testing.T) {
	d := []packngo.Device{
		{
			DeviceRaw: packngo.DeviceRaw{
				Hostname: "bar",
			},
		},
		{
			DeviceRaw: packngo.DeviceRaw{
				Hostname: "foo",
			},
		},
	}

	expected := []string{"bar", "foo"}

	if hostnames := devicesHostnames(d); !reflect.DeepEqual(expected, hostnames) {
		t.Fatalf("expected: %+v, got: %+v", expected, hostnames)
	}
}

func TestGetWorkerUserdataNoUserdataOnError(t *testing.T) {
	userdata, err := getWorkerUserdata("", "", []packngo.Device{})
	if err == nil {
		t.Error("if there is no devices to get user data from, error should be returned")
	}

	if userdata != "" {
		t.Error("if error is returned, no userdata should be returned either")
	}
}

func TestGetWorkerUserdataDuplicatedWorkers(t *testing.T) {
	d := []packngo.Device{
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "foo",
				Hostname: "worker",
				Facility: &packngo.Facility{},
			},
		},
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "bar",
				Hostname: "worker",
				Facility: &packngo.Facility{},
			},
		},
	}

	if _, err := getWorkerUserdata("", "", d); err == nil {
		t.Fatalf("if devices contains duplicates, error should be returned")
	}
}

func TestGetWorkerUserdataEmptyUserdata(t *testing.T) {
	device := packngo.Device{
		DeviceRaw: packngo.DeviceRaw{
			ID:       "foo",
			Hostname: "worker",
			Facility: &packngo.Facility{},
		},
	}

	d := []packngo.Device{
		device,
	}

	if _, err := getWorkerUserdata("", "", d); err == nil {
		t.Fatalf("if no device contains userdata, error should be returned")
	}
}

func TestGetWorkerUserdataFirstDevice(t *testing.T) {
	expected := "foo"
	expectedBase64 := base64.StdEncoding.EncodeToString([]byte(expected))

	d := []packngo.Device{
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "foo",
				UserData: expected,
				Hostname: "worker",
				Facility: &packngo.Facility{},
			},
		},
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "bar",
				UserData: "bar",
				Hostname: "worker2",
				Facility: &packngo.Facility{},
			},
		},
	}

	if userData, _ := getWorkerUserdata("", "", d); userData != expectedBase64 {
		t.Fatalf("expected: %q, got: %q", expectedBase64, userData)
	}
}

func TestGetWorkerUserdataReturnBase64(t *testing.T) {
	expected := "foo"

	d := []packngo.Device{
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "foo",
				UserData: expected,
				Hostname: "worker",
				Facility: &packngo.Facility{},
			},
		},
	}

	userData, _ := getWorkerUserdata("", "", d)
	if _, err := base64.StdEncoding.DecodeString(userData); err != nil {
		t.Fatalf("returned userdata should be valid base64 encoded string, got %q: %v", userData, err)
	}
}

func TestGetWorkerUserdataDuplicatedWorkersDifferentClusters(t *testing.T) {
	facility := "bar"

	device := packngo.Device{
		DeviceRaw: packngo.DeviceRaw{
			ID:       "foo",
			UserData: "foo",
			Hostname: "worker",
			Facility: &packngo.Facility{
				Code: "foo",
			},
		},
	}

	d := []packngo.Device{
		device,
		device,
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "bar",
				UserData: "foo",
				Hostname: "worker",
				Facility: &packngo.Facility{
					Code: facility,
				},
			},
		},
	}

	if _, err := getWorkerUserdata("", facility, d); err != nil {
		t.Fatalf("should ignore duplicated workers from other clusters")
	}
}

func TestGetWorkerUserdataDuplicatedWorkersIncludeHostnames(t *testing.T) {
	facility := "bar"

	hostnameA := "worker1"
	hostnameB := "worker2"

	d := []packngo.Device{
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "foo",
				Hostname: hostnameA,
				Facility: &packngo.Facility{
					Code: facility,
				},
			},
		},
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "bar",
				Hostname: hostnameA,
				Facility: &packngo.Facility{
					Code: facility,
				},
			},
		},
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "baz",
				Hostname: hostnameB,
				Facility: &packngo.Facility{
					Code: facility,
				},
			},
		},
		{
			DeviceRaw: packngo.DeviceRaw{
				ID:       "doh",
				Hostname: hostnameA,
				Facility: &packngo.Facility{
					Code: facility,
				},
			},
		},
	}

	_, err := getWorkerUserdata("", facility, d)
	if err == nil {
		t.Fatalf("expected duplicate device error")
	}

	if !strings.Contains(err.Error(), hostnameA) && !strings.Contains(err.Error(), hostnameB) {
		t.Fatalf("error should include all duplicated hostnames, got: %q", err)
	}
}
