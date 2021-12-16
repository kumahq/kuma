package sample

import (
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/validators"
	proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
)

const (
	TrafficRouteType model.ResourceType = "SampleTrafficRoute"
)

var _ model.Resource = &TrafficRouteResource{}

type TrafficRouteResource struct {
	Meta model.ResourceMeta
	Spec *proto.TrafficRoute
}

func NewTrafficRouteResource() *TrafficRouteResource {
	return &TrafficRouteResource{
		Spec: &proto.TrafficRoute{},
	}
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
	route, ok := spec.(*proto.TrafficRoute)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		if route == nil {
			t.Spec = &proto.TrafficRoute{}
		} else {
			t.Spec = route
		}
		return nil
	}
}

func (t *TrafficRouteResource) Validate() error {
	err := validators.ValidationError{}
	if t.Spec.Path == "" {
		err.AddViolation("path", "cannot be empty")
	}
	return err.OrNil()
}

func (t *TrafficRouteResource) Descriptor() model.ResourceTypeDescriptor {
	return TrafficRouteResourceTypeDescriptor
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

var TrafficRouteResourceTypeDescriptor model.ResourceTypeDescriptor

func init() {
	TrafficRouteResourceTypeDescriptor = model.ResourceTypeDescriptor{
		Name:         TrafficRouteType,
		Resource:     NewTrafficRouteResource(),
		ResourceList: &TrafficRouteResourceList{},
		ReadOnly:     false,
		AdminOnly:    false,
		Scope:        model.ScopeMesh,
		WsPath:       "sample-traffic-routes",
	}
	registry.RegisterType(TrafficRouteResourceTypeDescriptor)
}
