package builders

import "github.com/kumahq/kuma/v3/pkg/core/xds"

type RoutingBuilder struct {
	res *xds.Routing
}

func Routing() *RoutingBuilder {
	return &RoutingBuilder{
		res: &xds.Routing{
			OutboundTargets: xds.EndpointMap{},
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
