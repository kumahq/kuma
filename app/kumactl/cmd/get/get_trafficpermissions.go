package get

import (
	"context"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
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
				return errors.Wrapf(err, "failed to list TrafficPermissions")
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printTrafficPermissions(trafficPermissions.Items, cmd.OutOrStdout())
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

func printTrafficPermissions(trafficPermissions []*mesh.TrafficPermissionResource, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(trafficPermissions) <= i {
					return nil
				}
				trafficPermission := trafficPermissions[i]

				return []string{
					trafficPermission.GetMeta().GetMesh(), // MESH
					trafficPermission.GetMeta().GetName(), // NAME
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
