package main

import (
	"flag"

	"github.com/hoisie/mustache"
	"github.com/kumahq/kuma/tools/resource-gen/genutils"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	var flags flag.FlagSet
	endpointsTemplate := flags.String("endpoints-template", "", "OpenAPI endpoints template file")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		for _, file := range plugin.Files {
			if !file.Generate {
				continue
			}

			if err := generatePluginFile(plugin, file); err != nil {
				return err
			}
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

	rendered := mustache.RenderFile(openAPITemplate, info)

	filename := "api/v1alpha1/rest.yaml"
	g := p.NewGeneratedFile(filename, file.GoImportPath)
	g.P(rendered)

	return nil
}
