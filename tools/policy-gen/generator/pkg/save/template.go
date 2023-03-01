package save

import (
	"bytes"
	"go/format"
	"os"
	"text/template"
)

func GoTemplate(tmpl *template.Template, data any, outPath string) error {
	outBuf := bytes.Buffer{}
	if err := tmpl.Execute(&outBuf, data); err != nil {
		return err
	}

	out, err := format.Source(outBuf.Bytes())
	if err != nil {
		return err
	}

	return os.WriteFile(outPath, out, 0o600)
}

func PlainTemplate(tmpl *template.Template, data any, outPath string) error {
	outBuf := bytes.Buffer{}
	if err := tmpl.Execute(&outBuf, data); err != nil {
		return err
	}

	return os.WriteFile(outPath, outBuf.Bytes(), 0o600)
}
