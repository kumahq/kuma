package inspect

import (
	"context"
	"fmt"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

var dataplaneInspectTemplate = "{{ range .Items }}" +
	"{{ .AttachmentEntry | FormatAttachment }}:\n" +
	"{{ range $typ, $policies := .MatchedPolicies }}" +
	"  {{ $typ }}\n" +
	"    {{ range $policies }}{{ .Meta.Name }}\n{{ end }}" +
	"{{ end }}" +
	"\n" +
	"{{ end }}"

func newInspectDataplaneCmd(pctx *cmd.RootContext) *cobra.Command {
	tmpl, err := template.New("dataplane_inspect").Funcs(template.FuncMap{
		"FormatAttachment": attachmentToStr(true),
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
					return errors.Wrap(err, "failed to create a dataplane inspect client")
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
					return errors.Wrap(err, "failed to create a dataplane inspect client")
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
	return cmd
}
