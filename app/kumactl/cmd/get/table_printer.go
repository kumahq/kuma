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

// CustomTablePrinters are used to define different ways to print entities in table format.
var CustomTablePrinters = map[model.ResourceType]TablePrinter{
	mesh.DataplaneType: RowPrinter{
		Headers: []string{"MESH", "NAME", "TAGS", "ADDRESS", "AGE"},
		RowFn: func(rootTime time.Time, item model.Resource) []string {
			dataplane := item.(*mesh.DataplaneResource)
			address := dataplane.Spec.GetNetworking().GetAdvertisedAddress()
			if address == "" {
				address = dataplane.Spec.GetNetworking().GetAddress()
			}
			return []string{
				dataplane.Meta.GetMesh(),         // MESH
				dataplane.Meta.GetName(),         // NAME,
				dataplane.Spec.TagSet().String(), // TAGS
				address,                          // ADDRESS
				table.TimeSince(dataplane.Meta.GetModificationTime(), rootTime), // AGE
			}
		},
	},
	mesh.ExternalServiceType: RowPrinter{
		Headers: []string{"MESH", "NAME", "TAGS", "ADDRESS", "AGE"},
		RowFn: func(rootTime time.Time, item model.Resource) []string {
			externalService := item.(*mesh.ExternalServiceResource)
			return []string{
				externalService.Meta.GetMesh(),                                        // MESH
				externalService.Meta.GetName(),                                        // NAME,
				externalService.Spec.TagSet().String(),                                // TAGS
				externalService.Spec.Networking.Address,                               // ADDRESS
				table.TimeSince(externalService.Meta.GetModificationTime(), rootTime), // AGE
			}
		},
	},
	model.ScopeMesh: RowPrinter{
		Headers: []string{"NAME", "mTLS", "METRICS", "LOGGING", "TRACING", "LOCALITY", "ZONEEGRESS", "AGE"},
		RowFn: func(rootTime time.Time, item model.Resource) []string {
			mesh := item.(*mesh.MeshResource)

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
			zoneEgress := "off"
			if mesh.Spec.GetRouting().GetZoneEgress() {
				zoneEgress = "on"
			}
			return []string{
				mesh.GetMeta().GetName(), // NAME
				mtls,                     // mTLS
				metrics,                  // METRICS
				logging,                  // LOGGING
				tracing,                  // TRACING
				locality,                 // LOCALITY
				zoneEgress,               // ZONEEGRESS
				table.TimeSince(mesh.GetMeta().GetModificationTime(), rootTime), // AGE
			}
		},
	},
}

type TablePrinter interface {
	Print(time.Time, model.ResourceList, io.Writer) error
}

type RowPrinter struct {
	Headers []string
	RowFn   func(rootTime time.Time, item model.Resource) []string
}

func (rp RowPrinter) Print(rootTime time.Time, resources model.ResourceList, out io.Writer) error {
	items := resources.GetItems()
	data := printers.Table{
		Headers: rp.Headers,
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(items) <= i {
					return nil
				}
				return rp.RowFn(rootTime, items[i])
			}
		}(),
		Footer: table.PaginationFooter(resources),
	}
	return printers.NewTablePrinter().Print(data, out)
}

var BasicResourceTablePrinter = RowPrinter{
	Headers: []string{"MESH", "NAME", "AGE"},
	RowFn: func(rootTime time.Time, item model.Resource) []string {
		return []string{
			item.GetMeta().GetMesh(), // MESH
			item.GetMeta().GetName(), // NAME
			table.TimeSince(item.GetMeta().GetModificationTime(), rootTime), // AGE
		}
	},
}

var BasicGlobalResourceTablePrinter = RowPrinter{
	Headers: []string{"NAME", "AGE"},
	RowFn: func(rootTime time.Time, item model.Resource) []string {
		return []string{
			item.GetMeta().GetName(), // NAME
			table.TimeSince(item.GetMeta().GetModificationTime(), rootTime), // AGE
		}
	},
}

func ResolvePrinter(resourceType model.ResourceType, scope model.ResourceScope) TablePrinter {
	tablePrinter := CustomTablePrinters[resourceType]
	if tablePrinter == nil {
		switch scope {
		case model.ScopeMesh:
			tablePrinter = BasicResourceTablePrinter
		case model.ScopeGlobal:
			tablePrinter = BasicGlobalResourceTablePrinter
		}
	}
	return tablePrinter
}
