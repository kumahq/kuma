package table

import (
	"io"

	"github.com/hashicorp/go-multierror"
)

type Printer interface {
	Print(io.Writer) error
}

type Table struct {
	Headers    []string
	FooterFn   func(container interface{}) string
	RowForItem func(i int, container interface{}) ([]string, error)
}

func (p *Table) Print(data interface{}, out io.Writer) error {
	var allErr error
	table := NewWriter(out)
	defer table.Flush()

	if 0 < len(p.Headers) {
		if err := table.Headers(p.Headers...); err != nil {
			return err
		}
	}

	for i := 0; ; i++ {
		var row []string
		if p.RowForItem != nil {
			var err error
			row, err = p.RowForItem(i, data)
			if err != nil {
				allErr = multierror.Append(allErr, err)
			}
		}
		if row == nil {
			break
		}
		if err := table.Row(row...); err != nil {
			return err
		}
	}

	if p.FooterFn != nil {
		r := p.FooterFn(data)
		if r != "" {
			if err := table.Footer(p.FooterFn(data)); err != nil {
				return err
			}
		}
	}
	return allErr
}
