package generator

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_net "github.com/kumahq/kuma/pkg/util/net"
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

	vips := map[string][]string{}
	for _, dnsOutbound := range proxy.Routing.VipDomains {
		for _, domain := range dnsOutbound.Domains {
			vips[domain] = []string{dnsOutbound.Address}
			v6 := util_net.ToV6(dnsOutbound.Address)
			if v6 != dnsOutbound.Address { // It's already a v6
				vips[domain] = append(vips[domain], v6)
			}
		}
	}

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
