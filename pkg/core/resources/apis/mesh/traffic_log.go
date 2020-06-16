package mesh

import (
	"errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
)

const (
	TrafficLogType model.ResourceType = "TrafficLog"
)

var _ model.Resource = &TrafficLogResource{}

type TrafficLogResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.TrafficLog
}

func (t *TrafficLogResource) GetType() model.ResourceType {
	return TrafficLogType
}
func (t *TrafficLogResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *TrafficLogResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *TrafficLogResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *TrafficLogResource) SetSpec(spec model.ResourceSpec) error {
	status, ok := spec.(*mesh_proto.TrafficLog)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *status
		return nil
	}
}

var _ model.ResourceList = &TrafficLogResourceList{}

type TrafficLogResourceList struct {
	Items      []*TrafficLogResource
	Pagination model.Pagination
}

func (l *TrafficLogResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *TrafficLogResourceList) GetItemType() model.ResourceType {
	return TrafficLogType
}
func (l *TrafficLogResourceList) NewItem() model.Resource {
	return &TrafficLogResource{}
}
func (l *TrafficLogResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*TrafficLogResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*TrafficLogResource)(nil), r)
	}
}
func (l *TrafficLogResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(&TrafficLogResource{})
	registry.RegistryListType(&TrafficLogResourceList{})
}

func (t *TrafficLogResource) Sources() []*mesh_proto.Selector {
	return t.Spec.GetSources()
}

func (t *TrafficLogResource) Destinations() []*mesh_proto.Selector {
	return t.Spec.GetDestinations()
}
