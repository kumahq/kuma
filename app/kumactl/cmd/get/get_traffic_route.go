package get

import (
	"context"

	"github.com/pkg/errors"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

func newGetTrafficRouteCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traffic-route NAME",
		Short: "Show a single TrafficRoute resource",
		Long:  `Show a single TrafficRoute resource.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			name := args[0]
			currentMesh := pctx.CurrentMesh()
			trafficRoute := mesh.NewTrafficRouteResource()
			if err := rs.Get(context.Background(), trafficRoute, store.GetByKey(name, currentMesh)); err != nil {
				if store.IsResourceNotFound(err) {
					return errors.Errorf("No resources found in %s mesh", currentMesh)
				}
				return errors.Wrapf(err, "failed to get a traffic route in mesh %s", currentMesh)
			}
			trafficRoutes := &mesh.TrafficRouteResourceList{
				Items: []*mesh.TrafficRouteResource{trafficRoute},
			}
			switch format := output.Format(pctx.GetContext.Args.OutputFormat); format {
			case output.TableFormat:
				return printTrafficRoutes(pctx.Now(), trafficRoutes, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.Resource(trafficRoute), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}
