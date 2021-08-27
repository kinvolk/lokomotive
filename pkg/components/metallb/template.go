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

// Package metallb has code related to deployment of metallb component.
package metallb

const chartValuesTmpl = `
controller:
  enabled: true
  image:
    # This image is build on top of master which is
    # rebased on https://github.com/metallb/metallb/pull/593. Once merged, start using
    # upstream image.
    tag: v0.9.6-d8e5b333
  nodeSelector:
    "node.kubernetes.io/master": ""
  {{- with .ControllerNodeSelectors }}
  {{- range $key, $value := . }}
    {{ $key }}: "{{ $value }}"
  {{- end }}
  {{- end }}
  {{- if .ControllerTolerationsJSON }}
  tolerations: {{ .ControllerTolerationsJSON }}
  {{- end }}
speaker:
  enabled: true
  image:
    # This image is build on top of master which is
    # rebased on https://github.com/metallb/metallb/pull/593. Once merged, start using
    # upstream image.
    tag: v0.9.6-d8e5b333
  tolerateMaster: false
  {{- with .SpeakerNodeSelectors }}
  nodeSelector:
  {{- range $key, $value := . }}
    {{ $key }}: "{{ $value }}"
  {{- end }}
  {{- end }}
  {{- if .SpeakerTolerationsJSON }}
  tolerations: {{ .SpeakerTolerationsJSON }}
  {{- end }}

serviceMonitor: {{ .ServiceMonitor }}

configInline:
  peer-autodiscovery:
    from-labels:
    - my-asn: metallb.lokomotive.io/my-asn
      peer-asn: metallb.lokomotive.io/peer-asn
      peer-address: metallb.lokomotive.io/peer-address
      peer-port: metallb.lokomotive.io/peer-port
      source-address: metallb.lokomotive.io/src-address
      hold-time: metallb.lokomotive.io/hold-time
      router-id: metallb.lokomotive.io/router-id
    from-annotations:
    - my-asn: metallb.lokomotive.io/my-asn
      peer-address: metallb.lokomotive.io/peer-address
      peer-asn: metallb.lokomotive.io/peer-asn
  address-pools:
  {{- range $k, $v := .AddressPools }}
  - name: {{ $k }}
    protocol: bgp
    addresses:
    {{- range $a := $v }}
    - {{ $a }}
    {{- end }}
  {{- end }}
`
