package builders

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

type ProxyBuilder struct {
	res *xds.Proxy
}

func Proxy() *ProxyBuilder {
	return &ProxyBuilder{
		res: &xds.Proxy{
			APIVersion:        envoy_common.APIV3,
			Dataplane:         &core_mesh.DataplaneResource{},
			Metadata:          &xds.DataplaneMetadata{},
			Policies:          xds.MatchedPolicies{},
			Routing:           xds.Routing{},
			RuntimeExtensions: map[string]interface{}{},
			Zone:              "test-zone",
		},
	}
}

func (p *ProxyBuilder) Build() *xds.Proxy {
	return p.res
}

func (p *ProxyBuilder) With(fn func(*xds.Proxy)) *ProxyBuilder {
	fn(p.res)
	return p
}

func (p *ProxyBuilder) WithApiVersion(apiVersion core_xds.APIVersion) *ProxyBuilder {
	p.res.APIVersion = apiVersion
	return p
}

func (p *ProxyBuilder) WithZone(zone string) *ProxyBuilder {
	p.res.Zone = zone
	return p
}

func (p *ProxyBuilder) WithDataplane(dataplane *builders.DataplaneBuilder) *ProxyBuilder {
	p.res.Dataplane = dataplane.Build()
	return p
}

func (p *ProxyBuilder) WithMetadata(metadata *xds.DataplaneMetadata) *ProxyBuilder {
	p.res.Metadata = metadata
	return p
}

func (p *ProxyBuilder) WithSecretsTracker(secretsTracker core_xds.SecretsTracker) *ProxyBuilder {
	p.res.SecretsTracker = secretsTracker
	return p
}

func (p *ProxyBuilder) WithPolicies(policies *MatchedPoliciesBuilder) *ProxyBuilder {
	p.res.Policies = *policies.Build()
	return p
}

func (p *ProxyBuilder) WithOutbounds(outbounds xds.Outbounds) *ProxyBuilder {
	p.res.Outbounds = outbounds
	return p
}

func (p *ProxyBuilder) WithRouting(routing *RoutingBuilder) *ProxyBuilder {
	p.res.Routing = *routing.Build()
	return p
}

func (p *ProxyBuilder) WithID(id xds.ProxyId) *ProxyBuilder {
	p.res.Id = id
	return p
}
