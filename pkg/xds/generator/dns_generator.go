package generator

import (
	"cmp"
	"context"
	"encoding/json"
	"slices"

	core_system_names "github.com/kumahq/kuma/pkg/core/system_names"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/dns/dpapi"
	util_net "github.com/kumahq/kuma/pkg/util/net"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/dynconf"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

// OriginDNS is a marker to indicate by which ProxyGenerator resources were generated.
const OriginDNS = "dns"

type DNSGenerator struct{}

func (g DNSGenerator) Generate(_ context.Context, rs *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	tp := proxy.GetTransparentProxy()

	// DNS only makes sense when transparent proxy is used
	if !tp.Enabled() || !tp.Redirect.DNS.Enabled || proxy.Metadata.HasFeature(xds_types.FeatureBindOutbounds) {
		return nil, nil
	}
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
		if v6 != dnsOutbound.Address && tp.EnabledIPv6() {
			addresses = append(addresses, v6)
		}

		for _, domain := range dnsOutbound.Domains {
			vips[domain] = addresses
		}
	}
	getNameOrDefault := core_system_names.GetNameOrDefault(proxy.Metadata.HasFeature(xds_types.FeatureUnifiedResourceNaming))
	if proxy.Metadata.HasFeature(xds_types.FeatureEmbeddedDNS) {
		// This is purposefully set to 30s to avoid DNS cache stale with ExternalService and Kong Gateway see: https://github.com/kumahq/kuma/issues/13353.
		// https://github.com/kumahq/kuma/issues/13463
		dnsInfo := dpapi.DNSProxyConfig{TTL: 30, Records: []dpapi.DNSRecord{}}
		for name, addresses := range vips {
			dnsInfo.Records = append(dnsInfo.Records, dpapi.DNSRecord{
				Name: name,
				IPs:  addresses,
			})
		}

		slices.SortFunc(dnsInfo.Records, func(a, b dpapi.DNSRecord) int {
			return cmp.Compare(a.Name, b.Name)
		})
		bytes, err := json.Marshal(dnsInfo)
		if err != nil {
			return nil, err
		}
		err = dynconf.AddConfigRoute(proxy, rs, getNameOrDefault("dns", dpapi.PATH), dpapi.PATH, bytes)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	listener, err := envoy_listeners.
		NewInboundListenerBuilder(
			proxy.APIVersion,
			"127.0.0.1",
			uint32(tp.Redirect.DNS.Port),
			core_xds.SocketAddressProtocolUDP,
		).
		WithOverwriteName(
			getNameOrDefault(
				core_system_names.AsSystemName("dns"),
				names.GetDNSListenerName(),
			),
		).
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
