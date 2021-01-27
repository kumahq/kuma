package get

import (
	"io"
	"time"

	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func printDataplanes(rootTime time.Time, resources model.ResourceList, out io.Writer) error {
	dataplanes := resources.(*mesh.DataplaneResourceList)
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "TAGS", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(dataplanes.Items) <= i {
					return nil
				}
				dataplane := dataplanes.Items[i]

				return []string{
					dataplane.Meta.GetMesh(),                                        // MESH
					dataplane.Meta.GetName(),                                        // NAME,
					dataplane.Spec.TagSet().String(),                                // TAGS
					table.TimeSince(dataplane.Meta.GetModificationTime(), rootTime), // AGE

				}
			}
		}(),
		Footer: table.PaginationFooter(dataplanes),
	}
	return printers.NewTablePrinter().Print(data, out)
}
