package get

import (
	"context"
	"io"
	"time"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func newGetTrafficPermissionsCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
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
			if err := rs.List(context.Background(), &trafficPermissions, core_store.ListByMesh(pctx.CurrentMesh()), core_store.ListByPage(pctx.ListContext.Args.Size, pctx.ListContext.Args.Offset)); err != nil {
				return errors.Wrapf(err, "failed to list TrafficPermissions")
			}

			switch format := output.Format(pctx.GetContext.Args.OutputFormat); format {
			case output.TableFormat:
				return printTrafficPermissions(pctx.Now(), &trafficPermissions, cmd.OutOrStdout())
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

func printTrafficPermissions(rootTime time.Time, trafficPermissions *mesh.TrafficPermissionResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(trafficPermissions.Items) <= i {
					return nil
				}
				trafficPermission := trafficPermissions.Items[i]

				return []string{
					trafficPermission.GetMeta().GetMesh(),                                        // MESH
					trafficPermission.GetMeta().GetName(),                                        // NAME
					table.TimeSince(trafficPermission.GetMeta().GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(trafficPermissions),
	}
	return printers.NewTablePrinter().Print(data, out)
}
