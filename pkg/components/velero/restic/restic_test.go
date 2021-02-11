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

package restic_test

import (
	"testing"

	"github.com/kinvolk/lokomotive/pkg/components/util"
	"github.com/kinvolk/lokomotive/pkg/components/velero/restic"
)

//nolint:funlen
func TestResticConfig(t *testing.T) {
	tests := []struct {
		desc      string
		config    *restic.Configuration
		wantError bool
	}{
		{
			desc:      "effect should be NoExecute if TolerationSeconds is Set",
			wantError: true,
			config: &restic.Configuration{
				Credentials: "foo",
				BackupStorageLocation: &restic.BackupStorageLocation{
					Bucket:   "mybucket",
					Provider: "aws",
					Region:   "myregion",
				},
				Tolerations: []util.Toleration{
					{
						Key:               "key",
						Value:             "value",
						Effect:            "NoSchedule",
						Operator:          "Equal",
						TolerationSeconds: 3,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		diags := tc.config.Validate()
		if !tc.wantError && diags.HasErrors() {
			t.Fatalf("Valid config should not return error, got: %s", diags.Error())
		}
	}
}
