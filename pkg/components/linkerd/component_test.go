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

//nolint:testpackage
package linkerd

import "testing"

func Test_indent(t *testing.T) {
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
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := indent(tt.args.data, tt.args.indent); got != tt.want {
				t.Errorf("indent() = \n%v\nwant =\n%v", got, tt.want)
			}
		})
	}
}
