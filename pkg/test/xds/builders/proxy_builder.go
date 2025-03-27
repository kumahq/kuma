package builders

import (
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

type ProxyBuilder struct {
	res *core_xds.Proxy
}

func Proxy() *ProxyBuilder {
	return &ProxyBuilder{
		res: &core_xds.Proxy{
			APIVersion:        envoy_common.APIV3,
			Dataplane:         &core_mesh.DataplaneResource{},
			Metadata:          &core_xds.DataplaneMetadata{},
			Policies:          core_xds.MatchedPolicies{},
			Routing:           core_xds.Routing{},
			RuntimeExtensions: map[string]interface{}{},
			Zone:              "test-zone",
		},
	}
}

func (p *ProxyBuilder) Build() *core_xds.Proxy {
	return p.res
}

func (p *ProxyBuilder) With(fn func(*core_xds.Proxy)) *ProxyBuilder {
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

func (p *ProxyBuilder) WithMetadata(metadata *core_xds.DataplaneMetadata) *ProxyBuilder {
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

func (p *ProxyBuilder) WithOutbounds(outbounds xds_types.Outbounds) *ProxyBuilder {
	p.res.Outbounds = outbounds
	return p
}

func (p *ProxyBuilder) WithRouting(routing *RoutingBuilder) *ProxyBuilder {
	p.res.Routing = *routing.Build()
	return p
}

func (p *ProxyBuilder) WithID(id core_xds.ProxyId) *ProxyBuilder {
	p.res.Id = id
	return p
}

func (p *ProxyBuilder) WithInternalAddresses(addresses ...core_xds.InternalAddress) *ProxyBuilder {
	p.res.InternalAddresses = append(p.res.InternalAddresses, addresses...)
	return p
}
