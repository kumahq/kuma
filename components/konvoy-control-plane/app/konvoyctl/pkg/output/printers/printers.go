package printers

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output/json"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output/table"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output/yaml"

	"github.com/pkg/errors"
)

type Table = table.Table

var (
	NewTablePrinter = table.NewPrinter
)

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
