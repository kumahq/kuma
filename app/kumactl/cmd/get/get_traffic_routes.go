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

func newGetTrafficRoutesCmd(pctx *listContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traffic-routes",
		Short: "Show TrafficRoutes",
		Long:  `Show TrafficRoutes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			trafficRoutes := &mesh_core.TrafficRouteResourceList{}
			if err := rs.List(context.Background(), trafficRoutes, core_store.ListByMesh(pctx.CurrentMesh()), core_store.ListByPage(pctx.args.size, pctx.args.offset)); err != nil {
				return errors.Wrapf(err, "failed to list TrafficRoutes")
			}

			switch format := output.Format(pctx.getContext.args.outputFormat); format {
			case output.TableFormat:
				return printTrafficRoutes(pctx.Now(), trafficRoutes, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(trafficRoutes), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printTrafficRoutes(rootTime time.Time, trafficRoutes *mesh_core.TrafficRouteResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(trafficRoutes.Items) <= i {
					return nil
				}
				trafficroute := trafficRoutes.Items[i]

				return []string{
					trafficroute.Meta.GetMesh(),                                        // MESH
					trafficroute.Meta.GetName(),                                        // NAME
					table.TimeSince(trafficroute.Meta.GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(trafficRoutes),
	}
	return printers.NewTablePrinter().Print(data, out)
}
