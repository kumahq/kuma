package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	"text/template"

	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	structpb "google.golang.org/protobuf/types/known/structpb"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	_ "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/save"
	. "github.com/kumahq/kuma/tools/resource-gen/genutils"
)

// CustomResourceTemplate for creating a Kubernetes CRD to wrap a Kuma resource.
var CustomResourceTemplate = template.Must(template.New("custom-resource").Parse(`
// Generated by tools/resource-gen
// Run "make generate" to update this file.

{{ $pkg := printf "%s_proto" .Package }}
{{ $tk := "` + "`" + `" }}

// nolint:whitespace
package v1alpha1

import (
	"errors"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	{{ $pkg }} "github.com/kumahq/kuma/api/{{ .Package }}/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

{{range .Resources}}
{{- if not .SkipKubernetesWrappers }}

// +kubebuilder:object:root=true
{{- if .ScopeNamespace }}
// +kubebuilder:resource:categories=kuma,scope=Namespaced{{- if ne .ShortName ""}},shortName={{.ShortName}}{{- end}}
{{- else }}
// +kubebuilder:resource:categories=kuma,scope=Cluster{{- if ne .ShortName ""}},shortName={{.ShortName}}{{- end}}
{{- end}}
{{- range .AdditionalPrinterColumns }}
// +kubebuilder:printcolumn:{{ . }}
{{- end}}
type {{.ResourceType}} struct {
	metav1.TypeMeta   {{ $tk }}json:",inline"{{ $tk }}
	metav1.ObjectMeta {{ $tk }}json:"metadata,omitempty"{{ $tk }}

    // Mesh is the name of the Kuma mesh this resource belongs to.
	// It may be omitted for cluster-scoped resources.
	//
    // +kubebuilder:validation:Optional
	Mesh string {{ $tk }}json:"mesh,omitempty"{{ $tk }}

{{- if eq .ResourceType "DataplaneInsight" }}
	// Status is the status the Kuma resource.
    // +kubebuilder:validation:Optional
	Status   *apiextensionsv1.JSON {{ $tk }}json:"status,omitempty"{{ $tk }}
{{- else}}
	// Spec is the specification of the Kuma {{ .ProtoType }} resource.
    // +kubebuilder:validation:Optional
	Spec   *apiextensionsv1.JSON {{ $tk }}json:"spec,omitempty"{{ $tk }}
{{- end}}
}

// +kubebuilder:object:root=true
{{- if .ScopeNamespace }}
// +kubebuilder:resource:scope=Cluster
{{- else }}
// +kubebuilder:resource:scope=Namespaced
{{- end}}
type {{.ResourceType}}List struct {
	metav1.TypeMeta {{ $tk }}json:",inline"{{ $tk }}
	metav1.ListMeta {{ $tk }}json:"metadata,omitempty"{{ $tk }}
	Items           []{{.ResourceType}} {{ $tk }}json:"items"{{ $tk }}
}

{{- if not .SkipRegistration}}
func init() {
	SchemeBuilder.Register(&{{.ResourceType}}{}, &{{.ResourceType}}List{})
}
{{- end}}

func (cb *{{.ResourceType}}) GetObjectMeta() *metav1.ObjectMeta {
	return &cb.ObjectMeta
}

func (cb *{{.ResourceType}}) SetObjectMeta(m *metav1.ObjectMeta) {
	cb.ObjectMeta = *m
}

func (cb *{{.ResourceType}}) GetMesh() string {
	return cb.Mesh
}

func (cb *{{.ResourceType}}) SetMesh(mesh string) {
	cb.Mesh = mesh
}

func (cb *{{.ResourceType}}) GetSpec() (core_model.ResourceSpec, error) {
{{- if eq .ResourceType "DataplaneInsight" }}
	spec := cb.Status
{{- else}}
	spec := cb.Spec
{{- end}}
	m := {{$pkg}}.{{.ProtoType}}{}

    if spec == nil || len(spec.Raw) == 0 {
		return &m, nil
	}

	err := util_proto.FromJSON(spec.Raw, &m)
	return &m, err
}

func (cb *{{.ResourceType}}) SetSpec(spec core_model.ResourceSpec) {
	if spec == nil {
{{- if eq .ResourceType "DataplaneInsight" }}
		cb.Status = nil
{{- else }}
		cb.Spec = nil
{{- end }}
		return
	}

	s, ok := spec.(*{{$pkg}}.{{.ProtoType}}); 
	if !ok {
		panic(fmt.Sprintf("unexpected protobuf message type %T", spec))
	}

{{ if eq .ResourceType "DataplaneInsight" }}
	cb.Status = &apiextensionsv1.JSON{Raw: util_proto.MustMarshalJSON(s)}
{{- else}}
	cb.Spec = &apiextensionsv1.JSON{Raw: util_proto.MustMarshalJSON(s)}
{{- end}}
}

func (cb *{{.ResourceType}}) GetStatus() (core_model.ResourceStatus, error) {
	return nil, nil
}

func (cb *{{.ResourceType}}) SetStatus(_ core_model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (cb *{{.ResourceType}}) Scope() model.Scope {
{{- if .ScopeNamespace }}
	return model.ScopeNamespace
{{- else }}
	return model.ScopeCluster
{{- end }}
}

func (l *{{.ResourceType}}List) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

{{if not .SkipRegistration}}
func init() {
	registry.RegisterObjectType(&{{ $pkg }}.{{.ProtoType}}{}, &{{.ResourceType}}{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "{{.ResourceType}}",
		},
	})
	registry.RegisterListType(&{{ $pkg }}.{{.ProtoType}}{}, &{{.ResourceType}}List{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "{{.ResourceType}}List",
		},
	})
}
{{- end }} {{/* .SkipRegistration */}}
{{- end }} {{/* .SkipKubernetesWrappers */}}
{{- end }} {{/* Resources */}}
`))

// ResourceTemplate for creating a Kuma resource.
var ResourceTemplate = template.Must(template.New("resource").Funcs(map[string]any{"hasSuffix": strings.HasSuffix, "trimSuffix": strings.TrimSuffix}).Parse(`
// Generated by tools/resource-gen.
// Run "make generate" to update this file.

{{ $pkg := printf "%s_proto" .Package }}

// nolint:whitespace
package {{.Package}}

import (
	"errors"
	"fmt"

	{{$pkg}} "github.com/kumahq/kuma/api/{{.Package}}/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

{{range .Resources}}
{{ $baseType := trimSuffix (trimSuffix .ResourceType "Overview") "Insight" }}
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
		if protoType == nil {
			t.Spec = &{{$pkg}}.{{.ProtoType}}{}
		} else  {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *{{.ResourceName}}) GetStatus() model.ResourceStatus {
	return nil
}

func (t *{{.ResourceName}}) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *{{.ResourceName}}) Descriptor() model.ResourceTypeDescriptor {
	return {{.ResourceName}}TypeDescriptor 
}
{{- if and (hasSuffix .ResourceType "Overview") (ne $baseType "Service") }}

func (t *{{.ResourceName}}) SetOverviewSpec(resource model.Resource, insight model.Resource) error {
	t.SetMeta(resource.GetMeta())
	overview := &{{$pkg}}.{{.ProtoType}}{
		{{$baseType}}: resource.GetSpec().(*{{$pkg}}.{{$baseType}}),
	}
	if insight != nil {
		ins, ok := insight.GetSpec().(*{{$pkg}}.{{$baseType}}Insight)
		if !ok {
			return errors.New("failed to convert to insight type '{{$baseType}}Insight'")
		}
		overview.{{$baseType}}Insight = ins
	}
	return t.SetSpec(overview)
}
{{- end }}

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

func (l *{{.ResourceName}}List) SetPagination(p model.Pagination) {
	l.Pagination = p
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
		AllowToInspect: {{.AllowToInspect}},
		IsPolicy: {{.IsPolicy}},
		SingularDisplayName: "{{.SingularDisplayName}}",
		PluralDisplayName: "{{.PluralDisplayName}}",
		{{- if ne .ShortName "" }}
		ShortName: "{{.ShortName}}",{{- end}}
		IsExperimental: {{.IsExperimental}},
{{- if .HasInsights}}
		Insight: New{{.ResourceType}}InsightResource(),
		Overview: New{{.ResourceType}}OverviewResource(),
{{- end}}
	}

{{- if not .SkipRegistration}}
func init() {
	registry.RegisterType({{.ResourceName}}TypeDescriptor)
}
{{- end}}
{{end}}
`))

var rootDir = "."

func main() {
	var gen string
	var pkg string

	flag.StringVar(&gen, "generator", "", "the type of generator to run options: (type,crd,openapi)")
	flag.StringVar(&pkg, "package", "", "the name of the package to generate: (mesh, system)")
	flag.StringVar(&rootDir, "rootDir", "", "the root directory, default: .")

	flag.Parse()

	switch pkg {
	case "mesh", "system":
	default:
		log.Fatalf("package %s is not supported", pkg)
	}

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
	for _, t := range types {
		resourceInfo := ToResourceInfo(t.Descriptor())
		resources = append(resources, resourceInfo)
	}

	var generatorFn GeneratorFn

	switch gen {
	case "type":
		generatorFn = TemplateGeneratorFn(ResourceTemplate)
	case "crd":
		generatorFn = TemplateGeneratorFn(CustomResourceTemplate)
	case "openapi":
		generatorFn = openApiGenerator
	default:
		log.Fatalf("%s is not a valid generator option\n", gen)
	}

	err := generatorFn(pkg, resources)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
}

func openApiGenerator(pkg string, resources []ResourceInfo) error {
	// this is where the new types need to be added if we want to generate openAPI for it
	protoTypeToType := map[string]reflect.Type{
		"Mesh":        reflect.TypeOf(v1alpha1.Mesh{}),
		"MeshGateway": reflect.TypeOf(v1alpha1.MeshGateway{}),
	}

	for _, r := range resources {
		tpe, exists := protoTypeToType[r.ResourceType]
		if !exists {
			continue
		}
		reflector := jsonschema.Reflector{
			ExpandedStruct:            true,
			DoNotReference:            true,
			AllowAdditionalProperties: true,
			IgnoredTypes:              []any{structpb.Struct{}},
		}
		err := reflector.AddGoComments("github.com/kumahq/kuma/", path.Join(rootDir, "api/"))
		if err != nil {
			return err
		}
		schemaMap := orderedmap.New[string, *jsonschema.Schema]()
		schemaMap.Set("type", &jsonschema.Schema{Type: "string"})
		schemaMap.Set("name", &jsonschema.Schema{Type: "string"})
		if !r.Global {
			schemaMap.Set("mesh", &jsonschema.Schema{Type: "string"})
		}
		schemaMap.Set("labels", &jsonschema.Schema{Type: "object", AdditionalProperties: &jsonschema.Schema{Type: "string"}})
		properties := reflector.ReflectFromType(tpe).Properties
		for pair := properties.Oldest(); pair != nil; pair = pair.Next() {
			schemaMap.Set(pair.Key, pair.Value)
		}

		schema := jsonschema.Schema{
			Type:       "object",
			Required:   []string{"type", "name"},
			Properties: schemaMap,
		}

		if !r.Global {
			schema.Required = append(schema.Required, "mesh")
		}

		out, err := yaml.Marshal(schema)
		if err != nil {
			return err
		}

		outDir := path.Join(rootDir, "api", pkg, "v1alpha1", strings.ToLower(r.ResourceType))

		// Ensure the directory exists
		err = os.MkdirAll(outDir, 0o755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		err = os.WriteFile(path.Join(outDir, "schema.yaml"), out, 0o600)
		if err != nil {
			return err
		}

		templatePath := path.Join(rootDir, "tools", "openapi", "templates", "endpoints.yaml")
		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			return err
		}
		scope := "Mesh"
		if r.Global {
			scope = "Global"
		}
		opts := map[string]interface{}{
			"Package": "v1alpha1",
			"Name":    r.ResourceType,
			"Scope":   scope,
			"Path":    r.WsPath,
		}
		err = save.PlainTemplate(tmpl, opts, path.Join(outDir, "rest.yaml"))
		if err != nil {
			return err
		}
	}
	return nil
}

type GeneratorFn func(pkg string, resources []ResourceInfo) error

func TemplateGeneratorFn(tmpl *template.Template) GeneratorFn {
	return func(pkg string, resources []ResourceInfo) error {
		outBuf := bytes.Buffer{}
		if err := tmpl.Execute(&outBuf, struct {
			Package   string
			Resources []ResourceInfo
		}{
			Package:   pkg,
			Resources: resources,
		}); err != nil {
			return err
		}

		out, err := format.Source(outBuf.Bytes())
		if err != nil {
			return err
		}
		if _, err := os.Stdout.Write(out); err != nil {
			return err
		}
		return nil
	}
}
