package xds

import (
	"fmt"
	"strconv"
	"strings"

	xds "github.com/cncf/xds/go/xds/core/v3"
	v32 "github.com/cncf/xds/go/xds/type/matcher/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	network_input "github.com/envoyproxy/go-control-plane/envoy/extensions/matching/common_inputs/network/v3"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/proto"
)

var (
	DestinationPortInput = &xds.TypedExtensionConfig{
		Name:        "port",
		TypedConfig: proto.MustMarshalAny(&network_input.DestinationPortInput{}),
	}
	DestinationIPInput = &xds.TypedExtensionConfig{
		Name:        "ip",
		TypedConfig: proto.MustMarshalAny(&network_input.DestinationIPInput{}),
	}
	SNIInput = &xds.TypedExtensionConfig{
		Name:        "sni",
		TypedConfig: proto.MustMarshalAny(&network_input.ServerNameInput{}),
	}
	TransportProtocolInput = &xds.TypedExtensionConfig{
		Name:        "transport-protocol",
		TypedConfig: proto.MustMarshalAny(&network_input.TransportProtocolInput{}),
	}
	ApplicationProtocolInput = &xds.TypedExtensionConfig{
		Name:        "application-protocol",
		TypedConfig: proto.MustMarshalAny(&network_input.ApplicationProtocolInput{}),
	}
)

func createFieldMatcher(predicate []*v32.Matcher_MatcherList_Predicate, filterChainName string) *v32.Matcher_MatcherList_FieldMatcher {
	return &v32.Matcher_MatcherList_FieldMatcher{
		OnMatch: &v32.Matcher_OnMatch{
			OnMatch: &v32.Matcher_OnMatch_Action{
				Action: &xds.TypedExtensionConfig{
					Name:        filterChainName,
					TypedConfig: proto.MustMarshalAny(proto.String(filterChainName)),
				},
			},
		},
		Predicate: &v32.Matcher_MatcherList_Predicate{
			MatchType: &v32.Matcher_MatcherList_Predicate_OrMatcher{
				OrMatcher: &v32.Matcher_MatcherList_Predicate_PredicateList{
					Predicate: predicate,
				},
			},
		},
	}
}

func sniWildcardDomainMatcher(domain, filterChainName string) *v32.Matcher_MatcherList_FieldMatcher {
	predicate := []*v32.Matcher_MatcherList_Predicate{
		{
			MatchType: &v32.Matcher_MatcherList_Predicate_SinglePredicate_{
				SinglePredicate: &v32.Matcher_MatcherList_Predicate_SinglePredicate{
					Input: SNIInput,
					Matcher: &v32.Matcher_MatcherList_Predicate_SinglePredicate_ValueMatch{
						ValueMatch: &v32.StringMatcher{
							MatchPattern: &v32.StringMatcher_SafeRegex{
								SafeRegex: &v32.RegexMatcher{
									EngineType: &v32.RegexMatcher_GoogleRe2{},
									Regex:      fmt.Sprintf(".%s", domain),
								},
							},
						},
					},
				},
			},
		},
	}
	return createFieldMatcher(predicate, filterChainName)
}

func sniDomainMatcher(domain, filterChainName string) *v32.Matcher_MatcherList_FieldMatcher {
	predicate := []*v32.Matcher_MatcherList_Predicate{
		{
			MatchType: &v32.Matcher_MatcherList_Predicate_SinglePredicate_{
				SinglePredicate: &v32.Matcher_MatcherList_Predicate_SinglePredicate{
					Input: SNIInput,
					Matcher: &v32.Matcher_MatcherList_Predicate_SinglePredicate_ValueMatch{
						ValueMatch: &v32.StringMatcher{
							MatchPattern: &v32.StringMatcher_Exact{
								Exact: domain,
							},
						},
					},
				},
			},
		},
	}
	return createFieldMatcher(predicate, filterChainName)
}

func appProtocolMatcher(filterChainName string) *v32.Matcher_MatcherList_FieldMatcher {
	predicate := []*v32.Matcher_MatcherList_Predicate{
		{
			MatchType: &v32.Matcher_MatcherList_Predicate_SinglePredicate_{
				SinglePredicate: &v32.Matcher_MatcherList_Predicate_SinglePredicate{
					Input: ApplicationProtocolInput,
					Matcher: &v32.Matcher_MatcherList_Predicate_SinglePredicate_ValueMatch{
						ValueMatch: &v32.StringMatcher{
							MatchPattern: &v32.StringMatcher_Exact{
								Exact: "http/1.1",
							},
						},
					},
				},
			},
		},
		{
			MatchType: &v32.Matcher_MatcherList_Predicate_SinglePredicate_{
				SinglePredicate: &v32.Matcher_MatcherList_Predicate_SinglePredicate{
					Input: ApplicationProtocolInput,
					Matcher: &v32.Matcher_MatcherList_Predicate_SinglePredicate_ValueMatch{
						ValueMatch: &v32.StringMatcher{
							MatchPattern: &v32.StringMatcher_Exact{
								Exact: "h2c",
							},
						},
					},
				},
			},
		},
	}
	return createFieldMatcher(predicate, filterChainName)
}

func ipMatcher(ip, filterChainName string) *v32.Matcher_MatcherList_FieldMatcher {
	predicate := []*v32.Matcher_MatcherList_Predicate{
		{
			MatchType: &v32.Matcher_MatcherList_Predicate_SinglePredicate_{
				SinglePredicate: &v32.Matcher_MatcherList_Predicate_SinglePredicate{
					Input: DestinationIPInput,
					Matcher: &v32.Matcher_MatcherList_Predicate_SinglePredicate_CustomMatch{
						CustomMatch: &xds.TypedExtensionConfig{
							Name: "ip-matcher",
							TypedConfig: proto.MustMarshalAny(&v32.IPMatcher_IPRangeMatcher{
								Ranges: []*xds.CidrRange{
									{
										AddressPrefix: ip,
										PrefixLen:     proto.UInt32(32),
									},
								},
							}),
						},
					},
				},
			},
		},
	}
	return createFieldMatcher(predicate, filterChainName)
}

func cidrMatcher(cidr string, filterChainName string) *v32.Matcher_MatcherList_FieldMatcher {
	ip, mask := getIPandMask(cidr)
	predicate := []*v32.Matcher_MatcherList_Predicate{
		{
			MatchType: &v32.Matcher_MatcherList_Predicate_SinglePredicate_{
				SinglePredicate: &v32.Matcher_MatcherList_Predicate_SinglePredicate{
					Input: DestinationIPInput,
					Matcher: &v32.Matcher_MatcherList_Predicate_SinglePredicate_CustomMatch{
						CustomMatch: &xds.TypedExtensionConfig{
							Name: "ip-matcher",
							TypedConfig: proto.MustMarshalAny(&v32.IPMatcher_IPRangeMatcher{
								Ranges: []*xds.CidrRange{
									{
										AddressPrefix: ip,
										PrefixLen:     proto.UInt32(mask),
									},
								},
							}),
						},
					},
				},
			},
		},
	}
	return createFieldMatcher(predicate, filterChainName)
}

func portMatchers(matchers []*v32.Matcher_MatcherList_FieldMatcher) *v32.Matcher_OnMatch {
	return &v32.Matcher_OnMatch{
		OnMatch: &v32.Matcher_OnMatch_Matcher{
			Matcher: &v32.Matcher{
				MatcherType: &v32.Matcher_MatcherList_{
					MatcherList: &v32.Matcher_MatcherList{
						Matchers: matchers,
					},
				},
			},
		},
	}
}

func destinationPortsMatcher(allPortsMatcher *v32.Matcher_OnMatch, portsMatcher map[string]*v32.Matcher_OnMatch) *v32.Matcher_OnMatch {
	return &v32.Matcher_OnMatch{
		OnMatch: &v32.Matcher_OnMatch_Matcher{
			Matcher: &v32.Matcher{
				OnNoMatch: allPortsMatcher,
				MatcherType: &v32.Matcher_MatcherTree_{
					MatcherTree: &v32.Matcher_MatcherTree{
						Input: DestinationPortInput,
						TreeType: &v32.Matcher_MatcherTree_ExactMatchMap{
							ExactMatchMap: &v32.Matcher_MatcherTree_MatchMap{
								Map: portsMatcher,
							},
						},
					},
				},
			},
		},
	}
}

func addFilterChainToGenerate(
	filterChainName string,
	match api.Match,
	filterChainsAccumulator map[string]FilterChainConfiguration,
) {
	if value, found := filterChainsAccumulator[filterChainName]; found {
		if match.Protocol != api.ProtocolType("tcp") && match.Protocol != api.ProtocolType("tls") {
			routes := append(value.Routes, Route{ClusterName: clusterName(match), Domain: match.Value})
			filterChainsAccumulator[filterChainName] = FilterChainConfiguration{
				Protocol: core_mesh.ParseProtocol(string(value.Protocol)),
				Routes:   routes,
			}
		}
	} else {
		filterChainsAccumulator[filterChainName] = FilterChainConfiguration{
			Protocol: core_mesh.ParseProtocol(string(match.Protocol)),
			Routes: []Route{
				{
					ClusterName: clusterName(match),
					Domain:      match.Value,
				},
			},
		}
	}
}

func getValueMatchers(matchers MatchersPerType, filterChainsAccumulator map[string]FilterChainConfiguration) []*v32.Matcher_MatcherList_FieldMatcher {
	allPortsMatchers := []*v32.Matcher_MatcherList_FieldMatcher{}
	for _, match := range matchers[Domain] {
		name := filterChainName(match)
		addFilterChainToGenerate(name, match, filterChainsAccumulator)
		if match.Protocol == api.ProtocolType("tls") {
			allPortsMatchers = append(allPortsMatchers, sniDomainMatcher(match.Value, name))
		} else {
			allPortsMatchers = append(allPortsMatchers, appProtocolMatcher(name))
		}
	}
	for _, match := range matchers[IP] {
		name := filterChainName(match)
		addFilterChainToGenerate(name, match, filterChainsAccumulator)
		allPortsMatchers = append(allPortsMatchers, ipMatcher(match.Value, name))
	}
	for _, match := range matchers[WildcardDomain] {
		name := filterChainName(match)
		addFilterChainToGenerate(name, match, filterChainsAccumulator)
		if match.Protocol == api.ProtocolType("tls") {
			allPortsMatchers = append(allPortsMatchers, sniWildcardDomainMatcher(match.Value, name))
		} else {
			allPortsMatchers = append(allPortsMatchers, appProtocolMatcher(name))
		}
	}
	for _, match := range matchers[CIDR] {
		name := filterChainName(match)
		addFilterChainToGenerate(name, match, filterChainsAccumulator)
		allPortsMatchers = append(allPortsMatchers, cidrMatcher(match.Value, name))
	}
	return allPortsMatchers
}

func transportProtocolsMatcher(protocols map[string]*v32.Matcher_OnMatch) *v32.Matcher {
	return &v32.Matcher{
		MatcherType: &v32.Matcher_MatcherTree_{
			MatcherTree: &v32.Matcher_MatcherTree{
				Input: TransportProtocolInput,
				TreeType: &v32.Matcher_MatcherTree_ExactMatchMap{
					ExactMatchMap: &v32.Matcher_MatcherTree_MatchMap{
						Map: protocols,
					},
				},
			},
		},
	}
}

func generateProtocolMatchers(protocolMatchers MatchersPerPort, filterChainsAccumulator map[string]FilterChainConfiguration) *v32.Matcher_OnMatch {
	portsMatchers := map[string]*v32.Matcher_OnMatch{}
	for port, matchers := range protocolMatchers {
		// port 0 means traffic goes to all port so we resolv them at the end
		if port == 0 {
			continue
		}
		portMatcher := getValueMatchers(matchers, filterChainsAccumulator)
		portsMatchers[fmt.Sprint(port)] = portMatchers(portMatcher)
	}

	matchAllPorts := &v32.Matcher_OnMatch{}
	matchers, ok := protocolMatchers[0]
	if ok {
		allPortsMatchers := getValueMatchers(matchers, filterChainsAccumulator)
		matchAllPorts = &v32.Matcher_OnMatch{
			OnMatch: &v32.Matcher_OnMatch_Matcher{
				Matcher: &v32.Matcher{
					MatcherType: &v32.Matcher_MatcherList_{
						MatcherList: &v32.Matcher_MatcherList{
							Matchers: allPortsMatchers,
						},
					},
				},
			},
		}
	}
	return destinationPortsMatcher(matchAllPorts, portsMatchers)
}

func clusterName(match api.Match) string {
	if match.Port == nil {
		return fmt.Sprintf("meshpassthrough_%s_*", match.Value)
	}
	return fmt.Sprintf("meshpassthrough_%s_%d", match.Value, *match.Port)
}

func filterChainName(match api.Match) string {
	port := "*"
	if match.Port != nil {
		port = fmt.Sprintf("%d", *match.Port)
	}
	if match.Protocol == api.ProtocolType("tcp") || match.Protocol == api.ProtocolType("tls") {
		return fmt.Sprintf("meshpassthrough_%s_%s", match.Value, port)
	}
	return fmt.Sprintf("meshpassthrough_http_%s", port)
}

func getIPandMask(cidr string) (string, uint32) {
	parts := strings.Split(cidr, "/")
	prefixLength, err := strconv.Atoi(parts[1])
	// that shouldn't happened because we validate object when adding
	// TODO: add IPv6 support
	if err != nil {
		prefixLength = 32
	}
	return parts[0], uint32(prefixLength)
}

type FilterChainConfiguration struct {
	Protocol core_mesh.Protocol
	Routes   []Route
}

type Route struct {
	Name        string
	Domain      string
	ClusterName string
}

type FilterChainMatcherConfigurer struct {
	Conf api.Conf
}

func (c FilterChainMatcherConfigurer) Configure(listener *envoy_listener.Listener) map[string]FilterChainConfiguration {
	tls, rawBuffer := GetOrderedMatchers(c.Conf)
	filterChainsAccumulator := map[string]FilterChainConfiguration{}
	// TODO: add IPv6 support
	protocol := map[string]*v32.Matcher_OnMatch{}
	protocol["tls"] = generateProtocolMatchers(tls, filterChainsAccumulator)
	protocol["raw_buffer"] = generateProtocolMatchers(rawBuffer, filterChainsAccumulator)
	config := transportProtocolsMatcher(protocol)
	listener.FilterChainMatcher = config
	return filterChainsAccumulator
}
