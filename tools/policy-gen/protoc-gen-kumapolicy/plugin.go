package main

import (
	"bytes"
	"go/format"
	"html/template"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/compiler/protogen"
)

var PluginGoTemplate = template.Must(template.New("plugin-go").Parse(`
package {{ .Package }}

{{ $pkg := .Package }}
{{ $name := .Name }}

import (
	"github.com/kumahq/kuma/pkg/plugins/policies/core"
{{- range $idx, $version := .Versions}}
	api_{{ $version }} "github.com/kumahq/kuma/pkg/plugins/policies/{{ $pkg }}/api/{{ $version }}"
	k8s_{{ $version }} "github.com/kumahq/kuma/pkg/plugins/policies/{{ $pkg }}/k8s/{{ $version }}"
	plugin_{{ $version }} "github.com/kumahq/kuma/pkg/plugins/policies/{{ $pkg }}/plugin/{{ $version }}"
{{- end}}
)

func init() {
	{{- range $idx, $version := .Versions}}
	core.Register(
		api_{{ $version }}.{{ $name }}ResourceTypeDescriptor,
		k8s_{{ $version }}.AddToScheme,
		plugin_{{ $version }}.NewPlugin(),
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

	var infos []PolicyConfig
	for _, msg := range file.Messages {
		infos = append(infos, NewPolicyConfig(msg.Desc))
	}

	if len(infos) > 1 {
		return errors.Errorf("only one Kuma resource per proto file is allowed")
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
		Package  string
		Versions []string
		Name     string
	}{
		Package:  strings.ToLower(info.Name),
		Name:     info.Name,
		Versions: versions,
	}); err != nil {
		return err
	}

	pg, err := format.Source(outBuf.Bytes())
	if err != nil {
		return err
	}

	gviGenerator := gen.NewGeneratedFile("zz_generated.plugin.go", file.GoImportPath)
	if _, err := gviGenerator.Write(pg); err != nil {
		return err
	}
	return nil
}
