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

func newGetDataplanesCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataplanes",
		Short: "Show Dataplanes",
		Long:  `Show Dataplanes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			dataplanes := mesh.DataplaneResourceList{}
			if err := rs.List(context.Background(), &dataplanes, core_store.ListByMesh(pctx.CurrentMesh())); err != nil {
				return errors.Wrapf(err, "failed to list Dataplanes")
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printDataplanes(dataplanes.Items, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(&dataplanes), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printDataplanes(dataplanes []*mesh.DataplaneResource, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "TAGS"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(dataplanes) <= i {
					return nil
				}
				dataplane := dataplanes[i]

				return []string{
					dataplane.Meta.GetMesh(),       // MESH
					dataplane.Meta.GetName(),       // NAME,
					dataplane.Spec.Tags().String(), // TAGS
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
