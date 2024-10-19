package viewutil

import (
	"bytes"
	"html/template"
)

// Render compiles a template from text and renders it.
func Render(templateText string, model any) (template.HTML, error) {
	tmpl, err := template.New("").Parse(templateText)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, model); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
