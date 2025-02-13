package filters

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
)

type RequestRedirectConfigurer struct {
	requestRedirect api.RequestRedirect
	withPrefixMatch bool
}

func NewRequestRedirect(redirect api.RequestRedirect, withPrefixMatch bool) *RequestRedirectConfigurer {
	return &RequestRedirectConfigurer{
		requestRedirect: redirect,
		withPrefixMatch: withPrefixMatch,
	}
}

func (f *RequestRedirectConfigurer) Configure(envoyRoute *envoy_route.Route) error {
	redirect := f.requestRedirect

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
			if f.withPrefixMatch {
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

	switch redirect.StatusCode {
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

	return nil
}

func regexRewrite(s string) *envoy_type_matcher.RegexMatchAndSubstitute {
	return &envoy_type_matcher.RegexMatchAndSubstitute{
		Pattern: &envoy_type_matcher.RegexMatcher{
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
