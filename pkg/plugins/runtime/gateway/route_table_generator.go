package gateway

import (
	"context"
	"sort"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

// RouteTableGenerator generates Envoy xDS resources gateway routes from
// the current route table.
type RouteTableGenerator struct{}

// SupportsProtocol is always true for RouteTableGenerator.
func (*RouteTableGenerator) SupportsProtocol(mesh_proto.Gateway_Listener_Protocol) bool {
	return true
}

// GenerateHost generates xDS resources for the current route table.
func (r *RouteTableGenerator) GenerateHost(ctx xds_context.Context, info *GatewayResourceInfo) (*core_xds.ResourceSet, error) {
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

	destinations, err := makeRouteDestinations(&info.RouteTable)
	if err != nil {
		return nil, err
	}

	if err := resources.Add(r.generateClusters(ctx, info, destinations)); err != nil {
		return nil, err
	}

	if err := resources.Add(r.generateEndpointAssignments(ctx, info, destinations)); err != nil {
		return nil, err
	}

	// TODO(jpeach) Once we drop TrafficRoute, generate resources
	// for info.Resources.VirtualHost here instead of in generator.go

	return resources.Get(), nil
}

func (r *RouteTableGenerator) generateEndpointAssignments(
	ctx xds_context.Context,
	info *GatewayResourceInfo,
	destinations map[string]route.Destination,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	// TODO(jpeach) Don't generate load assignments for external services.

	for name, dest := range destinations {
		// The CLA cache needs an envoy.Cluster but only looks
		// at the fields we populate here.
		cluster := envoy.NewCluster(
			envoy.WithName(name),
			envoy.WithService(dest.Destination[mesh_proto.ServiceTag]),
			envoy.WithTags(dest.Destination),
		)

		loadAssignment, err := ctx.ControlPlane.CLACache.GetCLA(
			context.Background(),
			ctx.Mesh.Resource.GetMeta().GetName(),
			ctx.Mesh.Hash,
			cluster,
			info.Proxy.APIVersion,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to build LoadAssignment for cluster %q", name)
		}

		resources.Add(NewResource(name, loadAssignment))
	}

	return resources, nil
}

func (r *RouteTableGenerator) generateClusters(
	ctx xds_context.Context,
	info *GatewayResourceInfo,
	destinations map[string]route.Destination,
) (*core_xds.ResourceSet, error) {
	resources := ResourceAggregator{}

	for name, dest := range destinations {
		protocol := generator.InferServiceProtocol([]core_xds.Endpoint{{
			Tags: dest.Destination,
		}})

		builder := clusters.NewClusterBuilder(info.Proxy.APIVersion).Configure(
			clusters.EdsCluster(name),
			clusters.Timeout(protocol, nil /* TODO(jpeach) default timeout */),
			clusters.CircuitBreaker(nil /* TODO(jpeach) uses default */),
			clusters.OutlierDetection(nil /* TODO(jpeach) uses default */),
			clusters.HealthCheck(protocol, nil /* TODO(jpeach) uses default */),
			clusters.LB(nil /* TODO(jpeach) uses default Round Robin*/),
		)

		// Assuming this is an intra-Mesh service, enable mTLS.
		builder.Configure(
			clusters.ClientSideMTLS(ctx, dest.Destination[mesh_proto.ServiceTag], []envoy.Tags{dest.Destination}),
		)

		// TODO(jpeach) External services are "strict DNS" clusters and don't use mTLS.

		switch protocol {
		case mesh.ProtocolHTTP:
			builder.Configure(clusters.Http())
		case mesh.ProtocolHTTP2, mesh.ProtocolGRPC:
			builder.Configure(clusters.Http2())
		}

		if err := resources.Add(BuildResourceSet(builder)); err != nil {
			return nil, err
		}
	}

	return resources.Get(), nil
}

// makeRouteDestinations builds a map of all the destinations in the
// route table, indexed by cluster name. This de-duplicates the destinations
// by name and ensures we only have to generate the name once.
func makeRouteDestinations(table *route.Table) (map[string]route.Destination, error) {
	destinations := map[string]route.Destination{}

	for _, e := range table.Entries {
		if m := e.Mirror; m != nil {
			name, err := route.DestinationClusterName(m.Forward)
			if err != nil {
				return nil, err
			}

			destinations[name] = m.Forward
		}

		for _, d := range e.Action.Forward {
			name, err := route.DestinationClusterName(d)
			if err != nil {
				return nil, err
			}

			destinations[name] = d
		}
	}

	return destinations, nil
}
