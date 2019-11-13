package get

import (
	"context"
	"io"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newGetTrafficLogsCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traffic-logs",
		Short: "Show TrafficLogs",
		Long:  `Show TrafficLog entities.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			trafficLogging := mesh.TrafficLogResourceList{}
			if err := rs.List(context.Background(), &trafficLogging, core_store.ListByMesh(pctx.CurrentMesh())); err != nil {
				return errors.Wrapf(err, "failed to list TrafficLog")
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printTrafficLog(&trafficLogging, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(&trafficLogging), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printTrafficLog(trafficLogging *mesh.TrafficLogResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(trafficLogging.Items) <= i {
					return nil
				}
				trafficLogging := trafficLogging.Items[i]

				return []string{
					trafficLogging.GetMeta().GetMesh(), // MESH
					trafficLogging.GetMeta().GetName(), // NAME
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
