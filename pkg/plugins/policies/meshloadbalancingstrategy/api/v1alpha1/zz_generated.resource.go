// Generated by tools/policy-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package v1alpha1

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

//go:embed schema.yaml
var rawSchema []byte

func init() {
	var structuralSchema *schema.Structural
	var schemaObject spec.Schema
	var v1JsonSchemaProps *apiextensionsv1.JSONSchemaProps
	if rawSchema != nil {
		if err := yaml.Unmarshal(rawSchema, &schemaObject); err != nil {
			panic(err)
		}
		rawJson, err := yaml.YAMLToJSON(rawSchema)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(rawJson, &v1JsonSchemaProps); err != nil {
			panic(err)
		}
		var jsonSchemaProps apiextensions.JSONSchemaProps
		err = apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(v1JsonSchemaProps, &jsonSchemaProps, nil)
		if err != nil {
			panic(err)
		}
		structuralSchema, err = schema.NewStructural(&jsonSchemaProps)
		if err != nil {
			panic(err)
		}
	}
	rawSchema = nil
	MeshLoadBalancingStrategyResourceTypeDescriptor.Schema = &schemaObject
	MeshLoadBalancingStrategyResourceTypeDescriptor.StructuralSchema = structuralSchema
}

const (
	MeshLoadBalancingStrategyType model.ResourceType = "MeshLoadBalancingStrategy"
)

var _ model.Resource = &MeshLoadBalancingStrategyResource{}

type MeshLoadBalancingStrategyResource struct {
	Meta model.ResourceMeta
	Spec *MeshLoadBalancingStrategy
}

func NewMeshLoadBalancingStrategyResource() *MeshLoadBalancingStrategyResource {
	return &MeshLoadBalancingStrategyResource{
		Spec: &MeshLoadBalancingStrategy{},
	}
}

func (t *MeshLoadBalancingStrategyResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *MeshLoadBalancingStrategyResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *MeshLoadBalancingStrategyResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *MeshLoadBalancingStrategyResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*MeshLoadBalancingStrategy)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &MeshLoadBalancingStrategy{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *MeshLoadBalancingStrategyResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *MeshLoadBalancingStrategyResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *MeshLoadBalancingStrategyResource) Descriptor() model.ResourceTypeDescriptor {
	return MeshLoadBalancingStrategyResourceTypeDescriptor
}

func (t *MeshLoadBalancingStrategyResource) Validate() error {
	if v, ok := interface{}(t).(interface{ validate() error }); !ok {
		return nil
	} else {
		return v.validate()
	}
}

var _ model.ResourceList = &MeshLoadBalancingStrategyResourceList{}

type MeshLoadBalancingStrategyResourceList struct {
	Items      []*MeshLoadBalancingStrategyResource
	Pagination model.Pagination
}

func (l *MeshLoadBalancingStrategyResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *MeshLoadBalancingStrategyResourceList) GetItemType() model.ResourceType {
	return MeshLoadBalancingStrategyType
}

func (l *MeshLoadBalancingStrategyResourceList) NewItem() model.Resource {
	return NewMeshLoadBalancingStrategyResource()
}

func (l *MeshLoadBalancingStrategyResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*MeshLoadBalancingStrategyResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*MeshLoadBalancingStrategyResource)(nil), r)
	}
}

func (l *MeshLoadBalancingStrategyResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (l *MeshLoadBalancingStrategyResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var MeshLoadBalancingStrategyResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                         MeshLoadBalancingStrategyType,
	Resource:                     NewMeshLoadBalancingStrategyResource(),
	ResourceList:                 &MeshLoadBalancingStrategyResourceList{},
	Scope:                        model.ScopeMesh,
	KDSFlags:                     model.GlobalToAllZonesFlag | model.ZoneToGlobalFlag | model.GlobalToAllButOriginalZoneFlag,
	WsPath:                       "meshloadbalancingstrategies",
	KumactlArg:                   "meshloadbalancingstrategy",
	KumactlListArg:               "meshloadbalancingstrategies",
	AllowToInspect:               true,
	IsPolicy:                     true,
	IsExperimental:               false,
	SingularDisplayName:          "Mesh Load Balancing Strategy",
	PluralDisplayName:            "Mesh Load Balancing Strategies",
	IsPluginOriginated:           true,
	IsTargetRefBased:             true,
	HasToTargetRef:               true,
	HasFromTargetRef:             false,
	HasRulesTargetRef:            false,
	HasStatus:                    false,
	AllowedOnSystemNamespaceOnly: false,
	IsReferenceableInTo:          false,
	ShortName:                    "mlbs",
	InterpretFromEntriesAsRules:  false,
}
