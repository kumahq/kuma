package get

import (
	"context"
	"io"
	"time"

	"github.com/Kong/kuma/app/kumactl/pkg/output/table"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

func newGetHealthChecksCmd(pctx *listContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "healthchecks",
		Short: "Show HealthChecks",
		Long:  `Show HealthChecks.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			healthChecks := &mesh_core.HealthCheckResourceList{}
			if err := rs.List(context.Background(), healthChecks, core_store.ListByMesh(pctx.CurrentMesh()), core_store.ListByPage(pctx.args.size, pctx.args.offset)); err != nil {
				return errors.Wrapf(err, "failed to list HealthChecks")
			}

			switch format := output.Format(pctx.getContext.args.outputFormat); format {
			case output.TableFormat:
				return printHealthChecks(pctx.Now(), healthChecks, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(healthChecks), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printHealthChecks(rootTime time.Time, healthChecks *mesh_core.HealthCheckResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(healthChecks.Items) <= i {
					return nil
				}
				healthCheck := healthChecks.Items[i]

				return []string{
					healthCheck.Meta.GetMesh(),                                        // MESH
					healthCheck.Meta.GetName(),                                        // NAME
					table.TimeSince(healthCheck.Meta.GetModificationTime(), rootTime), //AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(healthChecks),
	}
	return printers.NewTablePrinter().Print(data, out)
}
