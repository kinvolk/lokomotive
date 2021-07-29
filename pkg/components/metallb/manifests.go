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

package metallb

var chartValuesTmpl = `
speaker:
  {{if .SpeakerNodeSelectors}}
  nodeSelector:
    {{- range $key, $value := .SpeakerNodeSelectors }}
    {{ $key }}: "{{ $value }}"
    {{- end }}
  {{end}}

  {{if .SpeakerTolerations}}
  tolerations: {{.SpeakerTolerationsJSON}}
  {{end}}

controller:
  {{if .ControllerNodeSelectors}}
  nodeSelector:
	{{- range $key, $value := .ControllerNodeSelectors }}
    {{ $key }}: "{{ $value }}"
    {{- end }}
  {{end}}

  {{if .ControllerTolerations}}
  tolerations: {{.ControllerTolerationsJSON}}
  {{end}}

enableMonitoring: {{ .ServiceMonitor }}

addressPools:
{{- range $k, $v := .AddressPools }}
- name: {{ $k }}
  protocol: bgp
  addresses:
  {{- range $a := $v }}
  - {{ $a }}
  {{- end }}
{{- end }}
`
