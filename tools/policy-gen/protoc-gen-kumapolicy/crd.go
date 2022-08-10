package main

import (
	"bytes"
	"fmt"
	"go/format"
	"html/template"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/compiler/protogen"
)

// CustomResourceTemplate for creating a Kubernetes CRD to wrap a Kuma resource.
var CustomResourceTemplate = template.Must(template.New("custom-resource").Parse(`
// Generated by tools/resource-gen
// Run "make generate" to update this file.

{{ $tk := "` + "`" + `" }}

// nolint:whitespace
package {{.PolicyVersion}}

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"


	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	policy "github.com/kumahq/kuma/pkg/plugins/policies/{{.Package}}/api/{{.PolicyVersion}}"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/model"
	{{- if not .SkipRegistration }}
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s/native/pkg/registry"
	{{- end }}
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=kuma,scope=Namespaced
type {{.Name}} struct {
	metav1.TypeMeta   {{ $tk }}json:",inline"{{ $tk }}
	metav1.ObjectMeta {{ $tk }}json:"metadata,omitempty"{{ $tk }}

	// Spec is the specification of the Kuma {{ .Name }} resource.
    // +kubebuilder:validation:Optional
	Spec   *policy.{{.Name}} {{ $tk }}json:"spec,omitempty"{{ $tk }}
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
type {{.Name}}List struct {
	metav1.TypeMeta {{ $tk }}json:",inline"{{ $tk }}
	metav1.ListMeta {{ $tk }}json:"metadata,omitempty"{{ $tk }}
	Items           []{{.Name}} {{ $tk }}json:"items"{{ $tk }}
}

func (cb *{{.Name}}) GetObjectMeta() *metav1.ObjectMeta {
	return &cb.ObjectMeta
}

func (cb *{{.Name}}) SetObjectMeta(m *metav1.ObjectMeta) {
	cb.ObjectMeta = *m
}

func (cb *{{.Name}}) GetMesh() string {
	if mesh, ok := cb.ObjectMeta.Labels[metadata.KumaMeshLabel]; ok {
		return mesh
	} else {
		return core_model.DefaultMesh
	}
}

func (cb *{{.Name}}) SetMesh(mesh string) {
	if cb.ObjectMeta.Labels == nil {
		cb.ObjectMeta.Labels = map[string]string{}
	}
	cb.ObjectMeta.Labels[metadata.KumaMeshLabel] = mesh
}

func (cb *{{.Name}}) GetSpec() (proto.Message, error) {
	return cb.Spec, nil
}

func (cb *{{.Name}}) SetSpec(spec proto.Message) {
	if spec == nil {
		cb.Spec = nil
		return
	}

	if _, ok := spec.(*policy.{{.Name}}); !ok {
		panic(fmt.Sprintf("unexpected protobuf message type %T", spec))
	}

	cb.Spec = spec.(*policy.{{.Name}})
}

func (cb *{{.Name}}) Scope() model.Scope {
	return model.ScopeNamespace
}

func (l *{{.Name}}List) GetItems() []model.KubernetesObject {
	result := make([]model.KubernetesObject, len(l.Items))
	for i := range l.Items {
		result[i] = &l.Items[i]
	}
	return result
}

{{if not .SkipRegistration}}
func init() {
	SchemeBuilder.Register(&{{.Name}}{}, &{{.Name}}List{})
	registry.RegisterObjectType(&policy.{{.Name}}{}, &{{.Name}}{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "{{.Name}}",
		},
	})
	registry.RegisterListType(&policy.{{.Name}}{}, &{{.Name}}List{
		TypeMeta: metav1.TypeMeta{
			APIVersion: GroupVersion.String(),
			Kind:       "{{.Name}}List",
		},
	})
}
{{- end }} {{/* .SkipRegistration */}}
`))

var GroupVersionInfoTemplate = template.Must(template.New("groupversion-info").Parse(`
// Package {{.PolicyVersion}} contains API Schema definitions for the mesh {{.PolicyVersion}} API group
// +groupName=kuma.io
package {{.PolicyVersion}}

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "kuma.io", Version: "{{.PolicyVersion}}"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
`))

func generateCRD(
	p *protogen.Plugin,
	file *protogen.File,
) error {
	var infos []PolicyConfig
	for _, msg := range file.Messages {
		infos = append(infos, NewPolicyConfig(msg.Desc))
	}

	if len(infos) > 1 {
		return errors.Errorf("only one Kuma resource per proto file is allowed")
	}

	info := infos[0]

	outBuf := bytes.Buffer{}
	if err := CustomResourceTemplate.Execute(&outBuf, struct {
		Name             string
		PolicyVersion    string
		Package          string
		SkipRegistration bool
	}{
		PolicyVersion:    string(file.GoPackageName),
		Package:          strings.ToLower(info.Name),
		Name:             info.Name,
		SkipRegistration: info.SkipRegistration,
	}); err != nil {
		return err
	}

	out, err := format.Source(outBuf.Bytes())
	if err != nil {
		return err
	}

	typesGenerator := p.NewGeneratedFile(fmt.Sprintf("k8s/%s/zz_generated.types.go", string(file.GoPackageName)), file.GoImportPath)
	if _, err := typesGenerator.Write(out); err != nil {
		return err
	}

	gviOutBuf := bytes.Buffer{}
	if err := GroupVersionInfoTemplate.Execute(&gviOutBuf, struct {
		PolicyVersion string
	}{
		PolicyVersion: string(file.GoPackageName),
	}); err != nil {
		return err
	}

	gvi, err := format.Source(gviOutBuf.Bytes())
	if err != nil {
		return err
	}

	gviGenerator := p.NewGeneratedFile(fmt.Sprintf("k8s/%s/groupversion_info.go", string(file.GoPackageName)), file.GoImportPath)
	if _, err := gviGenerator.Write(gvi); err != nil {
		return err
	}
	return nil
}
