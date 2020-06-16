package sample

import (
	"github.com/pkg/errors"

	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
	"github.com/Kong/kuma/pkg/core/validators"
	proto "github.com/Kong/kuma/pkg/test/apis/sample/v1alpha1"
)

const (
	TrafficRouteType model.ResourceType = "SampleTrafficRoute"
)

var _ model.Resource = &TrafficRouteResource{}

type TrafficRouteResource struct {
	Meta model.ResourceMeta
	Spec proto.TrafficRoute
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
	return &t.Spec
}
func (t *TrafficRouteResource) SetSpec(spec model.ResourceSpec) error {
	route, ok := spec.(*proto.TrafficRoute)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *route
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
	return &TrafficRouteResource{}
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
	registry.RegisterType(&TrafficRouteResource{})
	registry.RegistryListType(&TrafficRouteResourceList{})
}
