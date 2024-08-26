package generator

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_net "github.com/kumahq/kuma/pkg/util/net"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

// OriginDNS is a marker to indicate by which ProxyGenerator resources were generated.
const OriginDNS = "dns"

type DNSGenerator struct{}

func (g DNSGenerator) Generate(ctx context.Context, _ *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	dnsPort := proxy.Metadata.GetDNSPort()
	if dnsPort == 0 {
		return nil, nil
	}

	if proxy.Dataplane.Spec.GetNetworking().GetTransparentProxying() == nil {
		return nil, nil // DNS only makes sense when transparent proxy is used
	}
	ipV6Enabled := proxy.Dataplane.Spec.GetNetworking().GetTransparentProxying().IpFamilyMode != mesh_proto.Dataplane_Networking_TransparentProxying_IPv4

	outboundIPs := map[string]bool{}
	for _, outbound := range proxy.Outbounds {
		outboundIPs[outbound.GetAddress()] = true
	}

	vips := map[string][]string{}
	for _, dnsOutbound := range xdsCtx.Mesh.VIPDomains {
		if !outboundIPs[dnsOutbound.Address] {
			continue // if there is no outbound for given address, there is no point of providing DNS resolver
		}
		addresses := []string{dnsOutbound.Address}
		v6 := util_net.ToV6(dnsOutbound.Address)
		if v6 != dnsOutbound.Address && ipV6Enabled {
			addresses = append(addresses, v6)
		}

		for _, domain := range dnsOutbound.Domains {
			vips[domain] = addresses
		}
	}

	listener, err := envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, "127.0.0.1", dnsPort, core_xds.SocketAddressProtocolUDP).
		WithOverwriteName(names.GetDNSListenerName()).
		Configure(envoy_listeners.DNS(vips)).
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
