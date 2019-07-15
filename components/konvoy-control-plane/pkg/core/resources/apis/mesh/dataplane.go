package mesh

import (
	"errors"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
)

const (
	DataplaneStatusType model.ResourceType = "DataplaneStatus"
)

var _ model.Resource = &DataplaneStatusResource{}

type DataplaneStatusResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.DataplaneStatus
}

func (t *DataplaneStatusResource) GetType() model.ResourceType {
	return DataplaneStatusType
}
func (t *DataplaneStatusResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *DataplaneStatusResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *DataplaneStatusResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *DataplaneStatusResource) SetSpec(spec model.ResourceSpec) error {
	status, ok := spec.(*mesh_proto.DataplaneStatus)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *status
		return nil
	}
}

var _ model.ResourceList = &DataplaneStatusResourceList{}

type DataplaneStatusResourceList struct {
	Items []*DataplaneStatusResource
}

func (l *DataplaneStatusResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *DataplaneStatusResourceList) GetItemType() model.ResourceType {
	return DataplaneStatusType
}
func (l *DataplaneStatusResourceList) NewItem() model.Resource {
	return &DataplaneStatusResource{}
}
func (l *DataplaneStatusResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*DataplaneStatusResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*DataplaneStatusResource)(nil), r)
	}
}
