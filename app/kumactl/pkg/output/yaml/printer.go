package yaml

import (
	"bytes"
	"io"

	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
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
	// The common case is printing a single Kuma resource, in which
	// case showing the meta and then the spec is more readable than
	// showing fields in an arbitrary order. This partially addresses
	// https://github.com/kumahq/kuma/issues/679.
	if p.hasPrinted {
		if _, err := out.Write([]byte("---\n")); err != nil {
			return err
		}
	}
	p.hasPrinted = true
	switch obj := obj.(type) {
	case rest.Resource:
		if err := print(obj.GetMeta(), out); err != nil {
			return err
		}

		b, err := model.ToYAML(obj.GetSpec())
		if err != nil {
			return err
		}

		// Don't emit an empty YAML object that would cause a
		// subsequent parse failure.
		if len(b) == 0 || bytes.HasPrefix(b, []byte("{}")) {
			return nil
		}
		_, err = out.Write(b)
		return err
	default:
		return print(obj, out)
	}
}
