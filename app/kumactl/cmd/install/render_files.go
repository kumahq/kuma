package install

import (
	"bytes"
	"io"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"

	"github.com/Kong/kuma/app/kumactl/pkg/install/data"
)

func renderFiles(templates []data.File, args interface{}, newRenderer func(data.File) (templateRenderer, error)) ([]data.File, error) {
	renderedFiles := make([]data.File, len(templates))

	for i, template := range templates {
		renderer, err := newRenderer(template)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if err := renderer.Execute(&buf, args); err != nil {
			return nil, err
		}
		renderedFiles[i].Data = buf.Bytes()
	}

	return renderedFiles, nil
}

type templateRenderer interface {
	Execute(w io.Writer, data interface{}) error
}

func simpleTemplateRenderer(text data.File) (templateRenderer, error) {
	tmpl, err := template.New("").Funcs(sprig.TxtFuncMap()).Parse(string(text.Data))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse k8s resource template")
	}
	return tmpl, nil
}
