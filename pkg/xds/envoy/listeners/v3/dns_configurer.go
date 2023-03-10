package v3

import (
	"sort"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_data_dns "github.com/envoyproxy/go-control-plane/envoy/data/dns/v3"
	envoy_dns "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/dns_filter/v3"
	envoy_cares "github.com/envoyproxy/go-control-plane/envoy/extensions/network/dns_resolver/cares/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type DNSConfigurer struct {
	VIPs         map[string][]string
	EmptyDNSPort uint32
	EnvoyVersion *mesh_proto.EnvoyVersion
}

func (c *DNSConfigurer) Configure(listener *envoy_listener.Listener) error {
	pbst, err := util_proto.MarshalAnyDeterministic(c.dnsFilter())
	if err != nil {
		return err
	}

	listener.ListenerFilters = append(listener.ListenerFilters, &envoy_listener.ListenerFilter{
		Name: "envoy.filters.udp.dns_filter",
		ConfigType: &envoy_listener.ListenerFilter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	return nil
}

func (c *DNSConfigurer) dnsFilter() *envoy_dns.DnsFilterConfig {
	var virtualDomains []*envoy_data_dns.DnsTable_DnsVirtualDomain
	for domain, ips := range c.VIPs {
		virtualDomains = append(virtualDomains, &envoy_data_dns.DnsTable_DnsVirtualDomain{
			Name: domain,
			Endpoint: &envoy_data_dns.DnsTable_DnsEndpoint{
				EndpointConfig: &envoy_data_dns.DnsTable_DnsEndpoint_AddressList{
					AddressList: &envoy_data_dns.DnsTable_AddressList{
						Address: ips,
					},
				},
			},
			AnswerTtl: util_proto.Duration(time.Second * 30),
		})
	}
	sort.Stable(DnsTableByName(virtualDomains)) // for stable Envoy config

	resolverConfig := envoy_cares.CaresDnsResolverConfig{
		Resolvers: []*envoy_core.Address{
			{
				Address: &envoy_core.Address_SocketAddress{
					SocketAddress: &envoy_core.SocketAddress{
						Address: "127.0.0.1",
						PortSpecifier: &envoy_core.SocketAddress_PortValue{
							PortValue: c.EmptyDNSPort,
						},
					},
				},
			},
		},
	}

	return &envoy_dns.DnsFilterConfig{
		StatPrefix: "kuma_dns",
		ClientConfig: &envoy_dns.DnsFilterConfig_ClientContextConfig{
			// We configure upstream resolver to resolver that always returns that it could not find the domain (NXDOMAIN)
			// As for this moment there is no setting to disable upstream resolving.
			TypedDnsResolverConfig: &envoy_core.TypedExtensionConfig{
				Name:        "envoy.network.dns_resolver.cares",
				TypedConfig: util_proto.MustMarshalAny(&resolverConfig),
			},
			MaxPendingLookups: 256,
		},
		ServerConfig: &envoy_dns.DnsFilterConfig_ServerContextConfig{
			ConfigSource: &envoy_dns.DnsFilterConfig_ServerContextConfig_InlineDnsTable{
				InlineDnsTable: &envoy_data_dns.DnsTable{
					VirtualDomains:     virtualDomains,
					ExternalRetryCount: 0,
				},
			},
		},
	}
}

type DnsTableByName []*envoy_data_dns.DnsTable_DnsVirtualDomain

func (a DnsTableByName) Len() int      { return len(a) }
func (a DnsTableByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a DnsTableByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}
