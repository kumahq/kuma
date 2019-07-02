package api_server_test

import (
	api_server "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	sample_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/test/apis/sample/v1alpha1"
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
	SpecFactory: func() model.ResourceSpec {
		return &sample_proto.TrafficRoute{}
	},
	SampleSpec: sample_proto.TrafficRoute{},
}
