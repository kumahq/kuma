package get

import (
	"context"
	"io"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	"github.com/Kong/kuma/app/kumactl/pkg/output/table"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newGetMeshesCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meshes",
		Short: "Show Meshes",
		Long:  `Show Meshes.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}

			meshes := mesh.MeshResourceList{}
			if err := rs.List(context.Background(), &meshes); err != nil {
				return errors.Wrapf(err, "failed to list Meshes")
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
		Headers: []string{"NAME", "mTLS", "CA", "METRICS"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(meshes.Items) <= i {
					return nil
				}
				mesh := meshes.Items[i]

				ca := ""
				switch mesh.Spec.GetMtls().GetCa().GetType().(type) {
				case *mesh_proto.CertificateAuthority_Provided_:
					ca = "provided"
				case *mesh_proto.CertificateAuthority_Builtin_:
					ca = "builtin"
				default:
					ca = "unknown"
				}

				metrics := "off"
				switch {
				case mesh.HasPrometheusMetricsEnabled():
					metrics = "prometheus"
				}

				return []string{
					mesh.GetMeta().GetName(),                 // NAME
					table.OnOff(mesh.Spec.Mtls.GetEnabled()), // mTLS
					ca,                                       // CA
					metrics,                                  // METRICS
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
