package get

import (
	"context"
	"io"
	"time"

	"github.com/Kong/kuma/app/kumactl/pkg/output/table"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

func newGetProxyTemplatesCmd(pctx *listContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxytemplates",
		Short: "Show ProxyTemplates",
		Long:  `Show ProxyTemplates.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			proxyTemplates := &mesh_core.ProxyTemplateResourceList{}
			if err := rs.List(context.Background(), proxyTemplates, core_store.ListByMesh(pctx.CurrentMesh()), core_store.ListByPage(pctx.args.size, pctx.args.offset)); err != nil {
				return errors.Wrapf(err, "failed to list ProxyTemplates")
			}

			switch format := output.Format(pctx.getContext.args.outputFormat); format {
			case output.TableFormat:
				return printProxyTemplates(pctx.Now(), proxyTemplates, cmd.OutOrStdout())
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

func printProxyTemplates(rootTime time.Time, proxyTemplates *mesh_core.ProxyTemplateResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(proxyTemplates.Items) <= i {
					return nil
				}
				proxyTemplate := proxyTemplates.Items[i]

				return []string{
					proxyTemplate.GetMeta().GetMesh(),                                        // MESH
					proxyTemplate.GetMeta().GetName(),                                        // NAME
					table.TimeSince(proxyTemplate.GetMeta().GetModificationTime(), rootTime), //AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(proxyTemplates),
	}
	return printers.NewTablePrinter().Print(data, out)
}
