package xds

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

func regexMatcher(regex string) *envoy_type_matcher.RegexMatcher {
	return &envoy_type_matcher.RegexMatcher{
		Regex: regex,
	}
}

func routeHeadersMatch(envoyMatch *envoy_route.RouteMatch, headers []common_api.HeaderMatch) {
	// We ignore multiple matchers for the same name, though this is also
	// validated
	matchedNames := map[common_api.HeaderName]struct{}{}

	for _, header := range headers {
		if _, ok := matchedNames[header.Name]; ok {
			continue
		}
		matchedNames[header.Name] = struct{}{}

		var matcher envoy_route.HeaderMatcher

		matchType := common_api.HeaderMatchExact
		if header.Type != nil {
			matchType = *header.Type
		}

		switch matchType {
		case common_api.HeaderMatchExact:
			matcher = envoy_route.HeaderMatcher{
				Name: string(header.Name),
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_StringMatch{
					StringMatch: &envoy_type_matcher.StringMatcher{
						MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
							Exact: string(header.Value),
						},
					},
				},
			}
		case common_api.HeaderMatchPresent:
			matcher = envoy_route.HeaderMatcher{
				Name: string(header.Name),
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_PresentMatch{
					PresentMatch: true,
				},
			}
		case common_api.HeaderMatchRegularExpression:
			matcher = envoy_route.HeaderMatcher{
				Name: string(header.Name),
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_StringMatch{
					StringMatch: &envoy_type_matcher.StringMatcher{
						MatchPattern: &envoy_type_matcher.StringMatcher_SafeRegex{
							SafeRegex: regexMatcher(string(header.Value)),
						},
					},
				},
			}
		case common_api.HeaderMatchAbsent:
			matcher = envoy_route.HeaderMatcher{
				Name: string(header.Name),
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_PresentMatch{
					PresentMatch: false,
				},
			}
		case common_api.HeaderMatchPrefix:
			if header.Value != "" {
				matcher = envoy_route.HeaderMatcher{
					Name: string(header.Name),
					HeaderMatchSpecifier: &envoy_route.HeaderMatcher_StringMatch{
						StringMatch: &envoy_type_matcher.StringMatcher{
							MatchPattern: &envoy_type_matcher.StringMatcher_Prefix{
								Prefix: string(header.Value),
							},
						},
					},
				}
			} else {
				// the prefix matcher doesn't like empty string prefixes
				matcher = envoy_route.HeaderMatcher{
					Name: string(header.Name),
					HeaderMatchSpecifier: &envoy_route.HeaderMatcher_PresentMatch{
						PresentMatch: true,
					},
				}
			}
		default:
			panic("impossible")
		}

		envoyMatch.Headers = append(envoyMatch.Headers, &matcher)
	}
}
