package get

import (
	"context"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

func newGetHealthChecksCmd(pctx *getContext) *cobra.Command {
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
			if err := rs.List(context.Background(), healthChecks, core_store.ListByMesh(pctx.CurrentMesh())); err != nil {
				return errors.Wrapf(err, "failed to list HealthChecks")
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printHealthChecks(healthChecks.Items, cmd.OutOrStdout())
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

func printHealthChecks(healthChecks []*mesh_core.HealthCheckResource, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(healthChecks) <= i {
					return nil
				}
				healthCheck := healthChecks[i]

				return []string{
					healthCheck.Meta.GetMesh(), // MESH
					healthCheck.Meta.GetName(), // NAME
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
