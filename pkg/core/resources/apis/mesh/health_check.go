package mesh

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
)

const (
	HealthCheckType model.ResourceType = "HealthCheck"
)

var _ model.Resource = &HealthCheckResource{}

type HealthCheckResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.HealthCheck
}

func (t *HealthCheckResource) GetType() model.ResourceType {
	return HealthCheckType
}
func (t *HealthCheckResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *HealthCheckResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *HealthCheckResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *HealthCheckResource) SetSpec(spec model.ResourceSpec) error {
	template, ok := spec.(*mesh_proto.HealthCheck)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *template
		return nil
	}
}

func (t *HealthCheckResource) Validate() error {
	if t == nil {
		return nil
	}
	return t.Spec.Validate()
}

var _ model.ResourceList = &HealthCheckResourceList{}

type HealthCheckResourceList struct {
	Items []*HealthCheckResource
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
	return &HealthCheckResource{}
}
func (l *HealthCheckResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*HealthCheckResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*HealthCheckResource)(nil), r)
	}
}

func init() {
	registry.RegisterType(&HealthCheckResource{})
	registry.RegistryListType(&HealthCheckResourceList{})
}
