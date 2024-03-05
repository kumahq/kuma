package builders

import "github.com/kumahq/kuma/pkg/core/xds"

type RoutingBuilder struct {
	res *xds.Routing
}

func Routing() *RoutingBuilder {
	return &RoutingBuilder{
		res: &xds.Routing{
			TrafficRoutes:                  xds.RouteMap{},
			OutboundTargets:                xds.EndpointMap{},
			ExternalServiceOutboundTargets: xds.EndpointMap{},
		},
	}
}

func (r *RoutingBuilder) Build() *xds.Routing {
	return r.res
}

func (r *RoutingBuilder) With(fn func(*xds.Routing)) *RoutingBuilder {
	fn(r.res)
	return r
}

func (r *RoutingBuilder) WithOutboundTargets(outboundTargets *EndpointMapBuilder) *RoutingBuilder {
	r.res.OutboundTargets = outboundTargets.Build()
	return r
}

func (r *RoutingBuilder) WithExternalServiceOutboundTargets(externalServiceOutboundTargets *EndpointMapBuilder) *RoutingBuilder {
	r.res.ExternalServiceOutboundTargets = externalServiceOutboundTargets.Build()
	return r
}
