package gateway

import (
	"context"

	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

// ClusterGenerator generates Envoy clusters and their corresponding
// load assignments for both mesh services and external services.
type ClusterGenerator struct {
	Zone string
}

// GenerateHost generates clusters for all the services targeted in the current route table.
func (c *ClusterGenerator) GenerateClusters(ctx context.Context, xdsCtx xds_context.Context, info GatewayListenerInfo, hostInfo GatewayHostInfo) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	// If there is a service name conflict between external services
	// and mesh services, the external service takes priority since
	// it's easier to deterministically know it is present.
	//
	// XXX(jpeach) Mesh proxy code does the opposite. It generates
	// an array of endpoint and checks whether the first entry is from
	// an external service. Because the dataplane endpoints happen to be
	// generated first, the mesh service will have priority.
	for _, dest := range routeDestinationsMutable(hostInfo.Entries) {
		matched := match.ExternalService(info.ExternalServices.Items, mesh_proto.TagSelector(dest.Destination))
		service := dest.Destination[mesh_proto.ServiceTag]

		firstEndpointExternalService := route.HasExternalServiceEndpoint(xdsCtx.Mesh.Resource, info.OutboundEndpoints, *dest)

		// If there is Mesh property ZoneEgress enabled we want always to
		// direct the traffic through them. The condition is, the mesh must
		// have mTLS enabled and traffic through zoneEgress is enabled.
		isDirectExternalService := firstEndpointExternalService && !xdsCtx.Mesh.Resource.ZoneEgressEnabled()
		isExternalServiceThroughZoneEgress := firstEndpointExternalService && !isDirectExternalService

		var r *core_xds.Resource
		var err error

		if isDirectExternalService {
			log.V(1).Info("generating external service cluster",
				"service", service,
			)

			r, err = c.generateExternalCluster(ctx, xdsCtx.Mesh, info, matched, dest, hostInfo.Host.Tags)
		} else {
			log.Info("generating mesh cluster resource",
				"service", service,
			)

			upstreamServiceName := dest.Destination[mesh_proto.ServiceTag]
			if isExternalServiceThroughZoneEgress {
				upstreamServiceName = mesh_proto.ZoneEgressServiceName
			}

			r, err = c.generateMeshCluster(xdsCtx.Mesh.Resource, info, dest, upstreamServiceName, hostInfo.Host.Tags)
		}

		if err != nil {
			return nil, err
		}
		resources.Add(r)

		// Assign the generated unique cluster name to the
		// destination so that subsequent generator passes can
		// reference it.
		dest.Name = r.Name

		if isDirectExternalService {
			// External clusters don't get a load assignment.
			continue
		}

		// The CLA cache needs an envoy.Cluster but only looks
		// at the fields we populate here.
		cluster := envoy.NewCluster(
			envoy.WithName(dest.Name),
			envoy.WithService(dest.Destination[mesh_proto.ServiceTag]),
			envoy.WithTags(dest.Destination),
		)

		loadAssignment, err := xdsCtx.ControlPlane.CLACache.GetCLA(
			ctx,
			xdsCtx.Mesh.Resource.GetMeta().GetName(),
			xdsCtx.Mesh.Hash,
			cluster,
			info.Proxy.APIVersion,
			xdsCtx.Mesh.EndpointMap,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to build LoadAssignment for cluster %q", dest.Name)
		}

		resources.Add(NewResource(dest.Name, loadAssignment))
	}

	return resources, nil
}

func (c *ClusterGenerator) generateMeshCluster(
	mesh *core_mesh.MeshResource,
	info GatewayListenerInfo,
	dest *route.Destination,
	upstreamServiceName string,
	identifyingTags map[string]string,
) (*core_xds.Resource, error) {
	protocol := route.InferServiceProtocol([]core_xds.Endpoint{{
		Tags: dest.Destination,
	}}, dest.RouteProtocol)

	builder := newClusterBuilder(info.Proxy.APIVersion, protocol, dest).Configure(
		clusters.EdsCluster(dest.Destination[mesh_proto.ServiceTag]),
		clusters.LB(nil /* TODO(jpeach) uses default Round Robin*/),
		clusters.ClientSideMTLS(info.Proxy.SecretsTracker, mesh, upstreamServiceName, true, []envoy.Tags{dest.Destination}),
		clusters.ConnectionBufferLimit(DefaultConnectionBuffer),
	)

	// TODO(jpeach) Envoy configures retries and fault injection with
	// virtualhost filters, but Kuma models these as connection policies.
	// Source+Destination matching implies that we would need to know the
	// the destination cluster before deciding whether to enable the filter.
	// It's not clear whether that can be done.

	return buildClusterResource(dest, builder, identifyingTags)
}

func (c *ClusterGenerator) generateExternalCluster(
	ctx context.Context,
	meshCtx xds_context.MeshContext,
	info GatewayListenerInfo,
	externalServices []*core_mesh.ExternalServiceResource,
	dest *route.Destination,
	identifyingTags map[string]string,
) (*core_xds.Resource, error) {
	var endpoints []core_xds.Endpoint

	for _, ext := range externalServices {
		ep, err := topology.NewExternalServiceEndpoint(ctx, ext, meshCtx.Resource, meshCtx.DataSourceLoader, c.Zone)
		if err != nil {
			return nil, err
		}

		endpoints = append(endpoints, *ep)
	}

	protocol := route.InferServiceProtocol(endpoints, dest.RouteProtocol)

	return buildClusterResource(
		dest,
		newClusterBuilder(info.Proxy.APIVersion, protocol, dest).Configure(
			clusters.ProvidedEndpointCluster(dest.Destination[mesh_proto.ServiceTag], info.Proxy.Dataplane.IsIPv6(), endpoints...),
			clusters.ClientSideTLS(endpoints),
			clusters.ConnectionBufferLimit(DefaultConnectionBuffer),
		),
		identifyingTags,
	)
}

func newClusterBuilder(
	version envoy.APIVersion,
	protocol core_mesh.Protocol,
	dest *route.Destination,
) *clusters.ClusterBuilder {
	var timeout *mesh_proto.Timeout_Conf
	if timeoutResource := timeoutPolicyFor(dest); timeoutResource != nil {
		timeout = timeoutResource.Spec.GetConf()
	}

	builder := clusters.NewClusterBuilder(version).Configure(
		clusters.Timeout(timeout, protocol),
		clusters.CircuitBreaker(circuitBreakerPolicyFor(dest)),
		clusters.OutlierDetection(circuitBreakerPolicyFor(dest)),
		clusters.HealthCheck(protocol, healthCheckPolicyFor(dest)),
	)

	// TODO(jpeach) OutboundProxyGenerator unconditionally
	// gives mesh services a HTTP/2 transport. We ought to do
	// the same, but it doesn't work.
	switch protocol {
	case core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		builder.Configure(clusters.Http2FromEdge())
	case core_mesh.ProtocolHTTP:
		builder.Configure(clusters.Http())
	default:
	}

	return builder
}

// buildClusterResource closes the cluster builder and generates a fully
// qualified name for the cluster. Because the connection policies applied
// to a cluster can be different depending on the listener and the hostname,
// we can't just build a cluster using the service name and tags, we have to
// take listener tags into account.
func buildClusterResource(
	dest *route.Destination,
	c *clusters.ClusterBuilder,
	identifyingTags map[string]string,
) (*core_xds.Resource, error) {
	msg, err := c.Build()
	if err != nil {
		return nil, err
	}

	cluster := msg.(*envoy_cluster_v3.Cluster)

	name, err := route.DestinationClusterName(
		dest,
		identifyingTags,
	)
	if err != nil {
		return nil, err
	}

	cluster.Name = name

	return &core_xds.Resource{
		Name:     cluster.GetName(),
		Origin:   OriginGateway,
		Resource: cluster,
	}, nil
}

func routeDestinations(entries []route.Entry) []route.Destination {
	var destinations []route.Destination

	for _, dest := range routeDestinationsMutable(entries) {
		destinations = append(destinations, *dest)
	}

	return destinations
}

func routeDestinationsMutable(entries []route.Entry) []*route.Destination {
	var destinations []*route.Destination

	for _, e := range entries {
		if m := e.Mirror; m != nil {
			destinations = append(destinations, &m.Forward)
		}

		for i := range e.Action.Forward {
			destinations = append(destinations, &e.Action.Forward[i])
		}
	}

	return destinations
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

	return nil // TODO(jpeach) default circuit breaker policy
}

func healthCheckPolicyFor(dest *route.Destination) *core_mesh.HealthCheckResource {
	if policy, ok := dest.Policies[core_mesh.HealthCheckType]; ok {
		return policy.(*core_mesh.HealthCheckResource)
	}

	return nil // TODO(jpeach) default health check policy
}
