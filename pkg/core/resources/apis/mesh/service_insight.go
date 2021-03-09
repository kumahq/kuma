package mesh

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
)

const (
	ServiceInsightType model.ResourceType = "ServiceInsight"
)

var _ model.Resource = &ServiceInsightResource{}

type ServiceInsightResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.ServiceInsight
}

func NewServiceInsightResource() *ServiceInsightResource {
	return &ServiceInsightResource{
		Spec: &mesh_proto.ServiceInsight{},
	}
}

func (m *ServiceInsightResource) GetType() model.ResourceType {
	return ServiceInsightType
}

func (m *ServiceInsightResource) GetMeta() model.ResourceMeta {
	return m.Meta
}

func (m *ServiceInsightResource) SetMeta(meta model.ResourceMeta) {
	m.Meta = meta
}

func (m *ServiceInsightResource) GetSpec() model.ResourceSpec {
	return m.Spec
}

func (m *ServiceInsightResource) SetSpec(spec model.ResourceSpec) error {
	serviceInsight, ok := spec.(*mesh_proto.ServiceInsight)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		m.Spec = serviceInsight
		return nil
	}
}

func (m *ServiceInsightResource) Validate() error {
	return nil
}

func (m *ServiceInsightResource) Scope() model.ResourceScope {
	return model.ScopeMesh
}

var _ model.ResourceList = &ServiceInsightResourceList{}

type ServiceInsightResourceList struct {
	Items      []*ServiceInsightResource
	Pagination model.Pagination
}

func (l *ServiceInsightResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *ServiceInsightResourceList) GetItemType() model.ResourceType {
	return ServiceInsightType
}
func (l *ServiceInsightResourceList) NewItem() model.Resource {
	return NewServiceInsightResource()
}
func (l *ServiceInsightResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ServiceInsightResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ServiceInsightResource)(nil), r)
	}
}
func (l *ServiceInsightResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(NewServiceInsightResource())
	registry.RegistryListType(&ServiceInsightResourceList{})
}
