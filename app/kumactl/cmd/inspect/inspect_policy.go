package inspect

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/openapi/types"
	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var policyInspectTemplate = `Affected data plane proxies:

{{ range .Items }}{{ with IsSidecar . }}  {{ .DataplaneKey.Name }}{{ if . | PrintAttachments }}:
{{ range .Attachments }}    {{ . | FormatAttachment }}
{{ end }}{{ end }}
{{ end }}{{ with IsGateway . }}  {{ .Gateway.Name }}:
{{ range .Listeners }}    listener ({{ .Protocol }}:{{ .Port }}):
{{ range .Hosts }}      {{ .HostName }}:
{{ range .Routes }}        {{ .Route }}:
{{ range .Destinations }}          {{ FormatTags . }}
{{ end }}{{ end }}{{ end }}{{ end }}
{{ end }}{{ end }}`

func newInspectPolicyCmd(policyDesc core_model.ResourceTypeDescriptor, pctx *cmd.RootContext) *cobra.Command {
	legacyTmpl := template.Must(template.New("policy_inspect").Funcs(template.FuncMap{
		"IsSidecar": func(e api_server_types.PolicyInspectEntry) *api_server_types.PolicyInspectSidecarEntry {
			if concrete, ok := e.PolicyInspectEntryKind.(*api_server_types.PolicyInspectSidecarEntry); ok {
				return concrete
			}
			return nil
		},
		"IsGateway": func(e api_server_types.PolicyInspectEntry) *api_server_types.PolicyInspectGatewayEntry {
			if concrete, ok := e.PolicyInspectEntryKind.(*api_server_types.PolicyInspectGatewayEntry); ok {
				return concrete
			}
			return nil
		},
		"FormatAttachment": attachmentToStr(false),
		"PrintAttachments": func(sidecarEntry *api_server_types.PolicyInspectSidecarEntry) bool {
			if len(sidecarEntry.Attachments) == 1 && sidecarEntry.Attachments[0].Type == "dataplane" {
				return false
			}
			return true
		},
		"FormatTags": tagsToStr(false),
	}).Parse(policyInspectTemplate))

	var newApi bool
	var offset int
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s NAME", policyDesc.KumactlArg),
		Short: fmt.Sprintf("Inspect %s", policyDesc.Name),
		Long:  fmt.Sprintf("Inspect %s.", policyDesc.Name),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := pctx.CurrentPolicyInspectClient()
			if err != nil {
				return errors.Wrap(err, "failed to create a policy inspect client")
			}
			name := args[0]
			if !newApi {
				entryList, err := client.Inspect(cmd.Context(), policyDesc, pctx.CurrentMesh(), name)
				if err != nil {
					return err
				}
				return legacyTmpl.Execute(cmd.OutOrStdout(), entryList)
			}
			res, err := client.DataplanesForPolicy(cmd.Context(), policyDesc, pctx.CurrentMesh(), name)
			if err != nil {
				return err
			}
			format := output.Format(pctx.InspectContext.Args.OutputFormat)
			return printers.GenericPrint(format, res, printers.Table{
				Headers: []string{"Type", "Mesh", "Name"},
				FooterFn: func(container interface{}) string {
					return fmt.Sprintf("Total: %d", container.(types.InspectDataplanesForPolicyResponse).Total)
				},
				RowForItem: func(i int, container interface{}) ([]string, error) {
					items := container.(types.InspectDataplanesForPolicyResponse).Items
					if i >= len(items) {
						return nil, nil
					}
					itm := items[i]
					return []string{itm.Type, itm.Mesh, itm.Name}, nil
				},
			}, cmd.OutOrStdout())
		},
	}
	cmd.PersistentFlags().StringVarP(&pctx.Args.Mesh, "mesh", "m", "default", "mesh to use")
	cmd.PersistentFlags().IntVar(&offset, "offset", 0, "the offset for pagination")
	cmd.PersistentFlags().BoolVar(&newApi, "new-api", false, "use the newer version of the inspect api")
	return cmd
}

func attachmentToStr(upperCase bool) func(api_server_types.AttachmentEntry) string {
	return func(a api_server_types.AttachmentEntry) string {
		typeToStr := func(t string) string {
			if upperCase {
				return strings.ToUpper(t)
			}
			return t
		}
		switch a.Type {
		case "dataplane":
			return typeToStr(a.Type)
		case "service":
			return fmt.Sprintf("%s %s", typeToStr(a.Type), a.Name)
		default:
			return fmt.Sprintf("%s %s(%s)", typeToStr(a.Type), a.Name, a.Service)
		}
	}
}

func tagsToStr(upperCase bool) func(map[string]string) string {
	return func(destinationTags map[string]string) string {
		service := destinationTags[mesh_proto.ServiceTag]

		label := "service"
		if upperCase {
			label = strings.ToUpper(label)
		}
		return fmt.Sprintf("%s %s", label, service)
	}
}
