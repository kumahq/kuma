package get

import (
	"io"
	"time"

	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

type TablePrinter = func(time.Time, model.ResourceList, io.Writer) error

func BasicResourceTablePrinter(rootTime time.Time, resources model.ResourceList, out io.Writer) error {
	items := resources.GetItems()
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(items) <= i {
					return nil
				}
				item := items[i]

				return []string{
					item.GetMeta().GetMesh(), // MESH
					item.GetMeta().GetName(), // NAME
					table.TimeSince(item.GetMeta().GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(resources),
	}
	return printers.NewTablePrinter().Print(data, out)
}
