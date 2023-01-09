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
var schema = spec.Schema{}

func init() {
	if err := yaml.Unmarshal(rawSchema, &schema); err != nil {
		panic(err)
	}
}

const (
	MeshProxyPatchType model.ResourceType = "MeshProxyPatch"
)

var _ model.Resource = &MeshProxyPatchResource{}

type MeshProxyPatchResource struct {
	Meta model.ResourceMeta
	Spec *MeshProxyPatch
}

func NewMeshProxyPatchResource() *MeshProxyPatchResource {
	return &MeshProxyPatchResource{
		Spec: &MeshProxyPatch{},
	}
}

func (t *MeshProxyPatchResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *MeshProxyPatchResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *MeshProxyPatchResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *MeshProxyPatchResource) SetSpec(spec model.ResourceSpec) error {
	protoType, ok := spec.(*MeshProxyPatch)
	if !ok {
		return fmt.Errorf("invalid type %T for Spec", spec)
	} else {
		if protoType == nil {
			t.Spec = &MeshProxyPatch{}
		} else {
			t.Spec = protoType
		}
		return nil
	}
}

func (t *MeshProxyPatchResource) Descriptor() model.ResourceTypeDescriptor {
	return MeshProxyPatchResourceTypeDescriptor
}

func (t *MeshProxyPatchResource) Validate() error {
	if v, ok := interface{}(t).(interface{ validate() error }); !ok {
		return nil
	} else {
		return v.validate()
	}
}

var _ model.ResourceList = &MeshProxyPatchResourceList{}

type MeshProxyPatchResourceList struct {
	Items      []*MeshProxyPatchResource
	Pagination model.Pagination
}

func (l *MeshProxyPatchResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *MeshProxyPatchResourceList) GetItemType() model.ResourceType {
	return MeshProxyPatchType
}

func (l *MeshProxyPatchResourceList) NewItem() model.Resource {
	return NewMeshProxyPatchResource()
}

func (l *MeshProxyPatchResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*MeshProxyPatchResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*MeshProxyPatchResource)(nil), r)
	}
}

func (l *MeshProxyPatchResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

var MeshProxyPatchResourceTypeDescriptor = model.ResourceTypeDescriptor{
	Name:                MeshProxyPatchType,
	Resource:            NewMeshProxyPatchResource(),
	ResourceList:        &MeshProxyPatchResourceList{},
	Scope:               model.ScopeMesh,
	KDSFlags:            model.FromGlobalToZone,
	WsPath:              "meshproxypatches",
	KumactlArg:          "meshproxypatch",
	KumactlListArg:      "meshproxypatches",
	AllowToInspect:      true,
	IsPolicy:            true,
	IsExperimental:      false,
	SingularDisplayName: "Mesh Proxy Patch",
	PluralDisplayName:   "Mesh Proxy Patches",
	IsPluginOriginated:  true,
	Schema:              &schema,
}
