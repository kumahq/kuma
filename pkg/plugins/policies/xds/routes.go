package xds

import (
	"sort"

	envoy_route_api "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	// envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

type Routes struct {
	// service <- common virtual host
	// routes

}

type Match struct {
	Path    StringMatcher
	Headers map[string]StringMatcher
	From    map[string]string
	To      map[string]string
}

type MatchType uint32

const (
	Prefix MatchType = 0
	Exact            = 1
	Regex            = 2
)

type StringMatcher struct {
	Value     string
	MatchType MatchType
}

func GatherRoutes(vh *envoy_route_api.VirtualHost, isInbound bool) []*envoy_route_api.Route {
	vhRoutes := []*envoy_route_api.Route{}
		vhRoutes = append(vhRoutes, vh.Routes...)
		// for _, route := range vh.Routes{
			// allTags := envoy_metadata.ExtractListOfTags(route.Metadata)
			// if one match add to vh routes
			// vhRoutes = append(vhRoutes, route)
			// if len(allTags) == 0 {
			// 	vhRoutes = append(vhRoutes, route)
			// 	continue
			// }
			// for _, tags := range allTags{
			// 	subset := xds.SubsetFromTags(tags)
			// 	if rule := rules.Compute(subset); rule != nil {

			// 		//route and conf
			// 		vhRoutes = append(vhRoutes, route)
			// 	}
			// }
		// }
	// }
	// for _, route := range vhRoutes{
	// 	route
	// }

	// for _, routes := range virtualHosts {
	// 	for _, route := range routes {
	// 		pathMatcher := getPathMatcher(route.Match)
	// 		headerMatcher, headersOrder := getHeadersMatch(route.Match)
	// 		if isInbound {
	// 			value := headerMatcher[envoy_routes.TagsHeaderName]
				

	// 		} else {

	// 		}

	// 	}
	// }

	return vhRoutes
}

func getPathMatcher(matcher *envoy_route_api.RouteMatch) StringMatcher {
	stringMatcher := StringMatcher{}
	switch matcher.PathSpecifier.(type) {
	case *envoy_route_api.RouteMatch_Path:
		stringMatcher.Value = matcher.GetPath()
		stringMatcher.MatchType = Exact

	case *envoy_route_api.RouteMatch_Prefix:
		stringMatcher.Value = matcher.GetPrefix()
		stringMatcher.MatchType = Prefix
	case *envoy_route_api.RouteMatch_SafeRegex:
		if matcher.GetSafeRegex() != nil {
			stringMatcher.Value = matcher.GetSafeRegex().GetRegex()
			stringMatcher.MatchType = Regex
		}
	}
	return stringMatcher
}

func getHeadersMatch(matcher *envoy_route_api.RouteMatch) (map[string]StringMatcher, []string) {
	headersMatcher := map[string]StringMatcher{}
	var headers []string
	for _, hm := range matcher.Headers {
		headers = append(headers, hm.Name)
	}
	sort.Strings(headers)
	for _, hm := range matcher.Headers {
		stringMatcher := StringMatcher{}
		switch hm.HeaderMatchSpecifier.(type) {
		case *envoy_route_api.HeaderMatcher_PrefixMatch:
			stringMatcher.Value = hm.GetPrefixMatch()
			stringMatcher.MatchType = Prefix
		case *envoy_route_api.HeaderMatcher_ExactMatch:
			stringMatcher.Value = hm.GetExactMatch()
			stringMatcher.MatchType = Exact
		case *envoy_route_api.HeaderMatcher_SafeRegexMatch:
			if hm.GetSafeRegexMatch() != nil {
				stringMatcher.Value = hm.GetSafeRegexMatch().GetRegex()
				stringMatcher.MatchType = Regex
			}
		}
		headersMatcher[hm.Name] = stringMatcher
	}
	return headersMatcher, headers
}
