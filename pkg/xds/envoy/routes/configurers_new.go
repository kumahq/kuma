package routes

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/pkg/errors"
)

// TODO(bartsmykla): move types to better suiting place

type Rewrite struct {
	ReplaceFullPath *string

	ReplacePrefixMatch *string

	// HostToBackendHostname indicates that during forwarding, the host header
	// should be swapped with the hostname of the upstream host chosen by the
	// Envoy's cluster manager.
	HostToBackendHostname bool
}

// Redirection is an action that responds to a HTTP request with a HTTP
// redirect response.
type Redirection struct {
	Status      uint32 // HTTP status code.
	Scheme      string // URL scheme (optional).
	Host        string // URL host (optional).
	Port        uint32 // URL port (optional).
	PathRewrite *Rewrite

	StripQuery bool // Whether to strip the query string.
}

// TODO(bartsmykla): move types to better suiting place

type RedirectConfigurer struct {
	MatchPath    string
	NewPath      string
	Port         uint32
	AllowGetOnly bool
}

func (c RedirectConfigurer) Configure(virtualHost *envoy_route.VirtualHost) error {
	var headersMatcher []*envoy_route.HeaderMatcher
	if c.AllowGetOnly {
		matcher := envoy_type_matcher.StringMatcher{
			MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
				Exact: "GET",
			},
		}
		headersMatcher = []*envoy_route.HeaderMatcher{
			{
				Name: ":method",
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_StringMatch{
					StringMatch: &matcher,
				},
			},
		}
	}
	virtualHost.Routes = append(virtualHost.Routes, &envoy_route.Route{
		Match: &envoy_route.RouteMatch{
			PathSpecifier: &envoy_route.RouteMatch_Path{
				Path: c.MatchPath,
			},
			Headers: headersMatcher,
		},
		Action: &envoy_route.Route_Redirect{
			Redirect: &envoy_route.RedirectAction{
				PortRedirect: c.Port,
				PathRewriteSpecifier: &envoy_route.RedirectAction_PathRedirect{
					PathRedirect: c.NewPath,
				},
			},
		},
	})
	return nil
}

func RouteRedirectScheme(scheme string) RouteRedirectConfigurer {
	if scheme == "" {
		return RouteRedirectConfigurer(nil)
	}

	return RouteRedirectMustConfigureFunc(func(redirect *envoy_route.RedirectAction) {
		schemaRedirect := &envoy_route.RedirectAction_SchemeRedirect{
			SchemeRedirect: scheme,
		}

		redirect.SchemeRewriteSpecifier = schemaRedirect
	})
}

func RouteRedirectHost(host string) RouteRedirectConfigurer {
	if host == "" {
		return RouteRedirectConfigurer(nil)
	}

	return RouteRedirectMustConfigureFunc(func(redirect *envoy_route.RedirectAction) {
		redirect.HostRedirect = host
	})
}

func RouteRedirectPort(port, redirectPort uint32) RouteRedirectConfigurer {
	return RouteRedirectMustConfigureFunc(func(redirect *envoy_route.RedirectAction) {
		redirect.PortRedirect = port

		if redirectPort > 0 {
			redirect.PortRedirect = redirectPort
		}
	})
}

// RouteActionRedirect configures the route to automatically response
// with an HTTP redirection. This replaces any previous action specification.
func RouteActionRedirect(redirect *Redirection, port uint32) RouteConfigurer {
	if redirect == nil {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_route.Route) error {
		builder := RouteRedirectBuilder{}
		builder.Configure(
			RouteRedirectScheme(redirect.Scheme),
			RouteRedirectHost(redirect.Host),
			RouteRedirectPort(port, redirect.Port),
		)
		envoyRedirect := &envoy_route.RedirectAction{
			StripQuery:   redirect.StripQuery,
		}

		if redirect.Port > 0 {
			envoyRedirect.PortRedirect = redirect.Port
		}

		if rewrite := redirect.PathRewrite; rewrite != nil {
			if rewrite.ReplaceFullPath != nil {
				regexRewrite := &envoy_route.RedirectAction_RegexRewrite{
					RegexRewrite: &envoy_type_matcher.RegexMatchAndSubstitute{
						Pattern: &envoy_type_matcher.RegexMatcher{
							Regex: `.*`,
						},
						Substitution: *rewrite.ReplaceFullPath,
					},
				}

				envoyRedirect.PathRewriteSpecifier = regexRewrite
			}

			if rewrite.ReplacePrefixMatch != nil {
				prefixRewrite := &envoy_route.RedirectAction_PrefixRewrite{
					PrefixRewrite: *rewrite.ReplacePrefixMatch,
				}

				envoyRedirect.PathRewriteSpecifier = prefixRewrite
			}
		}

		switch redirect.Status {
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
			return errors.Errorf(
				"redirect status code %d is not supported",
				redirect.Status,
			)
		}

		r.Action = &envoy_route.Route_Redirect{
			Redirect: envoyRedirect,
		}

		return nil
	})
}
