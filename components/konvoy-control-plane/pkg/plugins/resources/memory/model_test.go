package memory_test

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
)

const (
	TrafficRouteType model.ResourceType = "TrafficRoute"
)

// TODO(yskopets): this would be a protobuf
type TrafficRoute struct {
	Path string `json:"path"`
}

var _ model.Resource = &TrafficRouteResource{}

type TrafficRouteResource struct {
	Meta model.ResourceMeta
	Spec TrafficRoute
}

func (t *TrafficRouteResource) GetType() model.ResourceType {
	return TrafficRouteType
}
func (t *TrafficRouteResource) GetMeta() model.ResourceMeta {
	return t.Meta
}
func (t *TrafficRouteResource) SetMeta(m model.ResourceMeta) {
	t.Meta = m
}
func (t *TrafficRouteResource) GetSpec() model.ResourceSpec {
	return &t.Spec
}

var _ model.ResourceList = &TrafficRouteResourceList{}

type TrafficRouteResourceList struct {
	Items []*TrafficRouteResource
}

func (l *TrafficRouteResourceList) GetItemType() model.ResourceType {
	return TrafficRouteType
}
func (l *TrafficRouteResourceList) NewItem() model.Resource {
	return &TrafficRouteResource{}
}
func (l *TrafficRouteResourceList) AddItem(r model.Resource) error {
	if trr, ok := r.(*TrafficRouteResource); ok {
		l.Items = append(l.Items, trr)
		return nil
	} else {
		return model.ErrorInvalidItemType((*TrafficRouteResource)(nil), r)
	}
}
