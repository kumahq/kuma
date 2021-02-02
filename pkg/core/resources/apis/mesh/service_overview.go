package mesh

import (
	"errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

const (
	ServiceOverviewType model.ResourceType = "ServiceOverview"
)

var _ model.Resource = &ServiceOverviewResource{}

type ServiceOverviewResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.ServiceInsight_DataplaneStat
}

func NewServiceOverviewResource() *ServiceOverviewResource {
	return &ServiceOverviewResource{
		Spec: &mesh_proto.ServiceInsight_DataplaneStat{},
	}
}

func (t *ServiceOverviewResource) GetType() model.ResourceType {
	return ServiceOverviewType
}

func (t *ServiceOverviewResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *ServiceOverviewResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *ServiceOverviewResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

func (t *ServiceOverviewResource) SetSpec(spec model.ResourceSpec) error {
	serviceOverview, ok := spec.(*mesh_proto.ServiceInsight_DataplaneStat)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = serviceOverview
		return nil
	}
}

func (t *ServiceOverviewResource) Validate() error {
	return nil
}

func (t *ServiceOverviewResource) Scope() model.ResourceScope {
	return model.ScopeMesh
}

var _ model.ResourceList = &ServiceOverviewResourceList{}

type ServiceOverviewResourceList struct {
	Items      []*ServiceOverviewResource
	Pagination model.Pagination
}

func (l *ServiceOverviewResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *ServiceOverviewResourceList) GetItemType() model.ResourceType {
	return ServiceOverviewType
}
func (l *ServiceOverviewResourceList) NewItem() model.Resource {
	return NewServiceOverviewResource()
}
func (l *ServiceOverviewResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ServiceOverviewResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ServiceOverviewResource)(nil), r)
	}
}

func (l *ServiceOverviewResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}
