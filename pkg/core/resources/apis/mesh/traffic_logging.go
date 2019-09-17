package mesh

import (
	"errors"
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/registry"
)

const (
	TrafficLoggingType model.ResourceType = "TrafficLogging"
)

var _ model.Resource = &TrafficLoggingResource{}

type TrafficLoggingResource struct {
	Meta model.ResourceMeta
	Spec mesh_proto.TrafficLogging
}

func (t *TrafficLoggingResource) GetType() model.ResourceType {
	return TrafficLoggingType
}
func (t *TrafficLoggingResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *TrafficLoggingResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *TrafficLoggingResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}
func (t *TrafficLoggingResource) SetSpec(spec model.ResourceSpec) error {
	status, ok := spec.(*mesh_proto.TrafficLogging)
	if !ok {
		return errors.New("invalid type of spec")
	} else {
		t.Spec = *status
		return nil
	}
}

var _ model.ResourceList = &TrafficLoggingResourceList{}

type TrafficLoggingResourceList struct {
	Items []*TrafficLoggingResource
}

func (l *TrafficLoggingResourceList) GetItems() []model.Resource {
	res := make([]model.Resource, len(l.Items))
	for i, elem := range l.Items {
		res[i] = elem
	}
	return res
}
func (l *TrafficLoggingResourceList) GetItemType() model.ResourceType {
	return TrafficLoggingType
}
func (l *TrafficLoggingResourceList) NewItem() model.Resource {
	return &TrafficLoggingResource{}
}
func (l *TrafficLoggingResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*TrafficLoggingResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*TrafficLoggingResource)(nil), r)
	}
}

func init() {
	registry.RegisterType(&TrafficLoggingResource{})
	registry.RegistryListType(&TrafficLoggingResourceList{})
}
