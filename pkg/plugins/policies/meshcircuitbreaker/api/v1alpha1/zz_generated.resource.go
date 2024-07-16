// Generated by tools/policy-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package v1alpha1

import (
	_ "embed"
	"errors"
	"fmt"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/core/resources/model"
)

//go:embed schema.yaml
var rawSchema []byte

func init() {
	var schema spec.Schema
	if rawSchema != nil {
		if err := yaml.Unmarshal(rawSchema, &schema); err != nil {
			panic(err)
		}
	}
	rawSchema = nil
	MeshCircuitBreakerResourceTypeDescriptor.Schema = &schema
}

const (
	MeshCircuitBreakerType model.ResourceType = "MeshCircuitBreaker"
)

var _ model.Resource = &MeshCircuitBreakerResource{}

type MeshCircuitBreakerResource struct {
	Meta model.ResourceMeta
	Spec *MeshCircuitBreaker
}

func NewMeshCircuitBreakerResource() *MeshCircuitBreakerResource {
	return &MeshCircuitBreakerResource{
		Spec: &MeshCircuitBreaker{},
	}
}

func (t *MeshCircuitBreakerResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *MeshCircuitBreakerResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *MeshCircuitBreakerResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *MeshCircuitBreakerResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*MeshCircuitBreaker)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &MeshCircuitBreaker{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *MeshCircuitBreakerResource) GetStatus() model.ResourceStatus {
	return nil
}

func (t *MeshCircuitBreakerResource) SetStatus(_ model.ResourceStatus) error {
	return errors.New("status not supported")
}

func (t *MeshCircuitBreakerResource) Descriptor() model.ResourceTypeDescriptor {
	return MeshCircuitBreakerResourceTypeDescriptor
}

func (t *MeshCircuitBreakerResource) Validate() error {
	if v, ok := interface{}(t).(interface{ validate() error }); !ok {
		return nil
	} else {
		return v.validate()
	}
}

var _ model.ResourceList = &MeshCircuitBreakerResourceList{}

type MeshCircuitBreakerResourceList struct {
	Items      []*MeshCircuitBreakerResource
	Pagination model.Pagination
}

func (l *MeshCircuitBreakerResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *MeshCircuitBreakerResourceList) GetItemType() model.ResourceType {
	return MeshCircuitBreakerType
}

func (l *MeshCircuitBreakerResourceList) NewItem() model.Resource {
	return NewMeshCircuitBreakerResource()
}

func (l *MeshCircuitBreakerResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*MeshCircuitBreakerResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*MeshCircuitBreakerResource)(nil), r)
	}
}

func (l *MeshCircuitBreakerResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (l *MeshCircuitBreakerResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var MeshCircuitBreakerResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                         MeshCircuitBreakerType,
	Resource:                     NewMeshCircuitBreakerResource(),
	ResourceList:                 &MeshCircuitBreakerResourceList{},
	Scope:                        model.ScopeMesh,
	KDSFlags:                     model.GlobalToAllZonesFlag | model.ZoneToGlobalFlag,
	WsPath:                       "meshcircuitbreakers",
	KumactlArg:                   "meshcircuitbreaker",
	KumactlListArg:               "meshcircuitbreakers",
	AllowToInspect:               true,
	IsPolicy:                     true,
	IsExperimental:               false,
	SingularDisplayName:          "Mesh Circuit Breaker",
	PluralDisplayName:            "Mesh Circuit Breakers",
	IsPluginOriginated:           true,
	IsTargetRefBased:             true,
	HasToTargetRef:               true,
	HasFromTargetRef:             true,
	HasStatus:                    false,
	AllowedOnSystemNamespaceOnly: false,
}
