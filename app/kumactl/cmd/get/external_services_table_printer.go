package get

import (
	"io"
	"time"

	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func printExternalServices(rootTime time.Time, resources model.ResourceList, out io.Writer) error {
	externalServices := resources.(*mesh.ExternalServiceResourceList)
	data := printers.Table{
		Headers: []string{"MESH", "NAME", "TAGS", "ADDRESS", "AGE"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(externalServices.Items) <= i {
					return nil
				}
				externalService := externalServices.Items[i]

				return []string{
					externalService.Meta.GetMesh(),                                        // MESH
					externalService.Meta.GetName(),                                        // NAME,
					externalService.Spec.TagSet().String(),                                // TAGS
					externalService.Spec.Networking.Address,                               // ADDRESS
					table.TimeSince(externalService.Meta.GetModificationTime(), rootTime), // AGE
				}
			}
		}(),
		Footer: table.PaginationFooter(externalServices),
	}
	return printers.NewTablePrinter().Print(data, out)
}
