package generator

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

// OriginInbound is a marker to indicate by which ProxyGenerator resources were generated.
const OriginDNS = "dns"

type DNSGenerator struct {
}

func (g DNSGenerator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	if proxy.APIVersion == envoy_common.APIV2 {
		return nil, nil // DNS Filter is not supported for API V2
	}

	vips := ctx.ControlPlane.DNSResolver.GetVIPs()
	ipToDomain := map[string]string{}
	for domain, ip := range vips {
		ipToDomain[ip] = domain
	}

	meshedVips := map[string]string{}
	for _, outbound := range proxy.Dataplane.Spec.GetNetworking().GetOutbound() {
		if domain, ok := ipToDomain[outbound.Address]; ok {
			meshedVips[domain + "." + ctx.ControlPlane.DNSResolver.GetDomain()] = outbound.Address
			endpoints := proxy.Routing.OutboundTargets[outbound.Tags[mesh_proto.ServiceTag]]
			for _, endpoint := range endpoints {
				if endpoint.ExternalService != nil && endpoint.ExternalService.Host != "" {
					meshedVips[endpoint.ExternalService.Host] = outbound.Address
				}
			}
		}
	}


	listener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
		Configure(envoy_listeners.InboundListener(names.GetDNSListenerName(), proxy.Dataplane.GetIP(), 5690, core_xds.SocketAddressProtocolUDP)). // todo parametrize port
		Configure(envoy_listeners.DNS(meshedVips)).
		Build()
	if err != nil {
		return nil, err
	}

	resources := core_xds.NewResourceSet()
	resources.Add(&core_xds.Resource{
		Name:     names.GetDNSListenerName(),
		Resource: listener,
		Origin:   OriginDNS,
	})
	return resources, nil
}
