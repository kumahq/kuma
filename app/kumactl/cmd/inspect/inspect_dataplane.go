package inspect

import (
	"context"
	"fmt"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var dataplaneInspectTemplate = `{{ with IsSidecar . }}{{ range $num, $item := .Items }}{{ .AttachmentEntry | FormatAttachment }}:
{{ range $typ, $policies := .MatchedPolicies }}  {{ $typ }}
    {{ range $policies }}{{ .Meta.Name }}
{{ end }}{{ end }}
{{ end }}{{ end }}{{ with IsGateway . }}MESHGATEWAY:
{{ range $typ, $policy := .Policies }}  {{ $typ }}
    {{ .Meta.Name }}
{{ end }}
{{ range .Listeners }}LISTENER ({{ .Protocol }}:{{ .Port }}):
{{ range .Hosts }}  {{ .HostName }}:
{{ range .Routes }}    ROUTE {{ .Route }}:
{{ range .Destinations }}      {{ FormatTags .Tags }}:
{{ range $typ, $policy := .Policies }}        {{ $typ }}
          {{ .Meta.Name }}
{{ end }}
{{ end }}{{ end }}{{ end }}{{ end }}{{ end }}`

func newInspectDataplaneCmd(pctx *cmd.RootContext) *cobra.Command {
	tmpl, err := template.New("dataplane_inspect").Funcs(template.FuncMap{
		"IsSidecar": func(e api_server_types.DataplaneInspectResponse) *api_server_types.DataplaneInspectEntryList {
			if concrete, ok := e.DataplaneInspectResponseKind.(*api_server_types.DataplaneInspectEntryList); ok {
				return concrete
			}
			return nil
		},
		"IsGateway": func(e api_server_types.DataplaneInspectResponse) *api_server_types.GatewayDataplaneInspectResult {
			if concrete, ok := e.DataplaneInspectResponseKind.(*api_server_types.GatewayDataplaneInspectResult); ok {
				return concrete
			}
			return nil
		},
		"FormatAttachment": attachmentToStr(true),
		"FormatTags":       tagsToStr(true),
	}).Parse(dataplaneInspectTemplate)
	if err != nil {
		panic("unable to parse template")
	}
	var configDump bool
	cmd := &cobra.Command{
		Use:   "dataplane NAME",
		Short: "Inspect Dataplane",
		Long:  "Inspect Dataplane.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if configDump {
				client, err := pctx.CurrentInspectEnvoyProxyClient(mesh.DataplaneResourceTypeDescriptor)
				if err != nil {
					return fmt.Errorf("failed to create a dataplane inspect client: %w", err)
				}
				bytes, err := client.ConfigDump(context.Background(), core_model.ResourceKey{Name: name, Mesh: pctx.CurrentMesh()})
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), string(bytes))
				return err
			} else {
				client, err := pctx.CurrentDataplaneInspectClient()
				if err != nil {
					return fmt.Errorf("failed to create a dataplane inspect client: %w", err)
				}
				entryList, err := client.InspectPolicies(context.Background(), pctx.CurrentMesh(), name)
				if err != nil {
					return err
				}
				return tmpl.Execute(cmd.OutOrStdout(), entryList)
			}
		},
	}
	cmd.PersistentFlags().BoolVar(&configDump, "config-dump", false, "if set then the command returns envoy config dump for provided dataplane")
	cmd.PersistentFlags().StringVarP(&pctx.Args.Mesh, "mesh", "m", "default", "mesh to use")
	return cmd
}
