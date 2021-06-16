package mesh

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	MeshInsightType model.ResourceType = "MeshInsight"
)

var _ model.Resource = &MeshInsightResource{}

type MeshInsightResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.MeshInsight
}

func NewMeshInsightResource() *MeshInsightResource {
	return &MeshInsightResource{
		Spec: &mesh_proto.MeshInsight{},
	}
}

func (m *MeshInsightResource) GetType() model.ResourceType {
	return MeshInsightType
}

func (m *MeshInsightResource) GetMeta() model.ResourceMeta {
	return m.Meta
}

func (m *MeshInsightResource) SetMeta(meta model.ResourceMeta) {
	m.Meta = meta
}

func (m *MeshInsightResource) GetSpec() model.ResourceSpec {
	return m.Spec
}

func (m *MeshInsightResource) SetSpec(spec model.ResourceSpec) error {
	meshInsight, ok := spec.(*mesh_proto.MeshInsight)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		m.Spec = meshInsight
		return nil
	}
}

func (m *MeshInsightResource) Validate() error {
	return nil
}

func (m *MeshInsightResource) Scope() model.ResourceScope {
	return model.ScopeGlobal
}

var _ model.ResourceList = &MeshInsightResourceList{}

type MeshInsightResourceList struct {
	Items      []*MeshInsightResource
	Pagination model.Pagination
}

func (l *MeshInsightResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *MeshInsightResourceList) GetItemType() model.ResourceType {
	return MeshInsightType
}

func (l *MeshInsightResourceList) NewItem() model.Resource {
	return NewMeshInsightResource()
}

func (l *MeshInsightResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*MeshInsightResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*MeshInsightResource)(nil), r)
	}
}

func (l *MeshInsightResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(NewMeshInsightResource())
	registry.RegistryListType(&MeshInsightResourceList{})
}
