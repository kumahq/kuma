package inspect

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var policyInspectTemplate = `Affected data plane proxies:

{{ range .Items }}{{ with IsSidecar . }}  {{ .DataplaneKey.Name }}{{ if . | PrintAttachments }}:
{{ range .Attachments }}    {{ . | FormatAttachment }}
{{ end }}{{ end }}
{{ end }}{{ end }}`

func newInspectPolicyCmd(policyDesc core_model.ResourceTypeDescriptor, pctx *cmd.RootContext) *cobra.Command {
	tmpl, err := template.New("policy_inspect").Funcs(template.FuncMap{
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
	}).Parse(policyInspectTemplate)
	if err != nil {
		panic("unable to parse template")
	}

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
			entryList, err := client.Inspect(context.Background(), policyDesc, pctx.CurrentMesh(), name)
			if err != nil {
				return err
			}
			return tmpl.Execute(cmd.OutOrStdout(), entryList)
		},
	}
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
