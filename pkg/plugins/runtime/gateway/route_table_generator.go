package gateway

import (
	"net/http"
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	policies_defaults "github.com/kumahq/kuma/pkg/plugins/policies/core/defaults"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

const emptyGatewayMsg = "This is a Kuma MeshGateway. No routes match this MeshGateway!\n"

// GenerateVirtualHost generates xDS resources for the current route table.
func GenerateVirtualHost(
	ctx xds_context.Context, info GatewayListenerInfo, hostname string, routes []route.Entry,
) (
	*envoy_virtual_hosts.VirtualHostBuilder, error,
) {
	vh := envoy_virtual_hosts.NewVirtualHostBuilder(info.Proxy.APIVersion, hostname).Configure(
		envoy_virtual_hosts.DomainNames(hostname),
	)

	// Ensure that we get TLS on HTTPS protocol listeners or crossMesh.
	if info.Listener.Protocol == mesh_proto.MeshGateway_Listener_HTTPS || info.Listener.CrossMesh {
		vh.Configure(
			envoy_virtual_hosts.RequireTLS(),
			// Set HSTS header to 1 year.
			envoy_virtual_hosts.SetResponseHeader(
				"Strict-Transport-Security",
				"max-age=31536000; includeSubDomains",
			),
		)
	}

	if len(routes) == 0 {
		routeBuilder := envoy_routes.NewRouteBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
			Configure(
				envoy_routes.RouteMatchPrefixPath("/"),
				envoy_routes.RouteActionDirectResponse(http.StatusNotFound, emptyGatewayMsg),
			)
		vh.Configure(envoy_routes.VirtualHostRoute(routeBuilder))

		return vh, nil
	}

	// TODO(jpeach) match the FaultInjection policy for this virtual host.

	// TODO(jpeach) apply additional virtual host configuration.

	// Sort routing table entries so the most specific match comes first.
	sort.Sort(route.Sorter(routes))

	for _, e := range routes {
		routeBuilder := envoy_routes.NewRouteBuilder(envoy_common.APIV3, e.Name).
			Configure(
				envoy_routes.RouteMatchExactPath(e.Match.ExactPath),
				envoy_routes.RouteMatchPrefixPath(e.Match.PrefixPath),
				envoy_routes.RouteMatchRegexPath(e.Match.RegexPath),
				envoy_routes.RouteMatchExactHeader(":method", e.Match.Method),

				route.RouteActionRedirect(e.Action.Redirect, info.Listener.Port),
				route.RouteActionForward(ctx, info.OutboundEndpoints, info.Proxy.Dataplane.Spec.TagSet(), e.Action.Forward),
				envoy_routes.RouteActionIdleTimeout(policies_defaults.DefaultGatewayStreamIdleTimeout),
			)

		if e.Rewrite != nil {
			routeBuilder.Configure(
				envoy_routes.RouteSetRewriteHostToBackendHostname(e.Rewrite.HostToBackendHostname),
				envoy_routes.RouteReplaceHostHeader(pointer.Deref(e.Rewrite.ReplaceHostname)),
			)
		}

		// Generate a retry policy for this route, if there is one.
		routeBuilder.Configure(
			retryRouteConfigurers(
				route.InferForwardingProtocol(e.Action.Forward),
				match.BestConnectionPolicyForDestination(e.Action.Forward, core_mesh.RetryType),
			)...,
		)

		if t := match.BestConnectionPolicyForDestination(e.Action.Forward, core_mesh.TimeoutType); t != nil {
			timeout := t.(*core_mesh.TimeoutResource)
			routeBuilder.Configure(
				envoy_routes.RouteActionRequestTimeout(timeout.Spec.GetConf().GetHttp().GetRequestTimeout().AsDuration()),
			)
		}

		if r := match.BestConnectionPolicyForDestination(e.Action.Forward, core_mesh.RateLimitType); r != nil {
			ratelimit := r.(*core_mesh.RateLimitResource)
			conf, err := v3.NewRateLimitConfiguration(v3.RateLimitConfigurationFromProto(ratelimit.Spec))
			if err != nil {
				return nil, err
			}

			routeBuilder.Configure(
				envoy_routes.RoutePerFilterConfig("envoy.filters.http.local_ratelimit", conf),
			)
		}

		for _, m := range e.Match.ExactHeader {
			routeBuilder.Configure(envoy_routes.RouteMatchExactHeader(m.Key, m.Value))
		}

		for _, m := range e.Match.RegexHeader {
			routeBuilder.Configure(envoy_routes.RouteMatchRegexHeader(m.Key, m.Value))
		}

		for _, m := range e.Match.PresentHeader {
			routeBuilder.Configure(envoy_routes.RouteMatchPresentHeader(m, true))
		}

		for _, m := range e.Match.AbsentHeader {
			routeBuilder.Configure(envoy_routes.RouteMatchPresentHeader(m, false))
		}

		for _, m := range e.Match.PrefixHeader {
			routeBuilder.Configure(envoy_routes.RouteMatchPrefixHeader(m.Key, m.Value))
		}

		for _, m := range e.Match.ExactQuery {
			routeBuilder.Configure(envoy_routes.RouteMatchExactQuery(m.Key, m.Value))
		}

		for _, m := range e.Match.RegexQuery {
			routeBuilder.Configure(envoy_routes.RouteMatchRegexQuery(m.Key, m.Value))
		}

		if rq := e.RequestHeaders; rq != nil {
			for _, h := range e.RequestHeaders.Replace {
				switch h.Key {
				case ":authority", "Host", "host":
					routeBuilder.Configure(envoy_routes.RouteReplaceHostHeader(h.Value))
				default:
					routeBuilder.Configure(envoy_routes.RouteAddRequestHeader(envoy_routes.RouteReplaceHeader(h.Key, h.Value)))
				}
			}

			for _, h := range e.RequestHeaders.Append {
				routeBuilder.Configure(envoy_routes.RouteAddRequestHeader(envoy_routes.RouteAppendHeader(h.Key, h.Value)))
			}

			for _, name := range e.RequestHeaders.Delete {
				routeBuilder.Configure(envoy_routes.RouteDeleteRequestHeader(name))
			}
		}
		if rq := e.ResponseHeaders; rq != nil {
			for _, h := range e.ResponseHeaders.Replace {
				routeBuilder.Configure(envoy_routes.RouteAddResponseHeader(envoy_routes.RouteReplaceHeader(h.Key, h.Value)))
			}

			for _, h := range e.ResponseHeaders.Append {
				routeBuilder.Configure(envoy_routes.RouteAddResponseHeader(envoy_routes.RouteAppendHeader(h.Key, h.Value)))
			}

			for _, name := range e.ResponseHeaders.Delete {
				routeBuilder.Configure(envoy_routes.RouteDeleteResponseHeader(name))
			}
		}

		// After configuring the route action, attempt to configure mirroring.
		// This only affects the forwarding action.
		if m := e.Mirror; m != nil {
			routeBuilder.Configure(route.RouteMirror(m.Percentage, m.Forward))
		}

		routeBuilder.Configure(route.RouteRewrite(e.Rewrite))

		vh.Configure(envoy_routes.VirtualHostRoute(routeBuilder))
	}

	return vh, nil
}

// retryRouteConfigurers returns the set of route configurers needed to implement the retry policy (if there is one).
func retryRouteConfigurers(protocol core_meta.Protocol, policy model.Resource) []envoy_routes.RouteConfigurer {
	retry, _ := policy.(*core_mesh.RetryResource)
	if retry == nil {
		return nil
	}

	return []envoy_routes.RouteConfigurer{
		envoy_routes.RouteActionRetryDefault(protocol),
		envoy_routes.RouteActionRetry(retry, protocol),
	}
}
