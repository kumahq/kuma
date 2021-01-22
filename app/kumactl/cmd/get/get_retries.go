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
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func newGetRetriesCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retries",
		Short: "Show Retries",
		Long:  `Show Retries.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			retries := &mesh_core.RetryResourceList{}
			if err := rs.List(
				context.Background(),
				retries,
				core_store.ListByMesh(pctx.CurrentMesh()),
				core_store.ListByPage(pctx.ListContext.Args.Size, pctx.ListContext.Args.Offset),
			); err != nil {
				return errors.Wrapf(err, "failed to list Retries")
			}

			switch format := output.Format(pctx.GetContext.Args.OutputFormat); format {
			case output.TableFormat:
				return printRetries(pctx.Now(), retries, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(
					rest_types.From.ResourceList(retries),
					cmd.OutOrStdout(),
				)
			}
		},
	}
	return cmd
}

func printRetries(
	rootTime time.Time,
	retries *mesh_core.RetryResourceList,
	out io.Writer,
) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(retries.Items) <= i {
					return nil
				}
				retry := retries.Items[i]

				return []string{
					retry.Meta.GetMesh(), // MESH
					retry.Meta.GetName(), // NAME
					table.TimeSince(retry.Meta.GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(retries),
	}
	return printers.NewTablePrinter().Print(data, out)
}
