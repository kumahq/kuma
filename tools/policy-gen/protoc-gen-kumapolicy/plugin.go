package main

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/kumahq/kuma/tools/resource-gen/genutils"
)

var PluginGoTemplate = template.Must(template.New("plugin-go").Parse(`
package {{ .Package }}

{{ $pkg := .Package }}
{{ $rName := .ResourceName }}

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core"
{{range $idx, $version := .Versions}}
	api_{{ $version }} "github.com/kumahq/kuma/pkg/plugins/policies/{{ $pkg }}/api/{{ $version }}"
	k8s_{{ $version }} "github.com/kumahq/kuma/pkg/plugins/policies/{{ $pkg }}/k8s/{{ $version }}"
{{- end}}
)

func init() {
	{{- range $idx, $version := .Versions}}
	core.Register(
		api_{{ $version }}.{{ $rName }}TypeDescriptor,
		k8s_{{ $version }}.AddToScheme,
	)
	{{- end}}
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
		return fmt.Errorf("only one Kuma resource per proto file is allowed")
	}

	info := infos[0]

	versions := []string{}
	for _, f := range files {
		versions = append(versions, string(f.GoPackageName))
	}

	if info.SkipRegistration {
		// Don't generate the plugin file if it doesn't exists
		return nil
	}
	outBuf := bytes.Buffer{}
	if err := PluginGoTemplate.Execute(&outBuf, struct {
		Package      string
		Versions     []string
		ResourceName string
	}{
		Package:      info.KumactlSingular,
		ResourceName: info.ResourceName,
		Versions:     versions,
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
