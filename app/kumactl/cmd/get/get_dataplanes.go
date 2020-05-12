package get

import (
	"context"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	"github.com/Kong/kuma/app/kumactl/pkg/output/table"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

func newGetDataplanesCmd(pctx *listContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataplanes",
		Short: "Show Dataplanes",
		Long:  `Show Dataplanes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			dataplanes := mesh.DataplaneResourceList{}
			if err := rs.List(context.Background(), &dataplanes, core_store.ListByMesh(pctx.CurrentMesh()), core_store.ListByPage(pctx.args.size, pctx.args.offset)); err != nil {
				return errors.Wrapf(err, "failed to list Dataplanes")
			}

			switch format := output.Format(pctx.getContext.args.outputFormat); format {
			case output.TableFormat:
				return printDataplanes(pctx.Now(), &dataplanes, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(&dataplanes), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printDataplanes(rootTime time.Time, dataplanes *mesh.DataplaneResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "TAGS", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(dataplanes.Items) <= i {
					return nil
				}
				dataplane := dataplanes.Items[i]

				return []string{
					dataplane.Meta.GetMesh(),                                        // MESH
					dataplane.Meta.GetName(),                                        // NAME,
					dataplane.Spec.Tags().String(),                                  // TAGS
					table.TimeSince(dataplane.Meta.GetModificationTime(), rootTime), // AGE

				}
			}
		}(),
		Footer: table.PaginationFooter(dataplanes),
	}
	return printers.NewTablePrinter().Print(data, out)
}
