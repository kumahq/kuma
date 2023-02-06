package xds

import (
	"strings"

	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func routeFilter(filter api.Filter, route *envoy_route.Route, matchesPrefix bool) {
	switch filter.Type {
	case api.RequestHeaderModifierType:
		requestHeaderModifier(*filter.RequestHeaderModifier, route)
	case api.ResponseHeaderModifierType:
		responseHeaderModifier(*filter.ResponseHeaderModifier, route)
	case api.RequestRedirectType:
		requestRedirect(*filter.RequestRedirect, route, matchesPrefix)
	case api.URLRewriteType:
		urlRewrite(*filter.URLRewrite, route, matchesPrefix)
	}
}

func headerValues(raw common_api.HeaderValue) []string {
	return strings.Split(string(raw), ",")
}

func headerModifiers(mod api.HeaderModifier) ([]*envoy_config_core.HeaderValueOption, []string) {
	var options []*envoy_config_core.HeaderValueOption

	for _, set := range mod.Set {
		for i, headerValue := range headerValues(set.Value) {
			replace := &envoy_config_core.HeaderValueOption{
				Append: util_proto.Bool(i > 0),
				Header: &envoy_config_core.HeaderValue{
					Key:   string(set.Name),
					Value: headerValue,
				},
			}
			options = append(options, replace)
		}
	}
	for _, add := range mod.Add {
		for _, headerValue := range headerValues(add.Value) {
			appendOption := &envoy_config_core.HeaderValueOption{
				Append: util_proto.Bool(true),
				Header: &envoy_config_core.HeaderValue{
					Key:   string(add.Name),
					Value: headerValue,
				},
			}
			options = append(options, appendOption)
		}
	}

	return options, mod.Remove
}

func requestHeaderModifier(mod api.HeaderModifier, envoyRoute *envoy_route.Route) {
	options, removes := headerModifiers(mod)

	envoyRoute.RequestHeadersToAdd = append(envoyRoute.RequestHeadersToAdd, options...)
	envoyRoute.RequestHeadersToRemove = append(envoyRoute.RequestHeadersToRemove, removes...)
}

func responseHeaderModifier(mod api.HeaderModifier, envoyRoute *envoy_route.Route) {
	options, removes := headerModifiers(mod)

	envoyRoute.ResponseHeadersToAdd = append(envoyRoute.ResponseHeadersToAdd, options...)
	envoyRoute.ResponseHeadersToRemove = append(envoyRoute.ResponseHeadersToRemove, removes...)
}

func requestRedirect(redirect api.RequestRedirect, envoyRoute *envoy_route.Route, withPrefixMatch bool) {
	envoyRedirect := &envoy_route.RedirectAction{}
	if redirect.Hostname != nil {
		envoyRedirect.HostRedirect = string(*redirect.Hostname)
	}
	if redirect.Port != nil {
		envoyRedirect.PortRedirect = uint32(*redirect.Port)
	}
	if redirect.Scheme != nil {
		envoyRedirect.SchemeRewriteSpecifier = &envoy_route.RedirectAction_SchemeRedirect{
			SchemeRedirect: *redirect.Scheme,
		}
	}
	if redirect.Path != nil {
		switch redirect.Path.Type {
		case api.ReplaceFullPathType:
			envoyRedirect.PathRewriteSpecifier = regexToSpecifier(regexRewrite(*redirect.Path.ReplaceFullPath))
		case api.ReplacePrefixMatchType:
			if withPrefixMatch {
				if envoyRoute.Match.GetPath() != "" {
					// We have the "exact /prefix" match case
					envoyRedirect.PathRewriteSpecifier = regexToSpecifier(regexRewrite(*redirect.Path.ReplacePrefixMatch))
				} else if envoyRoute.Match.GetPrefix() != "" {
					// We have the "prefix /prefix/" match case
					envoyRedirect.PathRewriteSpecifier = prefixToSpecifier(*redirect.Path.ReplacePrefixMatch)
				}
			}
		}
	}

	switch pointer.DerefOr(redirect.StatusCode, 301) {
	case 301:
		envoyRedirect.ResponseCode = envoy_route.RedirectAction_MOVED_PERMANENTLY
	case 302:
		envoyRedirect.ResponseCode = envoy_route.RedirectAction_FOUND
	case 303:
		envoyRedirect.ResponseCode = envoy_route.RedirectAction_SEE_OTHER
	case 307:
		envoyRedirect.ResponseCode = envoy_route.RedirectAction_TEMPORARY_REDIRECT
	case 308:
		envoyRedirect.ResponseCode = envoy_route.RedirectAction_PERMANENT_REDIRECT
	default:
		panic("impossible redirect")
	}

	envoyRoute.Action = &envoy_route.Route_Redirect{
		Redirect: envoyRedirect,
	}
}

func urlRewrite(rewrite api.URLRewrite, envoyRoute *envoy_route.Route, withPrefixMatch bool) {
	action := &envoy_route.RouteAction{}
	if rewrite.Hostname != nil {
		action.HostRewriteSpecifier = &envoy_route.RouteAction_HostRewriteLiteral{
			HostRewriteLiteral: string(*rewrite.Hostname),
		}
	}
	if rewrite.Path != nil {
		switch rewrite.Path.Type {
		case api.ReplaceFullPathType:
			action.RegexRewrite = regexRewrite(*rewrite.Path.ReplaceFullPath)
		case api.ReplacePrefixMatchType:
			if withPrefixMatch {
				if envoyRoute.Match.GetPath() != "" {
					// We have the "exact /prefix" match case
					action.RegexRewrite = regexRewrite(*rewrite.Path.ReplacePrefixMatch)
				} else if envoyRoute.Match.GetPrefix() != "" {
					// We have the "prefix /prefix/" match case
					action.PrefixRewrite = *rewrite.Path.ReplacePrefixMatch
				}
			}
		}
	}
	envoyRoute.Action = &envoy_route.Route_Route{
		Route: action,
	}
}

func regexRewrite(s string) *envoy_type_matcher.RegexMatchAndSubstitute {
	return &envoy_type_matcher.RegexMatchAndSubstitute{
		Pattern: &envoy_type_matcher.RegexMatcher{
			EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
				GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
			},
			Regex: `.*`,
		},
		Substitution: s,
	}
}

func regexToSpecifier(regMatch *envoy_type_matcher.RegexMatchAndSubstitute) *envoy_route.RedirectAction_RegexRewrite {
	return &envoy_route.RedirectAction_RegexRewrite{
		RegexRewrite: regMatch,
	}
}

func prefixToSpecifier(prefix string) *envoy_route.RedirectAction_PrefixRewrite {
	return &envoy_route.RedirectAction_PrefixRewrite{
		PrefixRewrite: prefix,
	}
}
