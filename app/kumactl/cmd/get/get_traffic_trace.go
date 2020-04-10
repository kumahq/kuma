package get

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

func newGetTrafficTraceCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traffic-trace NAME",
		Short: "Show Single TrafficTrace Resource",
		Long:  `Show Single TrafficTrace Resource.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			name := args[0]
			currentMesh := pctx.CurrentMesh()
			trafficTrace := &mesh.TrafficTraceResource{}
			if err := rs.Get(context.Background(), trafficTrace, store.GetByKey(name, currentMesh)); err != nil {
				if store.IsResourceNotFound(err) {
					return errors.Errorf("No resources found in %s mesh", currentMesh)
				}
				return errors.Wrapf(err, "failed to get mesh %s", currentMesh)
			}
			trafficTraces := []*mesh.TrafficTraceResource{trafficTrace}
			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printTrafficTraces(trafficTraces, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.Resource(trafficTrace), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}
