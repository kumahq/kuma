package json

import (
	"encoding/json"
	"io"

	"github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/pkg/output"
)

func NewPrinter() output.Printer {
	return &printer{}
}

var _ output.Printer = &printer{}

type printer struct{}

func (p *printer) Print(obj interface{}, out io.Writer) error {
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = out.Write(b)
	return err
}
