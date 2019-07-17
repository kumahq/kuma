package cmd

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output/printers"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	"github.com/spf13/cobra"
)

func newConfigControlPlanesListCmd(pctx *rootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List known Control Planes",
		Long:  `List known Control Planes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			controlPlanes := pctx.Config().ControlPlanes

			data := printers.Table{
				Headers: []string{"NAME", "ENVIRONMENT"},
				NextRow: func() func() []string {
					i := 0
					return func() []string {
						if len(controlPlanes) <= i {
							return nil
						}
						cp := controlPlanes[i]
						i++

						env := "non-k8s"
						if _, ok := cp.Coordinates.Type.(*config_proto.ControlPlaneCoordinates_Kubernetes_); ok {
							env = "k8s"
						}
						return []string{
							cp.Name, // NAME
							env,     // ENVIRONMENT
						}
					}
				}(),
			}
			return printers.NewTablePrinter().Print(data, cmd.OutOrStdout())
		},
	}
	return cmd
}
