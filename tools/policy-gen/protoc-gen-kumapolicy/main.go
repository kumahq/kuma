package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	var flags flag.FlagSet
	endpointsTemplate := flags.String("endpoints-template", "", "OpenAPI endpoints template file")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		filesToGenerate := []*protogen.File{}

		for _, file := range plugin.Files {
			if !file.Generate {
				continue
			}

			filesToGenerate = append(filesToGenerate, file)

			if err := generateResource(plugin, file); err != nil {
				return err
			}
			if err := generateDeepcopy(plugin, file); err != nil {
				return err
			}
			if err := generateCRD(plugin, file); err != nil {
				return err
			}
			if err := generateEndpoints(plugin, file, *endpointsTemplate); err != nil {
				return err
			}
			if err := generateHelpers(plugin, file); err != nil {
				return err
			}
		}

		if err := generatePluginFile(plugin, filesToGenerate); err != nil {
			return err
		}

		return nil
	})
}

func generateEndpoints(
	p *protogen.Plugin,
	file *protogen.File,
	openAPITemplate string,
) error {
	var infos []PolicyConfig
	for _, msg := range file.Messages {
		infos = append(infos, NewPolicyConfig(msg.Desc))
	}

	if len(infos) > 1 {
		return errors.Errorf("only one Kuma resource per proto file is allowed")
	}

	info := infos[0]
	tmpl, err := template.ParseFiles(openAPITemplate)
	if err != nil {
		return err
	}

	bf := &bytes.Buffer{}
	if err := tmpl.Execute(bf, struct {
		Path          string
		Name          string
		PolicyVersion string
	}{
		Path:          info.Path,
		Name:          info.Name,
		PolicyVersion: string(file.GoPackageName),
	}); err != nil {
		return err
	}

	filename := fmt.Sprintf("api/%s/rest.yaml", string(file.GoPackageName))
	g := p.NewGeneratedFile(filename, file.GoImportPath)
	if _, err := g.Write(bf.Bytes()); err != nil {
		return err
	}

	return nil
}
