package gateway

import (
	"net/http"
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

const emptyGatewayMsg = "This is a Kuma MeshGateway. No routes match this MeshGateway!\n"

// GenerateVirtualHost generates xDS resources for the current route table.
func GenerateVirtualHost(
	ctx xds_context.Context, info GatewayListenerInfo, host GatewayHost, routes []route.Entry,
) (
	*envoy_virtual_hosts.VirtualHostBuilder, error,
) {
	vh := envoy_virtual_hosts.NewVirtualHostBuilder(info.Proxy.APIVersion, host.Hostname).Configure(
		envoy_virtual_hosts.DomainNames(host.Hostname),
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
		routeBuilder := route.RouteBuilder{}

		routeBuilder.Configure(route.RouteMatchPrefixPath("/"))
		routeBuilder.Configure(route.RouteActionDirectResponse(http.StatusNotFound, emptyGatewayMsg))
		vh.Configure(route.VirtualHostRoute(&routeBuilder))

		return vh, nil
	}

	// TODO(jpeach) match the FaultInjection policy for this virtual host.

	// TODO(jpeach) apply additional virtual host configuration.

	// Sort routing table entries so the most specific match comes first.
	sort.Sort(route.Sorter(routes))

	for _, e := range routes {
		routeBuilder := route.RouteBuilder{}
		routeBuilder.Configure(
			route.RouteMatchExactPath(e.Match.ExactPath),
			route.RouteMatchPrefixPath(e.Match.PrefixPath),
			route.RouteMatchRegexPath(e.Match.RegexPath),
			route.RouteMatchExactHeader(":method", e.Match.Method),

			route.RouteActionRedirect(e.Action.Redirect, info.Listener.Port),
			route.RouteActionForward(ctx.Mesh.Resource, info.OutboundEndpoints, info.Proxy.Dataplane.Spec.TagSet(), e.Action.Forward),
			route.RouteSetRewriteHostToBackendHostname(e.Rewrite != nil && e.Rewrite.HostToBackendHostname),
			route.RouteActionIdleTimeout(DefaultStreamIdleTimeout),
		)

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
				route.RouteActionRequestTimeout(timeout.Spec.GetConf().GetHttp().GetRequestTimeout().AsDuration()),
			)
		}

		if r := match.BestConnectionPolicyForDestination(e.Action.Forward, core_mesh.RateLimitType); r != nil {
			ratelimit := r.(*core_mesh.RateLimitResource)
			conf, err := v3.NewRateLimitConfiguration(v3.RateLimitConfigurationFromProto(ratelimit.Spec))
			if err != nil {
				return nil, err
			}

			routeBuilder.Configure(
				route.RoutePerFilterConfig("envoy.filters.http.local_ratelimit", conf),
			)
		}

		for _, m := range e.Match.ExactHeader {
			routeBuilder.Configure(route.RouteMatchExactHeader(m.Key, m.Value))
		}

		for _, m := range e.Match.RegexHeader {
			routeBuilder.Configure(route.RouteMatchRegexHeader(m.Key, m.Value))
		}

		for _, m := range e.Match.PresentHeader {
			routeBuilder.Configure(route.RouteMatchPresentHeader(m, true))
		}

		for _, m := range e.Match.AbsentHeader {
			routeBuilder.Configure(route.RouteMatchPresentHeader(m, false))
		}

		for _, m := range e.Match.ExactQuery {
			routeBuilder.Configure(route.RouteMatchExactQuery(m.Key, m.Value))
		}

		for _, m := range e.Match.RegexQuery {
			routeBuilder.Configure(route.RouteMatchRegexQuery(m.Key, m.Value))
		}

		if rq := e.RequestHeaders; rq != nil {
			for _, h := range e.RequestHeaders.Replace {
				switch h.Key {
				case ":authority", "Host", "host":
					routeBuilder.Configure(route.RouteReplaceHostHeader(h.Value))
				default:
					routeBuilder.Configure(route.RouteAddRequestHeader(route.RouteReplaceHeader(h.Key, h.Value)))
				}
			}

			for _, h := range e.RequestHeaders.Append {
				routeBuilder.Configure(route.RouteAddRequestHeader(route.RouteAppendHeader(h.Key, h.Value)))
			}

			for _, name := range e.RequestHeaders.Delete {
				routeBuilder.Configure(route.RouteDeleteRequestHeader(name))
			}
		}
		if rq := e.ResponseHeaders; rq != nil {
			for _, h := range e.ResponseHeaders.Replace {
				routeBuilder.Configure(route.RouteAddResponseHeader(route.RouteReplaceHeader(h.Key, h.Value)))
			}

			for _, h := range e.ResponseHeaders.Append {
				routeBuilder.Configure(route.RouteAddResponseHeader(route.RouteAppendHeader(h.Key, h.Value)))
			}

			for _, name := range e.ResponseHeaders.Delete {
				routeBuilder.Configure(route.RouteDeleteResponseHeader(name))
			}
		}

		// After configuring the route action, attempt to configure mirroring.
		// This only affects the forwarding action.
		if m := e.Mirror; m != nil {
			routeBuilder.Configure(route.RouteMirror(m.Percentage, m.Forward))
		}

		routeBuilder.Configure(route.RouteRewrite(e.Rewrite))

		vh.Configure(route.VirtualHostRoute(&routeBuilder))
	}

	return vh, nil
}

// retryRouteConfigurers returns the set of route configurers needed to implement the retry policy (if there is one).
func retryRouteConfigurers(protocol core_mesh.Protocol, policy model.Resource) []route.RouteConfigurer {
	retry, _ := policy.(*core_mesh.RetryResource)
	if retry == nil {
		return nil
	}

	return []route.RouteConfigurer{
		route.RouteActionRetryDefault(protocol),
		route.RouteActionRetry(retry, protocol),
	}
}
