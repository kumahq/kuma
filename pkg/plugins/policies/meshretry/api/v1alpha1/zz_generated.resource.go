// Generated by tools/resource-gen.
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
	MeshRetryResourceTypeDescriptor.Schema = &schema
}

const (
	MeshRetryType model.ResourceType = "MeshRetry"
)

var _ model.Resource = &MeshRetryResource{}

type MeshRetryResource struct {
	Meta model.ResourceMeta
	Spec *MeshRetry
}

func NewMeshRetryResource() *MeshRetryResource {
	return &MeshRetryResource{
		Spec: &MeshRetry{},
	}
}

func (t *MeshRetryResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *MeshRetryResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *MeshRetryResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *MeshRetryResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*MeshRetry)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &MeshRetry{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *MeshRetryResource) Descriptor() model.ResourceTypeDescriptor {
	return MeshRetryResourceTypeDescriptor
}

func (t *MeshRetryResource) Validate() error {
	if v, ok := interface{}(t).(interface{ validate() error }); !ok {
		return nil
	} else {
		return v.validate()
	}
}

var _ model.ResourceList = &MeshRetryResourceList{}

type MeshRetryResourceList struct {
	Items      []*MeshRetryResource
	Pagination model.Pagination
}

func (l *MeshRetryResourceList) Descriptor() model.ResourceTypeDescriptor {
	return MeshRetryResourceTypeDescriptor
}

func (l *MeshRetryResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *MeshRetryResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*MeshRetryResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*MeshRetryResource)(nil), r)
	}
}

func (l *MeshRetryResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

var MeshRetryResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                MeshRetryType,
	Resource:            NewMeshRetryResource(),
	ResourceList:        &MeshRetryResourceList{},
	Scope:               model.ScopeMesh,
	KDSFlags:            model.FromGlobalToZone,
	WsPath:              "meshretries",
	KumactlArg:          "meshretry",
	KumactlListArg:      "meshretries",
	AllowToInspect:      true,
	IsPolicy:            true,
	IsExperimental:      false,
	SingularDisplayName: "Mesh Retry",
	PluralDisplayName:   "Mesh Retries",
	IsPluginOriginated:  true,
}
