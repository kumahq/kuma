package mesh

import (
	"errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	RateLimitType model.ResourceType = "RateLimit"
)

var _ model.Resource = &RateLimitResource{}

type RateLimitResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.RateLimit
}

func NewRateLimitResource() *RateLimitResource {
	return &RateLimitResource{
		Spec: &mesh_proto.RateLimit{},
	}
}

func (t *RateLimitResource) GetType() model.ResourceType {
	return RateLimitType
}
func (t *RateLimitResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *RateLimitResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *RateLimitResource) GetSpec() model.ResourceSpec {
	return t.Spec
}
func (t *RateLimitResource) SetSpec(spec model.ResourceSpec) error {
	status, ok := spec.(*mesh_proto.RateLimit)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = status
		return nil
	}
}
func (t *RateLimitResource) Scope() model.ResourceScope {
	return model.ScopeMesh
}

var _ model.ResourceList = &RateLimitResourceList{}

type RateLimitResourceList struct {
	Items      []*RateLimitResource
	Pagination model.Pagination
}

func (l *RateLimitResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *RateLimitResourceList) GetItemType() model.ResourceType {
	return RateLimitType
}
func (l *RateLimitResourceList) NewItem() model.Resource {
	return NewRateLimitResource()
}
func (l *RateLimitResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*RateLimitResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*RateLimitResource)(nil), r)
	}
}
func (l *RateLimitResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (t *RateLimitResource) Sources() []*mesh_proto.Selector {
	return t.Spec.GetSources()
}

func (t *RateLimitResource) Destinations() []*mesh_proto.Selector {
	return t.Spec.GetDestinations()
}

func init() {
	registry.RegisterType(NewRateLimitResource())
	registry.RegistryListType(&RateLimitResourceList{})
}
