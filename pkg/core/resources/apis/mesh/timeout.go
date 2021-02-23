package mesh

import (
	"errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	TimeoutType model.ResourceType = "Timeout"
)

var _ model.Resource = &TimeoutResource{}

type TimeoutResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.Timeout
}

func NewTimeoutResource() *TimeoutResource {
	return &TimeoutResource{
		Spec: &mesh_proto.Timeout{},
	}
}

func (t *TimeoutResource) Sources() []*mesh_proto.Selector {
	return t.Spec.Sources
}

func (t *TimeoutResource) Destinations() []*mesh_proto.Selector {
	return t.Spec.Destinations
}

func (t *TimeoutResource) GetType() model.ResourceType {
	return TimeoutType
}

func (t *TimeoutResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *TimeoutResource) SetMeta(meta model.ResourceMeta) {
	t.Meta = meta
}

func (t *TimeoutResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *TimeoutResource) SetSpec(value model.ResourceSpec) error {
	if spec, ok := value.(*mesh_proto.Timeout); ok {
		t.Spec = spec

		return nil
	}

	return errors.New("invalid type of spec")
}

func (t *TimeoutResource) Scope() model.ResourceScope {
	return model.ScopeMesh
}

var _ model.ResourceList = &TimeoutResourceList{}

type TimeoutResourceList struct {
	Items      []*TimeoutResource
	Pagination model.Pagination
}

func (r *TimeoutResourceList) GetItemType() model.ResourceType {
	return TimeoutType
}

func (r *TimeoutResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(r.Items))
	for i, elem := range r.Items {
		res[i] = elem
	}

	return res
}

func (r *TimeoutResourceList) NewItem() model.Resource {
	return NewTimeoutResource()
}

func (r *TimeoutResourceList) AddItem(value model.Resource) error {
	if resource, ok := value.(*TimeoutResource); ok {
		r.Items = append(r.Items, resource)

		return nil
	}

	return model.ErrorInvalidItemType((*TimeoutResource)(nil), r)
}

func (r *TimeoutResourceList) GetPagination() *model.Pagination {
	return &r.Pagination
}

func init() {
	registry.RegisterType(NewTimeoutResource())
	registry.RegistryListType(&TimeoutResourceList{})
}
