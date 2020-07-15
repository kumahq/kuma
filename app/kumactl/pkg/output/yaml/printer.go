package yaml

import (
	"io"

	"github.com/ghodss/yaml"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
)

func NewPrinter() output.Printer {
	return &printer{}
}

var _ output.Printer = &printer{}

type printer struct{}

func (p *printer) Print(obj interface{}, out io.Writer) error {
	b, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = out.Write(b)
	return err
}
