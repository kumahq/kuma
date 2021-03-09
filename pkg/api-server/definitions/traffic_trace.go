package definitions

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

var TrafficTraceWsDefinition = ResourceWsDefinition{
	Name: "Traffic Trace",
	Path: "traffic-traces",
	ResourceFactory: func() model.Resource {
		return mesh.NewTrafficTraceResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &mesh.TrafficTraceResourceList{}
	},
}
