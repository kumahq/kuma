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
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

func newGetCircuitBreakersCmd(pctx *listContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "circuit-breakers",
		Short: "Show CircuitBreakers",
		Long:  `Show CircuitBreaker entities.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			circuitBreakers := mesh.CircuitBreakerResourceList{}
			if err := rs.List(context.Background(), &circuitBreakers, core_store.ListByMesh(pctx.CurrentMesh()), core_store.ListByPage(pctx.args.size, pctx.args.offset)); err != nil {
				return errors.Wrapf(err, "failed to list CircuitBreaker")
			}

			switch format := output.Format(pctx.getContext.args.outputFormat); format {
			case output.TableFormat:
				return printCircuitBreakers(pctx.Now(), &circuitBreakers, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(&circuitBreakers), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printCircuitBreakers(rootTime time.Time, circuitBreakers *mesh.CircuitBreakerResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(circuitBreakers.Items) <= i {
					return nil
				}
				circuitBreaker := circuitBreakers.Items[i]

				return []string{
					circuitBreaker.GetMeta().GetMesh(),                                        // MESH
					circuitBreaker.GetMeta().GetName(),                                        // NAME
					table.TimeSince(circuitBreaker.GetMeta().GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(circuitBreakers),
	}
	return printers.NewTablePrinter().Print(data, out)
}
