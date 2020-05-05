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

// Package template contains the utility functions that help in rendering
// the go templates.
package template

import (
	"bytes"
	"text/template"
)

// Render applies a parsed template to the specified data object
// and returns the output as string or an error.
func Render(tmpl string, obj interface{}) (string, error) {
	t, err := template.New("render").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, obj); err != nil {
		return "", err
	}
	return buf.String(), nil
}
