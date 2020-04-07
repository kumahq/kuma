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

func newGetMeshCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mesh NAME",
		Short: "Show Single Mesh Resource",
		Long:  `Show Single Mesh Resource.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			name := args[0]
			currentMesh := name
			mesh := mesh.MeshResource{}
			if err := rs.Get(context.Background(), &mesh, store.GetByKey(name, currentMesh)); err != nil {
				if store.IsResourceNotFound(err) {
					return errors.Errorf("No resources found in %s mesh", currentMesh)
				}
				return errors.Wrapf(err, "failed to get mesh %s", currentMesh)
			}
			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printMesh(&mesh, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.Resource(&mesh), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printMesh(mesh *mesh.MeshResource, out io.Writer) error {
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
					mesh.GetMeta().GetMesh(), // MESH
					mesh.GetMeta().GetName(), // NAME,
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
