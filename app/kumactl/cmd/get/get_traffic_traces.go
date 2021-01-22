package get

import (
	"context"
	"io"
	"time"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func newGetTrafficTracesCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traffic-traces",
		Short: "Show TrafficTraces",
		Long:  `Show TrafficTrace entities.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			trafficTraces := mesh.TrafficTraceResourceList{}
			if err := rs.List(context.Background(), &trafficTraces, core_store.ListByMesh(pctx.CurrentMesh()), core_store.ListByPage(pctx.ListContext.Args.Size, pctx.ListContext.Args.Offset)); err != nil {
				return errors.Wrapf(err, "failed to list TrafficTrace")
			}

			switch format := output.Format(pctx.GetContext.Args.OutputFormat); format {
			case output.TableFormat:
				return printTrafficTraces(pctx.Now(), &trafficTraces, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(&trafficTraces), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printTrafficTraces(rootTime time.Time, trafficTraces *mesh.TrafficTraceResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(trafficTraces.Items) <= i {
					return nil
				}
				trafficTraces := trafficTraces.Items[i]

				return []string{
					trafficTraces.GetMeta().GetMesh(),                                        // MESH
					trafficTraces.GetMeta().GetName(),                                        // NAME
					table.TimeSince(trafficTraces.GetMeta().GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(trafficTraces),
	}
	return printers.NewTablePrinter().Print(data, out)
}
