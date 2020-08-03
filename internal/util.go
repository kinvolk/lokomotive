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

// Package internal contains the utility functions used across the codebase.
package internal

const (
	// NamespaceLabelKey acts a placeholder for the generic key name
	// `lokomotive.kinvolk.io/name`.
	// NOTE: In the subsequent versions if this value changes, it's very possible
	// that the change might not be backwards compatible.
	// In such a case we need to avoid updating this value and introduce
	// another key to ensure backwards compatibility.
	NamespaceLabelKey = "lokomotive.kinvolk.io/name"
)

// AppendNamespaceLabel appends the release namespace as value to the
// key `lokomotive.kinvolk.io/name`.
func AppendNamespaceLabel(namespace string, labels map[string]string) map[string]string {
	final := labels

	if final == nil {
		final = make(map[string]string)
	}

	if final[NamespaceLabelKey] == "" {
		final[NamespaceLabelKey] = namespace
	}

	return final
}
