package terraform

var backendTmpl = `
terraform { 
	{{ .Data }}
}
`
