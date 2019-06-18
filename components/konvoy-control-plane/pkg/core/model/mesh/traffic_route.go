package mesh

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/model"
)

const (
	TrafficRouteType model.ResourceType = "TrafficRoute"
)

// TODO(yskopets): this would be a protobuf
type TrafficRoute struct {
	Path string
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
func (t *TrafficRouteResource) GetSpec() model.ResourceSpec {
	return t.Spec
}

var _ model.ResourceList = &TrafficRouteResourceList{}

type TrafficRouteResourceList struct {
	Items []*TrafficRouteResource
}

func (l *TrafficRouteResourceList) GetItemType() model.ResourceType {
	return TrafficRouteType
}

func (l *TrafficRouteResourceList) GetItems() []model.Resource {
	items := make([]model.Resource, len(l.Items))
	for i := range l.Items {
		items[i] = l.Items[i]
	}
	return items
}
