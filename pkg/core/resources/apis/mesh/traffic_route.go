package mesh

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	TrafficRouteType model.ResourceType = "TrafficRoute"
)

var _ model.Resource = &TrafficRouteResource{}

type TrafficRouteResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.TrafficRoute
}

func NewTrafficRouteResource() *TrafficRouteResource {
	return &TrafficRouteResource{
		Spec: &mesh_proto.TrafficRoute{},
	}
}

func (t *TrafficRouteResource) GetType() model.ResourceType {
	return TrafficRouteType
}
func (t *TrafficRouteResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *TrafficRouteResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *TrafficRouteResource) GetSpec() model.ResourceSpec {
	return t.Spec
}
func (t *TrafficRouteResource) SetSpec(spec model.ResourceSpec) error {
	template, ok := spec.(*mesh_proto.TrafficRoute)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = template
		return nil
	}
}
func (t *TrafficRouteResource) Scope() model.ResourceScope {
	return model.ScopeMesh
}

var _ model.ResourceList = &TrafficRouteResourceList{}

type TrafficRouteResourceList struct {
	Items      []*TrafficRouteResource
	Pagination model.Pagination
}

func (l *TrafficRouteResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *TrafficRouteResourceList) GetItemType() model.ResourceType {
	return TrafficRouteType
}
func (l *TrafficRouteResourceList) NewItem() model.Resource {
	return NewTrafficRouteResource()
}
func (l *TrafficRouteResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*TrafficRouteResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*TrafficRouteResource)(nil), r)
	}
}
func (l *TrafficRouteResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(NewTrafficRouteResource())
	registry.RegistryListType(&TrafficRouteResourceList{})
}

func (t *TrafficRouteResource) Sources() []*mesh_proto.Selector {
	return t.Spec.GetSources()
}

func (t *TrafficRouteResource) Destinations() []*mesh_proto.Selector {
	return t.Spec.GetDestinations()
}
