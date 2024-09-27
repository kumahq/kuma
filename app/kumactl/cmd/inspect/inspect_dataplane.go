package inspect

import (
	"context"
	"fmt"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

const (
	InspectionTypePolicies   = "policies"
	InspectionTypeConfigDump = "config-dump"
	InspectionTypeStats      = "stats"
	InspectionTypeClusters   = "clusters"
	InspectionConfig         = "config"
)

var dataplaneInspectTemplate = `{{ with IsSidecar . }}{{ range $num, $item := .Items }}{{ .AttachmentEntry | FormatAttachment }}:
{{ range $typ, $policies := .MatchedPolicies }}  {{ $typ }}
    {{ range $policies }}{{ .Name }}
{{ end }}{{ end }}
{{ end }}{{ end }}{{ with IsGateway . }}MESHGATEWAY:
{{ range $typ, $policy := .Policies }}  {{ $typ }}
    {{ .Name }}
{{ end }}
{{ range .Listeners }}LISTENER ({{ .Protocol }}:{{ .Port }}):
{{ range .Hosts }}  {{ .HostName }}:
{{ range .Routes }}    ROUTE {{ .Route }}:
{{ range .Destinations }}      {{ FormatTags .Tags }}:
{{ range $typ, $policy := .Policies }}        {{ $typ }}
          {{ .Name }}
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
	var inspectionType string
	var shadow bool
	var include []string
	cmd := &cobra.Command{
		Use:   "dataplane NAME",
		Short: "Inspect Dataplane",
		Long:  "Inspect Dataplane.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if configDump {
				inspectionType = InspectionTypeConfigDump
			}

			if shadow && inspectionType != InspectionConfig {
				return errors.New("flag '--shadow' can be used only when '--type=config'")
			}
			if len(include) > 0 && inspectionType != InspectionConfig {
				return errors.New("flag '--include' can be used only when '--type=config'")
			}

			client, err := pctx.CurrentInspectEnvoyProxyClient(mesh.DataplaneResourceTypeDescriptor)
			if err != nil {
				return errors.Wrap(err, "failed to create a dataplane inspect client")
			}

			resourceKey := core_model.ResourceKey{Name: name, Mesh: pctx.CurrentMesh()}
			switch inspectionType {
			case InspectionTypePolicies:
				client, err := pctx.CurrentDataplaneInspectClient()
				if err != nil {
					return errors.Wrap(err, "failed to create a dataplane inspect client")
				}

				entryList, err := client.InspectPolicies(context.Background(), pctx.CurrentMesh(), name)
				if err != nil {
					return err
				}
				return tmpl.Execute(cmd.OutOrStdout(), entryList)
			case InspectionTypeConfigDump:
				bytes, err := client.ConfigDump(context.Background(), resourceKey, true)
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), string(bytes))
				return err
			case InspectionTypeStats:
				bytes, err := client.Stats(context.Background(), resourceKey)
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), string(bytes))
				return err
			case InspectionTypeClusters:
				bytes, err := client.Clusters(context.Background(), resourceKey)
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), string(bytes))
				return err
			case InspectionConfig:
				bytes, err := client.Config(context.Background(), resourceKey, shadow, include)
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(cmd.OutOrStdout(), string(bytes))
				return err
			default:
				return errors.New("invalid inspection type")
			}
		},
	}
	cmd.PersistentFlags().StringVar(&inspectionType, "type", InspectionTypePolicies, kuma_cmd.UsageOptions("inspection type", InspectionTypePolicies, InspectionTypeConfigDump, InspectionTypeStats, InspectionTypeClusters, InspectionConfig))
	cmd.PersistentFlags().BoolVar(&configDump, "config-dump", false, "if set then the command returns envoy config dump for provided dataplane")
	_ = cmd.PersistentFlags().MarkDeprecated("config-dump", "use --type=config-dump")
	cmd.PersistentFlags().StringVarP(&pctx.Args.Mesh, "mesh", "m", "default", "mesh to use")
	cmd.PersistentFlags().BoolVar(&shadow, "shadow", false, "when computing XDS config the CP takes into account policies with 'kuma.io/effect: shadow' label")
	cmd.PersistentFlags().StringSliceVar(&include, "include", []string{}, "an array of extra fields to include in the response. When `include=diff` the server computes a diff in JSONPatch format between the XDS config returned in 'xds' and the current proxy XDS config.")
	return cmd
}
