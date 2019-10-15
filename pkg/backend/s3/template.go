package s3

var backendConfigTmpl = `
backend "s3" {
  bucket = "{{ .Bucket }}"
  key    = "{{ .Key }}"
  region = "{{ .Region }}"
}
`
