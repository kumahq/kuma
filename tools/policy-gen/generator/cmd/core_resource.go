package cmd

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/parse"
	"github.com/kumahq/kuma/tools/policy-gen/generator/pkg/save"
)

func newCoreResource(rootArgs *args) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "core-resource",
		Short: "Generate a core model resource for the policy",
		Long:  "Generate a core model resource for the policy.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			policyName := filepath.Base(rootArgs.pluginDir)
			policyPath := filepath.Join(rootArgs.pluginDir, "api", rootArgs.version, policyName+".go")
			if _, err := os.Stat(policyPath); err != nil {
				return err
			}

			pconfig, err := parse.Policy(policyPath)
			if err != nil {
				return err
			}

			outPath := filepath.Join(filepath.Dir(policyPath), "zz_generated.resource.go")
			return save.GoTemplate(resourceTemplate, pconfig, outPath)
		},
	}

	return cmd
}

// resourceTemplate for creating a Kuma resource.
var resourceTemplate = template.Must(template.New("resource").Parse(`
// Generated by tools/policy-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package {{.Package}}

import (
	_ "embed"
	"fmt"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

//go:embed schema.yaml
var rawSchema []byte

func init() {
	var schema spec.Schema
	if err := yaml.Unmarshal(rawSchema, &schema); err != nil {
		panic(err)
	}
	rawSchema = nil
	{{.Name}}ResourceTypeDescriptor.Schema = &schema
}

const (
	{{.Name}}Type model.ResourceType = "{{.Name}}"
)

var _ model.Resource = &{{.Name}}Resource{}

type {{.Name}}Resource struct {
	Meta model.ResourceMeta
	Spec *{{.Name}}
}

func New{{.Name}}Resource() *{{.Name}}Resource {
	return &{{.Name}}Resource{
		Spec: &{{.Name}}{},
	}
}

func (t *{{.Name}}Resource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *{{.Name}}Resource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *{{.Name}}Resource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *{{.Name}}Resource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*{{.Name}})
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &{{.Name}}{}
		} else  {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *{{.Name}}Resource) Descriptor() model.ResourceTypeDescriptor {
	return {{.Name}}ResourceTypeDescriptor 
}

func (t *{{.Name}}Resource) Validate() error {
	if v, ok := interface{}(t).(interface{ validate() error }); !ok {
		return nil
	} else {
		return v.validate()
	}
}

var _ model.ResourceList = &{{.Name}}ResourceList{}

type {{.Name}}ResourceList struct {
	Items      []*{{.Name}}Resource
	Pagination model.Pagination
}

func (l *{{.Name}}ResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *{{.Name}}ResourceList) GetItemType() model.ResourceType {
	return {{.Name}}Type
}

func (l *{{.Name}}ResourceList) NewItem() model.Resource {
	return New{{.Name}}Resource()
}

func (l *{{.Name}}ResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*{{.Name}}Resource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*{{.Name}}Resource)(nil), r)
	}
}

func (l *{{.Name}}ResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

var {{.Name}}ResourceTypeDescriptor = model.ResourceTypeDescriptor{
		Name: {{.Name}}Type,
		Resource: New{{.Name}}Resource(),
		ResourceList: &{{.Name}}ResourceList{},
		Scope: model.ScopeMesh,
		KDSFlags: model.FromGlobalToZone,
		WsPath: "{{.Path}}",
		KumactlArg: "{{index .AlternativeNames 0}}",
		KumactlListArg: "{{.Path}}",
		AllowToInspect: true,
		IsPolicy: true,
		IsExperimental: false,
		SingularDisplayName: "{{.SingularDisplayName}}",
		PluralDisplayName: "{{.PluralDisplayName}}",
		IsPluginOriginated: true,
		IsTargetRefBased: true,
	}
`))
