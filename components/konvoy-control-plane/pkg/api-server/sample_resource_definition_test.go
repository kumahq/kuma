package api_server_test

import (
	api_server "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	sample_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/resources/apis/sample"
)

var TrafficRouteWsDefinition = api_server.ResourceWsDefinition{
	Name: "Traffic Route",
	Path: "traffic-routes",
	ResourceFactory: func() model.Resource {
		return &sample_model.TrafficRouteResource{}
	},
	ResourceListFactory: func() model.ResourceList {
		return &sample_model.TrafficRouteResourceList{}
	},
}
