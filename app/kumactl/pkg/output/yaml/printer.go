package yaml

import (
	"encoding/json"
	"io"

	"sigs.k8s.io/yaml/goyaml.v2"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
)

func NewPrinter() output.Printer {
	return &printer{}
}

var _ output.Printer = &printer{}

type printer struct {
	// hasPrinted is used to add yaml separators `---` when printing multiple objects with the same printer
	hasPrinted bool
}

func print(obj interface{}, out io.Writer) error {
	b, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = out.Write(b)
	return err
}

func (p *printer) Print(obj interface{}, out io.Writer) error {
	if p.hasPrinted {
		if _, err := out.Write([]byte("---\n")); err != nil {
			return err
		}
	}
	p.hasPrinted = true

	// The common case is printing a single Kuma resource, in which
	// case showing the meta and then the spec is more readable than
	// showing fields in an arbitrary order. This partially addresses
	// https://github.com/kumahq/kuma/issues/679.

	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return err
	}

	if obj == nil {
		var m interface{}
		if err := yaml.Unmarshal(b, &m); err != nil {
			return err
		}
		return print(m, out)
	}

	var m yaml.MapSlice
	if err := yaml.Unmarshal(b, &m); err != nil {
		return err
	}
	return print(m, out)
}
