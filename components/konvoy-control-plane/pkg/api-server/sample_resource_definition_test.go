package api_server_test

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server/definitions"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	sample_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/apis/sample"
)

var TrafficRouteWsDefinition = definitions.ResourceWsDefinition{
	Name: "Traffic Route",
	Path: "traffic-routes",
	ResourceFactory: func() model.Resource {
		return &sample_model.TrafficRouteResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &sample_model.TrafficRouteResourceList{}
	},
}
