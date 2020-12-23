package mesh

import (
	"errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	RetryType model.ResourceType = "Retry"
)

var _ model.Resource = &RetryResource{}

type RetryResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.Retry
}

func NewRetryResource() *RetryResource {
	return &RetryResource{
		Spec: &mesh_proto.Retry{},
	}
}

func (r *RetryResource) Sources() []*mesh_proto.Selector {
	return r.Spec.Sources
}

func (r *RetryResource) Destinations() []*mesh_proto.Selector {
	return r.Spec.Destinations
}

func (r *RetryResource) GetType() model.ResourceType {
	return RetryType
}

func (r *RetryResource) GetMeta() model.ResourceMeta {
	return r.Meta
}

func (r *RetryResource) SetMeta(meta model.ResourceMeta) {
	r.Meta = meta
}

func (r *RetryResource) GetSpec() model.ResourceSpec {
	return r.Spec
}

func (r *RetryResource) SetSpec(value model.ResourceSpec) error {
	if spec, ok := value.(*mesh_proto.Retry); ok {
		r.Spec = spec

		return nil
	}

	return errors.New("invalid type of spec")
}

func (r *RetryResource) Scope() model.ResourceScope {
	return model.ScopeMesh
}

var _ model.ResourceList = &RetryResourceList{}

type RetryResourceList struct {
	Items      []*RetryResource
	Pagination model.Pagination
}

func (r *RetryResourceList) GetItemType() model.ResourceType {
	return RetryType
}

func (r *RetryResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(r.Items))
	for i, elem := range r.Items {
		res[i] = elem
	}

	return res
}

func (r *RetryResourceList) NewItem() model.Resource {
	return NewRetryResource()
}

func (r *RetryResourceList) AddItem(value model.Resource) error {
	if resource, ok := value.(*RetryResource); ok {
		r.Items = append(r.Items, resource)

		return nil
	}

	return model.ErrorInvalidItemType((*RetryResource)(nil), r)
}

func (r *RetryResourceList) GetPagination() *model.Pagination {
	return &r.Pagination
}

func init() {
	registry.RegisterType(NewRetryResource())
	registry.RegistryListType(&RetryResourceList{})
}
