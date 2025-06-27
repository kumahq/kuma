package install

import (
	"bytes"
	"io"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/util/data"
)

type templateFilter interface {
	Filter(name string) bool
}

func renderFilesWithFilter(templates []data.File, args interface{}, newRenderer func(data.File) (templateRenderer, error), filter templateFilter) ([]data.File, error) {
	renderedFiles := make([]data.File, len(templates))

	for i, template := range templates {
		if !filter.Filter(template.FullPath) {
			continue
		}
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

// Template filters

type ExcludePrefixesFilter struct {
	Prefixes []string
}

func (f ExcludePrefixesFilter) Filter(name string) bool {
	for _, prefix := range f.Prefixes {
		if len(name) > len(prefix) && name[:len(prefix)] == prefix {
			return false
		}
	}
	return true
}

type NoneFilter struct{}

func (f NoneFilter) Filter(name string) bool {
	return true
}
