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

package helm

import (
	"sort"

	"helm.sh/helm/v3/pkg/release"
)

type releaseByVersionDesc []*release.Release

func (rs releaseByVersionDesc) Len() int           { return len(rs) }
func (rs releaseByVersionDesc) Less(i, j int) bool { return rs[i].Version < rs[j].Version }
func (rs releaseByVersionDesc) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

// HistoryClient allows mocking for tests.
type HistoryClient interface {
	Run(name string) ([]*release.Release, error)
}

// GetHistory returns at most max elements of the Helm history in descending
// version order, that is, the first element returned is the newest version of
// the release.
//
// See
// https://github.com/helm/helm/blob/041ce5a2c17a58be0fcd5f5e16fb3e7e95fea622/cmd/helm/history.go#L115-L135.
// When helm exposes this in the API we can get rid of this.
func GetHistory(client HistoryClient, name string, max int) ([]*release.Release, error) {
	history, err := client.Run(name)
	if err != nil {
		return nil, err
	}

	sort.Sort(sort.Reverse(releaseByVersionDesc(history)))

	if max > len(history) {
		max = len(history)
	}

	return history[:max], nil
}
