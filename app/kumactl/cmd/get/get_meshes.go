package get

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Kong/kuma/app/kumactl/pkg/output/table"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/app/kumactl/pkg/output"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	rest_types "github.com/Kong/kuma/pkg/core/resources/model/rest"
)

func newGetMeshesCmd(pctx *listContext) *cobra.Command {
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
			if err := rs.List(context.Background(), &meshes, core_store.ListByPage(pctx.args.size, pctx.args.offset)); err != nil {
				return errors.Wrapf(err, "failed to list Meshes")
			}

			switch format := output.Format(pctx.getContext.args.outputFormat); format {
			case output.TableFormat:
				return printMeshes(pctx.Now(), &meshes, cmd.OutOrStdout())
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

func printMeshes(rootTime time.Time, meshes *mesh.MeshResourceList, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"NAME", "mTLS", "METRICS", "LOGGING", "TRACING", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(meshes.Items) <= i {
					return nil
				}
				mesh := meshes.Items[i]

				mtls := "off"
				if mesh.MTLSEnabled() {
					backend := mesh.GetEnabledCertificateAuthorityBackend()
					mtls = fmt.Sprintf("%s/%s", backend.Type, backend.Name)
				}

				metrics := "off"
				if mesh.Spec.GetMetrics().GetEnabledBackend() != "" {
					backend := mesh.GetEnabledMetricsBackend()
					metrics = fmt.Sprintf("%s/%s", backend.Type, backend.Name)
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
					table.TimeSince(mesh.GetMeta().GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(meshes),
	}
	return printers.NewTablePrinter().Print(data, out)
}
