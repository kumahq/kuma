package mesh

import (
	"errors"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/registry"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
)

const (
	DataplaneInspectionType model.ResourceType = "DataplaneInspection"
)

var _ model.Resource = &DataplaneInspectionResource{}

type DataplaneInspectionResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.DataplaneInspection
}

func (t *DataplaneInspectionResource) GetType() model.ResourceType {
	return DataplaneInspectionType
}

func (t *DataplaneInspectionResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *DataplaneInspectionResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *DataplaneInspectionResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}

func (t *DataplaneInspectionResource) SetSpec(spec model.ResourceSpec) error {
	dataplaneInspection, ok := spec.(*mesh_proto.DataplaneInspection)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *dataplaneInspection
		return nil
	}
}

var _ model.ResourceList = &DataplaneInspectionResourceList{}

type DataplaneInspectionResourceList struct {
	Items []*DataplaneInspectionResource
}

func (l *DataplaneInspectionResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *DataplaneInspectionResourceList) GetItemType() model.ResourceType {
	return DataplaneInspectionType
}
func (l *DataplaneInspectionResourceList) NewItem() model.Resource {
	return &DataplaneInspectionResource{}
}
func (l *DataplaneInspectionResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*DataplaneInspectionResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*DataplaneInspectionResource)(nil), r)
	}
}

func init() {
	registry.RegisterType(&DataplaneInspectionResource{})
	registry.RegistryListType(&DataplaneInspectionResourceList{})
}

func NewDataplaneInspections(dataplanes DataplaneResourceList, insights DataplaneInsightResourceList) DataplaneInspectionResourceList {
	insightsByKey := map[model.ResourceKey]*DataplaneInsightResource{}
	for _, insight := range insights.Items {
		insightsByKey[model.MetaToResourceKey(insight.Meta)] = insight
	}

	var items []*DataplaneInspectionResource
	for _, dataplane := range dataplanes.Items {
		inspection := DataplaneInspectionResource{
			Meta: dataplane.Meta,
			Spec: mesh_proto.DataplaneInspection{
				Dataplane:        dataplane.Spec,
				DataplaneInsight: mesh_proto.DataplaneInsight{},
			},
		}
		insight, exists := insightsByKey[model.MetaToResourceKey(inspection.Meta)]
		if exists {
			inspection.Spec.DataplaneInsight = insight.Spec
		}
		items = append(items, &inspection)
	}
	return DataplaneInspectionResourceList{Items: items}
}

func (d *DataplaneInspectionResourceList) RetainMatchingTags(tags map[string]string) {
	result := []*DataplaneInspectionResource{}
	for _, inspection := range d.Items {
		if inspection.Spec.Dataplane.MatchTags(tags) {
			result = append(result, inspection)
		}
	}
	d.Items = result
}
