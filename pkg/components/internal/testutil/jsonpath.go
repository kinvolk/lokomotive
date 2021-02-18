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

package testutil

import (
	"reflect"
	"testing"

	"github.com/kinvolk/lokomotive/pkg/k8sutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/util/jsonpath"
)

// valFromObject takes a JSON path as a string and an object of type `unstructured.Unstructured`.
// This function returns an object of type `reflect.Value` at that JSON path.
func valFromObject(t *testing.T, jp string, obj *unstructured.Unstructured) reflect.Value {
	jPath := jsonpath.New("parse")
	if err := jPath.Parse(jp); err != nil {
		t.Fatalf("Parsing JSONPath: %v", err)
	}

	v, err := jPath.FindResults(obj.Object)
	if err != nil {
		t.Fatalf("Finding results using JSONPath in the YAML file: %v", err)
	}

	if len(v) == 0 || len(v[0]) == 0 {
		t.Fatalf("No result found")
	}

	return v[0][0]
}

// jsonPathValue extracts an object at a JSON path from a YAML config, and returns an interface
// object.
func jsonPathValue(t *testing.T, yamlConfig string, jsonPath string) interface{} {
	u, err := k8sutil.YAMLToUnstructured([]byte(yamlConfig))
	if err != nil {
		t.Fatalf("YAML to unstructured object: %v", err)
	}

	got := valFromObject(t, jsonPath, u)

	switch got.Kind() { //nolint:exhaustive
	case reflect.Interface:
		// TODO: Add type switch here for concrete types.
		return got.Interface()
	default:
		t.Fatalf("Extracted object has an unknown type: %v", got.Kind())
	}

	return nil
}

// MatchJSONPathStringValue is a helper function for component unit tests. It compares the string at
// a JSON path in a YAML config to the expected string.
func MatchJSONPathStringValue(t *testing.T, yamlConfig string, jsonPath string, expected string) {
	obj := jsonPathValue(t, yamlConfig, jsonPath)

	got, ok := obj.(string)
	if !ok {
		t.Fatalf("Value is not string: %#v", obj)
	}

	if got != expected {
		t.Fatalf("Expected: %s, Got: %s", expected, got)
	}
}

// MatchJSONPathInt64Value is a helper function for component unit tests. It compares the integer at
// a JSON path in a YAML config to the expected integer.
func MatchJSONPathInt64Value(t *testing.T, yamlConfig string, jsonPath string, expected int64) {
	obj := jsonPathValue(t, yamlConfig, jsonPath)

	got, ok := obj.(int64)
	if !ok {
		t.Fatalf("Value is not an integer: %#v", obj)
	}

	if got != expected {
		t.Fatalf("Expected: %d, Got: %d", expected, got)
	}
}
