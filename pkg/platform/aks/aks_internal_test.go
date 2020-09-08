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

package aks

import "testing"

func TestRenderRootModule(t *testing.T) {
	tests := []struct {
		desc    string
		config  *Config
		wantErr bool
	}{
		{
			desc: "One worker pool",
			config: &Config{
				WorkerPools: []workerPool{
					{
						Name:   "foo",
						VMSize: "bar",
						Count:  1,
					},
				},
			},
		},
		{
			desc: "Multiple worker pools",
			config: &Config{
				WorkerPools: []workerPool{
					{
						Name:   "foo",
						VMSize: "fake",
						Count:  1,
					},
					{
						Name:   "bar",
						VMSize: "fake",
						Count:  3,
					},
					{
						Name:   "baz",
						VMSize: "fake",
						Count:  5,
					},
				},
			},
		},
		{
			desc:    "No worker pools",
			config:  &Config{},
			wantErr: true,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			_, err := renderRootModule(test.config)
			if err != nil && !test.wantErr {
				t.Fatalf("Unexpected error: %v", err)
			}

			if err == nil && test.wantErr {
				t.Fatal("Expected an error but got none")
			}
		})
	}
}
