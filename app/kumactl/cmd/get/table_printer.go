package get

import (
	"time"

	"github.com/kumahq/kuma/v3/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/v3/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// CustomTablePrinters are used to define different ways to print entities in table format.
var CustomTablePrinters = map[model.ResourceType]RowPrinter{
	mesh.DataplaneType: {
		Headers: []string{"MESH", "NAME", "TAGS", "ADDRESS", "AGE"},
		RowFn: func(rootTime time.Time, item model.Resource) []string {
			dataplane := item.(*mesh.DataplaneResource)
			address := dataplane.Spec.GetNetworking().GetAddress()
			return []string{
				dataplane.Meta.GetMesh(),         // MESH
				dataplane.Meta.GetName(),         // NAME,
				dataplane.DisplayTags().String(), // TAGS
				address,                          // ADDRESS
				table.TimeSince(dataplane.Meta.GetModificationTime(), rootTime), // AGE
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
		RowForItem: func(i int, container any) ([]string, error) {
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
		FooterFn: func(container any) string {
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
