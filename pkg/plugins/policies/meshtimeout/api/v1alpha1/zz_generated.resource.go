// Generated by tools/policy-gen.
// Run "make generate" to update this file.

// nolint:whitespace
package v1alpha1

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
	MeshTimeoutResourceTypeDescriptor.Schema = &schema
}

const (
	MeshTimeoutType model.ResourceType = "MeshTimeout"
)

var _ model.Resource = &MeshTimeoutResource{}

type MeshTimeoutResource struct {
	Meta model.ResourceMeta
	Spec *MeshTimeout
}

func NewMeshTimeoutResource() *MeshTimeoutResource {
	return &MeshTimeoutResource{
		Spec: &MeshTimeout{},
	}
}

func (t *MeshTimeoutResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *MeshTimeoutResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *MeshTimeoutResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *MeshTimeoutResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*MeshTimeout)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &MeshTimeout{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *MeshTimeoutResource) Descriptor() model.ResourceTypeDescriptor {
	return MeshTimeoutResourceTypeDescriptor
}

func (t *MeshTimeoutResource) Validate() error {
	if v, ok := interface{}(t).(interface{ validate() error }); !ok {
		return nil
	} else {
		return v.validate()
	}
}

var _ model.ResourceList = &MeshTimeoutResourceList{}

type MeshTimeoutResourceList struct {
	Items      []*MeshTimeoutResource
	Pagination model.Pagination
}

func (l *MeshTimeoutResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *MeshTimeoutResourceList) GetItemType() model.ResourceType {
	return MeshTimeoutType
}

func (l *MeshTimeoutResourceList) NewItem() model.Resource {
	return NewMeshTimeoutResource()
}

func (l *MeshTimeoutResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*MeshTimeoutResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*MeshTimeoutResource)(nil), r)
	}
}

func (l *MeshTimeoutResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (l *MeshTimeoutResourceList) SetPagination(p model.Pagination) {
	l.Pagination = p
}

var MeshTimeoutResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                MeshTimeoutType,
	Resource:            NewMeshTimeoutResource(),
	ResourceList:        &MeshTimeoutResourceList{},
	Scope:               model.ScopeMesh,
	KDSFlags:            model.GlobalToAllZonesFlag | model.ZoneToGlobalFlag,
	WsPath:              "meshtimeouts",
	KumactlArg:          "meshtimeout",
	KumactlListArg:      "meshtimeouts",
	AllowToInspect:      true,
	IsPolicy:            true,
	IsExperimental:      false,
	SingularDisplayName: "Mesh Timeout",
	PluralDisplayName:   "Mesh Timeouts",
	IsPluginOriginated:  true,
	IsTargetRefBased:    true,
	HasToTargetRef:      true,
	HasFromTargetRef:    true,
}
