package v3

import (
	"sort"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_data_dns "github.com/envoyproxy/go-control-plane/envoy/data/dns/v3"
	envoy_dns "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/dns_filter/v3alpha"
	v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type DNSConfigurer struct {
	VIPs         map[string][]string
	EmptyDNSPort uint32
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

	return &envoy_dns.DnsFilterConfig{
		StatPrefix: "kuma_dns",
		ClientConfig: &envoy_dns.DnsFilterConfig_ClientContextConfig{
			// We configure upstream resolver to resolver that always returns that it could not find the domain (NXDOMAIN)
			// As for this moment there is no setting to disable upstream resolving.
			UpstreamResolvers: []*envoy_core.Address{
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
			MaxPendingLookups: 256,
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
									// This is just an indicator that Envoy will try to resolve any hostname using list of virtual domains
									Regex: ".*",
								},
							},
						},
					},
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
