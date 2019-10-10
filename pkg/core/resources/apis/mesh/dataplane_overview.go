package mesh

import (
	"errors"
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

const (
	DataplaneOverviewType model.ResourceType = "DataplaneOverview"
)

var _ model.Resource = &DataplaneOverviewResource{}

type DataplaneOverviewResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.DataplaneOverview
}

func (t *DataplaneOverviewResource) GetType() model.ResourceType {
	return DataplaneOverviewType
}

func (t *DataplaneOverviewResource) GetMeta() model.ResourceMeta {
	return t.Meta
}

func (t *DataplaneOverviewResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}

func (t *DataplaneOverviewResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}

func (t *DataplaneOverviewResource) SetSpec(spec model.ResourceSpec) error {
	dataplaneOverview, ok := spec.(*mesh_proto.DataplaneOverview)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *dataplaneOverview
		return nil
	}
}

var _ model.ResourceList = &DataplaneOverviewResourceList{}

type DataplaneOverviewResourceList struct {
	Items []*DataplaneOverviewResource
}

func (l *DataplaneOverviewResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}

func (l *DataplaneOverviewResourceList) GetItemType() model.ResourceType {
	return DataplaneOverviewType
}
func (l *DataplaneOverviewResourceList) NewItem() model.Resource {
	return &DataplaneOverviewResource{}
}
func (l *DataplaneOverviewResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*DataplaneOverviewResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*DataplaneOverviewResource)(nil), r)
	}
}

func NewDataplaneOverviews(dataplanes DataplaneResourceList, insights DataplaneInsightResourceList) DataplaneOverviewResourceList {
	insightsByKey := map[model.ResourceKey]*DataplaneInsightResource{}
	for _, insight := range insights.Items {
		insightsByKey[model.MetaToResourceKey(insight.Meta)] = insight
	}

	var items []*DataplaneOverviewResource
	for _, dataplane := range dataplanes.Items {
		overview := DataplaneOverviewResource{
			Meta: dataplane.Meta,
			Spec: mesh_proto.DataplaneOverview{
				Dataplane:        dataplane.Spec,
				DataplaneInsight: mesh_proto.DataplaneInsight{},
			},
		}
		insight, exists := insightsByKey[model.MetaToResourceKey(overview.Meta)]
		if exists {
			overview.Spec.DataplaneInsight = insight.Spec
		}
		items = append(items, &overview)
	}
	return DataplaneOverviewResourceList{Items: items}
}

func (d *DataplaneOverviewResourceList) RetainMatchingTags(tags map[string]string) {
	result := []*DataplaneOverviewResource{}
	for _, overview := range d.Items {
		if overview.Spec.Dataplane.MatchTags(tags) {
			result = append(result, overview)
		}
	}
	d.Items = result
}
