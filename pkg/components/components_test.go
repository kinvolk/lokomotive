package components

import (
	"reflect"
	"testing"
)

func Test_cleanConfigs(t *testing.T) {

	tests := []struct {
		name string
		args map[string]string
		want map[string]string
	}{
		{
			name: "one config eliminated with different extension",
			args: map[string]string{
				"a.txt":  "a: foo",
				"b.yaml": "b: bar",
				"c.yml":  "c: baz",
				"d.json": "d: json",
			},
			want: map[string]string{
				"b.yaml": "b: bar",
				"c.yml":  "c: baz",
				"d.json": "d: json",
			},
		},
		{
			name: "empty content files are removed",
			args: map[string]string{
				"a.yaml": "",
				"b.yaml": `
`,
				"c.yaml": "yaml: yes",
			},
			want: map[string]string{
				"c.yaml": "yaml: yes",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanConfigs(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cleanConfigs() = %v, want %v", got, tt.want)
			}
		})
	}
}
