package main

import (
	"flag"
	"fmt"

	"github.com/hoisie/mustache"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/kumahq/kuma/tools/resource-gen/genutils"
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
	var infos []genutils.ResourceInfo
	for _, msg := range file.Messages {
		infos = append(infos, genutils.ToResourceInfo(msg.Desc))
	}

	if len(infos) > 1 {
		return errors.Errorf("only one Kuma resource per proto file is allowed")
	}

	info := infos[0]

	rendered := mustache.RenderFile(openAPITemplate, struct {
		genutils.ResourceInfo
		PolicyVersion string
	}{
		ResourceInfo:  info,
		PolicyVersion: string(file.GoPackageName),
	})

	filename := fmt.Sprintf("api/%s/rest.yaml", string(file.GoPackageName))
	g := p.NewGeneratedFile(filename, file.GoImportPath)
	g.P(rendered)

	return nil
}
