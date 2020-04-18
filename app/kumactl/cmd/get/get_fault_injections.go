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

func newGetFaultInjectionsCmd(pctx *getContext) *cobra.Command {
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
			if err := rs.List(context.Background(), &faultInjections, core_store.ListByMesh(pctx.CurrentMesh())); err != nil {
				return errors.Wrapf(err, "failed to list FaultInjection")
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printFaultInjections(faultInjections.Items, cmd.OutOrStdout())
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

func printFaultInjections(faultInjections []*mesh.FaultInjectionResource, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(faultInjections) <= i {
					return nil
				}
				faultInjection := faultInjections[i]

				return []string{
					faultInjection.GetMeta().GetMesh(), // MESH
					faultInjection.GetMeta().GetName(), // NAME
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
