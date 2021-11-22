package v3

import (
	"sort"
	"time"

	"github.com/Masterminds/semver/v3"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_data_dns "github.com/envoyproxy/go-control-plane/envoy/data/dns/v3"
	envoy_dns "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/dns_filter/v3"
	v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/golang/protobuf/ptypes/any"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

/*
	Envoy DNS API had a version change from 'v3alpha' to 'v3' in Envoy v1.20.0.
	Since go-control-plane contains only the latest protos, we had to vendor old 'v3alpha'
	protos for backward compatibility with previous versions of the Envoy.

	Imported directory contains copies of dns_filter protos from
	'envoyproxy/go-control-plane v0.9.9-0.20210914001841-ec3541a22836'
*/
import (
	envoy_dns_v3alpha "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3/compatibility/v3alpha"
)

type DNSConfigurer struct {
	VIPs         map[string][]string
	EmptyDNSPort uint32
	EnvoyVersion *mesh_proto.EnvoyVersion
}

func (c *DNSConfigurer) Configure(listener *envoy_listener.Listener) error {
	version, _ := c.EnvoyVersion.ParseVersion()
	v, err := semver.NewVersion(version)
	if err != nil {
		return err
	}
	var pbst *any.Any
	if v.LessThan(semver.MustParse("1.20.0")) {
		if pbst, err = util_proto.MarshalAnyDeterministic(c.dnsFilterV3Alpha()); err != nil {
			return err
		}
	} else {
		if pbst, err = util_proto.MarshalAnyDeterministic(c.dnsFilter()); err != nil {
			return err
		}
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
			DnsResolutionConfig: &envoy_core.DnsResolutionConfig{
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

func (c *DNSConfigurer) dnsFilterV3Alpha() *envoy_dns_v3alpha.DnsFilterConfig {
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

	return &envoy_dns_v3alpha.DnsFilterConfig{
		StatPrefix: "kuma_dns",
		ClientConfig: &envoy_dns_v3alpha.DnsFilterConfig_ClientContextConfig{
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
		ServerConfig: &envoy_dns_v3alpha.DnsFilterConfig_ServerContextConfig{
			ConfigSource: &envoy_dns_v3alpha.DnsFilterConfig_ServerContextConfig_InlineDnsTable{
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
