package printers

import (
	"io"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/json"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/yaml"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
)

type Table = table.Table

func GenericPrint(format output.Format, data interface{}, table Table, out io.Writer) error {
	switch format {
	case output.JSONFormat:
		if rl, ok := data.(model.ResourceList); ok {
			data = rest_types.From.ResourceList(rl)
		}
		if r, ok := data.(model.Resource); ok {
			data = rest_types.From.Resource(r)
		}
		return json.NewPrinter().Print(data, out)
	case output.YAMLFormat:
		if rl, ok := data.(model.ResourceList); ok {
			data = rest_types.From.ResourceList(rl)
		}
		if r, ok := data.(model.Resource); ok {
			data = rest_types.From.Resource(r)
		}
		return yaml.NewPrinter().Print(data, out)
	case output.TableFormat:
		return table.Print(data, out)
	default:
		return errors.Errorf("unknown output format %q", format)
	}
}
