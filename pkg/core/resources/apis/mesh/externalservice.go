package mesh

import (
	"errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	ExternalServiceType model.ResourceType = "ExternalService"
)

var _ model.Resource = &ExternalServiceResource{}

type ExternalServiceResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.ExternalService
}

func (t *ExternalServiceResource) GetType() model.ResourceType {
	return ExternalServiceType
}
func (t *ExternalServiceResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *ExternalServiceResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *ExternalServiceResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *ExternalServiceResource) SetSpec(spec model.ResourceSpec) error {
	externalService, ok := spec.(*mesh_proto.ExternalService)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *externalService
		return nil
	}
}

func (t *ExternalServiceResource) Scope() model.ResourceScope {
	return model.ScopeMesh
}

var _ model.ResourceList = &ExternalServiceResourceList{}

type ExternalServiceResourceList struct {
	Items      []*ExternalServiceResource
	Pagination model.Pagination
}

func (l *ExternalServiceResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *ExternalServiceResourceList) GetItemType() model.ResourceType {
	return ExternalServiceType
}
func (l *ExternalServiceResourceList) NewItem() model.Resource {
	return &ExternalServiceResource{}
}
func (l *ExternalServiceResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ExternalServiceResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ExternalServiceResource)(nil), r)
	}
}
func (l *ExternalServiceResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(&ExternalServiceResource{})
	registry.RegistryListType(&ExternalServiceResourceList{})
}
