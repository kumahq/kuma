package get

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"

	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

func newGetTrafficLogCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traffic-log NAME",
		Short: "Show a single TrafficLog resource",
		Long:  `Show a single TrafficLog resource.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			name := args[0]
			currentMesh := pctx.CurrentMesh()
			trafficLog := &mesh.TrafficLogResource{}
			if err := rs.Get(context.Background(), trafficLog, store.GetByKey(name, currentMesh)); err != nil {
				if store.IsResourceNotFound(err) {
					return errors.Errorf("No resources found in %s mesh", currentMesh)
				}
				return errors.Wrapf(err, "failed to get mesh %s", currentMesh)
			}
			trafficLogs := &mesh.TrafficLogResourceList{
				Items: []*mesh.TrafficLogResource{trafficLog},
			}
			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printTrafficLogs(pctx.Now(), trafficLogs, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.Resource(trafficLog), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}
