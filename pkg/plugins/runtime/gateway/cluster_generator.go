package gateway

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

// ClusterGenerator generates Envoy clusters and their corresponding
// load assignments for both mesh services and external services.
type ClusterGenerator struct {
	DataSourceLoader datasource.Loader
	Zone             string
}

// SupportsProtocol is always true for generating clusters.
func (*ClusterGenerator) SupportsProtocol(mesh_proto.Gateway_Listener_Protocol) bool {
	return true
}

// GenerateHost generates clusters for all the services targeted in the current route table.
func (c *ClusterGenerator) GenerateHost(ctx xds_context.Context, info *GatewayResourceInfo) (*core_xds.ResourceSet, error) {
	resources := ResourceAggregator{}

	destinations, err := makeRouteDestinations(&info.RouteTable)
	if err != nil {
		return nil, err
	}

	// If there is a service name conflict between external services
	// and mesh services, the external service takes priority since
	// it's easier to deterministically know it is present.
	//
	// XXX(jpeach) Mesh proxy code does the opposite. It generates
	// an array of endpoint and checks whether the first entry is from
	// an external service. Because the dataplane endpoints happen to be
	// generated first, the mesh service will have priority.
	for name, dest := range destinations {
		matched := match.ExternalService(info.ExternalServices, mesh_proto.TagSelector(dest.Destination))
		if len(matched.Items) > 0 {
			if err := resources.Add(c.generateExternalCluster(ctx, info, matched, name, dest)); err != nil {
				return nil, err
			}

			// External clusters don't get a load assignment.
			continue
		}

		if err := resources.Add(c.generateMeshCluster(ctx, info, name, dest)); err != nil {
			return nil, err
		}

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

		resources.Get().Add(NewResource(name, loadAssignment))
	}

	return resources.Get(), nil
}

func (c *ClusterGenerator) generateMeshCluster(
	ctx xds_context.Context,
	info *GatewayResourceInfo,
	name string,
	dest route.Destination,
) (*core_xds.ResourceSet, error) {
	log.Info("generating mesh cluster resource",
		"name", name,
		"service", dest.Destination[mesh_proto.ServiceTag],
	)

	protocol := generator.InferServiceProtocol([]core_xds.Endpoint{{
		Tags: dest.Destination,
	}})

	// HTTP is a better default than "unknown".
	if protocol == core_mesh.ProtocolUnknown {
		protocol = core_mesh.ProtocolHTTP
	}

	builder := newClusterBuilder(info.Proxy.APIVersion, protocol, dest).Configure(
		clusters.EdsCluster(name),
		clusters.LB(nil /* TODO(jpeach) uses default Round Robin*/),
		clusters.ClientSideMTLS(ctx, dest.Destination[mesh_proto.ServiceTag], []envoy.Tags{dest.Destination}),
	)

	// TODO(jpeach) Envoy configures retries and fault injection with
	// virtualhost filters, but Kuma models these as connection policies.
	// Source+Destination matching implies that we would need to know the
	// the destination cluster before deciding whether to enable the filter.
	// It's not clear whether that can be done.

	return BuildResourceSet(builder)
}

func (c *ClusterGenerator) generateExternalCluster(
	ctx xds_context.Context,
	info *GatewayResourceInfo,
	service core_mesh.ExternalServiceResourceList,
	name string,
	dest route.Destination,
) (*core_xds.ResourceSet, error) {
	log.Info("generating external service cluster",
		"name", name,
		"service", dest.Destination[mesh_proto.ServiceTag],
	)

	var endpoints []core_xds.Endpoint

	for _, ext := range service.Items {
		ep, err := topology.NewExternalServiceEndpoint(ext, ctx.Mesh.Resource, c.DataSourceLoader, c.Zone)
		if err != nil {
			return nil, err
		}

		endpoints = append(endpoints, *ep)
	}

	protocol := generator.InferServiceProtocol(endpoints)

	// HTTP is a better default than "unknown".
	if protocol == core_mesh.ProtocolUnknown {
		protocol = core_mesh.ProtocolHTTP
	}

	return BuildResourceSet(
		newClusterBuilder(info.Proxy.APIVersion, protocol, dest).Configure(
			clusters.StrictDNSCluster(name, endpoints, info.Dataplane.IsIPv6()),
			clusters.ClientSideTLS(endpoints),
		),
	)
}

func newClusterBuilder(
	version envoy.APIVersion,
	protocol core_mesh.Protocol,
	dest route.Destination,
) *clusters.ClusterBuilder {
	builder := clusters.NewClusterBuilder(version).Configure(
		clusters.Timeout(protocol, timeoutPolicyFor(&dest)),
		clusters.CircuitBreaker(circuitBreakerPolicyFor(&dest)),
		clusters.OutlierDetection(circuitBreakerPolicyFor(&dest)),
		clusters.HealthCheck(protocol, healthCheckPolicyFor(&dest)),
	)

	// TODO(jpeach) OutboundProxyGenerator unconditionally
	// gives mesh services a HTTP/2 transport. We ought to do
	// the same, but it doesn't work.
	switch protocol {
	case core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		builder.Configure(clusters.Http2())
	case core_mesh.ProtocolHTTP:
		builder.Configure(clusters.Http())
	default:
	}

	return builder
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

func timeoutPolicyFor(dest *route.Destination) *core_mesh.TimeoutResource {
	if policy, ok := dest.Policies[core_mesh.TimeoutType]; ok {
		return policy.(*core_mesh.TimeoutResource)
	}

	return nil // TODO(jpeach) default timeout policy
}

func circuitBreakerPolicyFor(dest *route.Destination) *core_mesh.CircuitBreakerResource {
	if policy, ok := dest.Policies[core_mesh.CircuitBreakerType]; ok {
		return policy.(*core_mesh.CircuitBreakerResource)
	}

	return nil // TODO(jpeach) default breaker policy
}

func healthCheckPolicyFor(dest *route.Destination) *core_mesh.HealthCheckResource {
	if policy, ok := dest.Policies[core_mesh.HealthCheckType]; ok {
		return policy.(*core_mesh.HealthCheckResource)
	}

	return nil // TODO(jpeach) default breaker policy
}
