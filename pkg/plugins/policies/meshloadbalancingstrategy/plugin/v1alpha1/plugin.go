package v1alpha1

import (
	"context"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/plugin/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

var _ core_plugins.EgressPolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshLoadBalancingStrategyType, dataplane, resources)
}

func (p plugin) EgressMatchedPolicies(es *core_mesh.ExternalServiceResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.EgressMatchedPolicies(api.MeshLoadBalancingStrategyType, es, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.ZoneEgressProxy != nil {
		return p.configureEgress(rs, proxy)
	}

	policies, ok := proxy.Policies.Dynamic[api.MeshLoadBalancingStrategyType]
	if !ok {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	clusters := policies_xds.GatherClusters(rs)
	endpoints := policies_xds.GatherEndpoints(rs)
	routes := policies_xds.GatherRoutes(rs)

	if err := p.configureGateway(ctx, proxy, policies.ToRules, listeners.Gateway, clusters.Gateway, routes.Gateway, endpoints); err != nil {
		return err
	}

	return p.configureDPP(proxy, policies.ToRules, listeners, clusters, endpoints)
}

func (p plugin) configureDPP(
	proxy *core_xds.Proxy,
	toRules core_rules.ToRules,
	listeners policies_xds.Listeners,
	clusters policies_xds.Clusters,
	endpoints policies_xds.EndpointMap,
) error {
	serviceConfs := map[string]api.Conf{}

	for _, outbound := range proxy.Dataplane.Spec.Networking.GetOutbound() {
		oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
		serviceName := outbound.GetService()

		computed := toRules.Rules.Compute(core_rules.MeshService(serviceName))
		if computed == nil {
			continue
		}

		conf := computed.Conf.(api.Conf)

		if listener, ok := listeners.Outbound[oface]; ok {
			if err := p.configureListener(listener, nil, conf.LoadBalancer); err != nil {
				return err
			}
		}

		serviceConfs[serviceName] = conf
	}

	// when VIPs are enabled 2 listeners are pointing to the same cluster, that's why
	// we configure clusters in a separate loop to avoid configuring the same cluster twice
	for serviceName, conf := range serviceConfs {
		if cluster, ok := clusters.Outbound[serviceName]; ok {
			if err := p.configureCluster(cluster, conf.LoadBalancer); err != nil {
				return err
			}
		}
		for _, cluster := range clusters.OutboundSplit[serviceName] {
			if err := p.configureCluster(cluster, conf.LoadBalancer); err != nil {
				return err
			}
		}
		configureEndpoints(proxy.Dataplane, endpoints, serviceName, conf)
	}

	return nil
}

func configureEndpoints(
	dataplane *core_mesh.DataplaneResource,
	endpoints policies_xds.EndpointMap,
	serviceName string,
	conf api.Conf,
) {
	var zone string
	if inbounds := dataplane.Spec.GetNetworking().GetInbound(); len(inbounds) != 0 {
		zone = inbounds[0].GetTags()[mesh_proto.ZoneTag]
	}
	if conf.LocalityAwareness == nil || !pointer.Deref(conf.LocalityAwareness.Disabled) {
		for _, cla := range endpoints[serviceName] {
			for _, localityLbEndpoints := range cla.Endpoints {
				if localityLbEndpoints.Locality != nil && localityLbEndpoints.Locality.Zone != zone {
					localityLbEndpoints.Priority = 1
				}
			}
		}
	}
}

func (p plugin) configureGateway(
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
	rules core_rules.ToRules,
	gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	gatewayClusters map[string]*envoy_cluster.Cluster,
	gatewayRoutes map[string]*envoy_route.RouteConfiguration,
	endpoints policies_xds.EndpointMap,
) error {
	if !proxy.Dataplane.Spec.IsBuiltinGateway() {
		return nil
	}

	gatewayListenerInfos, err := gateway_plugin.GatewayListenerInfoFromProxy(context.TODO(), ctx.Mesh, proxy, ctx.ControlPlane.Zone)
	if err != nil {
		return err
	}

	conf := core_rules.ComputeConf[api.Conf](rules.Rules, core_rules.MeshSubset())
	if conf == nil {
		return nil
	}

	for _, listenerInfo := range gatewayListenerInfos {
		listener, ok := gatewayListeners[core_rules.InboundListener{
			Address: proxy.Dataplane.Spec.GetNetworking().GetAddress(),
			Port:    listenerInfo.Listener.Port,
		}]
		if !ok {
			continue
		}

		if err := p.configureListener(listener, gatewayRoutes, conf.LoadBalancer); err != nil {
			return err
		}

		for _, hostInfo := range listenerInfo.HostInfos {
			destinations := gateway_plugin.RouteDestinationsMutable(hostInfo.Entries)
			for _, dest := range destinations {
				clusterName, err := dest.Destination.DestinationClusterName(hostInfo.Host.Tags)
				if err != nil {
					continue
				}
				cluster, ok := gatewayClusters[clusterName]
				if !ok {
					continue
				}

				if err := p.configureCluster(cluster, conf.LoadBalancer); err != nil {
					return err
				}

				serviceName := dest.Destination[mesh_proto.ServiceTag]
				configureEndpoints(proxy.Dataplane, endpoints, serviceName, *conf)
			}
		}
	}

	return nil
}

func (p plugin) configureEgress(rs *core_xds.ResourceSet, proxy *core_xds.Proxy) error {
	localityAwareExternalServices := map[string][]core_xds.ServiceName{}

	for _, mr := range proxy.ZoneEgressProxy.MeshResourcesList {
		for es, dynamic := range mr.Dynamic {
			policies, ok := dynamic[api.MeshLoadBalancingStrategyType]
			if !ok {
				continue
			}

			if !p.isLocalityAware(policies.FromRules) {
				continue
			}

			meshName := mr.Mesh.GetMeta().GetName()
			localityAwareExternalServices[meshName] = append(localityAwareExternalServices[meshName], es)
		}
	}

	endpoints := policies_xds.GatherEgressEndpoints(rs)
	zone := proxy.ZoneEgressProxy.ZoneEgressResource.Spec.GetZone()

	for meshName, externalServices := range localityAwareExternalServices {
		for _, es := range externalServices {
			clusterName := envoy_names.GetMeshClusterName(meshName, es)
			cla, ok := endpoints[clusterName]
			if !ok {
				continue
			}
			for _, localityLbEndpoints := range cla.Endpoints {
				if localityLbEndpoints.Locality != nil && localityLbEndpoints.Locality.Zone != zone {
					localityLbEndpoints.Priority = 1
				}
			}
		}
	}

	return nil
}

// Zone egress is a single point for multiple clients. At this moment we don't support different
// configurations based on the client, that's why locality awareness is enabled if at least one
// client requires it to be enabled.
func (p plugin) isLocalityAware(fr core_rules.FromRules) bool {
	for _, rules := range fr.Rules {
		for _, r := range rules {
			conf := r.Conf.(api.Conf)
			if conf.LocalityAwareness == nil || !pointer.Deref(conf.LocalityAwareness.Disabled) {
				return true
			}
		}
	}
	return false
}

func (p plugin) configureListener(
	l *envoy_listener.Listener,
	routes map[string]*envoy_route.RouteConfiguration,
	lbConf *api.LoadBalancer,
) error {
	if lbConf == nil {
		return nil
	}

	var hashPolicy *[]api.HashPolicy

	switch lbConf.Type {
	case api.RingHashType:
		if lbConf.RingHash == nil {
			return nil
		}
		hashPolicy = lbConf.RingHash.HashPolicies
	case api.MaglevType:
		if lbConf.Maglev == nil {
			return nil
		}
		hashPolicy = lbConf.Maglev.HashPolicies
	default:
		return nil
	}

	if l.FilterChains == nil || len(l.FilterChains) != 1 {
		return errors.Errorf("expected exactly one filter chain, got %d", len(l.FilterChains))
	}

	return v3.UpdateHTTPConnectionManager(l.FilterChains[0], func(hcm *envoy_hcm.HttpConnectionManager) error {
		var routeConfig *envoy_route.RouteConfiguration
		switch r := hcm.RouteSpecifier.(type) {
		case *envoy_hcm.HttpConnectionManager_RouteConfig:
			routeConfig = r.RouteConfig
		case *envoy_hcm.HttpConnectionManager_Rds:
			routeConfig = routes[r.Rds.RouteConfigName]
		default:
			return errors.Errorf("unexpected RouteSpecifer %T", r)
		}

		hpc := &xds.HashPolicyConfigurer{HashPolicies: *hashPolicy}
		for _, vh := range routeConfig.VirtualHosts {
			for _, route := range vh.Routes {
				if err := hpc.Configure(route); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (p plugin) configureCluster(c *envoy_cluster.Cluster, lbConf *api.LoadBalancer) error {
	if lbConf == nil {
		return nil
	}
	return (&xds.LoadBalancerConfigurer{LoadBalancer: *lbConf}).Configure(c)
}
