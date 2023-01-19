package xds

import (
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
)

func regexMatcher(regex string) *envoy_type_matcher.RegexMatcher {
	return &envoy_type_matcher.RegexMatcher{
		Regex: regex,
		EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
			GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
		},
	}
}
