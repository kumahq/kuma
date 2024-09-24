package gateway

import (
	"context"

	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/match"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

// ClusterGenerator generates Envoy clusters and their corresponding
// load assignments for both mesh services and external services.
type ClusterGenerator struct {
	Zone string
}

// GenerateHost generates clusters for all the services targeted in the current route table.
func (c *ClusterGenerator) GenerateClusters(ctx context.Context, xdsCtx xds_context.Context, info GatewayListenerInfo, hostEntries []route.Entry, hostTags map[string]string) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	// If there is a service name conflict between external services
	// and mesh services, the external service takes priority since
	// it's easier to deterministically know it is present.
	//
	// XXX(jpeach) Mesh proxy code does the opposite. It generates
	// an array of endpoint and checks whether the first entry is from
	// an external service. Because the dataplane endpoints happen to be
	// generated first, the mesh service will have priority.
	for _, dest := range RouteDestinationsMutable(hostEntries) {
		var isExternalCluster bool
		var service string
		var r *core_xds.Resource
		var err error

		if dest.BackendRef != nil && dest.BackendRef.ReferencesRealResource() {
			r, service, err = c.generateRealBackendRefCluster(
				xdsCtx.Mesh,
				info.Proxy,
				dest.BackendRef.RealResourceBackendRef(),
				dest.RouteProtocol,
				xdsCtx.ControlPlane.SystemNamespace,
				hostTags,
			)
			if r == nil && err == nil {
				log.Info("skipping backendRef", "backendRef", dest.BackendRef)
				continue
			}
			isExternalService := xdsCtx.Mesh.IsExternalService(service)
			isExternalCluster = isExternalService && !xdsCtx.Mesh.Resource.ZoneEgressEnabled()
		} else {
			service = dest.Destination[mesh_proto.ServiceTag]

			if service == metadata.UnresolvedBackendServiceTag {
				dest.Name = metadata.UnresolvedBackendServiceTag
				continue
			}

			isExternalService := xdsCtx.Mesh.IsExternalService(service)
			if len(xdsCtx.Mesh.Resources.TrafficPermissions().Items) > 0 {
				isExternalService = route.HasExternalServiceEndpoint(xdsCtx.Mesh.Resource, info.OutboundEndpoints, *dest)
			}

			matched := match.ExternalService(info.ExternalServices.Items, mesh_proto.TagSelector(dest.Destination))

			// If there is Mesh property ZoneEgress enabled we want always to
			// direct the traffic through them. The condition is, the mesh must
			// have mTLS enabled and traffic through zoneEgress is enabled.
			isDirectExternalService := isExternalService && !xdsCtx.Mesh.Resource.ZoneEgressEnabled()
			isExternalCluster = isDirectExternalService && len(matched) > 0

			isExternalServiceThroughZoneEgress := isExternalService && xdsCtx.Mesh.Resource.ZoneEgressEnabled()

			if isExternalCluster {
				log.V(1).Info("generating external service cluster",
					"service", service,
				)

				r, err = c.generateExternalCluster(ctx, xdsCtx.Mesh, info, matched, dest, hostTags)
			} else {
				log.V(1).Info("generating mesh cluster resource",
					"service", service,
				)

				upstreamServiceName := dest.Destination[mesh_proto.ServiceTag]
				if isExternalServiceThroughZoneEgress {
					upstreamServiceName = mesh_proto.ZoneEgressServiceName
				}

				r, err = c.generateMeshCluster(xdsCtx.Mesh, info, dest, upstreamServiceName, hostTags)
			}
		}

		if err != nil {
			return nil, err
		}
		resources.Add(r)

		// Assign the generated unique cluster name to the
		// destination so that subsequent generator passes can
		// reference it.
		dest.Name = r.Name

		if isExternalCluster {
			// External clusters don't get a load assignment.
			continue
		}

		// The CLA cache needs an envoy.Cluster but only looks
		// at the fields we populate here.
		cluster := envoy.NewCluster(
			envoy.WithName(dest.Name),
			envoy.WithService(service),
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

func (c *ClusterGenerator) generateRealBackendRefCluster(
	meshCtx xds_context.MeshContext,
	proxy *core_xds.Proxy,
	backendRef *model.RealResourceBackendRef,
	routeProtocol core_mesh.Protocol,
	systemNamespace string,
	identifyingTags map[string]string,
) (*core_xds.Resource, string, error) {
	service, destProtocol, _, ok := meshroute.GetServiceProtocolPortFromRef(meshCtx, backendRef)
	if !ok {
		return nil, "", nil
	}
	protocol := route.InferServiceProtocol(destProtocol, routeProtocol)

	edsClusterBuilder := clusters.NewClusterBuilder(proxy.APIVersion, service).
		Configure(
			clusters.EdsCluster(),
			clusters.LB(nil /* TODO(jpeach) uses default Round Robin*/),
			clusters.ClientSideMultiIdentitiesMTLS(
				proxy.SecretsTracker,
				meshCtx.Resource,
				true, // TODO we just assume this atm?...
				meshroute.SniForBackendRef(backendRef, meshCtx, systemNamespace),
				meshroute.ServiceTagIdentities(backendRef, meshCtx),
			),
			clusters.ConnectionBufferLimit(DefaultConnectionBuffer),
		)

	switch protocol {
	case core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		edsClusterBuilder.Configure(clusters.Http2FromEdge())
	case core_mesh.ProtocolHTTP:
		edsClusterBuilder.Configure(clusters.Http())
	default:
	}
	// Note: these tags are just used to get a cluster name!
	tags := map[string]string{
		mesh_proto.ServiceTag: service,
	}
	cluster, err := buildClusterResource(tags, edsClusterBuilder, identifyingTags)
	cluster.ResourceOrigin = backendRef.Resource
	return cluster, service, err
}

func (c *ClusterGenerator) generateMeshCluster(
	meshCtx xds_context.MeshContext,
	info GatewayListenerInfo,
	dest *route.Destination,
	upstreamServiceName string,
	identifyingTags map[string]string,
) (*core_xds.Resource, error) {
	destProtocol := core_mesh.ParseProtocol(dest.Destination[mesh_proto.ProtocolTag])
	protocol := route.InferServiceProtocol(destProtocol, dest.RouteProtocol)

	builder := newClusterBuilder(info.Proxy.APIVersion, dest.Destination[mesh_proto.ServiceTag], protocol, dest).Configure(
		clusters.EdsCluster(),
		clusters.LB(nil /* TODO(jpeach) uses default Round Robin*/),
		clusters.ClientSideMTLS(info.Proxy.SecretsTracker, meshCtx.Resource, upstreamServiceName, true, []tags.Tags{dest.Destination}),
		clusters.ConnectionBufferLimit(DefaultConnectionBuffer),
	)

	// TODO(jpeach) Envoy configures retries and fault injection with
	// virtualhost filters, but Kuma models these as connection policies.
	// Source+Destination matching implies that we would need to know the
	// the destination cluster before deciding whether to enable the filter.
	// It's not clear whether that can be done.

	return buildClusterResource(dest.Destination, builder, identifyingTags)
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
	serviceName := dest.Destination[mesh_proto.ServiceTag]
	protocol := route.InferServiceProtocol(meshCtx.GetServiceProtocol(serviceName), dest.RouteProtocol)

	return buildClusterResource(
		dest.Destination,
		newClusterBuilder(info.Proxy.APIVersion, serviceName, protocol, dest).Configure(
			clusters.ProvidedEndpointCluster(info.Proxy.Dataplane.IsIPv6(), endpoints...),
			clusters.ClientSideTLS(endpoints),
			clusters.ConnectionBufferLimit(DefaultConnectionBuffer),
		),
		identifyingTags,
	)
}

func newClusterBuilder(
	version core_xds.APIVersion,
	name string,
	protocol core_mesh.Protocol,
	dest *route.Destination,
) *clusters.ClusterBuilder {
	var timeout *mesh_proto.Timeout_Conf
	if timeoutResource := timeoutPolicyFor(dest); timeoutResource != nil {
		timeout = timeoutResource.Spec.GetConf()
	}

	builder := clusters.NewClusterBuilder(version, name).Configure(
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
	destinationTags tags.Tags,
	c *clusters.ClusterBuilder,
	identifyingTags map[string]string,
) (*core_xds.Resource, error) {
	msg, err := c.Build()
	if err != nil {
		return nil, err
	}

	cluster := msg.(*envoy_cluster_v3.Cluster)

	name, err := destinationTags.DestinationClusterName(identifyingTags)
	if err != nil {
		return nil, err
	}

	cluster.Name = name

	return &core_xds.Resource{
		Name:     cluster.GetName(),
		Origin:   metadata.OriginGateway,
		Resource: cluster,
	}, nil
}

func routeDestinations(entries []route.Entry) []route.Destination {
	var destinations []route.Destination

	for _, dest := range RouteDestinationsMutable(entries) {
		destinations = append(destinations, *dest)
	}

	return destinations
}

func RouteDestinationsMutable(entries []route.Entry) []*route.Destination {
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
