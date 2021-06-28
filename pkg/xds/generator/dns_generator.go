package generator

import (
	"strings"

	"github.com/asaskevich/govalidator"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

// OriginDNS is a marker to indicate by which ProxyGenerator resources were generated.
const OriginDNS = "dns"

type DNSGenerator struct {
}

func (g DNSGenerator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	dnsPort := proxy.Metadata.GetDNSPort()
	emptyDnsPort := proxy.Metadata.GetEmptyDNSPort()
	if dnsPort == 0 || emptyDnsPort == 0 {
		return nil, nil
	}

	if proxy.Dataplane.Spec.GetNetworking().GetTransparentProxying() == nil {
		return nil, nil // DNS only makes sense when transparent proxy is used
	}

	vips := g.computeVIPs(ctx, proxy)
	listener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
		Configure(envoy_listeners.InboundListener(names.GetDNSListenerName(), "127.0.0.1", dnsPort, core_xds.SocketAddressProtocolUDP)).
		Configure(envoy_listeners.DNS(vips, emptyDnsPort)).
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

func (g DNSGenerator) computeVIPs(ctx xds_context.Context, proxy *core_xds.Proxy) map[string]string {
	domainsByIPs := ctx.ControlPlane.DNSResolver.GetVIPs().FQDNsByIPs()
	meshedVips := map[string]string{}
	for _, outbound := range proxy.Dataplane.Spec.GetNetworking().GetOutbound() {
		if domain, ok := domainsByIPs[outbound.Address]; ok {
			// add regular .mesh domain
			meshedVips[domain+"."+ctx.ControlPlane.DNSResolver.GetDomain()] = outbound.Address
			meshedVips[strings.ReplaceAll(domain, "_", ".")+"."+ctx.ControlPlane.DNSResolver.GetDomain()] = outbound.Address
			// add hostname from address in external service
			endpoints := proxy.Routing.OutboundTargets[outbound.Tags[mesh_proto.ServiceTag]]
			for _, endpoint := range endpoints {
				if govalidator.IsDNSName(endpoint.Target) {
					if endpoint.ExternalService != nil && endpoint.Target != "" {
						meshedVips[endpoint.Target] = outbound.Address
					}
				}
			}
		}
	}
	return meshedVips
}
