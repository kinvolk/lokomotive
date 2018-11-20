package utils

import (
	"bytes"
	"text/template"
)

// RenderTemplate takes in the text template and object which can render
// template and returns the rendered final as string
func RenderTemplate(tmpl string, obj interface{}) (string, error) {
	t := template.New("render")
	t, err := t.Parse(tmpl)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, obj); err != nil {
		return "", err
	}
	return buf.String(), nil
}
