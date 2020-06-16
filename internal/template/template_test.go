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

package template_test

import (
	"testing"

	"github.com/kinvolk/lokomotive/internal/template"
)

type testRenderer struct {
	Test string
}

func TestRenderSuccess(t *testing.T) {
	tmpl := `Rendered template is: {{ .Test }}`
	expected := "Rendered template is: Success"

	tr := &testRenderer{Test: "Success"}

	output, err := template.Render(tmpl, tr)
	if err != nil {
		t.Fatalf("error rendering template not expected, got: %q", err)
	}

	if output != expected {
		t.Fatalf("expected: %s, got: %s", expected, output)
	}
}

func TestRenderFail(t *testing.T) {
	tmpl := `Rendered template is: {{ .UnknownField }}`
	expected := "Rendered template is: Success"

	tr := &testRenderer{Test: "Fail"}

	output, err := template.Render(tmpl, tr)
	if err == nil {
		t.Fatalf("expected error in rendering template")
	}

	if output != "" {
		t.Fatalf("expected: %s, got: %s", expected, output)
	}
}
