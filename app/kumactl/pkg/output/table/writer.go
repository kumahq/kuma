package table

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type TableWriter interface {
	Headers(...string) error
	Row(...string) error
	Footer(string) error
	Flush() error
}

func NewWriter(output io.Writer) TableWriter {
	return &writer{
		out: tabwriter.NewWriter(output, 1, 0, 3, ' ', 0),
	}
}

var _ TableWriter = &writer{}

type writer struct {
	out *tabwriter.Writer
}

func (p *writer) Flush() error {
	return p.out.Flush()
}

func (p *writer) Headers(columns ...string) error {
	return p.Row(columns...)
}

func (p *writer) Row(columns ...string) error {
	_, err := fmt.Fprintf(p.out, "%s\n", strings.Join(columns, "\t"))
	return err
}

func (p *writer) Footer(footer string) error {
	_, err := fmt.Fprintf(p.out, "\n%s\n", footer)
	return err
}
