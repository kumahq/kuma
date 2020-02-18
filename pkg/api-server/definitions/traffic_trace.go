package definitions

import (
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
)

var TrafficTraceWsDefinition = ResourceWsDefinition{
	Name: "Traffic Trace",
	Path: "traffic-traces",
	ResourceFactory: func() model.Resource {
		return &mesh.TrafficTraceResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.TrafficTraceResourceList{}
	},
}
