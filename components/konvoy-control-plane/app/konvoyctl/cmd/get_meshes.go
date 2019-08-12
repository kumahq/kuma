package cmd

import (
	"context"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output/printers"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model/rest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
)

func newGetMeshesCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meshes",
		Short: "Show Meshes",
		Long:  `Show Meshes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			controlPlane, err := pctx.CurrentControlPlane()
			if err != nil {
				return err
			}
			rs, err := pctx.NewResourceStore(controlPlane)
			if err != nil {
				return errors.Wrapf(err, "Failed to create a client for a given Control Plane: %s", controlPlane)
			}

			meshes := mesh.MeshResourceList{}
			if err := rs.List(context.Background(), &meshes); err != nil {
				return errors.Wrapf(err, "Failed to list Meshes")
			}

			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printMeshes(&meshes, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.ResourceList(&meshes), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}

func printMeshes(meshes *mesh.MeshResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"NAME"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(meshes.Items) <= i {
					return nil
				}
				mesh := meshes.Items[i]

				return []string{
					mesh.GetMeta().GetName(), // NAME
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
