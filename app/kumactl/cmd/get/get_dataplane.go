package get

import (
	"context"
	"io"

	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"

	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/Kong/kuma/pkg/core/resources/store"
)

func newGetDataplaneCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataplane NAME",
		Short: "Show Single Dataplane Resource",
		Long:  `Show Single Dataplane Resource.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			name := args[0]
			currentMesh := pctx.CurrentMesh()
			dataplane := mesh.DataplaneResource{}
			if err := rs.Get(context.Background(), &dataplane, store.GetByKey(name, currentMesh)); err != nil {
				if store.IsResourceNotFound(err) {
					return errors.Errorf("No resources found in %s mesh", currentMesh)
				}
				return errors.Wrapf(err, "failed to get mesh %s", currentMesh)
			}
			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printDataplane(&dataplane, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.Resource(&dataplane), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printDataplane(dataplane *mesh.DataplaneResource, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if i == 1 {
					return nil
				}
				return []string{
					dataplane.GetMeta().GetMesh(), // MESH
					dataplane.GetMeta().GetName(), // NAME,
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
