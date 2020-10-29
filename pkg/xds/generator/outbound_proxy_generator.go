package generator

import (
	"context"
	"strings"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/validators"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"

	kuma_mesh "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

// OriginOutbound is a marker to indicate by which ProxyGenerator resources were generated.
const OriginOutbound = "outbound"

type OutboundProxyGenerator struct {
}

func (g OutboundProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	outbounds := proxy.Dataplane.Spec.Networking.GetOutbound()
	resources := model.NewResourceSet()
	if len(outbounds) == 0 {
		return resources, nil
	}
	clusters := envoy_common.Clusters{}

	for _, outbound := range outbounds {
		// Determine the list of destination subsets
		// For one outbound listener it may contain many subsets (ex. TrafficRoute to many destinations)
		subsets, err := g.determineSubsets(proxy, outbound)
		if err != nil {
			return nil, err
		}
		clusters.Add(subsets...)

		protocol := g.inferProtocol(proxy, subsets)

		// Generate listener
		listener, err := g.generateLDS(proxy, subsets, outbound, protocol)
		if err != nil {
			return nil, err
		}
		resources.Add(&model.Resource{
			Name:     listener.Name,
			Origin:   OriginOutbound,
			Resource: listener,
		})
	}

	// Generate clusters. It cannot be generated on the fly with outbound loop because we need to know all subsets of the cluster for every service.
	cdsResources, err := g.generateCDS(ctx, proxy, clusters)
	if err != nil {
		return nil, err
	}
	resources.AddSet(cdsResources)

	edsResources, err := g.generateEDS(ctx, proxy, clusters)
	if err != nil {
		return nil, err
	}
	resources.AddSet(edsResources)

	return resources, nil
}

func (_ OutboundProxyGenerator) generateLDS(proxy *model.Proxy, subsets []envoy_common.ClusterSubset, outbound *kuma_mesh.Dataplane_Networking_Outbound, protocol mesh_core.Protocol) (*envoy_api_v2.Listener, error) {
	oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
	meshName := proxy.Dataplane.Meta.GetMesh()
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	serviceName := outbound.GetTagsIncludingLegacy()[kuma_mesh.ServiceTag]
	outboundListenerName := envoy_names.GetOutboundListenerName(oface.DataplaneIP, oface.DataplanePort)
	filterChainBuilder := func() *envoy_listeners.FilterChainBuilder {
		filterChainBuilder := envoy_listeners.NewFilterChainBuilder()
		switch protocol {
		case mesh_core.ProtocolGRPC:
			filterChainBuilder.
				Configure(envoy_listeners.HttpConnectionManager(serviceName)).
				Configure(envoy_listeners.Tracing(proxy.TracingBackend)).
				Configure(envoy_listeners.HttpAccessLog(meshName, envoy_listeners.TrafficDirectionOutbound, sourceService, serviceName, proxy.Logs[serviceName], proxy)).
				Configure(envoy_listeners.HttpOutboundRoute(serviceName, subsets, proxy.Dataplane.Spec.TagSet())).
				Configure(envoy_listeners.GrpcStats())
		case mesh_core.ProtocolHTTP, mesh_core.ProtocolHTTP2:
			filterChainBuilder.
				Configure(envoy_listeners.HttpConnectionManager(serviceName)).
				Configure(envoy_listeners.Tracing(proxy.TracingBackend)).
				Configure(envoy_listeners.HttpAccessLog(meshName, envoy_listeners.TrafficDirectionOutbound, sourceService, serviceName, proxy.Logs[serviceName], proxy)).
				Configure(envoy_listeners.HttpOutboundRoute(serviceName, subsets, proxy.Dataplane.Spec.TagSet()))
		case mesh_core.ProtocolTCP:
			fallthrough
		default:
			// configuration for non-HTTP cases
			filterChainBuilder.
				Configure(envoy_listeners.TcpProxy(serviceName, subsets...)).
				Configure(envoy_listeners.NetworkAccessLog(meshName, envoy_listeners.TrafficDirectionOutbound, sourceService, serviceName, proxy.Logs[serviceName], proxy))
		}
		return filterChainBuilder
	}()
	listener, err := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.OutboundListener(outboundListenerName, oface.DataplaneIP, oface.DataplanePort)).
		Configure(envoy_listeners.FilterChain(filterChainBuilder)).
		Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener %s for service %s", outboundListenerName, serviceName)
	}
	return listener, nil
}

func (o OutboundProxyGenerator) generateCDS(ctx xds_context.Context, proxy *model.Proxy, clusters envoy_common.Clusters) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()
	for _, clusterName := range clusters.ClusterNames() {
		serviceName := clusters.Tags(clusterName)[0][kuma_mesh.ServiceTag]
		tags := clusters.Tags(clusterName)
		healthCheck := proxy.HealthChecks[serviceName]
		circuitBreaker := proxy.CircuitBreakers[serviceName]
		edsClusterBuilder := envoy_clusters.NewClusterBuilder().
			Configure(envoy_clusters.LbSubset(o.lbSubsets(tags))).
			Configure(envoy_clusters.OutlierDetection(circuitBreaker)).
			Configure(envoy_clusters.HealthCheck(healthCheck))

		if clusters.Get(clusterName).HasExternalService() {
			edsClusterBuilder = edsClusterBuilder.
				Configure(envoy_clusters.StrictDNSCluster(clusterName, proxy.OutboundTargets[serviceName])).
				Configure(envoy_clusters.ClientSideTLS(proxy.OutboundTargets[serviceName]))
			protocol := o.inferProtocol(proxy, clusters.Get(clusterName).Subsets())
			switch protocol {
			case mesh_core.ProtocolHTTP2, mesh_core.ProtocolGRPC:
				edsClusterBuilder = edsClusterBuilder.Configure(envoy_clusters.Http2())
			default:
			}
		} else {
			edsClusterBuilder = edsClusterBuilder.
				Configure(envoy_clusters.EdsCluster(clusterName)).
				Configure(envoy_clusters.ClientSideMTLS(ctx, proxy.Metadata, serviceName, tags)).
				Configure(envoy_clusters.Http2())
		}
		edsCluster, err := edsClusterBuilder.Build()
		if err != nil {
			return nil, err
		}
		resources.Add(&model.Resource{
			Name:     clusterName,
			Origin:   OriginOutbound,
			Resource: edsCluster,
		})
	}
	return resources, nil
}

func (_ OutboundProxyGenerator) lbSubsets(tagSets []envoy_common.Tags) [][]string {
	var result [][]string
	uniqueKeys := map[string]bool{}
	for _, tags := range tagSets {
		keys := tags.WithoutTag(kuma_mesh.ServiceTag).Keys()
		joined := strings.Join(keys, ",")
		if !uniqueKeys[joined] {
			uniqueKeys[joined] = true
			result = append(result, keys)
		}
	}
	return result
}

func (_ OutboundProxyGenerator) generateEDS(ctx xds_context.Context, proxy *model.Proxy, clusters envoy_common.Clusters) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()
	for _, clusterName := range clusters.ClusterNames() {
		// Endpoints for ExternalServices are specified in load assignment in DNS Cluster.
		// We are not allowed to add endpoints with DNS names through EDS.
		if !clusters.Get(clusterName).HasExternalService() {
			serviceName := clusters.Tags(clusterName)[0][kuma_mesh.ServiceTag]
			loadAssignment, err := proxy.CLACache.GetCLA(context.Background(), ctx.Mesh.Resource.Meta.GetMesh(), serviceName)
			if err != nil {
				return nil, err
			}
			resources.Add(&model.Resource{
				Name:     clusterName,
				Resource: loadAssignment,
			})
		}
	}
	return resources, nil
}

// inferProtocol infers protocol for the destination listener. It will only return HTTP when all endpoints are tagged with HTTP.
func (_ OutboundProxyGenerator) inferProtocol(proxy *model.Proxy, clusters []envoy_common.ClusterSubset) mesh_core.Protocol {
	var allEndpoints []model.Endpoint
	for _, cluster := range clusters {
		serviceName := cluster.Tags[kuma_mesh.ServiceTag]
		endpoints := model.EndpointList(proxy.OutboundTargets[serviceName])
		allEndpoints = append(allEndpoints, endpoints...)
	}
	return InferServiceProtocol(allEndpoints)
}

func (_ OutboundProxyGenerator) determineSubsets(proxy *model.Proxy, outbound *kuma_mesh.Dataplane_Networking_Outbound) (subsets []envoy_common.ClusterSubset, err error) {
	oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
	route := proxy.TrafficRoutes[oface]
	if route == nil { // should not happen since we always generate default route if TrafficRoute is not found
		return nil, errors.Errorf("no TrafficRoute for outbound %s", oface)
	}

	for j, destination := range route.Spec.Conf {
		service, ok := destination.Destination[kuma_mesh.ServiceTag]
		if !ok { // should not happen since we validate traffic route
			return nil, errors.Errorf("trafficroute{name=%q}.%s: mandatory tag %q is missing: %v", route.GetMeta().GetName(), validators.RootedAt("conf").Index(j).Field("destination"), kuma_mesh.ServiceTag, destination.Destination)
		}
		if destination.Weight == 0 {
			// 0 assumes no traffic is passed there. Envoy doesn't support 0 weight, so instead of passing it to Envoy we just skip such cluster.
			continue
		}

		subset := envoy_common.ClusterSubset{
			ClusterName: service,
			Weight:      destination.Weight,
			Tags:        destination.Destination,
		}

		// We assume that all the targets are either ExternalServices or not
		// therefore we check only the first one
		endpoints := proxy.OutboundTargets[service]
		if len(endpoints) > 0 {
			ep := endpoints[0]
			if ep.IsExternalService() {
				subset.IsExternalService = true
			}
		}

		subsets = append(subsets, subset)
	}
	return
}
