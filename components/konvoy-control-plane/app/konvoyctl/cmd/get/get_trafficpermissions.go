package get

import (
	"context"
	"io"

	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output/printers"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newGetTrafficPermissionsCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traffic-permissions",
		Short: "Show TrafficPermissions",
		Long:  `Show TrafficPermission entities.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			trafficPermissions := mesh.TrafficPermissionResourceList{}
			if err := rs.List(context.Background(), &trafficPermissions, core_store.ListByMesh(pctx.CurrentMesh())); err != nil {
				return errors.Wrapf(err, "Failed to list TrafficPermissions")
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printTrafficPermissions(&trafficPermissions, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(&trafficPermissions), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printTrafficPermissions(trafficPermissions *mesh.TrafficPermissionResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(trafficPermissions.Items) <= i {
					return nil
				}
				trafficPermission := trafficPermissions.Items[i]

				return []string{
					trafficPermission.GetMeta().GetMesh(), // MESH
					trafficPermission.GetMeta().GetName(), // NAME
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
