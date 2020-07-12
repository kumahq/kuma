package json

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
)

func NewPrinter() output.Printer {
	return &printer{}
}

var _ output.Printer = &printer{}

type printer struct{}

func (p *printer) Print(obj interface{}, out io.Writer) error {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}
	if _, err := out.Write(b); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out)
	return err
}
