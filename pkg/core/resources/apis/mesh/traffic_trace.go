package mesh

import (
	"errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
)

const (
	TrafficTraceType model.ResourceType = "TrafficTrace"
)

var _ model.Resource = &TrafficTraceResource{}

type TrafficTraceResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.TrafficTrace
}

func (t *TrafficTraceResource) GetType() model.ResourceType {
	return TrafficTraceType
}
func (t *TrafficTraceResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *TrafficTraceResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *TrafficTraceResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *TrafficTraceResource) SetSpec(spec model.ResourceSpec) error {
	status, ok := spec.(*mesh_proto.TrafficTrace)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *status
		return nil
	}
}

var _ model.ResourceList = &TrafficTraceResourceList{}

type TrafficTraceResourceList struct {
	Items      []*TrafficTraceResource
	Pagination model.Pagination
}

func (l *TrafficTraceResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *TrafficTraceResourceList) GetItemType() model.ResourceType {
	return TrafficTraceType
}
func (l *TrafficTraceResourceList) NewItem() model.Resource {
	return &TrafficTraceResource{}
}
func (l *TrafficTraceResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*TrafficTraceResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*TrafficTraceResource)(nil), r)
	}
}
func (l *TrafficTraceResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(&TrafficTraceResource{})
	registry.RegistryListType(&TrafficTraceResourceList{})
}

func (t *TrafficTraceResource) Selectors() []*mesh_proto.Selector {
	return t.Spec.GetSelectors()
}
