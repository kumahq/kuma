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

import (
	_ "github.com/kumahq/kuma/pkg/plugins/policies/{{ .Package }}/api/v1alpha1"
	_ "github.com/kumahq/kuma/pkg/plugins/policies/{{ .Package }}/k8s/v1alpha1"
)
`))

func generatePluginFile(
	gen *protogen.Plugin,
	file *protogen.File,
) error {
	var infos []genutils.ResourceInfo
	for _, msg := range file.Messages {
		infos = append(infos, genutils.ToResourceInfo(msg.Desc))
	}

	if len(infos) > 1 {
		return errors.Errorf("only one Kuma resource per proto file is allowed")
	}

	info := infos[0]

	outBuf := bytes.Buffer{}
	if err := PluginGoTemplate.Execute(&outBuf, struct {
		Package string
	}{
		Package: info.KumactlSingular,
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
