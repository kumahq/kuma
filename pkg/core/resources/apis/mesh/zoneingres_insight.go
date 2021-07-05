package mesh

import (
	"errors"

	"github.com/kumahq/kuma/pkg/core/resources/registry"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

const (
	ZoneIngressInsightType model.ResourceType = "ZoneIngressInsight"
)

var _ model.Resource = &ZoneIngressInsightResource{}

type ZoneIngressInsightResource struct {
	Meta model.ResourceMeta
	Spec *mesh_proto.ZoneIngressInsight
}

func NewZoneIngressInsightResource() *ZoneIngressInsightResource {
	return &ZoneIngressInsightResource{
		Spec: &mesh_proto.ZoneIngressInsight{},
	}
}

func (t *ZoneIngressInsightResource) GetType() model.ResourceType {
	return ZoneIngressInsightType
}
func (t *ZoneIngressInsightResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *ZoneIngressInsightResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *ZoneIngressInsightResource) GetSpec() model.ResourceSpec {
	return t.Spec
}
func (t *ZoneIngressInsightResource) SetSpec(spec model.ResourceSpec) error {
	status, ok := spec.(*mesh_proto.ZoneIngressInsight)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = status
		return nil
	}
}
func (t *ZoneIngressInsightResource) Validate() error {
	return nil
}
func (t *ZoneIngressInsightResource) Scope() model.ResourceScope {
	return model.ScopeGlobal
}

var _ model.ResourceList = &ZoneIngressInsightResourceList{}

type ZoneIngressInsightResourceList struct {
	Items      []*ZoneIngressInsightResource
	Pagination model.Pagination
}

func (l *ZoneIngressInsightResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *ZoneIngressInsightResourceList) GetItemType() model.ResourceType {
	return ZoneIngressInsightType
}
func (l *ZoneIngressInsightResourceList) NewItem() model.Resource {
	return NewZoneIngressInsightResource()
}
func (l *ZoneIngressInsightResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*ZoneIngressInsightResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*ZoneIngressInsightResource)(nil), r)
	}
}

func (l *ZoneIngressInsightResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func init() {
	registry.RegisterType(NewZoneIngressInsightResource())
	registry.RegistryListType(&ZoneIngressInsightResourceList{})
}
