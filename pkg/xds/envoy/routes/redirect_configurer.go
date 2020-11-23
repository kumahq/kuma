package routes

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
)

// Redirect for paths that match to matchPath returns 301 status code with new port and path
func Redirect(matchPath, newPath string, allowGetOnly bool, port uint32) VirtualHostBuilderOpt {
	return VirtualHostBuilderOptFunc(func(config *VirtualHostBuilderConfig) {
		config.Add(&RedirectConfigurer{
			matchPath:    matchPath,
			newPath:      newPath,
			port:         port,
			allowGetOnly: allowGetOnly,
		})
	})
}

type RedirectConfigurer struct {
	matchPath    string
	newPath      string
	port         uint32
	allowGetOnly bool
}

func (c RedirectConfigurer) Configure(virtualHost *envoy_route.VirtualHost) error {
	var headersMatcher []*envoy_route.HeaderMatcher
	if c.allowGetOnly {
		headersMatcher = []*envoy_route.HeaderMatcher{
			{
				Name: ":method",
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_ExactMatch{
					ExactMatch: "GET",
				},
			},
		}
	}
	virtualHost.Routes = append(virtualHost.Routes, &envoy_route.Route{
		Match: &envoy_route.RouteMatch{
			PathSpecifier: &envoy_route.RouteMatch_Path{
				Path: c.matchPath,
			},
			Headers: headersMatcher,
		},
		Action: &envoy_route.Route_Redirect{
			Redirect: &envoy_route.RedirectAction{
				PortRedirect: c.port,
				PathRewriteSpecifier: &envoy_route.RedirectAction_PathRedirect{
					PathRedirect: c.newPath,
				},
			},
		},
	})
	return nil
}
