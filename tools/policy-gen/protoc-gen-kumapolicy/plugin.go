package main

import (
	"bytes"
	"go/format"
	"html/template"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/kumahq/kuma/tools/resource-gen/genutils"
)

var PluginGoTemplate = template.Must(template.New("plugin-go").Parse(`
package {{ .Package }}

{{ $pkg := .Package }}

import (
	"k8s.io/apimachinery/pkg/runtime"
{{range $idx, $version := .Versions}}
	"github.com/kumahq/kuma/pkg/plugins/policies/{{ $pkg }}/k8s/{{ $version }}"
{{- end}}
)

func AddToScheme(s *runtime.Scheme) error {
{{- range $idx, $version := .Versions}}
	if err := {{ $version }}.AddToScheme(s); err != nil {
		return err
	}
{{- end}}
	return nil
}
`))

func generatePluginFile(
	gen *protogen.Plugin,
	files []*protogen.File,
) error {
	// each file represents different version of the same type, we can take the first file to figure out a type
	if len(files) == 0 {
		return nil
	}

	file := files[0]

	var infos []genutils.ResourceInfo
	for _, msg := range file.Messages {
		infos = append(infos, genutils.ToResourceInfo(msg.Desc))
	}

	if len(infos) > 1 {
		return errors.Errorf("only one Kuma resource per proto file is allowed")
	}

	info := infos[0]

	versions := []string{}
	for _, f := range files {
		versions = append(versions, string(f.GoPackageName))
	}

	outBuf := bytes.Buffer{}
	if err := PluginGoTemplate.Execute(&outBuf, struct {
		Package  string
		Versions []string
	}{
		Package:  info.KumactlSingular,
		Versions: versions,
	}); err != nil {
		return err
	}

	pg, err := format.Source(outBuf.Bytes())
	if err != nil {
		return err
	}

	gviGenerator := gen.NewGeneratedFile("plugin.go", file.GoImportPath)
	if _, err := gviGenerator.Write(pg); err != nil {
		return err
	}
	return nil
}
