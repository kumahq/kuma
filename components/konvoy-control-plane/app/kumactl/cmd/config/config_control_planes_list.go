package config

import (
	kumactl_cmd "github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/cmd"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/output/printers"
	"github.com/spf13/cobra"
)

func newConfigControlPlanesListCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List known Control Planes",
		Long:  `List known Control Planes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			controlPlanes := pctx.Config().ControlPlanes

			data := printers.Table{
				Headers: []string{"NAME", "API SERVER"},
				NextRow: func() func() []string {
					i := 0
					return func() []string {
						defer func() { i++ }()
						if len(controlPlanes) <= i {
							return nil
						}
						cp := controlPlanes[i]

						return []string{
							cp.GetName(), // NAME
							cp.GetCoordinates().GetApiServer().GetUrl(), // URL
						}
					}
				}(),
			}
			return printers.NewTablePrinter().Print(data, cmd.OutOrStdout())
		},
	}
}
