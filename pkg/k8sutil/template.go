package k8sutil

import (
	"bytes"
	"html/template"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

func GetKubernetesObjectFromTmpl(tmplData []byte, data interface{}) (runtime.Object, error) {
	tmpl, err := template.New("tmpTmpl").Parse(string(tmplData))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	if err = tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(buf.Bytes(), nil, nil)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
