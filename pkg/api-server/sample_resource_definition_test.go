package api_server_test

import (
	"github.com/kumahq/kuma/pkg/api-server/definitions"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	sample_model "github.com/kumahq/kuma/pkg/test/resources/apis/sample"
)

var SampleTrafficRouteWsDefinition = definitions.ResourceWsDefinition{
	Name: "Sample Traffic Route",
	Path: "sample-traffic-routes",
	ResourceFactory: func() model.Resource {
		return sample_model.NewTrafficRouteResource()
	},
	ResourceListFactory: func() model.ResourceList {
		return &sample_model.TrafficRouteResourceList{}
	},
	ReadOnly: false,
}
