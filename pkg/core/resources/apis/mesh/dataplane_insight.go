package mesh

import (
	"errors"

	"github.com/Kong/kuma/pkg/core/resources/registry"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

const (
	DataplaneInsightType model.ResourceType = "DataplaneInsight"
)

var _ model.Resource = &DataplaneInsightResource{}

type DataplaneInsightResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.DataplaneInsight
}

func (t *DataplaneInsightResource) GetType() model.ResourceType {
	return DataplaneInsightType
}
func (t *DataplaneInsightResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *DataplaneInsightResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *DataplaneInsightResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *DataplaneInsightResource) SetSpec(spec model.ResourceSpec) error {
	status, ok := spec.(*mesh_proto.DataplaneInsight)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *status
		return nil
	}
}
func (t *DataplaneInsightResource) Validate() error {
	return nil
}

var _ model.ResourceList = &DataplaneInsightResourceList{}

type DataplaneInsightResourceList struct {
	Items      []*DataplaneInsightResource
	Pagination model.Pagination
}

func (l *DataplaneInsightResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *DataplaneInsightResourceList) GetItemType() model.ResourceType {
	return DataplaneInsightType
}
func (l *DataplaneInsightResourceList) NewItem() model.Resource {
	return &DataplaneInsightResource{}
}
func (l *DataplaneInsightResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*DataplaneInsightResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*DataplaneInsightResource)(nil), r)
	}
}

func (l *DataplaneInsightResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(&DataplaneInsightResource{})
	registry.RegistryListType(&DataplaneInsightResourceList{})
}
