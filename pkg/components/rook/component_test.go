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

package rook //nolint:testpackage

import (
	"testing"
)

func TestConvertNodeSelector(t *testing.T) {
	m := map[string]string{
		"key1": "val1",
		"key3": "val3",
		"key4": "val4",
		"key2": "val2",
	}

	expectedOutput := "key1=val1; key2=val2; key3=val3; key4=val4; "
	actualOutput := convertNodeSelector(m)

	if expectedOutput != actualOutput {
		t.Fatalf("expected: %s, got: %s", expectedOutput, actualOutput)
	}
}
