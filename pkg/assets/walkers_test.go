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

package assets

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

const (
	tmpPattern = "lokomotive-test"
	testData   = "222"
)

func TestWriteFile(t *testing.T) {
	f, err := ioutil.TempFile("", tmpPattern)
	if err != nil {
		t.Fatalf("Creating temp file should succeed, got: %v", err)
	}

	defer os.Remove(f.Name())

	if err := writeFile(f.Name(), strings.NewReader(testData)); err != nil {
		t.Fatalf("Writing to file should succeed, got: %v", err)
	}

	d, err := ioutil.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Reading tmp file should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(testData, string(d)) {
		t.Fatalf("Expected: '%s', got '%s'", testData, string(d))
	}
}

func TestWriteFileTruncate(t *testing.T) {
	f, err := ioutil.TempFile("", tmpPattern)
	if err != nil {
		t.Fatalf("Creating temp file should succeed, got: %v", err)
	}

	defer os.Remove(f.Name())

	if err := writeFile(f.Name(), strings.NewReader("111111")); err != nil {
		t.Fatalf("Writing to file should succeed, got: %v", err)
	}

	if err := writeFile(f.Name(), strings.NewReader(testData)); err != nil {
		t.Fatalf("Updating file should succeed, got: %v", err)
	}

	d, err := ioutil.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("Reading tmp file should succeed, got: %v", err)
	}

	if !reflect.DeepEqual(testData, string(d)) {
		t.Fatalf("Expected: '%s', got '%s'", testData, string(d))
	}
}
