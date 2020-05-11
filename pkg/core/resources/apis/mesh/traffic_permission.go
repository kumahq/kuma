package mesh

import (
	"errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
)

const (
	TrafficPermissionType model.ResourceType = "TrafficPermission"
)

var _ model.Resource = &TrafficPermissionResource{}

type TrafficPermissionResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.TrafficPermission
}

func (t *TrafficPermissionResource) GetType() model.ResourceType {
	return TrafficPermissionType
}
func (t *TrafficPermissionResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *TrafficPermissionResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *TrafficPermissionResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *TrafficPermissionResource) SetSpec(spec model.ResourceSpec) error {
	status, ok := spec.(*mesh_proto.TrafficPermission)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *status
		return nil
	}
}

var _ model.ResourceList = &TrafficPermissionResourceList{}

type TrafficPermissionResourceList struct {
	Items      []*TrafficPermissionResource
	Pagination model.Pagination
}

func (l *TrafficPermissionResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *TrafficPermissionResourceList) GetItemType() model.ResourceType {
	return TrafficPermissionType
}
func (l *TrafficPermissionResourceList) NewItem() model.Resource {
	return &TrafficPermissionResource{}
}
func (l *TrafficPermissionResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*TrafficPermissionResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*TrafficPermissionResource)(nil), r)
	}
}
func (l *TrafficPermissionResourceList) GetPagination() *model.Pagination {
	return &l.Pagination
}

func (t *TrafficPermissionResource) Sources() []*mesh_proto.Selector {
	return t.Spec.GetSources()
}

func (t *TrafficPermissionResource) Destinations() []*mesh_proto.Selector {
	return t.Spec.GetDestinations()
}

func init() {
	registry.RegisterType(&TrafficPermissionResource{})
	registry.RegistryListType(&TrafficPermissionResourceList{})
}
