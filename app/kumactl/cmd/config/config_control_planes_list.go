package config

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/pkg/config/app/kumactl/v1alpha1"
)

func newConfigControlPlanesListCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List Control Planes",
		Long:  `List Control Planes.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			context, _ := pctx.CurrentContext()

			controlPlanes := pctx.Config().ControlPlanes

			data := printers.Table{
				Headers: []string{"ACTIVE", "NAME", "ADDRESS"},
				RowForItem: func(i int, container interface{}) ([]string, error) {
					cps := container.([]*v1alpha1.ControlPlane)
					if len(cps) <= i {
						return nil, nil
					}
					cp := cps[i]
					active := context != nil && context.ControlPlane == cp.Name

					return []string{
						table.Check(active), // ACTIVE
						cp.GetName(),        // NAME
						cp.GetCoordinates().GetApiServer().GetUrl(), // URL
					}, nil
				},
			}
			return data.Print(controlPlanes, cmd.OutOrStdout())
		},
	}
}
