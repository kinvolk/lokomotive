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

package contour

var template = map[string]string{
	"02-service-envoy.yaml": `
---
apiVersion: v1
kind: Service
metadata:
  name: envoy
  namespace: projectcontour
  annotations:
    # This annotation puts the AWS ELB into "TCP" mode so that it does not
    # do HTTP negotiation for HTTPS connections at the ELB edge.
    # The downside of this is the remote IP address of all connections will
    # appear to be the internal address of the ELB. See docs/proxy-proto.md
    # for information about enabling the PROXY protocol on the ELB to recover
    # the original remote IP address.
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: tcp
    {{- if .IngressHosts }}
    external-dns.alpha.kubernetes.io/hostname: "{{ .IngressHostsRaw }}"
    {{- end }}
spec:
  externalTrafficPolicy: Local
  ports:
  - port: 80
    name: http
    protocol: TCP
  - port: 443
    name: https
    protocol: TCP
  selector:
    app: envoy
  type: LoadBalancer
`,
}
