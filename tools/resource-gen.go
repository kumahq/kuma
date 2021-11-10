package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"html/template"
	"os"
	"path"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/kumahq/kuma/api/mesh"
	_ "github.com/kumahq/kuma/api/mesh/v1alpha1"
)

// ResourceTemplate for creating a Kuma resource.
var ResourceTemplate = template.Must(template.New("resource").Parse(`
// Generated by tools/resource-gen.
// Run "make generate" to update this file.

{{ $pkg := printf "%s_proto" .Package }}

// nolint:whitespace
package {{.Package}}

import (
	"fmt"

	{{$pkg}} "github.com/kumahq/kuma/api/{{.Package}}/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

{{range .Resources}}
const (
	{{.ResourceType}}Type model.ResourceType = "{{.ResourceType}}"
)

var _ model.Resource = &{{.ResourceName}}{}

type {{.ResourceName}} struct {
	Meta model.ResourceMeta
	Spec *{{$pkg}}.{{.ProtoType}}
}

func New{{.ResourceName}}() *{{.ResourceName}} {
	return &{{.ResourceName}}{
		Spec: &{{$pkg}}.{{.ProtoType}}{},
	}
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
func (t *{{$in.ResourceName}}) {{.}}() []*{{$pkg}}.Selector {
	return t.Spec.Get{{.}}()
}
{{end}}
{{end}}

func (t *{{.ResourceName}}) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*{{$pkg}}.{{.ProtoType}})
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		t.Spec = protoType
		return nil
	}
}

func (t *{{.ResourceName}}) Descriptor() model.ResourceTypeDescriptor {
	return {{.ResourceName}}TypeDescriptor 
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

var {{.ResourceName}}TypeDescriptor = model.ResourceTypeDescriptor{
		Name: {{.ResourceType}}Type,
		Resource: New{{.ResourceName}}(),
		ResourceList: &{{.ResourceName}}List{},
		ReadOnly: {{.WsReadOnly}},
		AdminOnly: {{.WsAdminOnly}},
		Scope: {{if .Global}}model.ScopeGlobal{{else}}model.ScopeMesh{{end}},
		{{- if ne .KdsDirection ""}}
		KDSFlags: {{.KdsDirection}},
		{{- end}}
		WsPath: "{{.WsPath}}",
		KumactlArg: "{{.KumactlSingular}}",
		KumactlListArg: "{{.KumactlPlural}}",
	}

{{- if not .SkipRegistration}}
func init() {
	registry.RegisterType({{.ResourceName}}TypeDescriptor)
}
{{- end}}
{{end}}
`))

// KumaResourceForMessage fetches the Kuma resource option out of a message.
func KumaResourceForMessage(m protoreflect.MessageType) *mesh.KumaResourceOptions {
	ext := proto.GetExtension(m.Descriptor().Options(), mesh.E_Resource)
	var resOption *mesh.KumaResourceOptions
	if r, ok := ext.(*mesh.KumaResourceOptions); ok {
		resOption = r
	}

	return resOption
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

type ResourceInfo struct {
	ResourceName     string
	ResourceType     string
	ProtoType        string
	Selectors        []string
	SkipRegistration bool
	SkipValidation   bool
	Global           bool
	KumactlSingular  string
	KumactlPlural    string
	WsReadOnly       bool
	WsAdminOnly      bool
	WsPath           string
	KdsDirection     string
}

func ToResourceInfo(m protoreflect.MessageType) ResourceInfo {
	r := KumaResourceForMessage(m)

	out := ResourceInfo{
		ResourceType:     r.Type,
		ResourceName:     r.Name,
		ProtoType:        string(m.Descriptor().Name()),
		Selectors:        SelectorsForMessage(m.Descriptor()),
		SkipRegistration: r.SkipRegistration,
		SkipValidation:   r.SkipValidation,
		Global:           r.Global,
	}
	if r.Ws != nil {
		pluralResourceName := r.Ws.Plural
		if pluralResourceName == "" {
			pluralResourceName = r.Ws.Name + "s"
		}
		out.WsReadOnly = r.Ws.ReadOnly
		out.WsAdminOnly = r.Ws.AdminOnly
		out.WsPath = pluralResourceName
		if !r.Ws.ReadOnly {
			out.KumactlSingular = r.Ws.Name
			out.KumactlPlural = pluralResourceName
			// Keep the typo to preserve backward compatibility
			if out.KumactlSingular == "health-check" {
				out.KumactlSingular = "healthcheck"
				out.KumactlPlural = "healthchecks"
			}
		}
	}
	switch {
	case r.Kds == nil || (!r.Kds.SendToZone && !r.Kds.SendToGlobal):
		out.KdsDirection = ""
	case r.Kds.SendToGlobal && r.Kds.SendToZone:
		out.KdsDirection = "model.FromZoneToGlobal | model.FromGlobalToZone"
	case r.Kds.SendToGlobal:
		out.KdsDirection = "model.FromZoneToGlobal"
	case r.Kds.SendToZone:
		out.KdsDirection = "model.FromGlobalToZone"
	}

	if p := m.Descriptor().Parent(); p != nil {
		if _, ok := p.(protoreflect.MessageDescriptor); ok {
			out.ProtoType = fmt.Sprintf("%s_%s", p.Name(), m.Descriptor().Name())
		}
	}
	return out
}

// ProtoMessageFunc ...
type ProtoMessageFunc func(protoreflect.MessageType) bool

// OnKumaResourceMessage ...
func OnKumaResourceMessage(pkg string, f ProtoMessageFunc) ProtoMessageFunc {
	return func(m protoreflect.MessageType) bool {
		r := KumaResourceForMessage(m)
		if r == nil {
			return true
		}

		if r.Package == pkg {
			return f(m)
		}

		return true
	}
}

func main() {
	var gen string
	flag.StringVar(&gen, "generator", "", "the type of generator to run options: (type,ws,kumactl,kds)")
	flag.Parse()

	for _, pkg := range []string{"system", "mesh"} {
		var types []protoreflect.MessageType
		protoregistry.GlobalTypes.RangeMessages(
			OnKumaResourceMessage(pkg, func(m protoreflect.MessageType) bool {
				types = append(types, m)
				return true
			}))

		// Sort by name so the output is deterministic.
		sort.Slice(types, func(i, j int) bool {
			return types[i].Descriptor().FullName() < types[j].Descriptor().FullName()
		})

		var resources []ResourceInfo
		apisPath := fmt.Sprintf("pkg/core/resources/apis/%s", pkg)
		for _, t := range types {
			resourceInfo := ToResourceInfo(t)
			resources = append(resources, resourceInfo)
		}

		globalTemplates := map[string]*template.Template{}
		switch gen {
		case "type":
			globalTemplates[path.Join(apisPath, "generated_resources.go")] = ResourceTemplate
		default:
			panic(fmt.Sprintf("%s is not a valid generator option\n", gen))
		}

		for filename, tmpl := range globalTemplates {
			outBuf := bytes.Buffer{}
			if err := tmpl.Execute(&outBuf, struct {
				Package   string
				Resources []ResourceInfo
			}{
				Package:   pkg,
				Resources: resources,
			}); err != nil {
				panic(fmt.Sprintf("template error: %s", err))
			}
			out, err := format.Source(outBuf.Bytes())
			if err != nil {
				panic(fmt.Sprintf("%s\n", err))
			}
			if err := os.WriteFile(filename, out, os.FileMode(0644)); err != nil {
				panic(fmt.Sprintf("%s\n", err))
			}
		}
	}
}
