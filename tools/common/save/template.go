package save

import (
	"bytes"
	"go/format"
	"os"
	"path/filepath"
	"strings"
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

func PlainFileTemplate(src, dst string, data any) error {
	tmpl, err := template.
		New(filepath.Base(src)).
		Funcs(FuncMap).
		ParseFiles(src)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	return os.WriteFile(dst, buf.Bytes(), 0o600)
}

var FuncMap = template.FuncMap{
	"ternary":    ternary,
	"hasPrefix":  strings.HasPrefix,
	"hasSuffix":  strings.HasSuffix,
	"trimSuffix": strings.TrimSuffix,
}

func ternary(cond bool, a, b any) any {
	if cond {
		return a
	}
	return b
}
