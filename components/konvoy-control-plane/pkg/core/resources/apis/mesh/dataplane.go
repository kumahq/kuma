package mesh

import (
	"errors"
	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
)

const (
	DataplaneType model.ResourceType = "Dataplane"
)

var _ model.Resource = &DataplaneResource{}

type DataplaneResource struct {
	Meta   model.ResourceMeta
	Status mesh_proto.DataplaneStatus
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
	return &t.Status
}
func (t *DataplaneResource) SetSpec(spec model.ResourceSpec) error {
	status, ok := spec.(*mesh_proto.DataplaneStatus)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Status = *status
		return nil
	}
}

var _ model.ResourceList = &DataplaneResourceList{}

type DataplaneResourceList struct {
	Items []*DataplaneResource
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
