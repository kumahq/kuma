package get

import (
	"fmt"
	"io"
	"time"

	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func printMeshes(rootTime time.Time, resources model.ResourceList, out io.Writer) error {
	meshes := resources.(*mesh.MeshResourceList)
	data := printers.Table{
		Headers: []string{"NAME", "mTLS", "METRICS", "LOGGING", "TRACING", "LOCALITY", "AGE"},
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
				locality := "off"
				if mesh.Spec.GetRouting().GetLocalityAwareLoadBalancing() {
					locality = "on"
				}
				return []string{
					mesh.GetMeta().GetName(), // NAME
					mtls,                     // mTLS
					metrics,                  // METRICS
					logging,                  // LOGGING
					tracing,                  // TRACING
					locality,                 // LOCALITY
					table.TimeSince(mesh.GetMeta().GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(meshes),
	}
	return printers.NewTablePrinter().Print(data, out)
}
