package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"html/template"
	"os"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/kumahq/kuma/api/mesh"
	_ "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

// Package to generate code into.
var Package string

// Output of type templating.
var Output = &bytes.Buffer{}

// ResourceTemplate for creating a Kuma resource.
var ResourceTemplate = template.Must(template.New("resource").Parse(`
const (
	{{.ResourceType}}Type model.ResourceType = "{{.ResourceType}}"
)

var _ model.Resource = &{{.ResourceName}}{}

type {{.ResourceName}} struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.{{.ProtoType}}
}

func New{{.ResourceName}}() *{{.ResourceName}} {
	return &{{.ResourceName}}{
		Spec: &mesh_proto.{{.ProtoType}}{},
	}
}

func (t *{{.ResourceName}}) GetType() model.ResourceType {
	return {{.ResourceType}}Type
}

func (t *{{.ResourceName}}) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *{{.ResourceName}}) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *{{.ResourceName}}) GetSpec() model.ResourceSpec {
	return t.Spec
}

{{if .SkipValidation}}
func (t *{{.ResourceName}}) Validate() error {
	return nil
}
{{end}}

{{with $in := .}}
{{range .Selectors}}
func (t *{{$in.ResourceName}}) {{.}}() []*mesh_proto.Selector {
	return t.Spec.Get{{.}}()
}
{{end}}
{{end}}

func (t *{{.ResourceName}}) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*mesh_proto.{{.ProtoType}})
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		t.Spec = protoType
		return nil
	}
}

func (t *{{.ResourceName}}) Scope() model.ResourceScope {
{{if .Global}}
	return model.ScopeGlobal
{{else}}
	return model.ScopeMesh
{{end}}
}

var _ model.ResourceList = &{{.ResourceName}}List{}

type {{.ResourceName}}List struct {
	Items      []*{{.ResourceName}}
	Pagination model.Pagination
}

func (l *{{.ResourceName}}List) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *{{.ResourceName}}List) GetItemType() model.ResourceType {
	return {{.ResourceType}}Type
}

func (l *{{.ResourceName}}List) NewItem() model.Resource {
	return New{{.ResourceName}}()
}

func (l *{{.ResourceName}}List) AddItem(r model.Resource) error {
	if trr, ok := r.(*{{.ResourceName}}); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*{{.ResourceName}})(nil), r)
	}
}

func (l *{{.ResourceName}}List) GetPagination() *model.Pagination {
	return &l.Pagination
}

{{if not .SkipRegistration}}
func init() {
	registry.RegisterType(New{{.ResourceName}}())
	registry.RegistryListType(&{{.ResourceName}}List{})
}
{{end}}
`))

// KumaResourceForMessage fishes the Kuma resource option out of a message.
func KumaResourceForMessage(m protoreflect.MessageType) *mesh.KumaResourceOptions {
	ext := proto.GetExtension(m.Descriptor().Options(), mesh.E_Resource)
	if r, ok := ext.(*mesh.KumaResourceOptions); ok {
		return r
	}

	return nil
}

// SelectorsForMessage finds all the top-level fields in the message are
// repeated selectors. We want to generate convenience accessors for these.
func SelectorsForMessage(m protoreflect.MessageDescriptor) []string {
	var selectors []string
	fields := m.Fields()

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		m := field.Message()
		if m != nil && m.FullName() == "kuma.mesh.v1alpha1.Selector" {
			fieldName := string(field.Name())
			selectors = append(selectors, strings.Title(fieldName))
		}
	}

	return selectors
}

// GenerateResource generates a Kuma resource for this message.
func GenerateResource(m protoreflect.MessageType) bool {
	r := KumaResourceForMessage(m)

	values := struct {
		ResourceName     string
		ResourceType     string
		ProtoType        string
		Selectors        []string
		SkipRegistration bool
		SkipValidation   bool
		Global           bool
	}{
		ResourceType:     r.Type,
		ResourceName:     r.Name,
		ProtoType:        string(m.Descriptor().Name()),
		Selectors:        SelectorsForMessage(m.Descriptor().(protoreflect.MessageDescriptor)),
		SkipRegistration: r.SkipRegistration,
		SkipValidation:   r.SkipValidation,
		Global:           r.Global,
	}

	if p := m.Descriptor().Parent(); p != nil {
		if _, ok := p.(protoreflect.MessageDescriptor); ok {
			values.ProtoType = fmt.Sprintf("%s_%s", p.Name(), m.Descriptor().Name())
		}
	}

	if err := ResourceTemplate.Execute(Output, values); err != nil {
		panic(fmt.Sprintf("template error: %s", err))
	}

	return true
}

// ProtoMessageFunc ...
type ProtoMessageFunc func(protoreflect.MessageType) bool

// OnKumaResourceMessage ...
func OnKumaResourceMessage(f ProtoMessageFunc) ProtoMessageFunc {
	return func(m protoreflect.MessageType) bool {
		r := KumaResourceForMessage(m)
		if r == nil {
			return true
		}

		if r.Package == Package {
			return f(m)
		}

		return true
	}
}

func main() {
	flag.StringVar(&Package, "package", "mesh", "Package to generate")
	flag.Parse()

	fmt.Fprintf(Output, `
// Generated by tools/resource-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package %s

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/%s/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

`, Package, Package)

	var types []protoreflect.MessageType

	protoregistry.GlobalTypes.RangeMessages(
		OnKumaResourceMessage(func(m protoreflect.MessageType) bool {
			types = append(types, m)
			return true
		}))

	// Sort by name so the output is deterministic.
	sort.Slice(types, func(i, j int) bool {
		return types[i].Descriptor().FullName() < types[j].Descriptor().FullName()
	})

	for _, t := range types {
		GenerateResource(t)
	}

	out, err := format.Source(Output.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s", string(out))
}
