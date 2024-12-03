package get

import (
	"fmt"
	"time"

	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

// CustomTablePrinters are used to define different ways to print entities in table format.
var CustomTablePrinters = map[model.ResourceType]RowPrinter{
	mesh.DataplaneType: {
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
	mesh.ExternalServiceType: {
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
	model.ScopeMesh: {
		Headers: []string{"NAME", "mTLS", "LOCALITY", "ZONEEGRESS", "AGE"},
		RowFn: func(rootTime time.Time, item model.Resource) []string {
			mesh := item.(*mesh.MeshResource)

			mtls := "off"
			if mesh.MTLSEnabled() {
				backend := mesh.GetEnabledCertificateAuthorityBackend()
				mtls = fmt.Sprintf("%s/%s", backend.Type, backend.Name)
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
				locality,                 // LOCALITY
				zoneEgress,               // ZONEEGRESS
				table.TimeSince(mesh.GetMeta().GetModificationTime(), rootTime), // AGE
			}
		},
	},
}

type RowPrinter struct {
	Headers []string
	RowFn   func(rootTime time.Time, item model.Resource) []string
	Now     time.Time
}

func (rp RowPrinter) AsTable() printers.Table {
	return printers.Table{
		Headers: rp.Headers,
		RowForItem: func(i int, container interface{}) ([]string, error) {
			rl, ok := container.(model.ResourceList)
			if ok {
				items := rl.GetItems()
				if len(items) <= i {
					return nil, nil
				}
				return rp.RowFn(rp.Now, items[i]), nil
			} else {
				if i != 0 {
					return nil, nil
				}
				return rp.RowFn(rp.Now, container.(model.Resource)), nil
			}
		},
		FooterFn: func(container interface{}) string {
			rl, ok := container.(model.ResourceList)
			if !ok {
				return ""
			}
			return table.PaginationFooter(rl)
		},
	}
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

func ResolvePrinter(resourceType model.ResourceType, scope model.ResourceScope, now time.Time) printers.Table {
	tablePrinter, ok := CustomTablePrinters[resourceType]
	if !ok {
		switch scope {
		case model.ScopeMesh:
			tablePrinter = BasicResourceTablePrinter
		case model.ScopeGlobal:
			tablePrinter = BasicGlobalResourceTablePrinter
		}
	}
	tablePrinter.Now = now
	return tablePrinter.AsTable()
}
