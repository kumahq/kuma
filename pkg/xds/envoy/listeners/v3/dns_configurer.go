package v3

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_data_dns "github.com/envoyproxy/go-control-plane/envoy/data/dns/v3"
	envoy_dns "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/dns_filter/v3alpha"
	v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	"github.com/kumahq/kuma/pkg/util/proto"
)

type DNSConfigurer struct {
	VIPs map[string]string
}

func (c *DNSConfigurer) Configure(listener *envoy_listener.Listener) error {
	pbst, err := proto.MarshalAnyDeterministic(c.dnsFilter())
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
	for domain, ip := range c.VIPs {
		virtualDomains = append(virtualDomains, &envoy_data_dns.DnsTable_DnsVirtualDomain{
			Name: domain,
			Endpoint: &envoy_data_dns.DnsTable_DnsEndpoint{
				EndpointConfig: &envoy_data_dns.DnsTable_DnsEndpoint_AddressList{
					AddressList: &envoy_data_dns.DnsTable_AddressList{
						Address: []string{
							ip,
						},
					},
				},
			},
			AnswerTtl: nil,
		})
	}

	return &envoy_dns.DnsFilterConfig{
		StatPrefix: "dns-resolver",
		ClientConfig: &envoy_dns.DnsFilterConfig_ClientContextConfig{
			UpstreamResolvers: []*envoy_core.Address{
				{
					Address: &envoy_core.Address_SocketAddress{
						SocketAddress: &envoy_core.SocketAddress{
							Address: "127.0.0.1",
							PortSpecifier: &envoy_core.SocketAddress_PortValue{
								PortValue: 5691,
							},
						},
					},
				},
			},
			MaxPendingLookups: 1,
		},
		ServerConfig: &envoy_dns.DnsFilterConfig_ServerContextConfig{
			ConfigSource: &envoy_dns.DnsFilterConfig_ServerContextConfig_InlineDnsTable{
				InlineDnsTable: &envoy_data_dns.DnsTable{
					VirtualDomains:     virtualDomains,
					ExternalRetryCount: 0,
					KnownSuffixes: []*v3.StringMatcher{
						{
							MatchPattern: &v3.StringMatcher_SafeRegex{
								SafeRegex: &v3.RegexMatcher{
									EngineType: &v3.RegexMatcher_GoogleRe2{},
									Regex:      ".*",
								},
							},
						},
					},
				},
			},
		},
	}
}
