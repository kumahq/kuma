package printers

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/json"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/yaml"
)

type Table = table.Table

var NewTablePrinter = table.NewPrinter

func NewGenericPrinter(format output.Format) (output.Printer, error) {
	switch format {
	case output.JSONFormat:
		return json.NewPrinter(), nil
	case output.YAMLFormat:
		return yaml.NewPrinter(), nil
	default:
		return nil, errors.Errorf("unknown output format %q", format)
	}
}
