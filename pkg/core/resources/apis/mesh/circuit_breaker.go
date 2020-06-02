package mesh

import (
	"errors"

	"github.com/Kong/kuma/pkg/core/resources/registry"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

const (
	CircuitBreakerType model.ResourceType = "CircuitBreaker"
)

var _ model.Resource = &CircuitBreakerResource{}

type CircuitBreakerResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.CircuitBreaker
}

func (c *CircuitBreakerResource) GetType() model.ResourceType {
	return CircuitBreakerType
}

func (c *CircuitBreakerResource) GetMeta() model.ResourceMeta {
	return c.Meta
}

func (c *CircuitBreakerResource) SetMeta(m model.ResourceMeta) {
	c.Meta = m
}

func (c *CircuitBreakerResource) GetSpec() model.ResourceSpec {
	return &c.Spec
}

func (c *CircuitBreakerResource) SetSpec(spec model.ResourceSpec) error {
	circuitBreaker, ok := spec.(*mesh_proto.CircuitBreaker)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		c.Spec = *circuitBreaker
		return nil
	}
}

var _ model.ResourceList = &CircuitBreakerResourceList{}

type CircuitBreakerResourceList struct {
	Items      []*CircuitBreakerResource
	Pagination model.Pagination
}

func (l *CircuitBreakerResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *CircuitBreakerResourceList) GetItemType() model.ResourceType {
	return CircuitBreakerType
}

func (l *CircuitBreakerResourceList) NewItem() model.Resource {
	return &CircuitBreakerResource{}
}

func (l *CircuitBreakerResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*CircuitBreakerResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*CircuitBreakerResource)(nil), r)
	}
}

func (l *CircuitBreakerResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(&CircuitBreakerResource{})
	registry.RegistryListType(&CircuitBreakerResourceList{})
}

func (c *CircuitBreakerResource) Sources() []*mesh_proto.Selector {
	return c.Spec.GetSources()
}

func (c *CircuitBreakerResource) Destinations() []*mesh_proto.Selector {
	return c.Spec.GetDestinations()
}
