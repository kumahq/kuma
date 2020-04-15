package get

import (
	"context"
	"io"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
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
				return printMeshes(meshes.Items, cmd.OutOrStdout())
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

func printMeshes(meshes []*mesh.MeshResource, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"NAME", "mTLS", "METRICS", "LOGGING", "TRACING"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(meshes) <= i {
					return nil
				}
				mesh := meshes[i]

				mtls := "off"
				if mesh.Spec.Mtls.GetEnabled() {
					switch mesh.Spec.GetMtls().GetCa().GetType().(type) {
					case *mesh_proto.CertificateAuthority_Provided_:
						mtls = "provided"
					case *mesh_proto.CertificateAuthority_Builtin_:
						mtls = "builtin"
					}
				}

				metrics := "off"
				switch {
				case mesh.HasPrometheusMetricsEnabled():
					metrics = "prometheus"
				}
				logging := "off"
				if mesh.Spec.GetLogging() != nil {
					logging = mesh.GetLoggingBackends()
					if len(logging) == 0 {
						logging = "off"
					}
				}
				tracing := "off"
				if mesh.Spec.GetTracing() != nil {
					tracing = mesh.GetTracingBackends()
					if len(tracing) == 0 {
						tracing = "off"
					}
				}
				return []string{
					mesh.GetMeta().GetName(), // NAME
					mtls,                     // mTLS
					metrics,                  // METRICS
					logging,                  // LOGGING
					tracing,                  // TRACING
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}
