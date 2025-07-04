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
{{- if not .HasStatus }}
	"errors"
{{- end }}
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/core/resources/model"
{{- if .IsDestination }}
    "github.com/kumahq/kuma/pkg/core/resources/apis/core"
{{- end }}
)

{{- if not .SkipRegistration }}
//go:embed schema.yaml
{{- end }}
var rawSchema []byte

func init() {
	var structuralSchema *schema.Structural
	var v1JsonSchemaProps *apiextensionsv1.JSONSchemaProps
	var validator *validate.SchemaValidator
	if rawSchema != nil {
		if err := yaml.Unmarshal(rawSchema, &v1JsonSchemaProps); err != nil {
			panic(err)
		}
		var jsonSchemaProps apiextensions.JSONSchemaProps
		err := apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(v1JsonSchemaProps, &jsonSchemaProps, nil)
		if err != nil {
			panic(err)
		}
		structuralSchema, err = schema.NewStructural(&jsonSchemaProps)
		if err != nil {
			panic(err)
		}
		schemaObject := structuralSchema.ToKubeOpenAPI()
		validator = validate.NewSchemaValidator(schemaObject, nil, "", strfmt.Default)
	}
	rawSchema = nil
	{{.Name}}ResourceTypeDescriptor.Validator = validator
	{{.Name}}ResourceTypeDescriptor.StructuralSchema = structuralSchema
}

const (
	{{.Name}}Type model.ResourceType = "{{.Name}}"
)

var _ model.Resource = &{{.Name}}Resource{}

type {{.Name}}Resource struct {
	Meta model.ResourceMeta
	Spec *{{.Name}}
{{- if .HasStatus }}
	Status *{{.Name}}Status
{{- end }}
}

func New{{.Name}}Resource() *{{.Name}}Resource {
	return &{{.Name}}Resource{
		Spec: &{{.Name}}{},
{{- if .HasStatus }}
		Status: &{{.Name}}Status{},
{{- end }}
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

{{ if .HasStatus }}
func (t *{{.Name}}Resource) GetStatus() model.ResourceStatus {
	return t.Status
}

func (t *{{.Name}}Resource) SetStatus(status model.ResourceStatus) error {
	protoType, ok := status.(*{{.Name}}Status)
	if !ok {
		return fmt.Errorf("invalid type %T for Status", status)
	} else {
		if protoType == nil {
			t.Status = &{{.Name}}Status{}
		} else  {
			t.Status = protoType
		}
		return nil
	}
}
{{ else }}
func (t *{{.Name}}Resource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *{{.Name}}Resource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}
{{ end }}

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

func (l *{{.Name}}ResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

{{- if .IsDestination }}
var _ core.Destination = &{{.Name}}Resource{}
{{- end }}

var {{.Name}}ResourceTypeDescriptor = model.ResourceTypeDescriptor{
		Name: {{.Name}}Type,
		Resource: New{{.Name}}Resource(),
		ResourceList: &{{.Name}}ResourceList{},
		Scope: model.Scope{{.Scope}},
		KDSFlags: {{.KDSFlags}},
		WsPath: "{{.Path}}",
		KumactlArg: "{{.NameLower}}",
		KumactlListArg: "{{.Path}}",
		AllowToInspect: {{.IsPolicy}},
		IsPolicy: {{.IsPolicy}},
        IsDestination: {{.IsDestination}},
		IsExperimental: false,
		SingularDisplayName: "{{.SingularDisplayName}}",
		PluralDisplayName: "{{.PluralDisplayName}}",
		IsPluginOriginated: true,
		IsTargetRefBased: {{.IsPolicy}},
		HasToTargetRef: {{.HasTo}},
		HasFromTargetRef: {{.HasFrom}},
        HasRulesTargetRef: {{.HasRules}},
		HasStatus: {{.HasStatus}},
		AllowedOnSystemNamespaceOnly: {{.AllowedOnSystemNamespaceOnly}},
		IsReferenceableInTo: {{.IsReferenceableInTo}},
		ShortName: "{{.ShortName}}",
		IsFromAsRules: {{.IsFromAsRules}},
	}
`))
