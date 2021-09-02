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
    tag: v0.9.6-86016b74
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
    # This image is based on the branch https://github.com/kinvolk/metallb/tree/imran/multiple-peers-patch
    # which in turn is based on top of https://github.com/metallb/metallb/pull/593.
    # Commit: https://github.com/kinvolk/metallb/commit/86016b748a520f45403fb81abd380e65b0c39f27

    # The reason for using this image is mentioned in the commit message above and also an issue is opened
    # in the Cloud Provider Equinix Metal for public discussion.
    # In any case, the direction of the public discussion would determine to either continue using this
    # custom image or not.

    # During upgrade of MetalLB or when the base PR #593 is merged, this commit must be cherry-picked to
    # avoid regression.
    tag: v0.9.6-86016b74
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
    from-annotations:
    - my-asn: metal.equinix.com/node-asn
      peer-address: metal.equinix.com/peer-ip
      peer-asn: metal.equinix.com/peer-asn
      source-address: metal.equinix.com/src-ip
      peer-port: metallb.lokomotive.io/peer-port
      hold-time: metallb.lokomotive.io/hold-time
      router-id: metallb.lokomotive.io/router-id
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
