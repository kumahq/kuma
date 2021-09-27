package gateway

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

// RouteTableGenerator generates Envoy xDS resources gateway routes from
// the current route table.
type RouteTableGenerator struct{}

// SupportsProtocol is always true for RouteTableGenerator.
func (*RouteTableGenerator) SupportsProtocol(mesh_proto.Gateway_Listener_Protocol) bool {
	return true
}

// GenerateHost generates xDS resources for the current route table.
func (*RouteTableGenerator) GenerateHost(ctx xds_context.Context, info *GatewayResourceInfo) (*core_xds.ResourceSet, error) {
	resources := ResourceAggregator{}

	// Sort routing table entries so the most specific match comes first.
	sort.Sort(route.Sorter(info.RouteTable.Entries))

	for _, e := range info.RouteTable.Entries {
		routeBuilder := route.RouteBuilder{}

		routeBuilder.Configure(
			route.RouteMatchExactPath(e.Match.ExactPath),
			route.RouteMatchPrefixPath(e.Match.PrefixPath),
			route.RouteMatchRegexPath(e.Match.RegexPath),
			route.RouteMatchExactHeader(":method", e.Match.Method),

			route.RouteActionRedirect(e.Action.Redirect),
			route.RouteActionForward(e.Action.Forward),
		)

		for _, m := range e.Match.ExactHeader {
			routeBuilder.Configure(route.RouteMatchExactHeader(m.Key, m.Value))
		}

		for _, m := range e.Match.RegexHeader {
			routeBuilder.Configure(route.RouteMatchRegexHeader(m.Key, m.Value))
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
					routeBuilder.Configure(route.RouteReplaceRequestHeader(h.Key, h.Value))
				}
			}

			for _, h := range e.RequestHeaders.Append {
				routeBuilder.Configure(route.RouteAppendRequestHeader(h.Key, h.Value))
			}

			for _, name := range e.RequestHeaders.Delete {
				routeBuilder.Configure(route.RouteDeleteRequestHeader(name))
			}
		}

		// After configuring the route action, attempt to configure mirroring.
		// This only affects the forwarding action.
		if m := e.Mirror; m != nil {
			routeBuilder.Configure(route.RouteMirror(m.Percentage, m.Forward))
		}

		info.Resources.VirtualHost.Configure(route.VirtualHostRoute(&routeBuilder))
	}

	// TODO(jpeach) walk the routing table and generate clusters
	// (including the mirror cluster). Hopefully we can use the
	// envoy.Cluster type for this.

	// TODO(jpeach) Once we drop TrafficRoute, generate resources
	// for info.Resources.VirtualHost here instead of in generator.go

	return resources.Get(), nil
}
