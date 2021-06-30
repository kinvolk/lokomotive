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

package helm_test

import (
	"testing"

	"github.com/kinvolk/lokomotive/pkg/helm"

	"github.com/google/go-cmp/cmp"
	"helm.sh/helm/v3/pkg/release"
)

type mockHistory []*release.Release

func newMockHistory(rs []*release.Release) mockHistory {
	return rs
}

func (m mockHistory) Run(name string) ([]*release.Release, error) {
	return m, nil
}

func TestGetHistory(t *testing.T) { //nolint:funlen
	h := newMockHistory([]*release.Release{
		{
			Name:    "testRelease",
			Version: 6,
		},
		{
			Name:    "testRelease",
			Version: 2,
		},
		{
			Name:    "testRelease",
			Version: 4,
		},
		{
			Name:    "testRelease",
			Version: 3,
		},
		{
			Name:    "testRelease",
			Version: 5,
		},
	})

	type testCase struct {
		name     string
		max      int
		expected []*release.Release
	}

	cases := []testCase{
		{
			name: "returns_at_most_given_max_elements",
			max:  2,
			expected: []*release.Release{
				{
					Name:    "testRelease",
					Version: 6,
				},
				{
					Name:    "testRelease",
					Version: 5,
				},
			},
		},
		{
			name: "returns_releases_with_descending_order_by_version",
			max:  10,
			expected: []*release.Release{
				{
					Name:    "testRelease",
					Version: 6,
				},
				{
					Name:    "testRelease",
					Version: 5,
				},
				{
					Name:    "testRelease",
					Version: 4,
				},
				{
					Name:    "testRelease",
					Version: 3,
				},
				{
					Name:    "testRelease",
					Version: 2,
				},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := helm.GetHistory(h, "testRelease", tc.max)
			if err != nil {
				t.Fatalf("Getting history: %v", err)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Fatalf("Unexpected history (-want +got)\n%s", diff)
			}
		})
	}
}
