package config

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
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
				NextRow: func() func() []string {
					i := 0
					return func() []string {
						defer func() { i++ }()
						if len(controlPlanes) <= i {
							return nil
						}
						cp := controlPlanes[i]

						active := context != nil && context.ControlPlane == cp.Name

						return []string{
							table.Check(active), // ACTIVE
							cp.GetName(),        // NAME
							cp.GetCoordinates().GetApiServer().GetUrl(), // URL
						}
					}
				}(),
			}
			return printers.NewTablePrinter().Print(data, cmd.OutOrStdout())
		},
	}
}
