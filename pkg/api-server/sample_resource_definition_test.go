package api_server_test

import (
	"github.com/kumahq/kuma/pkg/api-server/definitions"
	sample_model "github.com/kumahq/kuma/pkg/test/resources/apis/sample"
)

var SampleTrafficRouteWsDefinition = definitions.ResourceWsDefinition{
	Type:     sample_model.TrafficRouteType,
	Path:     "sample-traffic-routes",
	ReadOnly: false,
}
