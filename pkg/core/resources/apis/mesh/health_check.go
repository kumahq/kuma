package mesh

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	HealthCheckType model.ResourceType = "HealthCheck"
)

var _ model.Resource = &HealthCheckResource{}

type HealthCheckResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.HealthCheck
}

func NewHealthCheckResource() *HealthCheckResource {
	return &HealthCheckResource{
		Spec: &mesh_proto.HealthCheck{},
	}
}

func (r *HealthCheckResource) GetType() model.ResourceType {
	return HealthCheckType
}
func (r *HealthCheckResource) GetMeta() model.ResourceMeta {
	return r.Meta
}
func (r *HealthCheckResource) SetMeta(m model.ResourceMeta) {
	r.Meta = m
}
func (r *HealthCheckResource) GetSpec() model.ResourceSpec {
	return r.Spec
}
func (r *HealthCheckResource) SetSpec(value model.ResourceSpec) error {
	spec, ok := value.(*mesh_proto.HealthCheck)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		r.Spec = spec
		return nil
	}
}
func (t *HealthCheckResource) Sources() []*mesh_proto.Selector {
	return t.Spec.GetSources()
}
func (t *HealthCheckResource) Destinations() []*mesh_proto.Selector {
	return t.Spec.GetDestinations()
}
func (t *HealthCheckResource) Scope() model.ResourceScope {
	return model.ScopeMesh
}

var _ model.ResourceList = &HealthCheckResourceList{}

type HealthCheckResourceList struct {
	Items      []*HealthCheckResource
	Pagination model.Pagination
}

func (l *HealthCheckResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *HealthCheckResourceList) GetItemType() model.ResourceType {
	return HealthCheckType
}
func (l *HealthCheckResourceList) NewItem() model.Resource {
	return NewHealthCheckResource()
}
func (l *HealthCheckResourceList) AddItem(r model.Resource) error {
	if item, ok := r.(*HealthCheckResource); ok {
		l.Items = append(l.Items, item)
		return nil
	} else {
		return model.ErrorInvalidItemType((*HealthCheckResource)(nil), r)
	}
}
func (l *HealthCheckResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(NewHealthCheckResource())
	registry.RegistryListType(&HealthCheckResourceList{})
}
