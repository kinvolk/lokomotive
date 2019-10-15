package local

var backendConfigTmpl = `
{{- if .Path }}
backend "local" {
  path = "{{ .Path }}"
}
{{- end }}
`
