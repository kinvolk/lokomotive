package s3

var backendConfigTmpl = `
backend "s3" {
  bucket = "{{ .Bucket }}"
  key    = "{{ .Key }}"
  region = "{{ .Region }}"
  {{- if .AWSCredsPath }}
  shared_credentials_file = "{{ .AWSCredsPath }}"
  {{- end }}
  {{- if .DynamoDBTable }}
  dynamodb_table = "{{ .DynamoDBTable }}"
  {{- end }}
}
`
