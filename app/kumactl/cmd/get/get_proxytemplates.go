package get

import (
	"context"
	"io"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newGetProxyTemplatesCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxytemplates",
		Short: "Show ProxyTemplates",
		Long:  `Show ProxyTemplates.`,
		Example: `# Get all proxy template
		kumactl get proxytemplates
		
		# Get all proxy templae of specific Mesh
		kumactl get proxytemplate -ojson | jq '.items[] | select(.mesh == "my-service-mesh" | .name)'
		`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			proxyTemplates := &mesh_core.ProxyTemplateResourceList{}
			if err := rs.List(context.Background(), proxyTemplates, core_store.ListByMesh(pctx.CurrentMesh())); err != nil {
				return errors.Wrapf(err, "failed to list ProxyTemplates")
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return PrintProxyTemplates(proxyTemplates, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(proxyTemplates), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func PrintProxyTemplates(proxyTemplates *mesh_core.ProxyTemplateResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(proxyTemplates.Items) <= i {
					return nil
				}
				proxyTemplate := proxyTemplates.Items[i]

				return []string{
					proxyTemplate.GetMeta().GetMesh(), // MESH
					proxyTemplate.GetMeta().GetName(), // NAME
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
