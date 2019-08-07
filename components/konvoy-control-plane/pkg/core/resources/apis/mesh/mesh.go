package mesh

import (
	"errors"
	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/registry"
)

const (
	MeshType model.ResourceType = "Mesh"
)

var _ model.Resource = &MeshResource{}

type MeshResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.Mesh
}

func (t *MeshResource) GetType() model.ResourceType {
	return MeshType
}
func (t *MeshResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *MeshResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *MeshResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *MeshResource) SetSpec(spec model.ResourceSpec) error {
	mesh, ok := spec.(*mesh_proto.Mesh)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *mesh
		return nil
	}
}

var _ model.ResourceList = &MeshResourceList{}

type MeshResourceList struct {
	Items []*MeshResource
}

func (l *MeshResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *MeshResourceList) GetItemType() model.ResourceType {
	return MeshType
}
func (l *MeshResourceList) NewItem() model.Resource {
	return &MeshResource{}
}
func (l *MeshResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*MeshResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*MeshResource)(nil), r)
	}
}

func init() {
	registry.RegisterType(MeshType, &MeshResource{})
	registry.RegistryListType(MeshType, &MeshResourceList{})
}
