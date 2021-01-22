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

func newGetFaultInjectionsCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fault-injections",
		Short: "Show FaultInjections",
		Long:  `Show FaultInjection entities.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			faultInjections := mesh.FaultInjectionResourceList{}
			if err := rs.List(context.Background(), &faultInjections, core_store.ListByMesh(pctx.CurrentMesh()), core_store.ListByPage(pctx.ListContext.Args.Size, pctx.ListContext.Args.Offset)); err != nil {
				return errors.Wrapf(err, "failed to list FaultInjection")
			}

			switch format := output.Format(pctx.GetContext.Args.OutputFormat); format {
			case output.TableFormat:
				return printFaultInjections(pctx.Now(), &faultInjections, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(&faultInjections), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printFaultInjections(rootTime time.Time, faultInjections *mesh.FaultInjectionResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(faultInjections.Items) <= i {
					return nil
				}
				faultInjection := faultInjections.Items[i]

				return []string{
					faultInjection.GetMeta().GetMesh(),                                        // MESH
					faultInjection.GetMeta().GetName(),                                        // NAME
					table.TimeSince(faultInjection.GetMeta().GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(faultInjections),
	}
	return printers.NewTablePrinter().Print(data, out)
}
