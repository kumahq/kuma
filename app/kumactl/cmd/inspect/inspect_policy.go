package inspect

import (
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var policyInspectTemplate = "Affected data plane proxies:\n\n" +
	"{{ range .Items }}" +
	"  {{ .DataplaneKey.Name }}" +
	"{{ if . | PrintAttachments }}" +
	":\n" +
	"{{ range .Attachments }}" +
	"    {{ . | FormatAttachment }}\n" +
	"{{ end }}" +
	"{{ end }}" +
	"\n" +
	"{{ end }}"

func newInspectPolicyCmd(policyDesc core_model.ResourceTypeDescriptor, pctx *cmd.RootContext) *cobra.Command {
	tmpl, err := template.New("policy_inspect").Funcs(template.FuncMap{
		"FormatAttachment": attachmentToStr(false),
		"PrintAttachments": func(e api_server_types.PolicyInspectEntry) bool {
			if len(e.Attachments) == 1 && e.Attachments[0].Type == "dataplane" {
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
