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
package internal_test

import (
	"testing"

	"github.com/kinvolk/lokomotive/internal"
)

const namespace = "test"

func TestApppendNamespaceLabelSuccess(t *testing.T) {
	m := map[string]string{
		"key": "value",
	}

	final := internal.AppendNamespaceLabel(namespace, m)

	if !(final[internal.NamespaceLabelKey] == namespace) {
		t.Errorf("expected %s key to have the value %s", internal.NamespaceLabelKey, namespace)
	}
}

func TestApppendNamespaceLabelDontAllowOverride(t *testing.T) {
	value := "old_value"
	m := map[string]string{
		internal.NamespaceLabelKey: value,
	}

	final := internal.AppendNamespaceLabel(namespace, m)

	if final[internal.NamespaceLabelKey] == namespace {
		t.Errorf("expected %s key to have the value %s", internal.NamespaceLabelKey, value)
	}
}

func TestApppendNamespaceLabelNilMap(t *testing.T) {
	var m map[string]string

	final := internal.AppendNamespaceLabel(namespace, m)

	if !(final[internal.NamespaceLabelKey] == namespace) {
		t.Errorf("expected %s key to have the value %s", internal.NamespaceLabelKey, namespace)
	}
}

func TestApppendNamespaceLabelEmptyMap(t *testing.T) {
	m := map[string]string{}

	final := internal.AppendNamespaceLabel(namespace, m)

	if !(final[internal.NamespaceLabelKey] == namespace) {
		t.Errorf("expected %s key to have the value %s", internal.NamespaceLabelKey, namespace)
	}
}

func TestMergeMapsSuccess(t *testing.T) {
	m1 := map[string]string{
		"test": "good",
		"one":  "two",
	}

	m2 := map[string]string{
		"test": "bad",
	}

	final := internal.MergeMaps(m1, m2)

	if len(final) != 2 {
		t.Errorf("expected length of map after merging to be %d, got: %d", len(m1), len(final))
	}

	if final["test"] != "good" {
		t.Errorf("expected value of key `test` as %s, got: %s", m1["test"], final["test"])
	}
}

func TestMergeMapsNil(t *testing.T) {
	var m1 map[string]string

	var m2 map[string]string

	final := internal.MergeMaps(m1, m2)

	if final == nil {
		t.Errorf("expected map to be empty but not nil, got: %v", final)
	}
}

func TestIndent(t *testing.T) {
	type args struct {
		data   string
		indent int
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{
				data: `foo:
  bar:
  - baz`,
				indent: 2,
			},
			want: `  foo:
    bar:
    - baz`,
		},
		{
			args: args{
				data:   "singleline",
				indent: 3,
			},
			want: "   singleline",
		},
		{
			args: args{
				data: `a:
  b: foobar
`,
				indent: 4,
			},
			want: `    a:
      b: foobar
`,
		},
		{
			args: args{
				data: `[default]
aws_access_key=test_key
aws_secret_access_key=secret_key`,
				indent: 6,
			},
			want: `      [default]
      aws_access_key=test_key
      aws_secret_access_key=secret_key`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := internal.Indent(tt.args.data, tt.args.indent); got != tt.want {
				t.Errorf("indent() = \n%v\nwant =\n%v", got, tt.want)
			}
		})
	}
}
