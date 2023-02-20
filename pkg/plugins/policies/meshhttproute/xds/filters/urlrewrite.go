package filters

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
)

type URLRewriteConfigurer struct {
	urlRewrite      api.URLRewrite
	withPrefixMatch bool
}

func NewURLRewrite(rewrite api.URLRewrite, withPrefixMatch bool) *URLRewriteConfigurer {
	return &URLRewriteConfigurer{
		urlRewrite:      rewrite,
		withPrefixMatch: withPrefixMatch,
	}
}

func (f *URLRewriteConfigurer) Configure(envoyRoute *envoy_route.Route) error {
	rewrite := f.urlRewrite

	return UpdateRouteAction(envoyRoute, func(action *envoy_route.RouteAction) error {
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
				if f.withPrefixMatch {
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

		return nil
	})
}
