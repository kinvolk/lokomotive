package contour

const envoyServiceTmpl = `
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
`
