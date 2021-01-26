package get

import (
	"io"
	"time"

	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func printZones(rootTime time.Time, resources model.ResourceList, out io.Writer) error {
	zones := resources.(*system.ZoneResourceList)
	data := printers.Table{
		Headers: []string{"NAME", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(zones.Items) <= i {
					return nil
				}
				zone := zones.Items[i]

				return []string{
					zone.GetMeta().GetName(), // NAME
					table.TimeSince(zone.GetMeta().GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(zones),
	}
	return printers.NewTablePrinter().Print(data, out)
}
