package certmanager

var values = `
# Namespace in which this nginx-ingress be deployed
namespace: {{.Namespace}}
email: {{.Email}}
`
