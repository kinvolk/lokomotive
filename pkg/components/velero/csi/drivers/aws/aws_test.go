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

package aws_test

import (
	"testing"

	"github.com/kinvolk/lokomotive/pkg/components/velero/csi/drivers/aws"
)

func TestEmptyConfig(t *testing.T) {
	c := &aws.Configuration{}

	_, err := c.Values()
	if err == nil {
		t.Fatalf("Empty config should return error")
	}
}

func TestValuesSuccess(t *testing.T) {
	c := &aws.Configuration{
		BackupStorageLocation: &aws.BackupStorageLocation{
			Bucket: "mybucket",
			Region: "myregion",
		},
		VolumeSnapshotLocation: &aws.VolumeSnapshotLocation{
			Region: "myregion",
		},
		Credentials: "mycredentials",
	}

	values, err := c.Values()
	if err != nil {
		t.Fatalf("Valid config should not return error, got: %s", err)
	}

	if values == "" {
		t.Fatalf("Valid config should not return error, got empty")
	}
}

func TestValidateSuccess(t *testing.T) {
	c := &aws.Configuration{
		BackupStorageLocation: &aws.BackupStorageLocation{
			Bucket: "mybucket",
			Region: "myregion",
		},
		VolumeSnapshotLocation: &aws.VolumeSnapshotLocation{
			Region: "myregion",
		},
		Credentials: "mycredentials",
	}

	if diags := c.Validate(); diags.HasErrors() {
		t.Fatalf("Valid config should not return error, got: %s", diags.Error())
	}
}

func TestValidateEmptyBackupStorageLocation(t *testing.T) {
	c := &aws.Configuration{
		BackupStorageLocation: &aws.BackupStorageLocation{},
		VolumeSnapshotLocation: &aws.VolumeSnapshotLocation{
			Region: "myregion",
		},
		Credentials: "mycredentials",
	}

	if diags := c.Validate(); !diags.HasErrors() {
		t.Fatalf("Empty BackupStorageLocation should return error, got: %s", diags.Error())
	}
}

func TestValidateEmptyVolumeSnapshotLocation(t *testing.T) {
	c := &aws.Configuration{
		BackupStorageLocation: &aws.BackupStorageLocation{
			Bucket: "mybucket",
			Region: "myregion",
		},
		VolumeSnapshotLocation: &aws.VolumeSnapshotLocation{},
		Credentials:            "mycredentials",
	}

	if diags := c.Validate(); !diags.HasErrors() {
		t.Fatalf("Empty VolumeSnapshotLocation should return error, got: %s", diags.Error())
	}
}

func TestValidateEmptyCredentials(t *testing.T) {
	c := &aws.Configuration{
		BackupStorageLocation: &aws.BackupStorageLocation{
			Bucket: "mybucket",
			Region: "myregion",
		},
		VolumeSnapshotLocation: &aws.VolumeSnapshotLocation{
			Region: "myregion",
		},
	}

	if diags := c.Validate(); !diags.HasErrors() {
		t.Fatalf("Empty Credentials should return error, got: %s", diags.Error())
	}
}
