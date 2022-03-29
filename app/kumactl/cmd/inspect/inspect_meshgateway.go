package inspect

import (
	"context"
	"fmt"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

var inspectMeshGatewayTemplate = `{{ range $num, $item := .Selectors }}SELECTOR {{ $num }}:
  GATEWAY:
{{ range $typ, $policy := .Policies }}    {{ $typ }}
      {{ .Meta.Name }}
{{ end }}
{{ range .Listeners }}  LISTENER ({{ .Protocol }}:{{ .Port }}):
{{ range .Hosts }}    {{ .HostName }}:
{{ range .Routes }}      ROUTE {{ .Route }}:
{{ range .Destinations }}        {{ FormatTags .Tags }}:
{{ range $typ, $policy := .Policies }}          {{ $typ }}
            {{ .Meta.Name }}
{{ end }}
{{ end }}{{ end }}{{ end }}{{ end }}{{ end }}`

func newInspectMeshGatewayCmd(pctx *cmd.RootContext) *cobra.Command {
	tmpl, err := template.New("meshgateway_inspect").Funcs(template.FuncMap{
		"FormatTags": tagsToStr(true),
	}).Parse(inspectMeshGatewayTemplate)
	if err != nil {
		panic(fmt.Sprintf("unable to parse template %v", err))
	}
	var configDump bool
	cmd := &cobra.Command{
		Use:   "meshgateway NAME",
		Short: "Inspect MeshGateway",
		Long:  "Inspect MeshGateway.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			client, err := pctx.CurrentMeshGatewayInspectClient()
			if err != nil {
				return errors.Wrap(err, "failed to create a dataplane inspect client")
			}
			selectorList, err := client.InspectPolicies(context.Background(), pctx.CurrentMesh(), name)
			if err != nil {
				return err
			}
			return tmpl.Execute(cmd.OutOrStdout(), selectorList)
		},
	}
	cmd.PersistentFlags().BoolVar(&configDump, "config-dump", false, "if set then the command returns envoy config dump for provided dataplane")
	return cmd
}
