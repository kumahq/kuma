package mesh

import (
	"errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
)

const (
	DataplaneType model.ResourceType = "Dataplane"
)

var _ model.Resource = &DataplaneResource{}

type DataplaneResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.Dataplane
}

func (t *DataplaneResource) GetType() model.ResourceType {
	return DataplaneType
}
func (t *DataplaneResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *DataplaneResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *DataplaneResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *DataplaneResource) SetSpec(spec model.ResourceSpec) error {
	dataplane, ok := spec.(*mesh_proto.Dataplane)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *dataplane
		return nil
	}
}

var _ model.ResourceList = &DataplaneResourceList{}

type DataplaneResourceList struct {
	Items      []*DataplaneResource
	Pagination model.Pagination
}

func (l *DataplaneResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *DataplaneResourceList) GetItemType() model.ResourceType {
	return DataplaneType
}
func (l *DataplaneResourceList) NewItem() model.Resource {
	return &DataplaneResource{}
}
func (l *DataplaneResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*DataplaneResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*DataplaneResource)(nil), r)
	}
}
func (l *DataplaneResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(&DataplaneResource{})
	registry.RegistryListType(&DataplaneResourceList{})
}
