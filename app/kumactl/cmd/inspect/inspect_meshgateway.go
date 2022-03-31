package inspect

import (
	"context"
	"fmt"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

var inspectMeshGatewayTemplate = `{{ range $num, $item := .Items }}{{ .DataplaneKey.Name }}
{{ end }}`

func newInspectMeshGatewayCmd(pctx *cmd.RootContext) *cobra.Command {
	tmpl, err := template.New("meshgateway_inspect").Funcs(template.FuncMap{
		"FormatTags": tagsToStr(true),
	}).Parse(inspectMeshGatewayTemplate)
	if err != nil {
		panic(fmt.Sprintf("unable to parse template %v", err))
	}
	cmd := &cobra.Command{
		Use:   "meshgateway NAME",
		Short: "Inspect MeshGateway",
		Long:  "List Dataplanes matched by this MeshGateway.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			client, err := pctx.CurrentMeshGatewayInspectClient()
			if err != nil {
				return errors.Wrap(err, "failed to create a dataplane inspect client")
			}
			dataplanes, err := client.InspectDataplanes(context.Background(), pctx.CurrentMesh(), name)
			if err != nil {
				return err
			}
			return tmpl.Execute(cmd.OutOrStdout(), dataplanes)
		},
	}
	return cmd
}
